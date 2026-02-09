package evaluator

import (
	"github.com/glemzurg/go-tlaplus/internal/simulator/ast"
	"github.com/glemzurg/go-tlaplus/internal/simulator/object"
)

// evalRealInfix evaluates an arithmetic infix expression.
func evalRealInfix(node *ast.RealInfixExpression, bindings *Bindings) *EvalResult {
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

	var result *object.Number

	switch node.Operator {
	case "+":
		result = leftNum.Add(rightNum)
	case "-":
		result = leftNum.Sub(rightNum)
	case "*":
		result = leftNum.Mul(rightNum)
	case "÷", "/":
		if rightNum.IsZero() {
			return NewEvalError("division by zero")
		}
		result = leftNum.Div(rightNum)
	case "%":
		mod, err := leftNum.Mod(rightNum)
		if err != nil {
			return NewEvalError("modulo error: %v", err)
		}
		result = mod
	case "^":
		pow, err := leftNum.Pow(rightNum)
		if err != nil {
			return NewEvalError("power error: %v", err)
		}
		result = pow
	default:
		return NewEvalError("unknown arithmetic operator: %s", node.Operator)
	}

	return NewEvalResult(result)
}
