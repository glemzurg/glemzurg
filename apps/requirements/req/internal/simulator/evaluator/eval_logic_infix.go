package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// _OP_AND is the Unicode conjunction operator (logical AND).
const _OP_AND = "∧"

// _OP_OR is the Unicode disjunction operator (logical OR).
const _OP_OR = "∨"

// evalLogicInfix evaluates a logic infix expression (∧, ∨, ⇒, ≡).
func evalLogicInfix(node *ast.LogicInfixExpression, bindings *Bindings) *EvalResult {
	leftResult := EvalAST(node.Left, bindings)
	if leftResult.IsError() {
		return leftResult
	}

	leftBool, ok := leftResult.Value.(*object.Boolean)
	if !ok {
		return NewEvalError("left operand must be Boolean, got %s", leftResult.Value.Type())
	}

	// Short-circuit evaluation for AND and OR
	switch node.Operator {
	case _OP_AND, "/\\":
		if !leftBool.Value() {
			return NewEvalResult(FALSE)
		}
	case _OP_OR, "\\/":
		if leftBool.Value() {
			return NewEvalResult(TRUE)
		}
	}

	rightResult := EvalAST(node.Right, bindings)
	if rightResult.IsError() {
		return rightResult
	}

	rightBool, ok := rightResult.Value.(*object.Boolean)
	if !ok {
		return NewEvalError("right operand must be Boolean, got %s", rightResult.Value.Type())
	}

	var result bool
	switch node.Operator {
	case _OP_AND, "/\\":
		result = leftBool.Value() && rightBool.Value()
	case _OP_OR, "\\/":
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
