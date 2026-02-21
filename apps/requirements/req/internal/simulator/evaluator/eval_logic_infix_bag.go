package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// evalLogicInfixBag evaluates a bag comparison (⊑).
func evalLogicInfixBag(node *ast.LogicInfixBag, bindings *Bindings) *EvalResult {
	leftResult := Eval(node.Left, bindings)
	if leftResult.IsError() {
		return leftResult
	}

	rightResult := Eval(node.Right, bindings)
	if rightResult.IsError() {
		return rightResult
	}

	leftBag, ok := leftResult.Value.(*object.Bag)
	if !ok {
		return NewEvalError("left operand must be Bag, got %s", leftResult.Value.Type())
	}

	rightBag, ok := rightResult.Value.(*object.Bag)
	if !ok {
		return NewEvalError("right operand must be Bag, got %s", rightResult.Value.Type())
	}

	var result bool
	switch node.Operator {
	case "⊏", "\\sqsubset":
		result = leftBag.IsProperSubBagOf(rightBag)
	case "⊑", "\\sqsubseteq":
		result = leftBag.IsSubBagOf(rightBag)
	case "⊐", "\\sqsupset":
		result = leftBag.IsProperSuperBagOf(rightBag)
	case "⊒", "\\sqsupseteq":
		result = leftBag.IsSuperBagOf(rightBag)
	default:
		return NewEvalError("unknown bag comparison operator: %s", node.Operator)
	}
	return NewEvalResult(nativeBoolToBoolean(result))
}
