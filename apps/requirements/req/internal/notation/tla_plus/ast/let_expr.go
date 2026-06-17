package ast

import (
	"bytes"
	"fmt"
)

// LetExpr is a local binding: LET name == value IN body.
type LetExpr struct {
	Variable string     `validate:"required"`
	Value    Expression `validate:"required"`
	Body     Expression `validate:"required"`
}

func (e *LetExpr) expressionNode() {}

func (e *LetExpr) String() string {
	var out bytes.Buffer
	out.WriteString("LET ")
	out.WriteString(e.Variable)
	out.WriteString(" == ")
	out.WriteString(e.Value.String())
	out.WriteString(" IN ")
	out.WriteString(e.Body.String())
	return out.String()
}

func (e *LetExpr) ASCII() string {
	var out bytes.Buffer
	out.WriteString("LET ")
	out.WriteString(e.Variable)
	out.WriteString(" == ")
	out.WriteString(e.Value.ASCII())
	out.WriteString(" IN ")
	out.WriteString(e.Body.ASCII())
	return out.String()
}

func (e *LetExpr) Validate() error {
	if err := _validate.Struct(e); err != nil {
		return err
	}
	if err := e.Value.Validate(); err != nil {
		return fmt.Errorf("value: %w", err)
	}
	if err := e.Body.Validate(); err != nil {
		return fmt.Errorf("body: %w", err)
	}
	return nil
}
