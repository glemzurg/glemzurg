package actions

import (
	"fmt"
	"maps"
	"math/rand"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

// ActionHasParameterSimulation reports whether any action parameter carries simulator sampling metadata.
func ActionHasParameterSimulation(action model_state.Action) bool {
	for _, param := range action.Parameters {
		if param.Simulation != nil && param.Simulation.HasSimulation() {
			return true
		}
	}
	return false
}

// ActionSimulationRequiresMet reports whether every parameter with simulation rules has
// at least one rule whose requires all hold (so the action is eligible for sampling).
func ActionSimulationRequiresMet(
	action model_state.Action,
	bindings *evaluator.Bindings,
) (bool, error) {
	for _, param := range action.Parameters {
		if param.Simulation == nil || len(param.Simulation.Rules) == 0 {
			continue
		}
		eligible, err := EligibleSimulationRules(param, bindings)
		if err != nil {
			return false, fmt.Errorf("parameter %q simulation: %w", param.Name, err)
		}
		if len(eligible) == 0 {
			return false, nil
		}
	}
	return true, nil
}

// EligibleSimulationRules returns rules that have a specification and whose requires all hold.
func EligibleSimulationRules(
	param model_state.Parameter,
	bindings *evaluator.Bindings,
) ([]model_state.ParameterSimulationRule, error) {
	if param.Simulation == nil {
		return nil, nil
	}
	var eligible []model_state.ParameterSimulationRule
	for i, rule := range param.Simulation.Rules {
		if !rule.HasSpecification() {
			continue
		}
		ok, err := ruleRequiresMet(rule, bindings)
		if err != nil {
			return nil, fmt.Errorf("rule %d: %w", i, err)
		}
		if ok {
			eligible = append(eligible, rule)
		}
	}
	return eligible, nil
}

func ruleRequiresMet(rule model_state.ParameterSimulationRule, bindings *evaluator.Bindings) (bool, error) {
	for _, req := range rule.Requires {
		ok, err := evaluateSimulationAssessment(req, bindings)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

func evaluateSimulationAssessment(req model_logic.Logic, bindings *evaluator.Bindings) (bool, error) { //nolint:unparam // err reserved for lowered-expression failures
	expr := req.Spec.Expression
	if expr == nil {
		if req.Spec.Specification == "" {
			return true, nil
		}
		return false, fmt.Errorf("expression not lowered")
	}
	result := evaluator.Eval(expr, bindings)
	if result.IsError() {
		return false, fmt.Errorf("evaluation error: %s", result.Error.Inspect())
	}
	return isTrueBoolean(result.Value), nil
}

// EvaluateSimulationSpecification evaluates a parameter's simulation into a runtime value.
// Among eligible rules (requires satisfied), one is chosen uniformly at random.
func EvaluateSimulationSpecification(
	param model_state.Parameter,
	bindings *evaluator.Bindings,
	rng *rand.Rand,
) (object.Object, error) {
	if param.Simulation == nil || len(param.Simulation.Rules) == 0 {
		return nil, fmt.Errorf("parameter %q has no simulation rules", param.Name)
	}
	eligible, err := EligibleSimulationRules(param, bindings)
	if err != nil {
		return nil, err
	}
	if len(eligible) == 0 {
		return nil, fmt.Errorf("parameter %q has no eligible simulation rules", param.Name)
	}
	rule := eligible[0]
	if len(eligible) > 1 {
		if rng == nil {
			return nil, fmt.Errorf("parameter %q has multiple eligible simulation rules but no RNG", param.Name)
		}
		rule = eligible[rng.Intn(len(eligible))]
	}
	return evaluateRuleSpecification(param, rule, bindings)
}

func evaluateRuleSpecification(
	param model_state.Parameter,
	rule model_state.ParameterSimulationRule,
	bindings *evaluator.Bindings,
) (object.Object, error) {
	if rule.Specification == nil {
		return nil, fmt.Errorf("parameter %q simulation rule has no specification", param.Name)
	}
	expr := rule.Specification.Spec.Expression
	if expr == nil {
		return nil, fmt.Errorf("parameter %q simulation specification not lowered", param.Name)
	}
	result := evaluator.Eval(expr, bindings)
	if result.IsError() {
		return nil, fmt.Errorf("evaluation error: %s", result.Error.Inspect())
	}
	return CoerceValueForDataType(param.DataType, result.Value), nil
}

// BuildSimulationBindings creates evaluator bindings for sampling on a surface action.
func BuildSimulationBindings(
	builder *state.BindingsBuilder,
	classNameMap map[identity.Key]string,
	instance *state.ClassInstance,
) *evaluator.Bindings {
	if instance != nil {
		return builder.BuildWithClassInstancesForInstance(classNameMap, instance)
	}
	return builder.BuildWithClassInstances(classNameMap)
}

// SurfaceEventSamplingContext carries dependencies for simulator-authored parameter sampling.
type SurfaceEventSamplingContext struct {
	Binder               *ParameterBinder
	Sampler              *ParameterSampler
	Bindings             *evaluator.Bindings
	UsedSimulationParams map[identity.Key]bool
	RNG                  *rand.Rand
}

// SampleSurfaceEventPayload samples event parameters for a surface-scoped transition or do-action.
func SampleSurfaceEventPayload(
	event model_state.Event,
	action *model_state.Action,
	ctx SurfaceEventSamplingContext,
) (map[string]object.Object, error) {
	matched := matchActionParametersByEventNames(event.ParameterNames, action)
	result := make(map[string]object.Object)

	for _, param := range matched {
		if param.Simulation != nil && len(param.Simulation.Rules) > 0 {
			value, err := EvaluateSimulationSpecification(param, ctx.Bindings, ctx.RNG)
			if err != nil {
				return nil, err
			}
			result[param.Name] = value
			if ctx.UsedSimulationParams != nil {
				ctx.UsedSimulationParams[param.Key] = true
			}
			continue
		}
		var sampled map[string]object.Object
		var err error
		if action != nil && ctx.Sampler != nil {
			owner := ParameterOwnerFromAction(*action)
			if owner.NeedsRequiresAwareSampling([]model_state.Parameter{param}) {
				sampled, err = ctx.Sampler.SampleParameters(owner, []model_state.Parameter{param}, ctx.RNG)
			} else {
				sampled = ctx.Binder.GenerateRandomParameters([]model_state.Parameter{param}, ctx.RNG)
			}
		} else {
			sampled = ctx.Binder.GenerateRandomParameters([]model_state.Parameter{param}, ctx.RNG)
		}
		if err != nil {
			return nil, err
		}
		maps.Copy(result, sampled)
	}

	return remapSampledToEventNames(event.ParameterNames, matched, result), nil
}
