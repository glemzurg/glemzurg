package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// evalLogicRealComparison evaluates a numeric comparison (<, >, ≤, ≥).
func evalLogicRealComparison(node *ast.LogicRealComparison, bindings *Bindings) *EvalResult {
	leftResult := Eval(node.Left, bindings)
	if leftResult.IsError() {
		return leftResult
	}

	rightResult := Eval(node.Right, bindings)
	if rightResult.IsError() {
		return rightResult
	}

	leftNum, ok := leftResult.Value.(*object.Number)
	if !ok {
		return NewEvalError("left operand must be Number, got %s", leftResult.Value.Type())
	}

	rightNum, ok := rightResult.Value.(*object.Number)
	if !ok {
		return NewEvalError("right operand must be Number, got %s", rightResult.Value.Type())
	}

	cmp := leftNum.Cmp(rightNum)
	var result bool

	switch node.Operator {
	case "<":
		result = cmp < 0
	case ">":
		result = cmp > 0
	case "≤", "<=":
		result = cmp <= 0
	case "≥", ">=":
		result = cmp >= 0
	default:
		return NewEvalError("unknown comparison operator: %s", node.Operator)
	}

	return NewEvalResult(nativeBoolToBoolean(result))
}
