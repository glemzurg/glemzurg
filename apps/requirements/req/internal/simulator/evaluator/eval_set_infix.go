package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// evalSetInfix evaluates a set infix expression (∪, ∩, \).
func evalSetInfix(node *ast.SetInfix, bindings *Bindings) *EvalResult {
	leftResult := Eval(node.Left, bindings)
	if leftResult.IsError() {
		return leftResult
	}

	rightResult := Eval(node.Right, bindings)
	if rightResult.IsError() {
		return rightResult
	}

	leftSet, ok := leftResult.Value.(*object.Set)
	if !ok {
		return NewEvalError("left operand must be Set, got %s", leftResult.Value.Type())
	}

	rightSet, ok := rightResult.Value.(*object.Set)
	if !ok {
		return NewEvalError("right operand must be Set, got %s", rightResult.Value.Type())
	}

	var result *object.Set
	switch node.Operator {
	case "∪", "\\union":
		result = leftSet.Union(rightSet)
	case "∩", "\\intersect":
		result = leftSet.Intersection(rightSet)
	case "\\", "\\setminus":
		result = leftSet.Difference(rightSet)
	default:
		return NewEvalError("unknown set operator: %s", node.Operator)
	}

	return NewEvalResult(result)
}
