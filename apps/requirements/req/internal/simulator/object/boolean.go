package object

import (
	"fmt"
)

// Boolean wraps a bool value.
type Boolean struct {
	value bool
}

// NewBoolean creates a new Boolean.
func NewBoolean(value bool) *Boolean {
	return &Boolean{value: value}
}

func (b *Boolean) Type() ObjectType { return TypeBoolean }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.value) }
func (b *Boolean) Value() bool      { return b.value }

func (b *Boolean) SetValue(source Object) error {
	src, ok := source.(*Boolean)
	if !ok {
		return fmt.Errorf("cannot assign %T to Boolean", source)
	}
	b.value = src.value
	return nil
}

func (b *Boolean) Clone() Object {
	return &Boolean{value: b.value}
}
