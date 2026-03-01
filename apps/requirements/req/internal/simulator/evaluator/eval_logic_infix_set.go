package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// evalLogicInfixSet evaluates a set comparison (=, ≠, ⊆, ⊂, ⊇, ⊃).
func evalLogicInfixSet(node *ast.LogicInfixSet, bindings *Bindings) *EvalResult {
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

	var result bool
	switch node.Operator {
	case "=":
		result = leftSet.Equals(rightSet)
	case "≠", "/=", "#":
		result = !leftSet.Equals(rightSet)
	case "⊆", "\\subseteq":
		result = leftSet.IsSubsetOf(rightSet)
	case "⊂", "\\subset":
		result = leftSet.IsSubsetOf(rightSet) && !leftSet.Equals(rightSet)
	case "⊇", "\\supseteq":
		result = rightSet.IsSubsetOf(leftSet)
	case "⊃", "\\supset":
		result = rightSet.IsSubsetOf(leftSet) && !leftSet.Equals(rightSet)
	default:
		return NewEvalError("unknown set comparison operator: %s", node.Operator)
	}

	return NewEvalResult(nativeBoolToBoolean(result))
}
