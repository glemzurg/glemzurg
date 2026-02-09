package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// evalBagInfix evaluates a bag infix expression (⊕, ⊖).
func evalBagInfix(node *ast.BagInfix, bindings *Bindings) *EvalResult {
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

	var result *object.Bag
	switch node.Operator {
	case "⊕", "(+)":
		result = leftBag.Sum(rightBag)
	case "⊖", "(-)":
		result = leftBag.Difference(rightBag)
	default:
		return NewEvalError("unknown bag operator: %s", node.Operator)
	}

	return NewEvalResult(result)
}
