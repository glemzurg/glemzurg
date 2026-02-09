package evaluator

import (
	"github.com/glemzurg/go-tlaplus/internal/simulator/ast"
	"github.com/glemzurg/go-tlaplus/internal/simulator/object"
)

// evalCase evaluates a CASE expression.
func evalCase(node *ast.ExpressionCase, bindings *Bindings) *EvalResult {
	// Evaluate each branch condition
	for _, branch := range node.Branches {
		condResult := Eval(branch.Condition, bindings)
		if condResult.IsError() {
			return condResult
		}

		condBool, ok := condResult.Value.(*object.Boolean)
		if !ok {
			return NewEvalError("CASE branch condition must be Boolean, got %s", condResult.Value.Type())
		}

		if condBool.Value() {
			return Eval(branch.Result, bindings)
		}
	}

	// No branch matched - check for OTHER clause
	if node.Other != nil {
		return Eval(node.Other, bindings)
	}

	return NewEvalError("CASE: no branch matched and no OTHER clause")
}
