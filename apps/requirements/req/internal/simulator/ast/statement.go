package ast

import "bytes"

// Assignment represents a single state assignment.
// Pattern: identifier' = expression
type Assignment struct {
	Target *Identifier `validate:"required"`
	Value  Expression  `validate:"required"`
}

func (a *Assignment) statementNode() {}

func (a *Assignment) String() (value string) {
	var out bytes.Buffer
	out.WriteString(a.Target.String())
	out.WriteString("' = ")
	out.WriteString(a.Value.String())
	return out.String()
}

func (a *Assignment) Ascii() (value string) {
	var out bytes.Buffer
	out.WriteString(a.Target.Ascii())
	out.WriteString("' = ")
	out.WriteString(a.Value.Ascii())
	return out.String()
}

func (a *Assignment) Validate() error {
	if err := _validate.Struct(a); err != nil {
		return err
	}
	if err := a.Target.Validate(); err != nil {
		return err
	}
	if err := a.Value.Validate(); err != nil {
		return err
	}
	return nil
}
