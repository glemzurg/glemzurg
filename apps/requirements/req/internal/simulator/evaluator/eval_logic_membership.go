package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// evalLogicMembership evaluates a set membership test (∈, ∉).
func evalLogicMembership(node *ast.Membership, bindings *Bindings) *EvalResult {
	leftResult := EvalAST(node.Left, bindings)
	if leftResult.IsError() {
		return leftResult
	}

	rightResult := EvalAST(node.Right, bindings)
	if rightResult.IsError() {
		return rightResult
	}

	set, ok := rightResult.Value.(*object.Set)
	if !ok {
		return NewEvalError("membership test requires Set, got %s", rightResult.Value.Type())
	}

	contains := set.Contains(leftResult.Value)

	switch node.Operator {
	case "∈", "\\in":
		return NewEvalResult(nativeBoolToBoolean(contains))
	case "∉", "\\notin":
		return NewEvalResult(nativeBoolToBoolean(!contains))
	default:
		return NewEvalError("unknown membership operator: %s", node.Operator)
	}
}
