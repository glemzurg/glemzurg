package ast

import (
	"fmt"
	"strings"
)

// NumberBase represents the base/radix of a number literal.
type NumberBase int

const (
	BaseDecimal NumberBase = 10 // Default decimal (0-9)
	BaseBinary  NumberBase = 2  // Binary with \b or \B prefix
	BaseOctal   NumberBase = 8  // Octal with \o or \O prefix
	BaseHex     NumberBase = 16 // Hexadecimal with \h or \H prefix
)

// NumberLiteral represents a numeric literal in TLA+.
// It can be:
//   - An integer: "42", "007", "\b1010", "\o17", "\hFF"
//   - A decimal: "3.14", ".5", "0.123"
//
// The `-` negation and `/` fraction operators are NOT part of this node;
// they are separate operator nodes in the AST.
//
// Leading zeros are preserved for exact round-trip reconstruction.
type NumberLiteral struct {
	// Base is the number base (decimal, binary, octal, hex)
	Base NumberBase

	// BasePrefix stores the original prefix for non-decimal bases
	// e.g., "\\b", "\\B", "\\o", "\\O", "\\h", "\\H"
	// Empty for decimal numbers.
	BasePrefix string

	// IntegerPart is the digits before the decimal point (if any).
	// For ".5", this is empty. For "3.14", this is "3". For "007", this is "007".
	// Preserves leading zeros.
	IntegerPart string

	// HasDecimalPoint indicates whether there's a decimal point.
	// True for "3.14" and ".5", false for "42".
	HasDecimalPoint bool

	// FractionalPart is the digits after the decimal point.
	// For "3.14", this is "14". For "42", this is empty.
	// Preserves trailing zeros (e.g., "3.140" has FractionalPart "140").
	FractionalPart string
}

func (n *NumberLiteral) expressionNode() {}

// String returns the TLA+ representation of the number literal.
// This enables exact round-trip: parse -> AST -> String() == original input.
func (n *NumberLiteral) String() string {
	var sb strings.Builder

	// Write base prefix if not decimal
	if n.BasePrefix != "" {
		sb.WriteString(n.BasePrefix)
	}

	// Write integer part (may be empty for ".5")
	sb.WriteString(n.IntegerPart)

	// Write decimal point and fractional part if present
	if n.HasDecimalPoint {
		sb.WriteString(".")
		sb.WriteString(n.FractionalPart)
	}

	return sb.String()
}

// Ascii returns the ASCII representation (same as String for numbers).
func (n *NumberLiteral) Ascii() string {
	return n.String()
}

// Validate checks that the NumberLiteral is well-formed.
func (n *NumberLiteral) Validate() error {
	// Must have at least integer part or fractional part
	if n.IntegerPart == "" && n.FractionalPart == "" {
		return fmt.Errorf("NumberLiteral: must have integer part or fractional part")
	}

	// If has decimal point, must have fractional part
	if n.HasDecimalPoint && n.FractionalPart == "" {
		return fmt.Errorf("NumberLiteral: decimal point requires fractional part")
	}

	// Non-decimal bases cannot have decimal points
	if n.Base != BaseDecimal && n.HasDecimalPoint {
		return fmt.Errorf("NumberLiteral: non-decimal bases cannot have decimal points")
	}

	// Validate digits match the base
	if err := n.validateDigits(); err != nil {
		return err
	}

	return nil
}

// validateDigits checks that all digits are valid for the number's base.
func (n *NumberLiteral) validateDigits() error {
	allDigits := n.IntegerPart + n.FractionalPart

	for _, ch := range allDigits {
		valid := false
		switch n.Base {
		case BaseDecimal:
			valid = ch >= '0' && ch <= '9'
		case BaseBinary:
			valid = ch == '0' || ch == '1'
		case BaseOctal:
			valid = ch >= '0' && ch <= '7'
		case BaseHex:
			valid = (ch >= '0' && ch <= '9') ||
				(ch >= 'a' && ch <= 'f') ||
				(ch >= 'A' && ch <= 'F')
		}
		if !valid {
			return fmt.Errorf("NumberLiteral: invalid digit '%c' for base %d", ch, n.Base)
		}
	}

	return nil
}

// IsInteger returns true if this is an integer (no decimal point).
func (n *NumberLiteral) IsInteger() bool {
	return !n.HasDecimalPoint
}

// IsDecimal returns true if this has a decimal point.
func (n *NumberLiteral) IsDecimal() bool {
	return n.HasDecimalPoint
}

// NewNumberLiteral creates a decimal integer NumberLiteral.
func NewNumberLiteral(integerPart string) *NumberLiteral {
	return &NumberLiteral{
		Base:        BaseDecimal,
		IntegerPart: integerPart,
	}
}

// NewDecimalNumberLiteral creates a decimal number with fractional part.
func NewDecimalNumberLiteral(integerPart, fractionalPart string) *NumberLiteral {
	return &NumberLiteral{
		Base:            BaseDecimal,
		IntegerPart:     integerPart,
		HasDecimalPoint: true,
		FractionalPart:  fractionalPart,
	}
}

// NewBinaryNumberLiteral creates a binary number literal.
func NewBinaryNumberLiteral(prefix, digits string) *NumberLiteral {
	return &NumberLiteral{
		Base:        BaseBinary,
		BasePrefix:  prefix,
		IntegerPart: digits,
	}
}

// NewOctalNumberLiteral creates an octal number literal.
func NewOctalNumberLiteral(prefix, digits string) *NumberLiteral {
	return &NumberLiteral{
		Base:        BaseOctal,
		BasePrefix:  prefix,
		IntegerPart: digits,
	}
}

// NewHexNumberLiteral creates a hexadecimal number literal.
func NewHexNumberLiteral(prefix, digits string) *NumberLiteral {
	return &NumberLiteral{
		Base:        BaseHex,
		BasePrefix:  prefix,
		IntegerPart: digits,
	}
}

// NewIntLiteral creates a NumberLiteral from an integer value.
// For negative values, creates a NumericPrefixExpression with negation.
// This is a convenience function for tests and programmatic AST construction.
func NewIntLiteral(value int) Expression {
	if value < 0 {
		return NewNegation(&NumberLiteral{
			Base:        BaseDecimal,
			IntegerPart: fmt.Sprintf("%d", -value),
		})
	}
	return &NumberLiteral{
		Base:        BaseDecimal,
		IntegerPart: fmt.Sprintf("%d", value),
	}
}
