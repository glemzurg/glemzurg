package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// evalIfElse evaluates IF condition THEN expr ELSE expr.
func evalIfElse(node *ast.IfThenElse, bindings *Bindings) *EvalResult {
	condResult := EvalAST(node.Condition, bindings)
	if condResult.IsError() {
		return condResult
	}

	condBool, ok := condResult.Value.(*object.Boolean)
	if !ok {
		return NewEvalError("IF condition must be Boolean, got %s", condResult.Value.Type())
	}

	if condBool.Value() {
		return EvalAST(node.Then, bindings)
	}
	return EvalAST(node.Else, bindings)
}
