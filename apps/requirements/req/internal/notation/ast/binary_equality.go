package ast

import "bytes"

const (
	EqualityOperatorEqual    = "="
	EqualityOperatorNotEqual = "≠"
)

var _asciiEqualityOperatorMap = map[string]string{
	EqualityOperatorEqual:    `=`,
	EqualityOperatorNotEqual: `/=`,
}

// BinaryEquality is a binary equality comparison operation between two values of any type.
// Unlike BinarySetComparison which only handles sets, this handles equality for all types:
// numbers, strings, booleans, sets, tuples, records, etc.
type BinaryEquality struct {
	Operator string     `validate:"required,oneof== ≠"` // The equality operator: = or ≠
	Left     Expression `validate:"required"`           // Left operand (any type)
	Right    Expression `validate:"required"`           // Right operand (any type)
}

func (le *BinaryEquality) expressionNode() {}

func (le *BinaryEquality) String() (value string) {
	var out bytes.Buffer
	out.WriteString(le.Left.String())
	out.WriteString(" " + le.Operator + " ")
	out.WriteString(le.Right.String())
	return out.String()
}

func (le *BinaryEquality) Ascii() (value string) {
	var out bytes.Buffer
	out.WriteString(le.Left.Ascii())
	out.WriteString(" " + _asciiEqualityOperatorMap[le.Operator] + " ")
	out.WriteString(le.Right.Ascii())
	return out.String()
}

func (le *BinaryEquality) Validate() error {
	if err := _validate.Struct(le); err != nil {
		return err
	}
	if err := le.Left.Validate(); err != nil {
		return err
	}
	if err := le.Right.Validate(); err != nil {
		return err
	}
	return nil
}

// LogicEquality is an alias for backwards compatibility.
// Deprecated: Use BinaryEquality instead.
type LogicEquality = BinaryEquality
