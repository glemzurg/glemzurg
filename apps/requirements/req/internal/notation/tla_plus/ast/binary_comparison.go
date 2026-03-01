package ast

import (
	"bytes"
	"fmt"
)

// Comparison operators
const (
	ComparisonLessThan           = "<"  // Less than
	ComparisonGreaterThan        = ">"  // Greater than
	ComparisonLessThanOrEqual    = "≤"  // Less than or equal (Unicode)
	ComparisonGreaterThanOrEqual = "≥"  // Greater than or equal (Unicode)
)

// comparisonAscii maps Unicode operators to ASCII equivalents.
var comparisonAscii = map[string]string{
	ComparisonLessThan:           "<",
	ComparisonGreaterThan:        ">",
	ComparisonLessThanOrEqual:    "=<",
	ComparisonGreaterThanOrEqual: ">=",
}

// BinaryComparison is a numeric comparison expression (<, >, ≤, ≥).
type BinaryComparison struct {
	Left     Expression `validate:"required"` // The left operand (must be numeric)
	Operator string     `validate:"required,oneof=< > ≤ ≥"`
	Right    Expression `validate:"required"` // The right operand (must be numeric)
}

func (b *BinaryComparison) expressionNode() {}

func (b *BinaryComparison) String() (value string) {
	var out bytes.Buffer
	out.WriteString(b.Left.String())
	out.WriteString(" ")
	out.WriteString(b.Operator)
	out.WriteString(" ")
	out.WriteString(b.Right.String())
	return out.String()
}

func (b *BinaryComparison) Ascii() (value string) {
	var out bytes.Buffer
	out.WriteString(b.Left.Ascii())
	out.WriteString(" ")
	if ascii, ok := comparisonAscii[b.Operator]; ok {
		out.WriteString(ascii)
	} else {
		out.WriteString(b.Operator)
	}
	out.WriteString(" ")
	out.WriteString(b.Right.Ascii())
	return out.String()
}

func (b *BinaryComparison) Validate() error {
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

// LogicRealComparison is an alias for backwards compatibility.
// Deprecated: Use BinaryComparison instead.
type LogicRealComparison = BinaryComparison

// Backwards compatibility constants.
// Deprecated: Use Comparison* constants instead.
const (
	RealComparisonLessThan           = ComparisonLessThan
	RealComparisonGreaterThan        = ComparisonGreaterThan
	RealComparisonLessThanOrEqual    = ComparisonLessThanOrEqual
	RealComparisonGreaterThanOrEqual = ComparisonGreaterThanOrEqual
)
