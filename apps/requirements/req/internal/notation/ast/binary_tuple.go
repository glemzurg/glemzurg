package ast

import (
	"bytes"
	"fmt"
)

// Tuple operators
const (
	TupleOperatorConcat = "∘" // Concatenation (U+2218 RING OPERATOR)
)

// tupleOperatorAscii maps Unicode operators to ASCII equivalents.
var tupleOperatorAscii = map[string]string{
	TupleOperatorConcat: `\o`,
}

// TupleConcat is a tuple concatenation expression with two or more operands.
type TupleConcat struct {
	Operator string       `validate:"required,oneof=∘"`
	Operands []Expression `validate:"required,min=2"` // At least two tuple operands (must be Tuple)
}

func (t *TupleConcat) expressionNode() {}

func (t *TupleConcat) String() (value string) {
	var out bytes.Buffer
	for i, operand := range t.Operands {
		if i > 0 {
			out.WriteString(" ")
			out.WriteString(t.Operator)
			out.WriteString(" ")
		}
		out.WriteString(operand.String())
	}
	return out.String()
}

func (t *TupleConcat) Ascii() (value string) {
	var out bytes.Buffer
	ascii := t.Operator
	if a, ok := tupleOperatorAscii[t.Operator]; ok {
		ascii = a
	}
	for i, operand := range t.Operands {
		if i > 0 {
			out.WriteString(" ")
			out.WriteString(ascii)
			out.WriteString(" ")
		}
		out.WriteString(operand.Ascii())
	}
	return out.String()
}

func (t *TupleConcat) Validate() error {
	if err := _validate.Struct(t); err != nil {
		return err
	}
	for i, operand := range t.Operands {
		if operand == nil {
			return fmt.Errorf("Operands[%d]: is nil", i)
		}
		if err := operand.Validate(); err != nil {
			return fmt.Errorf("Operands[%d]: %w", i, err)
		}
	}
	return nil
}

// TupleInfixExpression is an alias for backwards compatibility.
// Deprecated: Use TupleConcat instead.
type TupleInfixExpression = TupleConcat
