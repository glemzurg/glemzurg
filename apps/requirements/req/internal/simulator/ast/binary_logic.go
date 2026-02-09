package ast

import "bytes"

const (
	LogicOperatorAnd     = "∧"
	LogicOperatorOr      = "∨"
	LogicOperatorImplies = "⇒"
	LogicOperatorEquiv   = "≡"
)

var binaryLogicAscii = map[string]string{
	LogicOperatorAnd:     `/\`,
	LogicOperatorOr:      `\/`,
	LogicOperatorImplies: `=>`,
	LogicOperatorEquiv:   `<=>`,
}

// BinaryLogic is a binary logic expression (∧, ∨, ⇒, ≡).
type BinaryLogic struct {
	Operator string     `validate:"required,oneof=∧ ∨ ⇒ ≡"` // The logic operator: ∧, ∨, ⇒, ≡
	Left     Expression `validate:"required"`                 // Must be Boolean
	Right    Expression `validate:"required"`                 // Must be Boolean
}

func (b *BinaryLogic) expressionNode() {}

func (b *BinaryLogic) String() (value string) {
	var out bytes.Buffer
	out.WriteString(b.Left.String())
	out.WriteString(" " + b.Operator + " ")
	out.WriteString(b.Right.String())
	return out.String()
}

func (b *BinaryLogic) Ascii() (value string) {
	var out bytes.Buffer
	out.WriteString(b.Left.Ascii())
	out.WriteString(" " + binaryLogicAscii[b.Operator] + " ")
	out.WriteString(b.Right.Ascii())
	return out.String()
}

func (b *BinaryLogic) Validate() error {
	if err := _validate.Struct(b); err != nil {
		return err
	}
	if err := b.Left.Validate(); err != nil {
		return err
	}
	if err := b.Right.Validate(); err != nil {
		return err
	}
	return nil
}

// LogicInfixExpression is an alias for backwards compatibility.
// Deprecated: Use BinaryLogic instead.
type LogicInfixExpression = BinaryLogic
