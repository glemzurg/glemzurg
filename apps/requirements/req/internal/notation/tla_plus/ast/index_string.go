package ast

import (
	"bytes"
)

// StringIndex represents string indexing.
// Pattern: string[index]
type StringIndex struct {
	Str   Expression `validate:"required"` // Must be String
	Index Expression `validate:"required"` // Must be Natural
}

func (s *StringIndex) expressionNode() {}

func (s *StringIndex) String() (value string) {
	var out bytes.Buffer
	out.WriteString(s.Str.String())
	out.WriteString("[")
	out.WriteString(s.Index.String())
	out.WriteString("]")
	return out.String()
}

func (s *StringIndex) Ascii() (value string) {
	var out bytes.Buffer
	out.WriteString(s.Str.Ascii())
	out.WriteString("[")
	out.WriteString(s.Index.Ascii())
	out.WriteString("]")
	return out.String()
}

func (s *StringIndex) Validate() error {
	if err := _validate.Struct(s); err != nil {
		return err
	}
	if err := s.Str.Validate(); err != nil {
		return err
	}
	if err := s.Index.Validate(); err != nil {
		return err
	}
	return nil
}
