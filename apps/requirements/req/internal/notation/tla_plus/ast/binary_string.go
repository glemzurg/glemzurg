package ast

import (
	"bytes"
	"fmt"
)

// String operators
const (
	StringOperatorConcat = "∘" // Concatenation (U+2218 RING OPERATOR)
)

// stringOperatorAscii maps Unicode operators to ASCII equivalents.
var stringOperatorAscii = map[string]string{
	StringOperatorConcat: `\o`,
}

// StringConcat is a string concatenation expression with two or more operands.
type StringConcat struct {
	Operator string       `validate:"required,oneof=∘"`
	Operands []Expression `validate:"required,min=2"` // At least two string operands (must be String)
}

func (s *StringConcat) expressionNode() {}

func (s *StringConcat) String() (value string) {
	var out bytes.Buffer
	for i, operand := range s.Operands {
		if i > 0 {
			out.WriteString(" ")
			out.WriteString(s.Operator)
			out.WriteString(" ")
		}
		out.WriteString(operand.String())
	}
	return out.String()
}

func (s *StringConcat) Ascii() (value string) {
	var out bytes.Buffer
	ascii := s.Operator
	if a, ok := stringOperatorAscii[s.Operator]; ok {
		ascii = a
	}
	for i, operand := range s.Operands {
		if i > 0 {
			out.WriteString(" ")
			out.WriteString(ascii)
			out.WriteString(" ")
		}
		out.WriteString(operand.Ascii())
	}
	return out.String()
}

func (s *StringConcat) Validate() error {
	if err := _validate.Struct(s); err != nil {
		return err
	}
	for i, operand := range s.Operands {
		if operand == nil {
			return fmt.Errorf("Operands[%d]: is nil", i)
		}
		if err := operand.Validate(); err != nil {
			return fmt.Errorf("Operands[%d]: %w", i, err)
		}
	}
	return nil
}

// StringInfixExpression is an alias for backwards compatibility.
// Deprecated: Use StringConcat instead.
type StringInfixExpression = StringConcat
