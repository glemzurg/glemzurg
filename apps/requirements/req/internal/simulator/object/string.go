package object

import (
	"fmt"
)

// String wraps a string value.
type String struct {
	value string
}

// NewString creates a new String.
func NewString(value string) *String {
	return &String{value: value}
}

func (s *String) Type() ObjectType { return TypeString }
func (s *String) Inspect() string  { return fmt.Sprintf("%q", s.value) }
func (s *String) Value() string    { return s.value }

func (s *String) SetValue(source Object) error {
	src, ok := source.(*String)
	if !ok {
		return fmt.Errorf("cannot assign %T to String", source)
	}
	s.value = src.value
	return nil
}

func (s *String) Clone() Object {
	return &String{value: s.value}
}
