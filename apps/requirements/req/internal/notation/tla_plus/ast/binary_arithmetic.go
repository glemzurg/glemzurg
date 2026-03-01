package ast

import (
	"bytes"
	"fmt"
)

// Arithmetic operators
const (
	ArithmeticOperatorAdd      = "+"  // Addition
	ArithmeticOperatorSubtract = "-"  // Subtraction
	ArithmeticOperatorMultiply = "*"  // Multiplication
	ArithmeticOperatorPower    = "^"  // Exponentiation
	ArithmeticOperatorDivide   = "÷"  // Division (Unicode)
	ArithmeticOperatorModulo   = "%"  // Modulo
)

// arithmeticOperatorAscii maps Unicode operators to ASCII equivalents.
var arithmeticOperatorAscii = map[string]string{
	ArithmeticOperatorAdd:      "+",
	ArithmeticOperatorSubtract: "-",
	ArithmeticOperatorMultiply: "*",
	ArithmeticOperatorPower:    "^",
	ArithmeticOperatorDivide:   "\\div",
	ArithmeticOperatorModulo:   "%",
}

// BinaryArithmetic is an arithmetic binary expression (+, -, *, /, ^, %).
type BinaryArithmetic struct {
	Left     Expression `validate:"required"` // The left operand (must be numeric)
	Operator string     `validate:"required,oneof=+ - * ^ ÷ %"`
	Right    Expression `validate:"required"` // The right operand (must be numeric)
}

func (b *BinaryArithmetic) expressionNode() {}

func (b *BinaryArithmetic) String() (value string) {
	var out bytes.Buffer
	out.WriteString(b.Left.String())
	out.WriteString(" ")
	out.WriteString(b.Operator)
	out.WriteString(" ")
	out.WriteString(b.Right.String())
	return out.String()
}

func (b *BinaryArithmetic) Ascii() (value string) {
	var out bytes.Buffer
	out.WriteString(b.Left.Ascii())
	out.WriteString(" ")
	if ascii, ok := arithmeticOperatorAscii[b.Operator]; ok {
		out.WriteString(ascii)
	} else {
		out.WriteString(b.Operator)
	}
	out.WriteString(" ")
	out.WriteString(b.Right.Ascii())
	return out.String()
}

func (b *BinaryArithmetic) Validate() error {
	if err := _validate.Struct(b); err != nil {
		return err
	}
	if err := b.Left.Validate(); err != nil {
		return fmt.Errorf("Left: %w", err)
	}
	if err := b.Right.Validate(); err != nil {
		return fmt.Errorf("Right: %w", err)
	}
	return nil
}

// RealInfixExpression is an alias for backwards compatibility.
// Deprecated: Use BinaryArithmetic instead.
type RealInfixExpression = BinaryArithmetic

// Backwards compatibility constants
// Deprecated: Use ArithmeticOperator* constants instead.
const (
	RealOperatorAdd      = ArithmeticOperatorAdd
	RealOperatorSubtract = ArithmeticOperatorSubtract
	RealOperatorMultiply = ArithmeticOperatorMultiply
	RealOperatorPower    = ArithmeticOperatorPower
	RealOperatorDivide   = ArithmeticOperatorDivide
	RealOperatorModulo   = ArithmeticOperatorModulo
)
