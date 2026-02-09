package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// evalIfElse evaluates IF condition THEN expr ELSE expr.
func evalIfElse(node *ast.ExpressionIfElse, bindings *Bindings) *EvalResult {
	condResult := Eval(node.Condition, bindings)
	if condResult.IsError() {
		return condResult
	}

	condBool, ok := condResult.Value.(*object.Boolean)
	if !ok {
		return NewEvalError("IF condition must be Boolean, got %s", condResult.Value.Type())
	}

	if condBool.Value() {
		return Eval(node.Then, bindings)
	}
	return Eval(node.Else, bindings)
}
