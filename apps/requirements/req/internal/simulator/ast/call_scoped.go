package ast

import (
	"bytes"
	"fmt"
)

// ScopedCall is a function call with hierarchical scoping.
//
// Scoping levels (hierarchical - each level requires all levels below it):
//   - Domain!Subdomain!Class!FunctionName(record) - fully scoped (requires Subdomain and Class)
//   - Subdomain!Class!FunctionName(record)        - subdomain scope (requires Class)
//   - Class!FunctionName(record)                  - class scope
//   - FunctionName(record)                        - function only (class scope implied)
//   - _FunctionName(record)                       - model scope (ModelScope=true)
//
// Validation rules:
//   - If Domain is set, Subdomain and Class must also be set
//   - If Subdomain is set, Class must also be set
//   - All three can be nil (function only or model scope)
//
// The leading _ for model scope is not standard TLA+ syntax.
// FunctionName is always required.
type ScopedCall struct {
	ModelScope   bool        // If true, this is a model-level call (!FunctionName)
	Domain       *Identifier // Optional domain scope
	Subdomain    *Identifier // Optional subdomain scope
	Class        *Identifier // Optional class scope
	FunctionName *Identifier `validate:"required"` // Required function name
	Parameter    Expression  `validate:"required"` // Required record parameter (must be Record)
}

func (c *ScopedCall) expressionNode() {}

func (c *ScopedCall) String() (value string) {
	var out bytes.Buffer

	// Model scope uses leading _
	if c.ModelScope {
		out.WriteString("_")
		out.WriteString(c.FunctionName.String())
	} else {
		// Build the scoped function name
		if c.Domain != nil {
			out.WriteString(c.Domain.String())
			out.WriteString("!")
		}
		if c.Subdomain != nil {
			out.WriteString(c.Subdomain.String())
			out.WriteString("!")
		}
		if c.Class != nil {
			out.WriteString(c.Class.String())
			out.WriteString("!")
		}
		out.WriteString(c.FunctionName.String())
	}

	// Build the parameter
	out.WriteString("(")
	out.WriteString(c.Parameter.String())
	out.WriteString(")")

	return out.String()
}

func (c *ScopedCall) Ascii() (value string) {
	var out bytes.Buffer

	// Model scope uses leading _
	if c.ModelScope {
		out.WriteString("_")
		out.WriteString(c.FunctionName.Ascii())
	} else {
		// Build the scoped function name
		if c.Domain != nil {
			out.WriteString(c.Domain.Ascii())
			out.WriteString("!")
		}
		if c.Subdomain != nil {
			out.WriteString(c.Subdomain.Ascii())
			out.WriteString("!")
		}
		if c.Class != nil {
			out.WriteString(c.Class.Ascii())
			out.WriteString("!")
		}
		out.WriteString(c.FunctionName.Ascii())
	}

	// Build the parameter
	out.WriteString("(")
	out.WriteString(c.Parameter.Ascii())
	out.WriteString(")")

	return out.String()
}

func (c *ScopedCall) Validate() error {
	if err := _validate.Struct(c); err != nil {
		return err
	}

	// Validate scoping rules:
	// - ModelScope can only be true if Domain, Subdomain, and Class are all nil
	if c.ModelScope {
		if c.Domain != nil {
			return fmt.Errorf("Domain: must be nil when ModelScope is true")
		}
		if c.Subdomain != nil {
			return fmt.Errorf("Subdomain: must be nil when ModelScope is true")
		}
		if c.Class != nil {
			return fmt.Errorf("Class: must be nil when ModelScope is true")
		}
	}

	// Validate hierarchical scoping:
	// - If Domain is set, Subdomain and Class must also be set
	// - If Subdomain is set, Class must also be set
	if c.Domain != nil {
		if c.Subdomain == nil {
			return fmt.Errorf("Subdomain: required when Domain is set")
		}
		if c.Class == nil {
			return fmt.Errorf("Class: required when Domain is set")
		}
	}
	if c.Subdomain != nil && c.Class == nil {
		return fmt.Errorf("Class: required when Subdomain is set")
	}

	// Validate individual identifiers
	if c.Domain != nil {
		if err := c.Domain.Validate(); err != nil {
			return fmt.Errorf("Domain: %w", err)
		}
	}
	if c.Subdomain != nil {
		if err := c.Subdomain.Validate(); err != nil {
			return fmt.Errorf("Subdomain: %w", err)
		}
	}
	if c.Class != nil {
		if err := c.Class.Validate(); err != nil {
			return fmt.Errorf("Class: %w", err)
		}
	}
	if err := c.FunctionName.Validate(); err != nil {
		return fmt.Errorf("FunctionName: %w", err)
	}

	// Validate parameter
	if err := c.Parameter.Validate(); err != nil {
		return fmt.Errorf("Parameter: %w", err)
	}

	return nil
}

// CallExpression is an alias for backwards compatibility.
// Deprecated: Use ScopedCall instead.
type CallExpression = ScopedCall
