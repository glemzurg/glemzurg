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
	implicitEnumRequireSubKeyPrefix = "implicit_enum_"
	logicOwnerKindAction            = "action"
	logicOwnerKindQuery             = "query"
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
// explicit and implicit requires rather than type-only random generation.
func (o ParameterOwner) NeedsRequiresAwareSampling(paramDefs []model_state.Parameter) bool {
	return len(o.Requires) > 0 || o.HasImplicitEnumRequiresFor(paramDefs)
}

// EffectiveRequiresFor returns explicit requires plus synthesized enum-membership
// assessments for enumeration parameters in paramDefs not already constrained.
func (o ParameterOwner) EffectiveRequiresFor(paramDefs []model_state.Parameter) ([]model_logic.Logic, error) {
	return effectiveRequires(o.Key, o.Kind, paramDefs, o.Requires)
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
	effectiveRequires, err := o.EffectiveRequiresFor(o.Parameters)
	if err != nil {
		return err
	}
	paramNames := parameterNames(o.Parameters)
	if err := validateRequiresSamplingSupport(effectiveRequires, paramNames); err != nil {
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

// AssessRequires evaluates effective requires for paramDefs against bindings.
func (o ParameterOwner) AssessRequires(
	paramDefs []model_state.Parameter,
	bindings *evaluator.Bindings,
) ([]RequireAssessmentFailure, error) {
	requires, err := o.EffectiveRequiresFor(paramDefs)
	if err != nil {
		return nil, err
	}
	if err := evalLetBindings(requires, bindings, o.Kind, o.Name, "requires"); err != nil {
		return nil, err
	}

	var failures []RequireAssessmentFailure
	for i, req := range requires {
		if req.Type == model_logic.LogicTypeLet {
			continue
		}
		expr := req.Spec.Expression
		if expr == nil {
			return nil, fmt.Errorf("%s %s requires[%d]: expression not lowered", o.Kind, o.Name, i)
		}
		if model_bridge.ContainsAnyPrimedME(expr) {
			return nil, fmt.Errorf("%s %s requires[%d]: Requires must not contain primed variables", o.Kind, o.Name, i)
		}

		result := evaluator.Eval(expr, bindings)
		if result.IsError() {
			return nil, fmt.Errorf("%s %s requires[%d] evaluation error: %s", o.Kind, o.Name, i, result.Error.Inspect())
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

	effectiveRequires, err := owner.EffectiveRequiresFor(paramDefs)
	if err != nil {
		return nil, err
	}
	if len(effectiveRequires) == 0 {
		return s.binder.GenerateRandomParameters(paramDefs, rng), nil
	}

	paramNames := parameterNames(paramDefs)
	if err := validateRequiresSamplingSupport(effectiveRequires, paramNames); err != nil {
		var unsupported *UnsupportedRequiresSamplingError
		if errors.As(err, &unsupported) {
			unsupported.ActionName = owner.Name
		}
		return nil, err
	}

	constraints := extractParameterConstraints(effectiveRequires)
	result := s.binder.GenerateRandomParameters(paramDefs, rng)
	nullableByName := parameterNullableByName(paramDefs)
	applyParameterConstraints(result, constraints, rng, s.namedSetValues, nullableByName)
	enforceParameterNullability(result, paramDefs, rng)
	return result, nil
}

func effectiveRequires(
	ownerKey identity.Key,
	ownerKind string,
	params []model_state.Parameter,
	explicitRequires []model_logic.Logic,
) ([]model_logic.Logic, error) {
	implicit, err := implicitEnumRequires(ownerKey, ownerKind, params, explicitRequires)
	if err != nil {
		return nil, err
	}
	if len(implicit) == 0 {
		return explicitRequires, nil
	}
	combined := make([]model_logic.Logic, 0, len(explicitRequires)+len(implicit))
	combined = append(combined, explicitRequires...)
	combined = append(combined, implicit...)
	return combined, nil
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
