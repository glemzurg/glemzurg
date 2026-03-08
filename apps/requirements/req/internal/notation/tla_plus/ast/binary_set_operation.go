package ast

import "bytes"

// Set operation operators.
const (
	SetOperatorUnion        = `∪`
	SetOperatorIntersection = `∩`
	SetOperatorDifference   = `\`
)

var setOperationASCII = map[string]string{
	SetOperatorUnion:        `\union`,
	SetOperatorIntersection: `\intersect`,
	SetOperatorDifference:   `\`,
}

// BinarySetOperation is a binary operation between two sets that produces a set.
type BinarySetOperation struct {
	Operator string     `validate:"required,oneof=∪ ∩ \\"` // The set operator, e.g., ∪, ∩, \
	Left     Expression `validate:"required"`              // Must be Set
	Right    Expression `validate:"required"`              // Must be Set
}

func (si *BinarySetOperation) expressionNode() {}

func (si *BinarySetOperation) String() (value string) {
	var out bytes.Buffer
	out.WriteString(si.Left.String())
	out.WriteString(" " + si.Operator + " ")
	out.WriteString(si.Right.String())
	return out.String()
}

func (si *BinarySetOperation) ASCII() (value string) {
	var out bytes.Buffer
	out.WriteString(si.Left.ASCII())
	out.WriteString(" " + setOperationASCII[si.Operator] + " ")
	out.WriteString(si.Right.ASCII())
	return out.String()
}

func (si *BinarySetOperation) Validate() error {
	if err := _validate.Struct(si); err != nil {
		return err
	}
	if err := si.Left.Validate(); err != nil {
		return err
	}
	if err := si.Right.Validate(); err != nil {
		return err
	}
	return nil
}

// SetInfix is an alias for backwards compatibility.
//
// Deprecated: Use BinarySetOperation instead.
type SetInfix = BinarySetOperation
