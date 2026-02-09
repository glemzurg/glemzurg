package ast

import (
	"bytes"
	"fmt"
)

// Fraction represents a fraction expression in TLA+ using the `/` operator.
// This is NOT division - it creates a fractional real number.
// For example: 3/4 creates the fraction three-fourths.
//
// The `/` operator has higher precedence than `-` (negation), so:
//   - `-3/4` parses as `-(3/4)`
//   - `3/-4` parses as `3/(-4)`
type Fraction struct {
	Numerator   Expression `validate:"required"` // The numerator (top)
	Denominator Expression `validate:"required"` // The denominator (bottom)
}

func (f *Fraction) expressionNode() {}

func (f *Fraction) String() string {
	var out bytes.Buffer
	out.WriteString(f.Numerator.String())
	out.WriteString("/")
	out.WriteString(f.Denominator.String())
	return out.String()
}

func (f *Fraction) Ascii() string {
	return f.String()
}

func (f *Fraction) Validate() error {
	if err := _validate.Struct(f); err != nil {
		return err
	}
	if err := f.Numerator.Validate(); err != nil {
		return fmt.Errorf("Numerator: %w", err)
	}
	if err := f.Denominator.Validate(); err != nil {
		return fmt.Errorf("Denominator: %w", err)
	}
	return nil
}

// NewFraction creates a new fraction expression.
func NewFraction(numerator, denominator Expression) *Fraction {
	return &Fraction{
		Numerator:   numerator,
		Denominator: denominator,
	}
}

// FractionExpr is an alias for backwards compatibility.
// Deprecated: Use Fraction instead.
type FractionExpr = Fraction

// NewFractionExpr is an alias for backwards compatibility.
// Deprecated: Use NewFraction instead.
var NewFractionExpr = NewFraction
