package ast

import "bytes"

// Bag comparison operators
const (
	BagComparisonProperSubBag = "⊏" // Proper subbag (U+228F SQUARE IMAGE OF)
	BagComparisonSubBag       = "⊑" // Subbag or equal (U+2291 SQUARE IMAGE OF OR EQUAL TO)
	BagComparisonProperSupBag = "⊐" // Proper superbag (U+2290 SQUARE ORIGINAL OF)
	BagComparisonSupBag       = "⊒" // Superbag or equal (U+2292 SQUARE ORIGINAL OF OR EQUAL TO)
)

var bagComparisonAscii = map[string]string{
	BagComparisonProperSubBag: `\sqsubset`,
	BagComparisonSubBag:       `\sqsubseteq`,
	BagComparisonProperSupBag: `\sqsupset`,
	BagComparisonSupBag:       `\sqsupseteq`,
}

// Backwards compatibility constants.
// Deprecated: Use BagComparison* constants instead.
const (
	LogicBagOperatorSubBag = BagComparisonSubBag
)

// BinaryBagComparison is a binary comparison operation between two bags.
type BinaryBagComparison struct {
	Operator string     `validate:"required,oneof=⊏ ⊑ ⊐ ⊒"` // The comparison operator
	Left     Expression `validate:"required"`                // Must be Bag
	Right    Expression `validate:"required"`                // Must be Bag
}

func (ib *BinaryBagComparison) expressionNode() {}

func (ib *BinaryBagComparison) String() (value string) {
	var out bytes.Buffer
	out.WriteString(ib.Left.String())
	out.WriteString(" " + ib.Operator + " ")
	out.WriteString(ib.Right.String())
	return out.String()
}

func (ib *BinaryBagComparison) Ascii() (value string) {
	var out bytes.Buffer
	out.WriteString(ib.Left.Ascii())
	out.WriteString(" " + bagComparisonAscii[ib.Operator] + " ")
	out.WriteString(ib.Right.Ascii())
	return out.String()
}

func (ib *BinaryBagComparison) Validate() error {
	if err := _validate.Struct(ib); err != nil {
		return err
	}
	if err := ib.Left.Validate(); err != nil {
		return err
	}
	if err := ib.Right.Validate(); err != nil {
		return err
	}
	return nil
}

// LogicInfixBag is an alias for backwards compatibility.
// Deprecated: Use BinaryBagComparison instead.
type LogicInfixBag = BinaryBagComparison
