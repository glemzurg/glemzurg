package ast

import "bytes"

const (
	LogicOperatorNot = "¬"
)

var unaryLogicAscii = map[string]string{
	LogicOperatorNot: `~`,
}

// UnaryLogic is a unary logic expression (¬).
type UnaryLogic struct {
	Operator string     `validate:"required,oneof=¬"` // The logic operator: ¬
	Right    Expression `validate:"required"`         // Must be Boolean
}

func (u *UnaryLogic) expressionNode() {}

func (u *UnaryLogic) String() (value string) {
	var out bytes.Buffer
	out.WriteString(u.Operator)
	out.WriteString(u.Right.String())
	return out.String()
}

func (u *UnaryLogic) Ascii() (value string) {
	var out bytes.Buffer
	out.WriteString(unaryLogicAscii[u.Operator])
	out.WriteString(u.Right.Ascii())
	return out.String()
}

func (u *UnaryLogic) Validate() error {
	if err := _validate.Struct(u); err != nil {
		return err
	}
	if err := u.Right.Validate(); err != nil {
		return err
	}
	return nil
}

// LogicPrefixExpression is an alias for backwards compatibility.
// Deprecated: Use UnaryLogic instead.
type LogicPrefixExpression = UnaryLogic
