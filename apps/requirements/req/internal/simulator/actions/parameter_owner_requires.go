package actions

import (
	"errors"
	"fmt"
	"math/rand"
	"slices"
	"sort"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/invariants"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/model_bridge"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

const (
	implicitEnumRequireSubKeyPrefix      = "implicit_enum_"
	implicitReferenceRequireSubKeyPrefix = "implicit_ref_"
	logicOwnerKindAction                 = "action"
	logicOwnerKindQuery                  = "query"
	maxParameterSampleAttempts           = 10
)

// ParameterOwner is an action or query that owns typed parameters and requires.
// All simulator requires handling (implicit enum synthesis, sampling, validation,
// and execution) flows through this type so action and query behavior stay aligned.
type ParameterOwner struct {
	Key        identity.Key
	Kind       string
	Name       string
	Parameters []model_state.Parameter
	Requires   []model_logic.Logic
}

// ParameterOwnerFromAction wraps an action as a parameter owner.
func ParameterOwnerFromAction(action model_state.Action) ParameterOwner {
	return ParameterOwner{
		Key:        action.Key,
		Kind:       logicOwnerKindAction,
		Name:       action.Name,
		Parameters: action.Parameters,
		Requires:   action.Requires,
	}
}

// ParameterOwnerFromQuery wraps a query as a parameter owner.
func ParameterOwnerFromQuery(query model_state.Query) ParameterOwner {
	return ParameterOwner{
		Key:        query.Key,
		Kind:       logicOwnerKindQuery,
		Name:       query.Name,
		Parameters: query.Parameters,
		Requires:   query.Requires,
	}
}

// NeedsRequiresAwareSampling reports whether paramDefs should be sampled using
// explicit requires, parameter invariants, or implicit requires rather than type-only random generation.
func (o ParameterOwner) NeedsRequiresAwareSampling(paramDefs []model_state.Parameter) bool {
	return len(o.Requires) > 0 || o.HasImplicitEnumRequiresFor(paramDefs) || hasParameterInvariantAssessments(paramDefs)
}

// EffectiveRequiresFor returns explicit owner requires plus synthesized implicit requires.
// Parameter invariants are excluded; use SamplingLogicsFor for value generation.
func (o ParameterOwner) EffectiveRequiresFor(paramDefs []model_state.Parameter) ([]model_logic.Logic, error) {
	return effectiveOwnerRequires(o.Key, o.Kind, paramDefs, o.Requires)
}

// SamplingLogicsFor returns owner requires plus parameter invariants for constraint extraction.
// Nullable parameter invariants are auto-wrapped to apply only when the value is set.
func (o ParameterOwner) SamplingLogicsFor(paramDefs []model_state.Parameter) ([]model_logic.Logic, error) {
	ownerRequires, err := o.EffectiveRequiresFor(paramDefs)
	if err != nil {
		return nil, err
	}
	paramInv, err := parameterInvariantAssessmentsForSampling(o.Key, paramDefs)
	if err != nil {
		return nil, err
	}
	if len(paramInv) == 0 {
		return ownerRequires, nil
	}
	combined := make([]model_logic.Logic, 0, len(ownerRequires)+len(paramInv))
	combined = append(combined, ownerRequires...)
	combined = append(combined, paramInv...)
	return combined, nil
}

// HasImplicitEnumRequiresFor reports whether paramDefs include parsed enumeration
// rules that would synthesize at least one implicit require.
func (o ParameterOwner) HasImplicitEnumRequiresFor(paramDefs []model_state.Parameter) bool {
	return hasImplicitEnumRequires(paramDefs, o.Requires)
}

// ValidateRequiresSamplingSupport checks that effective requires can drive random
// parameter generation for all owner parameters.
func (o ParameterOwner) ValidateRequiresSamplingSupport(className string) error {
	if len(o.Parameters) == 0 {
		return nil
	}
	paramNames := parameterNames(o.Parameters)
	if err := validateRequiresSamplingSupport(o.Requires, paramNames); err != nil {
		var unsupported *UnsupportedRequiresSamplingError
		if errors.As(err, &unsupported) {
			unsupported.ClassName = className
			unsupported.ActionName = o.Name
			return unsupported
		}
		return err
	}
	return nil
}

// RequireAssessmentFailure records one requires assessment that did not evaluate to TRUE.
type RequireAssessmentFailure struct {
	Index   int
	Logic   model_logic.Logic
	Message string
}

// AssessRequires evaluates owner requires for paramDefs against bindings.
// Parameter invariants are assessed separately via AssessParameterInvariants.
func (o ParameterOwner) AssessRequires(
	paramDefs []model_state.Parameter,
	bindings *evaluator.Bindings,
) ([]RequireAssessmentFailure, error) {
	return assessLogics(o, paramDefs, bindings, "requires", func() ([]model_logic.Logic, error) {
		return o.EffectiveRequiresFor(paramDefs)
	})
}

// AssessParameterInvariants evaluates parameter invariants for paramDefs against bindings.
// Nullable parameters skip invariant checks when the bound value is NULL/absent.
func (o ParameterOwner) AssessParameterInvariants(
	paramDefs []model_state.Parameter,
	bindings *evaluator.Bindings,
) ([]RequireAssessmentFailure, error) {
	var logics []model_logic.Logic
	for _, param := range paramDefs {
		if param.Nullable {
			if val, ok := bindings.GetValue(param.Name); !ok || object.IsNull(val) {
				continue
			}
		}
		for _, inv := range param.Invariants {
			if inv.Type == model_logic.LogicTypeLet {
				logics = append(logics, inv)
				continue
			}
			if inv.Type != model_logic.LogicTypeAssessment || inv.Spec.Expression == nil {
				continue
			}
			if invariants.IsParameterEqualityInvariant(inv.Spec.Expression) {
				continue
			}
			logics = append(logics, inv)
		}
	}
	return assessLogics(o, paramDefs, bindings, "parameter invariant", func() ([]model_logic.Logic, error) {
		return logics, nil
	})
}

func assessLogics(
	o ParameterOwner,
	_ []model_state.Parameter,
	bindings *evaluator.Bindings,
	kindLabel string,
	logicsFn func() ([]model_logic.Logic, error),
) ([]RequireAssessmentFailure, error) {
	requires, err := logicsFn()
	if err != nil {
		return nil, err
	}
	if len(requires) == 0 {
		return nil, nil
	}
	if err := evalLetBindings(requires, bindings, o.Kind, o.Name, kindLabel); err != nil {
		return nil, err
	}

	var failures []RequireAssessmentFailure
	for i, req := range requires {
		if req.Type == model_logic.LogicTypeLet {
			continue
		}
		failure, failed, err := assessOneLogic(o, bindings, kindLabel, i, req)
		if err != nil {
			return nil, err
		}
		if failed {
			failures = append(failures, failure)
		}
	}
	return failures, nil
}

func assessOneLogic(
	o ParameterOwner,
	bindings *evaluator.Bindings,
	kindLabel string,
	index int,
	req model_logic.Logic,
) (RequireAssessmentFailure, bool, error) {
	expr := req.Spec.Expression
	if expr == nil {
		return RequireAssessmentFailure{}, false, fmt.Errorf("%s %s %s[%d]: expression not lowered", o.Kind, o.Name, kindLabel, index)
	}
	if model_bridge.ContainsAnyPrimedME(expr) {
		return RequireAssessmentFailure{}, false, fmt.Errorf("%s %s %s[%d]: must not contain primed variables", o.Kind, o.Name, kindLabel, index)
	}

	if pattern, ok := detectPeerFieldDistinctFromParam(expr); ok {
		if assessPeerFieldDistinctFromParam(pattern, bindings) {
			return RequireAssessmentFailure{}, false, nil
		}
		return RequireAssessmentFailure{
			Index:   index,
			Logic:   req,
			Message: fmt.Sprintf("expression returned %s", "FALSE"),
		}, true, nil
	}

	result := evaluator.Eval(expr, bindings)
	if result.IsError() {
		return RequireAssessmentFailure{}, false, fmt.Errorf("%s %s %s[%d] evaluation error: %s", o.Kind, o.Name, kindLabel, index, result.Error.Inspect())
	}
	if isTrueBoolean(result.Value) {
		return RequireAssessmentFailure{}, false, nil
	}

	msg := _EXPRESSION_RETURNED_NIL
	if result.Value != nil {
		msg = fmt.Sprintf("expression returned %s", result.Value.Inspect())
	}
	return RequireAssessmentFailure{Index: index, Logic: req, Message: msg}, true, nil
}

// ParameterInvariantViolations converts parameter invariant assessment failures into violations.
func (o ParameterOwner) ParameterInvariantViolations(
	failures []RequireAssessmentFailure,
	instanceID instance.ID,
) invariants.ViolationErrors {
	var violations invariants.ViolationErrors
	for _, failure := range failures {
		violations = append(violations, invariants.NewParameterInvariantViolation(
			o.Key, o.Name, failure.Index, failure.Logic.Spec.Specification, instanceID, failure.Message,
		))
	}
	return violations
}

// ActionRequiresViolations converts assessment failures into action-require violations.
func (o ParameterOwner) ActionRequiresViolations(
	failures []RequireAssessmentFailure,
	instanceID instance.ID,
) invariants.ViolationErrors {
	var violations invariants.ViolationErrors
	for _, failure := range failures {
		violations = append(violations, invariants.NewActionRequiresViolation(
			o.Key, o.Name, failure.Index, failure.Logic.Spec.Specification, instanceID, failure.Message,
		))
	}
	return violations
}

// RequireAssessmentError returns the first failure as an error for query-style callers.
func (o ParameterOwner) RequireAssessmentError(failures []RequireAssessmentFailure) error {
	if len(failures) == 0 {
		return nil
	}
	failure := failures[0]
	return fmt.Errorf(
		"%s %s precondition failed: requires[%d] = %s",
		o.Kind, o.Name, failure.Index, failure.Logic.Spec.Specification,
	)
}

type parameterSamplingPrep struct {
	samplingLogics []model_logic.Logic
	constraints    parameterConstraints
	nullableByName map[string]bool
	generate       func(paramDefs []model_state.Parameter, rng *rand.Rand) map[string]object.Object
}

func (s *ParameterSampler) prepareRequiresSampling(
	owner ParameterOwner,
	paramDefs []model_state.Parameter,
) (prep *parameterSamplingPrep, requiresAware bool, err error) {
	samplingLogics, err := owner.SamplingLogicsFor(paramDefs)
	if err != nil {
		return nil, false, err
	}
	if len(samplingLogics) == 0 {
		return nil, false, nil
	}

	paramNames := parameterNames(paramDefs)
	if err := validateRequiresSamplingSupport(owner.Requires, paramNames); err != nil {
		var unsupported *UnsupportedRequiresSamplingError
		if errors.As(err, &unsupported) {
			unsupported.ActionName = owner.Name
		}
		return nil, false, err
	}

	generate := s.binder.GenerateRandomParameters
	if s.generateOverride != nil {
		generate = s.generateOverride
	}
	return &parameterSamplingPrep{
		samplingLogics: samplingLogics,
		constraints:    extractParameterConstraints(samplingLogics),
		nullableByName: parameterNullableByName(paramDefs),
		generate:       generate,
	}, true, nil
}

func (s *ParameterSampler) sampleUntilRequiresSatisfied(
	owner ParameterOwner,
	paramDefs []model_state.Parameter,
	prep *parameterSamplingPrep,
	rng *rand.Rand,
) (map[string]object.Object, error) {
	if err := s.errIfNamedSetDomainExhausted(owner, paramDefs); err != nil {
		return nil, err
	}

	var lastAttempt string
	var lastRejectReason string

	for range maxParameterSampleAttempts {
		result, rejectReason := s.generateConstrainedSample(paramDefs, prep, rng)
		if rejectReason != "" {
			lastAttempt = formatSampledParameters(result)
			lastRejectReason = rejectReason
			continue
		}
		enforceParameterNullability(result, paramDefs, rng)
		coerceSampledParameters(paramDefs, result)
		failures, err := owner.samplingAssessmentFailures(
			paramDefs, prep.samplingLogics, result, s.namedSetValues,
		)
		if err != nil {
			return nil, err
		}
		if len(failures) == 0 {
			return result, nil
		}
		lastAttempt = formatSampledParameters(result)
		lastRejectReason = formatSamplingFailures(failures)
	}
	return nil, &ParameterSampleExhaustedError{
		Owner:            owner,
		Attempts:         maxParameterSampleAttempts,
		LastAttempt:      lastAttempt,
		LastRejectReason: lastRejectReason,
	}
}

func (s *ParameterSampler) errIfNamedSetDomainExhausted(
	owner ParameterOwner,
	paramDefs []model_state.Parameter,
) error {
	// Action selection should have excluded this already; fail closed if not.
	ok, err := s.NamedSetSampleDomainsAvailable(owner, paramDefs)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	return &ParameterSampleExhaustedError{
		Owner:            owner,
		Attempts:         0,
		LastRejectReason: "named-set sample domain empty (all allowed values already used)",
	}
}

func (s *ParameterSampler) generateConstrainedSample(
	paramDefs []model_state.Parameter,
	prep *parameterSamplingPrep,
	rng *rand.Rand,
) (result map[string]object.Object, rejectReason string) {
	result = prep.generate(paramDefs, rng)
	if s.generateOverride != nil {
		return result, ""
	}
	applyParameterConstraints(result, prep.constraints, rng, s.namedSetValues, prep.nullableByName, s.peerFieldDistinctLookup)
	if samplingPeerFieldDistinctConflict(result, prep.constraints, s.peerFieldDistinctLookup) {
		return result, "peer field value already used by another instance"
	}
	return result, ""
}

// SampleParameters generates values for paramDefs using effective requires constraints.
func (s *ParameterSampler) SampleParameters(
	owner ParameterOwner,
	paramDefs []model_state.Parameter,
	rng *rand.Rand,
) (map[string]object.Object, error) {
	if len(paramDefs) == 0 {
		return map[string]object.Object{}, nil
	}

	prep, requiresAware, err := s.prepareRequiresSampling(owner, paramDefs)
	if err != nil {
		return nil, err
	}
	if !requiresAware {
		return s.binder.GenerateRandomParameters(paramDefs, rng), nil
	}
	return s.sampleUntilRequiresSatisfied(owner, paramDefs, prep, rng)
}

// ParameterSampleExhaustedError reports requires-aware parameter sampling exhaustion.
type ParameterSampleExhaustedError struct {
	Owner            ParameterOwner
	Attempts         int
	LastAttempt      string
	LastRejectReason string
}

func (e *ParameterSampleExhaustedError) Error() string {
	msg := fmt.Sprintf(
		"%s %q: failed to sample parameters satisfying requires after %d attempts",
		e.Owner.Kind, e.Owner.Name, e.Attempts,
	)
	if e.LastAttempt != "" {
		msg += fmt.Sprintf("; last attempt: %s", e.LastAttempt)
	}
	if e.LastRejectReason != "" {
		msg += fmt.Sprintf("; reject reason: %s", e.LastRejectReason)
	}
	return msg
}

func formatSampledParameters(values map[string]object.Object) string {
	if len(values) == 0 {
		return "{}"
	}
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	parts := make([]string, 0, len(names))
	for _, name := range names {
		val := values[name]
		if val == nil {
			parts = append(parts, fmt.Sprintf("%s=nil", name))
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%s", name, val.Inspect()))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

func formatSamplingFailures(failures []RequireAssessmentFailure) string {
	if len(failures) == 0 {
		return ""
	}
	failure := failures[0]
	spec := failure.Logic.Spec.Specification
	if spec == "" {
		spec = fmt.Sprintf("requires[%d]", failure.Index)
	}
	return fmt.Sprintf("%s (%s)", spec, failure.Message)
}

func (o ParameterOwner) samplingAssessmentFailures(
	paramDefs []model_state.Parameter,
	samplingLogics []model_logic.Logic,
	values map[string]object.Object,
	namedSetValues map[string]object.Object,
) ([]RequireAssessmentFailure, error) {
	sampledNames := parameterNames(paramDefs)
	logics := logicsReferencingOnlySampledParams(samplingLogics, sampledNames)
	if len(logics) == 0 {
		return nil, nil
	}

	bindings := evaluator.NewBindings()
	for name, value := range values {
		bindings.Set(name, value, evaluator.NamespaceLocal)
	}
	for name, value := range namedSetValues {
		bindings.Set(name, value, evaluator.NamespaceGlobal)
	}

	return assessLogics(o, paramDefs, bindings, "requires", func() ([]model_logic.Logic, error) {
		return logics, nil
	})
}

func logicsReferencingOnlySampledParams(
	logics []model_logic.Logic,
	sampledNames map[string]bool,
) []model_logic.Logic {
	if len(sampledNames) == 0 {
		return nil
	}
	filtered := make([]model_logic.Logic, 0, len(logics))
	for _, logic := range logics {
		if logic.Type != model_logic.LogicTypeAssessment || !logic.Spec.ParseOk() {
			continue
		}
		// Class extents are not bound during parameter sampling; structural constraints
		// (e.g. set-minus peer field) already drive generation for those requires.
		if expressionReferencesClassExtent(logic.Spec.Expression) {
			continue
		}
		if !expressionReferencesOnlyParams(logic.Spec.Expression, sampledNames) {
			continue
		}
		filtered = append(filtered, logic)
	}
	return filtered
}

func expressionReferencesClassExtent(expr me.Expression) bool {
	if expr == nil {
		return false
	}
	if _, ok := expr.(*me.ClassRef); ok {
		return true
	}
	if slices.ContainsFunc(expressionChildNodes(expr), expressionReferencesClassExtent) {
		return true
	}
	// Membership set side is not walked by expressionChildNodes (element only);
	// set-minus used-codes and peer quantifiers live there.
	if membership, ok := expr.(*me.Membership); ok {
		return expressionReferencesClassExtent(membership.Set)
	}
	if setMap, ok := expr.(*me.SetMap); ok {
		return expressionReferencesClassExtent(setMap.Set) || expressionReferencesClassExtent(setMap.Transform)
	}
	if setFilter, ok := expr.(*me.SetFilter); ok {
		return expressionReferencesClassExtent(setFilter.Set) || expressionReferencesClassExtent(setFilter.Predicate)
	}
	return false
}

func expressionReferencesOnlyParams(expr me.Expression, paramNames map[string]bool) bool {
	if expr == nil {
		return true
	}
	refs := referencedParamNames(expr)
	for name := range refs {
		if !paramNames[name] {
			return false
		}
	}
	return true
}

func referencedParamNames(expr me.Expression) map[string]bool {
	refs := make(map[string]bool)
	collectReferencedParamNames(expr, refs)
	return refs
}

func collectReferencedParamNames(expr me.Expression, refs map[string]bool) {
	if expr == nil {
		return
	}
	if localVar, ok := expr.(*me.LocalVar); ok {
		refs[localVar.Name] = true
		return
	}
	for _, child := range expressionChildNodes(expr) {
		collectReferencedParamNames(child, refs)
	}
}

func effectiveOwnerRequires(
	ownerKey identity.Key,
	ownerKind string,
	params []model_state.Parameter,
	explicitRequires []model_logic.Logic,
) ([]model_logic.Logic, error) {
	combined := make([]model_logic.Logic, 0, len(explicitRequires)+len(params))
	combined = append(combined, explicitRequires...)

	implicitEnum, err := implicitEnumRequires(ownerKey, ownerKind, params, combined)
	if err != nil {
		return nil, err
	}
	combined = append(combined, implicitEnum...)

	implicitRef, err := implicitReferenceRequires(ownerKey, ownerKind, params, combined)
	if err != nil {
		return nil, err
	}
	combined = append(combined, implicitRef...)
	return combined, nil
}

// parameterInvariantAssessmentsForSampling returns parameter invariant assessments for sampling.
// Nullable parameters get an automatic NULL guard unless the author already wrote one.
func parameterInvariantAssessmentsForSampling(ownerKey identity.Key, params []model_state.Parameter) ([]model_logic.Logic, error) {
	if _, err := identity.ParseKey(ownerKey.ParentKey); err != nil {
		return nil, fmt.Errorf("parameter invariants: owner %q: %w", ownerKey.String(), err)
	}

	var assessments []model_logic.Logic
	for _, param := range params {
		for _, inv := range param.Invariants {
			if inv.Type != model_logic.LogicTypeAssessment || inv.Spec.Expression == nil {
				continue
			}
			if invariants.IsParameterEqualityInvariant(inv.Spec.Expression) {
				continue
			}
			logic := inv
			if param.Nullable && !invariants.LogicSpecHasNullableWhenUnsetGuard(inv.Spec) {
				spec := logic_spec.ExpressionSpec{
					Notation:      inv.Spec.Notation,
					Specification: invariants.NullableWhenSetSpecification(param.Name, inv.Spec.Specification),
					Expression:    invariants.WrapNullableWhenSetExpression(param.Name, inv.Spec.Expression),
				}
				logic = model_logic.NewLogic(
					inv.Key,
					inv.Type,
					inv.Description,
					inv.Target,
					spec,
					inv.TargetTypeSpec,
				)
			}
			assessments = append(assessments, logic)
		}
	}
	return assessments, nil
}

func hasParameterInvariantAssessments(params []model_state.Parameter) bool {
	for _, param := range params {
		for _, inv := range param.Invariants {
			if inv.Type != model_logic.LogicTypeAssessment || inv.Spec.Expression == nil {
				continue
			}
			if invariants.IsParameterEqualityInvariant(inv.Spec.Expression) {
				continue
			}
			return true
		}
	}
	return false
}

func paramsCoveredByConstraints(logics []model_logic.Logic) map[string]bool {
	covered := make(map[string]bool)
	constraints := extractParameterConstraints(logics)
	for name := range constraints.enumValues {
		covered[name] = true
	}
	if constraints.nullableElseMembership != nil {
		covered[constraints.nullableElseMembership.paramName] = true
	}
	if constraints.paramInNamedSetMinusPeerField != nil {
		covered[constraints.paramInNamedSetMinusPeerField.paramName] = true
	}
	if constraints.paramInNamedSet != nil {
		covered[constraints.paramInNamedSet.paramName] = true
	}
	if constraints.nullableElseMirror != nil {
		covered[constraints.nullableElseMirror.driverParam] = true
		covered[constraints.nullableElseMirror.followerParam] = true
	}
	if constraints.nullableElseExclusionEquality != nil {
		covered[constraints.nullableElseExclusionEquality.driverParam] = true
		covered[constraints.nullableElseExclusionEquality.followerParam] = true
	}
	if constraints.nullableElseEquality != nil {
		covered[constraints.nullableElseEquality.driverParam] = true
		covered[constraints.nullableElseEquality.followerParam] = true
	}
	if constraints.nullableElseBooleanConstant != nil {
		covered[constraints.nullableElseBooleanConstant.driverParam] = true
		covered[constraints.nullableElseBooleanConstant.followerParam] = true
	}
	if constraints.nullableElseTuple != nil {
		for _, name := range constraints.nullableElseTuple.paramNames {
			covered[name] = true
		}
	}
	if constraints.tupleInSet != nil {
		for _, name := range constraints.tupleInSet.paramNames {
			covered[name] = true
		}
	}
	return covered
}

func implicitReferenceRequires(
	ownerKey identity.Key,
	ownerKind string,
	params []model_state.Parameter,
	logics []model_logic.Logic,
) ([]model_logic.Logic, error) {
	covered := paramsCoveredByConstraints(logics)
	classKey, err := identity.ParseKey(ownerKey.ParentKey)
	if err != nil {
		return nil, fmt.Errorf("implicit reference requires: owner %q: %w", ownerKey.String(), err)
	}

	ctx := &convert.LowerContext{
		ClassKey:   classKey,
		Parameters: parameterNames(params),
	}
	pf := convert.NewExpressionParseFunc(ctx)

	var implicit []model_logic.Logic
	ordinal := 0
	for _, param := range params {
		if !model_data_type.ContainsReferenceConstraint(param.DataType) {
			continue
		}
		if covered[param.Name] {
			continue
		}

		spec, err := logic_spec.NewExpressionSpec(model_logic.NotationTLAPlus, tlaLiteralTrue, pf)
		if err != nil {
			return nil, fmt.Errorf(
				"implicit reference require for parameter %q (owner %q): %w",
				param.Name, ownerKey.String(), err,
			)
		}

		requireKey, err := newOwnerRequireKey(ownerKey, ownerKind, fmt.Sprintf("%s%d", implicitReferenceRequireSubKeyPrefix, ordinal))
		if err != nil {
			return nil, err
		}
		ordinal++

		implicit = append(implicit, model_logic.NewLogic(
			requireKey,
			model_logic.LogicTypeAssessment,
			fmt.Sprintf("Reference parameter %q has no formal constraint; simulation accepts any value.", param.Name),
			"",
			spec,
			nil,
		))
	}

	return implicit, nil
}

func implicitEnumRequires(
	ownerKey identity.Key,
	ownerKind string,
	params []model_state.Parameter,
	explicitRequires []model_logic.Logic,
) ([]model_logic.Logic, error) {
	if len(params) == 0 {
		return nil, nil
	}

	explicitEnum := extractParameterConstraints(explicitRequires).enumValues
	paramNames := parameterNames(params)
	classKey, err := identity.ParseKey(ownerKey.ParentKey)
	if err != nil {
		return nil, fmt.Errorf("implicit enum requires: owner %q: %w", ownerKey.String(), err)
	}

	ctx := &convert.LowerContext{
		ClassKey:   classKey,
		Parameters: paramNames,
	}
	pf := convert.NewExpressionParseFunc(ctx)

	var implicit []model_logic.Logic
	ordinal := 0
	for _, param := range params {
		if explicitEnum != nil {
			if _, covered := explicitEnum[param.Name]; covered {
				continue
			}
		}
		values := model_data_type.EnumerationValues(param.DataType)
		if len(values) == 0 {
			continue
		}

		specText := enumMembershipSpecification(param.Name, param.DataType, values, param.Nullable)
		spec, err := logic_spec.NewExpressionSpec(model_logic.NotationTLAPlus, specText, pf)
		if err != nil {
			return nil, fmt.Errorf(
				"implicit enum require for parameter %q (owner %q): %w",
				param.Name, ownerKey.String(), err,
			)
		}

		requireKey, err := newOwnerRequireKey(ownerKey, ownerKind, fmt.Sprintf("%s%d", implicitEnumRequireSubKeyPrefix, ordinal))
		if err != nil {
			return nil, err
		}
		ordinal++

		implicit = append(implicit, model_logic.NewLogic(
			requireKey,
			model_logic.LogicTypeAssessment,
			fmt.Sprintf("Parameter %q must be one of the allowed enumeration values.", param.Name),
			"",
			spec,
			nil,
		))
	}

	return implicit, nil
}

func hasImplicitEnumRequires(params []model_state.Parameter, explicitRequires []model_logic.Logic) bool {
	explicitEnum := extractParameterConstraints(explicitRequires).enumValues
	for _, param := range params {
		if explicitEnum != nil {
			if _, covered := explicitEnum[param.Name]; covered {
				continue
			}
		}
		if len(model_data_type.EnumerationValues(param.DataType)) > 0 {
			return true
		}
	}
	return false
}

func newOwnerRequireKey(ownerKey identity.Key, ownerKind, subKey string) (identity.Key, error) {
	switch ownerKind {
	case logicOwnerKindAction:
		return identity.NewActionRequireKey(ownerKey, subKey)
	case logicOwnerKindQuery:
		return identity.NewQueryRequireKey(ownerKey, subKey)
	default:
		return identity.Key{}, fmt.Errorf("unsupported owner kind %q for implicit enum requires", ownerKind)
	}
}

func enumMembershipSpecification(
	paramName string,
	dataType *model_data_type.DataType,
	values []string,
	nullable bool,
) string {
	var membership string
	if model_data_type.HasBooleanTypeSpec(dataType) {
		membership = fmt.Sprintf(`%s \in BOOLEAN`, paramName)
	} else {
		membership = fmt.Sprintf(`%s \in %s`, paramName, formatTLAPlusStringSet(values))
	}
	if !nullable {
		return membership
	}
	return fmt.Sprintf(`_GZ!WhenNotNull(%s, %s)`, paramName, membership)
}

func formatTLAPlusStringSet(values []string) string {
	quoted := make([]string, len(values))
	for i, value := range values {
		quoted[i] = `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`
	}
	return "{" + strings.Join(quoted, ", ") + "}"
}
