package ast

import (
	"bytes"
	"fmt"
)

// BuiltinCall represents a call to a builtin function.
// Pattern: _Module!Function(args...)
// Examples: _Seq!Head(tuple), _Bags!CopiesIn(elem, bag)
type BuiltinCall struct {
	Name string       `validate:"required"` // e.g., "_Seq!Head"
	Args []Expression `validate:"required"` // Arguments (can be empty slice)
}

func (b *BuiltinCall) expressionNode() {}

func (b *BuiltinCall) String() string {
	var out bytes.Buffer
	out.WriteString(b.Name)
	out.WriteString("(")
	for i, arg := range b.Args {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(arg.String())
	}
	out.WriteString(")")
	return out.String()
}

func (b *BuiltinCall) Ascii() string {
	var out bytes.Buffer
	out.WriteString(b.Name)
	out.WriteString("(")
	for i, arg := range b.Args {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(arg.Ascii())
	}
	out.WriteString(")")
	return out.String()
}

func (b *BuiltinCall) Validate() error {
	if err := _validate.Struct(b); err != nil {
		return err
	}
	for i, arg := range b.Args {
		if arg == nil {
			return fmt.Errorf("Args[%d]: is nil", i)
		}
		if err := arg.Validate(); err != nil {
			return fmt.Errorf("Args[%d]: %w", i, err)
		}
	}
	return nil
}
