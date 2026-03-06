package ast

import (
	"bytes"
	"fmt"
)

// RecordTypeField represents a field in a record type declaration.
type RecordTypeField struct {
	Name *Identifier `validate:"required"` // The field name
	Type Expression  `validate:"required"` // The field type expression
}

// RecordTypeExpr represents a record type with typed field declarations.
// Pattern: [name: STRING, age: Int]
// This is distinct from RecordInstance which uses |-> for value bindings.
type RecordTypeExpr struct {
	Fields []*RecordTypeField `validate:"required,min=1"` // At least one field
}

func (r *RecordTypeExpr) expressionNode() {}

func (r *RecordTypeExpr) String() (value string) {
	var out bytes.Buffer
	out.WriteString("[")
	for i, field := range r.Fields {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(field.Name.String())
		out.WriteString(": ")
		out.WriteString(field.Type.String())
	}
	out.WriteString("]")
	return out.String()
}

func (r *RecordTypeExpr) Ascii() (value string) {
	var out bytes.Buffer
	out.WriteString("[")
	for i, field := range r.Fields {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(field.Name.Ascii())
		out.WriteString(": ")
		out.WriteString(field.Type.Ascii())
	}
	out.WriteString("]")
	return out.String()
}

func (r *RecordTypeExpr) Validate() error {
	if err := _validate.Struct(r); err != nil {
		return err
	}
	for i, field := range r.Fields {
		if field == nil {
			return fmt.Errorf("Fields[%d]: is nil", i)
		}
		if field.Name == nil {
			return fmt.Errorf("Fields[%d].Name: is required", i)
		}
		if err := field.Name.Validate(); err != nil {
			return fmt.Errorf("Fields[%d].Name: %w", i, err)
		}
		if field.Type == nil {
			return fmt.Errorf("Fields[%d].Type: is required", i)
		}
		if err := field.Type.Validate(); err != nil {
			return fmt.Errorf("Fields[%d].Type: %w", i, err)
		}
	}
	return nil
}
