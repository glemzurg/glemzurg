package evaluator

import (
	"strconv"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// evalNumberLiteral evaluates a number literal.
// Returns a Number object.
func evalNumberLiteral(node *ast.NumberLiteral) *EvalResult {
	// Parse the integer part (may be empty for ".5")
	intVal := int64(0)
	if node.IntegerPart != "" {
		var err error
		intVal, err = parseIntegerPart(node)
		if err != nil {
			return NewEvalError("invalid number literal: %v", err)
		}
	}

	// If no decimal point, return as integer
	if !node.HasDecimalPoint {
		return NewEvalResult(object.NewNatural(intVal))
	}

	// Decimal representation: convert to Real
	// whole.precision = (whole * denom + precision) / denom
	precision := node.FractionalPart
	precisionVal, err := strconv.ParseInt(precision, 10, 64)
	if err != nil {
		return NewEvalError("invalid fractional part in number literal: %s", precision)
	}

	// Calculate denominator: 10^len(precision)
	denom := int64(1)
	for range len(precision) {
		denom *= 10
	}

	numerator := intVal*denom + precisionVal
	return NewEvalResult(object.NewReal(numerator, denom))
}

// parseIntegerPart parses the integer part of a NumberLiteral according to its base.
func parseIntegerPart(node *ast.NumberLiteral) (int64, error) {
	switch node.Base {
	case ast.BaseDecimal:
		return strconv.ParseInt(node.IntegerPart, 10, 64)
	case ast.BaseBinary:
		return strconv.ParseInt(node.IntegerPart, 2, 64)
	case ast.BaseOctal:
		return strconv.ParseInt(node.IntegerPart, 8, 64)
	case ast.BaseHex:
		return strconv.ParseInt(node.IntegerPart, 16, 64)
	default:
		return 0, nil
	}
}

// evalNumericPrefixExpression evaluates a numeric prefix expression (negation).
func evalNumericPrefixExpression(node *ast.NumericPrefixExpression, bindings *Bindings) *EvalResult {
	// Evaluate the right operand
	rightResult := Eval(node.Right, bindings)
	if rightResult.IsError() {
		return rightResult
	}

	rightObj := rightResult.Value

	// Handle negation
	if node.Operator == "-" {
		num, ok := rightObj.(*object.Number)
		if !ok {
			return NewEvalError("cannot negate non-numeric value: %T", rightObj)
		}
		return NewEvalResult(num.Neg())
	}

	return NewEvalError("unknown numeric prefix operator: %s", node.Operator)
}

// evalFractionExpr evaluates a fraction expression (a/b).
func evalFractionExpr(node *ast.FractionExpr, bindings *Bindings) *EvalResult {
	// Evaluate numerator
	numResult := Eval(node.Numerator, bindings)
	if numResult.IsError() {
		return numResult
	}

	// Evaluate denominator
	denomResult := Eval(node.Denominator, bindings)
	if denomResult.IsError() {
		return denomResult
	}

	// Both must be numbers
	numNum, ok := numResult.Value.(*object.Number)
	if !ok {
		return NewEvalError("fraction numerator must be numeric, got %T", numResult.Value)
	}

	denomNum, ok := denomResult.Value.(*object.Number)
	if !ok {
		return NewEvalError("fraction denominator must be numeric, got %T", denomResult.Value)
	}

	if denomNum.IsZero() {
		return NewEvalError("division by zero")
	}

	return NewEvalResult(numNum.Div(denomNum))
}

// evalParenExpr evaluates a parenthesized expression.
func evalParenExpr(node *ast.ParenExpr, bindings *Bindings) *EvalResult {
	return Eval(node.Inner, bindings)
}
