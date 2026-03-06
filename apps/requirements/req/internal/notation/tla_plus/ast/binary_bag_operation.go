package ast

import "bytes"

// Bag operation operators
const (
	BagOperatorUnion       = `⊕` // Bag union (U+2295 CIRCLED PLUS)
	BagOperatorSubtraction = `⊖` // Bag subtraction (U+2296 CIRCLED MINUS)
)

var bagOperationAscii = map[string]string{
	BagOperatorUnion:       `(+)`,
	BagOperatorSubtraction: `(-)`,
}

// BinaryBagOperation is a binary operation between two bags that produces a bag.
type BinaryBagOperation struct {
	Operator string     `validate:"required,oneof=⊕ ⊖"` // The bag operator, e.g., ⊕, ⊖
	Left     Expression `validate:"required"`           // Must be Bag
	Right    Expression `validate:"required"`           // Must be Bag
}

func (bi *BinaryBagOperation) expressionNode() {}

func (bi *BinaryBagOperation) String() (value string) {
	var out bytes.Buffer
	out.WriteString(bi.Left.String())
	out.WriteString(" " + bi.Operator + " ")
	out.WriteString(bi.Right.String())
	return out.String()
}

func (bi *BinaryBagOperation) Ascii() (value string) {
	var out bytes.Buffer
	out.WriteString(bi.Left.Ascii())
	out.WriteString(" " + bagOperationAscii[bi.Operator] + " ")
	out.WriteString(bi.Right.Ascii())
	return out.String()
}

func (bi *BinaryBagOperation) Validate() error {
	if err := _validate.Struct(bi); err != nil {
		return err
	}
	if err := bi.Left.Validate(); err != nil {
		return err
	}
	if err := bi.Right.Validate(); err != nil {
		return err
	}
	return nil
}

// BagInfix is an alias for backwards compatibility.
// Deprecated: Use BinaryBagOperation instead.
type BagInfix = BinaryBagOperation
