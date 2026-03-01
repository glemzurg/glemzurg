package actions

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/model_bridge"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/parser"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

// GuardEvaluator evaluates guard conditions on class instances.
// Guards are boolean expressions that must evaluate to TRUE for
// the guard to pass.
type GuardEvaluator struct {
	bindingsBuilder *state.BindingsBuilder
}

// NewGuardEvaluator creates a new guard evaluator.
func NewGuardEvaluator(bindingsBuilder *state.BindingsBuilder) *GuardEvaluator {
	return &GuardEvaluator{
		bindingsBuilder: bindingsBuilder,
	}
}

// EvaluateGuard checks if all guard expressions evaluate to TRUE
// for the given instance. Returns false if any expression is not TRUE.
func (g *GuardEvaluator) EvaluateGuard(
	guard model_state.Guard,
	instance *state.ClassInstance,
) (bool, error) {
	bindings := g.bindingsBuilder.BuildForInstance(instance)

	if guard.Logic.Spec.Specification == "" {
		return true, nil // No specification means guard always passes
	}

	expr, err := parser.ParseExpression(guard.Logic.Spec.Specification)
	if err != nil {
		return false, fmt.Errorf("guard %s parse error: %w", guard.Name, err)
	}

	if model_bridge.ContainsAnyPrimed(expr) {
		return false, fmt.Errorf("guard %s: guards must not contain primed variables", guard.Name)
	}

	result := evaluator.Eval(expr, bindings)

	if result.IsError() {
		return false, fmt.Errorf("guard %s evaluation error: %s", guard.Name, result.Error.Inspect())
	}

	if !isTrueBoolean(result.Value) {
		return false, nil // Guard not satisfied (not an error)
	}

	return true, nil // Guard expression passed
}
