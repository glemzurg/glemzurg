package ast

import "fmt"

// Parenthesized represents a parenthesized expression in TLA+.
// This node is necessary to preserve parentheses for exact round-trip reconstruction.
// For example, "(1)" should reconstruct back to "(1)", not just "1".
type Parenthesized struct {
	// Inner is the expression inside the parentheses.
	Inner Expression
}

func (p *Parenthesized) expressionNode() {}

// String returns the TLA+ representation with parentheses.
func (p *Parenthesized) String() string {
	if p.Inner == nil {
		return "()"
	}
	return fmt.Sprintf("(%s)", p.Inner.String())
}

// Ascii returns the ASCII representation (same as String).
func (p *Parenthesized) Ascii() string {
	return p.String()
}

// Validate checks that the Parenthesized is well-formed.
func (p *Parenthesized) Validate() error {
	if p.Inner == nil {
		return fmt.Errorf("Parenthesized: inner expression cannot be nil")
	}
	return nil
}

// NewParenthesized creates a new parenthesized expression.
func NewParenthesized(inner Expression) *Parenthesized {
	return &Parenthesized{Inner: inner}
}

// ParenExpr is an alias for backwards compatibility.
// Deprecated: Use Parenthesized instead.
type ParenExpr = Parenthesized

// NewParenExpr is an alias for backwards compatibility.
// Deprecated: Use NewParenthesized instead.
var NewParenExpr = NewParenthesized
