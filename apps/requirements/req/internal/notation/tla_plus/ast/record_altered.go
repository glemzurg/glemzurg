package ast

import (
	"bytes"
	"fmt"
)

// FieldAlteration represents a field being altered with a new expression value.
type FieldAlteration struct {
	Field      *FieldIdentifier `validate:"required"` // The field being altered (e.g., !.val)
	Expression Expression       `validate:"required"` // The new value expression
}

// RecordAltered represents an EXCEPT expression that alters fields of a record.
// Pattern: [base EXCEPT !.field1 = expr1, !.field2 = expr2, ...]
// The base can be any expression, enabling chaining: [[r EXCEPT !.x = 1] EXCEPT !.y = 2].
type RecordAltered struct {
	Base        Expression         `validate:"required"`       // The base expression (typically an identifier or another RecordAltered)
	Alterations []*FieldAlteration `validate:"required,min=1"` // At least one field alteration
}

func (r *RecordAltered) expressionNode() {}

func (r *RecordAltered) String() (value string) {
	var out bytes.Buffer
	out.WriteString("[")
	out.WriteString(r.Base.String())
	out.WriteString(" EXCEPT ")
	for i, alt := range r.Alterations {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(alt.Field.String())
		out.WriteString(" = ")
		out.WriteString(alt.Expression.String())
	}
	out.WriteString("]")
	return out.String()
}

func (r *RecordAltered) ASCII() (value string) {
	var out bytes.Buffer
	out.WriteString("[")
	out.WriteString(r.Base.ASCII())
	out.WriteString(" EXCEPT ")
	for i, alt := range r.Alterations {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(alt.Field.ASCII())
		out.WriteString(" = ")
		out.WriteString(alt.Expression.ASCII())
	}
	out.WriteString("]")
	return out.String()
}

func (r *RecordAltered) Validate() error {
	if err := _validate.Struct(r); err != nil {
		return err
	}
	if err := r.Base.Validate(); err != nil {
		return fmt.Errorf("base: %w", err)
	}
	for i, alt := range r.Alterations {
		if alt == nil {
			return fmt.Errorf("alterations[%d]: is nil", i)
		}
		if alt.Field == nil {
			return fmt.Errorf("Alterations[%d].Field: is required", i)
		}
		if err := alt.Field.Validate(); err != nil {
			return fmt.Errorf("Alterations[%d].Field: %w", i, err)
		}
		if alt.Expression == nil {
			return fmt.Errorf("Alterations[%d].Expression: is required", i)
		}
		if err := alt.Expression.Validate(); err != nil {
			return fmt.Errorf("Alterations[%d].Expression: %w", i, err)
		}
	}
	return nil
}
