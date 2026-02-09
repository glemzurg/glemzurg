package object

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type BooleanSuite struct {
	suite.Suite
}

func TestBooleanSuite(t *testing.T) {
	suite.Run(t, new(BooleanSuite))
}

func (s *BooleanSuite) TestNewBoolean() {
	b := NewBoolean(true)
	s.True(b.Value())
	s.Equal(TypeBoolean, b.Type())

	b = NewBoolean(false)
	s.False(b.Value())
}

func (s *BooleanSuite) TestInspect() {
	s.Equal("true", NewBoolean(true).Inspect())
	s.Equal("false", NewBoolean(false).Inspect())
}

func (s *BooleanSuite) TestSetValue() {
	b := NewBoolean(false)
	source := NewBoolean(true)

	err := b.SetValue(source)
	s.NoError(err)
	s.True(b.Value())
}

func (s *BooleanSuite) TestSetValueIncompatible() {
	b := NewBoolean(false)
	source := NewInteger(42)

	err := b.SetValue(source)
	s.Error(err)
}

func (s *BooleanSuite) TestClone() {
	original := NewBoolean(true)
	clone := original.Clone().(*Boolean)

	s.Equal(original.Value(), clone.Value())
	s.Equal(original.Type(), clone.Type())

	clone.value = false
	s.True(original.Value())
}

