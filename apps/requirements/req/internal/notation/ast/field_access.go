package ast

import (
	"bytes"
	"fmt"
)

// FieldAccess is a field access expression (base.member).
// The base can be any expression (identifier, existing value, or another field access).
// When Base is nil and Identifier is nil, it uses "!" to indicate existing value in EXCEPT context.
//
// For backwards compatibility:
// - If Identifier is set (non-nil), it is used as the base
// - If Base is set (non-nil), it is used as the base
// - If both are nil, "!" (existing value) is the base
type FieldAccess struct {
	// Base is the expression being accessed (can be any expression for chaining).
	// Takes precedence over Identifier if both are set.
	Base Expression

	// Identifier is for backwards compatibility with the old structure.
	// Deprecated: Use Base instead.
	Identifier *Identifier

	Member string `validate:"required"` // The member name
}

func (f *FieldAccess) expressionNode() {}

func (f *FieldAccess) String() (value string) {
	var out bytes.Buffer
	if f.Base != nil {
		out.WriteString(f.Base.String())
	} else if f.Identifier != nil {
		out.WriteString(f.Identifier.String())
	} else {
		out.WriteString("!")
	}
	out.WriteString(".")
	out.WriteString(f.Member)
	return out.String()
}

func (f *FieldAccess) Ascii() (value string) { return f.String() }

func (f *FieldAccess) Validate() error {
	if err := _validate.Struct(f); err != nil {
		return err
	}
	// Validate Base if present
	if f.Base != nil {
		if validator, ok := f.Base.(interface{ Validate() error }); ok {
			if err := validator.Validate(); err != nil {
				return fmt.Errorf("Base: %w", err)
			}
		}
	} else if f.Identifier != nil {
		// Backwards compatibility: validate Identifier
		if err := f.Identifier.Validate(); err != nil {
			return fmt.Errorf("Identifier: %w", err)
		}
	}
	return nil
}

// GetBase returns the effective base expression for this field access.
// It handles backwards compatibility with the old Identifier field.
func (f *FieldAccess) GetBase() Expression {
	if f.Base != nil {
		return f.Base
	}
	if f.Identifier != nil {
		return f.Identifier
	}
	return nil // nil means use existing value (@)
}

// FieldIdentifier is an alias for backwards compatibility.
// Deprecated: Use FieldAccess instead.
type FieldIdentifier = FieldAccess
