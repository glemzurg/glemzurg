package actions

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/invariants"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/model_bridge"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

const (
	implicitEnumRequireSubKeyPrefix      = "implicit_enum_"
	implicitReferenceRequireSubKeyPrefix = "implicit_ref_"
	logicOwnerKindAction                 = "action"
	logicOwnerKindQuery                  = "query"
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
		expr := req.Spec.Expression
		if expr == nil {
			return nil, fmt.Errorf("%s %s %s[%d]: expression not lowered", o.Kind, o.Name, kindLabel, i)
		}
		if model_bridge.ContainsAnyPrimedME(expr) {
			return nil, fmt.Errorf("%s %s %s[%d]: must not contain primed variables", o.Kind, o.Name, kindLabel, i)
		}

		result := evaluator.Eval(expr, bindings)
		if result.IsError() {
			return nil, fmt.Errorf("%s %s %s[%d] evaluation error: %s", o.Kind, o.Name, kindLabel, i, result.Error.Inspect())
		}
		if isTrueBoolean(result.Value) {
			continue
		}

		msg := _EXPRESSION_RETURNED_NIL
		if result.Value != nil {
			msg = fmt.Sprintf("expression returned %s", result.Value.Inspect())
		}
		failures = append(failures, RequireAssessmentFailure{Index: i, Logic: req, Message: msg})
	}
	return failures, nil
}

// ParameterInvariantViolations converts parameter invariant assessment failures into violations.
func (o ParameterOwner) ParameterInvariantViolations(
	failures []RequireAssessmentFailure,
	instanceID state.InstanceID,
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
	instanceID state.InstanceID,
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

// SampleParameters generates values for paramDefs using effective requires constraints.
func (s *ParameterSampler) SampleParameters(
	owner ParameterOwner,
	paramDefs []model_state.Parameter,
	rng *rand.Rand,
) (map[string]object.Object, error) {
	if len(paramDefs) == 0 {
		return map[string]object.Object{}, nil
	}

	samplingLogics, err := owner.SamplingLogicsFor(paramDefs)
	if err != nil {
		return nil, err
	}
	if len(samplingLogics) == 0 {
		return s.binder.GenerateRandomParameters(paramDefs, rng), nil
	}

	paramNames := parameterNames(paramDefs)
	if err := validateRequiresSamplingSupport(owner.Requires, paramNames); err != nil {
		var unsupported *UnsupportedRequiresSamplingError
		if errors.As(err, &unsupported) {
			unsupported.ActionName = owner.Name
		}
		return nil, err
	}

	constraints := extractParameterConstraints(samplingLogics)
	result := s.binder.GenerateRandomParameters(paramDefs, rng)
	nullableByName := parameterNullableByName(paramDefs)
	applyParameterConstraints(result, constraints, rng, s.namedSetValues, nullableByName)
	enforceParameterNullability(result, paramDefs, rng)
	return result, nil
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
	if constraints.nullableElseMirror != nil {
		covered[constraints.nullableElseMirror.driverParam] = true
		covered[constraints.nullableElseMirror.followerParam] = true
	}
	if constraints.nullableElseEquality != nil {
		covered[constraints.nullableElseEquality.driverParam] = true
		covered[constraints.nullableElseEquality.followerParam] = true
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

		spec, err := logic_spec.NewExpressionSpec(model_logic.NotationTLAPlus, "TRUE", pf)
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

		specText := enumMembershipSpecification(param.Name, values, param.Nullable)
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

func enumMembershipSpecification(paramName string, values []string, nullable bool) string {
	membership := fmt.Sprintf(`%s \in %s`, paramName, formatTLAPlusStringSet(values))
	if !nullable {
		return membership
	}
	return fmt.Sprintf(`IF %s = NULL THEN TRUE ELSE %s`, paramName, membership)
}

func formatTLAPlusStringSet(values []string) string {
	quoted := make([]string, len(values))
	for i, value := range values {
		quoted[i] = `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`
	}
	return "{" + strings.Join(quoted, ", ") + "}"
}
