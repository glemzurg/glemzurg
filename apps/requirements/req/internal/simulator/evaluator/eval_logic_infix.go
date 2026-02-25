package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// evalLogicInfix evaluates a logic infix expression (∧, ∨, ⇒, ≡).
func evalLogicInfix(node *ast.LogicInfixExpression, bindings *Bindings) *EvalResult {
	leftResult := Eval(node.Left, bindings)
	if leftResult.IsError() {
		return leftResult
	}

	leftBool, ok := leftResult.Value.(*object.Boolean)
	if !ok {
		return NewEvalError("left operand must be Boolean, got %s", leftResult.Value.Type())
	}

	// Short-circuit evaluation for AND and OR
	switch node.Operator {
	case "∧", "/\\":
		if !leftBool.Value() {
			return NewEvalResult(FALSE)
		}
	case "∨", "\\/":
		if leftBool.Value() {
			return NewEvalResult(TRUE)
		}
	}

	rightResult := Eval(node.Right, bindings)
	if rightResult.IsError() {
		return rightResult
	}

	rightBool, ok := rightResult.Value.(*object.Boolean)
	if !ok {
		return NewEvalError("right operand must be Boolean, got %s", rightResult.Value.Type())
	}

	var result bool
	switch node.Operator {
	case "∧", "/\\":
		result = leftBool.Value() && rightBool.Value()
	case "∨", "\\/":
		result = leftBool.Value() || rightBool.Value()
	case "⇒", "=>":
		// A => B is equivalent to !A || B
		result = !leftBool.Value() || rightBool.Value()
	case "≡", "<=>":
		// A <=> B is equivalent to A == B
		result = leftBool.Value() == rightBool.Value()
	default:
		return NewEvalError("unknown logic operator: %s", node.Operator)
	}

	return NewEvalResult(nativeBoolToBoolean(result))
}
