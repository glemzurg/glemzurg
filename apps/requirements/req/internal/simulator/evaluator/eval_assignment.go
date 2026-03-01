package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
)

// evalAssignment evaluates a statement (target' = expression).
// This is one of the two valid root nodes.
func evalAssignment(node *ast.Assignment, bindings *Bindings) *EvalResult {
	// Evaluate the value expression
	valueResult := Eval(node.Value, bindings)
	if valueResult.IsError() {
		return valueResult
	}

	// Prime the target binding
	targetName := node.Target.Value
	bindings.SetPrimed(targetName, valueResult.Value)

	// Return success with the primed bindings
	return NewEvalResultWithPrimed(
		EMPTY_SET, // Success indicator (no NULL in TLA+)
		bindings.GetPrimedBindings(),
	)
}
