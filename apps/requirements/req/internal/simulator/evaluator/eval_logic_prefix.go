package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// evalLogicPrefix evaluates a logic prefix expression (¬).
func evalLogicPrefix(node *ast.LogicPrefixExpression, bindings *Bindings) *EvalResult {
	result := Eval(node.Right, bindings)
	if result.IsError() {
		return result
	}

	boolVal, ok := result.Value.(*object.Boolean)
	if !ok {
		return NewEvalError("operand must be Boolean, got %s", result.Value.Type())
	}

	switch node.Operator {
	case "¬", "~":
		return NewEvalResult(nativeBoolToBoolean(!boolVal.Value()))
	default:
		return NewEvalError("unknown logic prefix operator: %s", node.Operator)
	}
}
