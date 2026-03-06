package ast

import (
	"bytes"
)

// TupleIndex represents tuple/array indexing.
// Pattern: tuple[index]
type TupleIndex struct {
	Tuple Expression `validate:"required"` // Must be Tuple
	Index Expression `validate:"required"` // Must be Natural
}

func (e *TupleIndex) expressionNode() {}

func (e *TupleIndex) String() (value string) {
	var out bytes.Buffer
	out.WriteString(e.Tuple.String())
	out.WriteString("[")
	out.WriteString(e.Index.String())
	out.WriteString("]")
	return out.String()
}

func (e *TupleIndex) Ascii() (value string) {
	var out bytes.Buffer
	out.WriteString(e.Tuple.Ascii())
	out.WriteString("[")
	out.WriteString(e.Index.Ascii())
	out.WriteString("]")
	return out.String()
}

func (e *TupleIndex) Validate() error {
	if err := _validate.Struct(e); err != nil {
		return err
	}
	if err := e.Tuple.Validate(); err != nil {
		return err
	}
	if err := e.Index.Validate(); err != nil {
		return err
	}
	return nil
}

// ExpressionTupleIndex is an alias for backwards compatibility.
// Deprecated: Use TupleIndex instead.
type ExpressionTupleIndex = TupleIndex
