package ast

import "bytes"

// Negation operator
const (
	NegationOperator = "-" // Unary negation
)

// UnaryNegation is a unary negation expression (-).
type UnaryNegation struct {
	Operator string     `validate:"required,oneof=-"` // The operator (-)
	Right    Expression `validate:"required"`         // The operand
}

func (u *UnaryNegation) expressionNode() {}

func (u *UnaryNegation) String() string {
	var out bytes.Buffer
	out.WriteString(u.Operator)
	out.WriteString(u.Right.String())
	return out.String()
}

func (u *UnaryNegation) Ascii() string {
	return u.String()
}

func (u *UnaryNegation) Validate() error {
	if err := _validate.Struct(u); err != nil {
		return err
	}
	if err := u.Right.Validate(); err != nil {
		return err
	}
	return nil
}

// NewNegation creates a negation expression.
func NewNegation(operand Expression) *UnaryNegation {
	return &UnaryNegation{
		Operator: NegationOperator,
		Right:    operand,
	}
}

// NumericPrefixExpression is an alias for backwards compatibility.
// Deprecated: Use UnaryNegation instead.
type NumericPrefixExpression = UnaryNegation

// Backwards compatibility constant.
// Deprecated: Use NegationOperator instead.
const NumericOperatorNegate = NegationOperator
