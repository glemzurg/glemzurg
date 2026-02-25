package object

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type StringSuite struct {
	suite.Suite
}

func TestStringSuite(t *testing.T) {
	suite.Run(t, new(StringSuite))
}

func (s *StringSuite) TestNewString() {
	str := NewString("hello")
	s.Equal("hello", str.Value())
	s.Equal(TypeString, str.Type())
}

func (s *StringSuite) TestInspect() {
	s.Equal(`"hello"`, NewString("hello").Inspect())
	s.Equal(`""`, NewString("").Inspect())
	s.Equal(`"with spaces"`, NewString("with spaces").Inspect())
}

func (s *StringSuite) TestSetValue() {
	str := NewString("")
	source := NewString("hello")

	err := str.SetValue(source)
	s.NoError(err)
	s.Equal("hello", str.Value())
}

func (s *StringSuite) TestSetValueIncompatible() {
	str := NewString("")
	source := NewInteger(42)

	err := str.SetValue(source)
	s.Error(err)
}

func (s *StringSuite) TestClone() {
	original := NewString("hello")
	clone := original.Clone().(*String)

	s.Equal(original.Value(), clone.Value())
	s.Equal(original.Type(), clone.Type())

	clone.value = "world"
	s.Equal("hello", original.Value())
}

