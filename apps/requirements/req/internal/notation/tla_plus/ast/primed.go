package ast

import (
	"bytes"
	"fmt"
)

// Primed represents a primed expression (x') in TLA+.
// Primed variables refer to the "next state" value of a variable.
// The base expression must evaluate to a variable name.
type Primed struct {
	Base Expression `validate:"required"` // The expression being primed (typically an identifier)
}

func (p *Primed) expressionNode() {}

func (p *Primed) String() (value string) {
	var out bytes.Buffer
	out.WriteString(p.Base.String())
	out.WriteString("'")
	return out.String()
}

func (p *Primed) Ascii() (value string) { return p.String() }

func (p *Primed) Validate() error {
	if err := _validate.Struct(p); err != nil {
		return err
	}
	if validator, ok := p.Base.(interface{ Validate() error }); ok {
		if err := validator.Validate(); err != nil {
			return fmt.Errorf("Base: %w", err)
		}
	}
	return nil
}
