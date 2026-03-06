package ast

import (
	"bytes"
	"fmt"
)

// CartesianProduct represents a Cartesian product of two or more sets.
// Pattern: S1 \X S2 or S1 × S2
// In TLA+, the Cartesian product of sets produces the set of all tuples.
// As a type expression, it represents a tuple type: Int \X STRING = TupleType{IntegerType, StringType}.
type CartesianProduct struct {
	Operands []Expression `validate:"required,min=2"` // At least two operands
}

func (cp *CartesianProduct) expressionNode() {}

func (cp *CartesianProduct) String() (value string) {
	var out bytes.Buffer
	for i, op := range cp.Operands {
		if i > 0 {
			out.WriteString(" × ")
		}
		out.WriteString(op.String())
	}
	return out.String()
}

func (cp *CartesianProduct) Ascii() (value string) {
	var out bytes.Buffer
	for i, op := range cp.Operands {
		if i > 0 {
			out.WriteString(" \\X ")
		}
		out.WriteString(op.Ascii())
	}
	return out.String()
}

func (cp *CartesianProduct) Validate() error {
	if err := _validate.Struct(cp); err != nil {
		return err
	}
	for i, op := range cp.Operands {
		if op == nil {
			return fmt.Errorf("Operands[%d]: is nil", i)
		}
		if err := op.Validate(); err != nil {
			return fmt.Errorf("Operands[%d]: %w", i, err)
		}
	}
	return nil
}
