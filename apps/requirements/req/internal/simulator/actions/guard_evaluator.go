package actions

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/model_bridge"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/parser"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

// GuardEvaluator evaluates guard conditions on class instances.
// Guards are TLA+ boolean expressions that must all be TRUE for
// the guard to pass. Multiple TlaGuard entries are ANDed together.
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

	for i, guardStr := range guard.TlaGuard {
		expr, err := parser.ParseExpression(guardStr)
		if err != nil {
			return false, fmt.Errorf("guard %s[%d] parse error: %w", guard.Name, i, err)
		}

		if model_bridge.ContainsAnyPrimed(expr) {
			return false, fmt.Errorf("guard %s[%d]: guards must not contain primed variables", guard.Name, i)
		}

		result := evaluator.Eval(expr, bindings)

		if result.IsError() {
			return false, fmt.Errorf("guard %s[%d] evaluation error: %s", guard.Name, i, result.Error.Inspect())
		}

		if !isTrueBoolean(result.Value) {
			return false, nil // Guard not satisfied (not an error)
		}
	}

	return true, nil // All guard expressions passed
}
