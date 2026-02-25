package ast

import (
	"bytes"
	"fmt"
)

// FieldBinding represents a field being bound to an expression value.
type FieldBinding struct {
	Field      *Identifier `validate:"required"` // The field name
	Expression Expression  `validate:"required"` // The value expression
}

// RecordInstance represents a record literal with field bindings.
// Pattern: [a ↦ 1, b ↦ 2, c ↦ 3]
type RecordInstance struct {
	Bindings []*FieldBinding `validate:"required,min=1"` // At least one field binding
}

func (r *RecordInstance) expressionNode() {}

func (r *RecordInstance) String() (value string) {
	var out bytes.Buffer
	out.WriteString("[")
	for i, binding := range r.Bindings {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(binding.Field.String())
		out.WriteString(" ↦ ")
		out.WriteString(binding.Expression.String())
	}
	out.WriteString("]")
	return out.String()
}

func (r *RecordInstance) Ascii() (value string) {
	var out bytes.Buffer
	out.WriteString("[")
	for i, binding := range r.Bindings {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(binding.Field.Ascii())
		out.WriteString(" |-> ")
		out.WriteString(binding.Expression.Ascii())
	}
	out.WriteString("]")
	return out.String()
}

func (r *RecordInstance) Validate() error {
	if err := _validate.Struct(r); err != nil {
		return err
	}
	for i, binding := range r.Bindings {
		if binding == nil {
			return fmt.Errorf("Bindings[%d]: is nil", i)
		}
		if binding.Field == nil {
			return fmt.Errorf("Bindings[%d].Field: is required", i)
		}
		if err := binding.Field.Validate(); err != nil {
			return fmt.Errorf("Bindings[%d].Field: %w", i, err)
		}
		if binding.Expression == nil {
			return fmt.Errorf("Bindings[%d].Expression: is required", i)
		}
		if err := binding.Expression.Validate(); err != nil {
			return fmt.Errorf("Bindings[%d].Expression: %w", i, err)
		}
	}
	return nil
}
