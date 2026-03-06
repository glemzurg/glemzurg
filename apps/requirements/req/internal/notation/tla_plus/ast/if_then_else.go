package ast

import (
	"bytes"
)

// IfThenElse represents an IF-THEN-ELSE expression.
// Pattern: IF logic THEN expression ELSE expression
type IfThenElse struct {
	Condition Expression `validate:"required"` // Must be Boolean
	Then      Expression `validate:"required"`
	Else      Expression `validate:"required"`
}

func (e *IfThenElse) expressionNode() {}

func (e *IfThenElse) String() (value string) {
	var out bytes.Buffer
	out.WriteString("IF ")
	out.WriteString(e.Condition.String())
	out.WriteString(" THEN ")
	out.WriteString(e.Then.String())
	out.WriteString(" ELSE ")
	out.WriteString(e.Else.String())
	return out.String()
}

func (e *IfThenElse) Ascii() (value string) {
	var out bytes.Buffer
	out.WriteString("IF ")
	out.WriteString(e.Condition.Ascii())
	out.WriteString(" THEN ")
	out.WriteString(e.Then.Ascii())
	out.WriteString(" ELSE ")
	out.WriteString(e.Else.Ascii())
	return out.String()
}

func (e *IfThenElse) Validate() error {
	if err := _validate.Struct(e); err != nil {
		return err
	}
	if err := e.Condition.Validate(); err != nil {
		return err
	}
	if err := e.Then.Validate(); err != nil {
		return err
	}
	if err := e.Else.Validate(); err != nil {
		return err
	}
	return nil
}

// ExpressionIfElse is an alias for backwards compatibility.
// Deprecated: Use IfThenElse instead.
type ExpressionIfElse = IfThenElse
