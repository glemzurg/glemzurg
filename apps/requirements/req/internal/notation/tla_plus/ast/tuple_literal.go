package ast

import (
	"bytes"
	"fmt"
)

// TupleLiteral represents a tuple with zero or more elements.
// Pattern: <<3, 7, 3>> or <<>> for empty tuple
type TupleLiteral struct {
	Elements []Expression // Can be empty
}

func (t *TupleLiteral) expressionNode() {}

func (t *TupleLiteral) String() (value string) {
	var out bytes.Buffer
	out.WriteString("⟨")
	for i, el := range t.Elements {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(el.String())
	}
	out.WriteString("⟩")
	return out.String()
}

func (t *TupleLiteral) Ascii() (value string) {
	var out bytes.Buffer
	out.WriteString("<<")
	for i, el := range t.Elements {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(el.Ascii())
	}
	out.WriteString(">>")
	return out.String()
}

func (t *TupleLiteral) Validate() error {
	if err := _validate.Struct(t); err != nil {
		return err
	}
	for i, el := range t.Elements {
		if el == nil {
			return fmt.Errorf("Elements[%d]: is nil", i)
		}
		if err := el.Validate(); err != nil {
			return fmt.Errorf("Elements[%d]: %w", i, err)
		}
	}
	return nil
}
