package object

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type NumberSuite struct {
	suite.Suite
}

func TestNumberSuite(t *testing.T) {
	suite.Run(t, new(NumberSuite))
}

func (s *NumberSuite) TestNewNatural() {
	n := NewNatural(42)
	s.Equal("42", n.Inspect())
	s.Equal(TypeNumber, n.Type())
	s.Equal(KindNatural, n.Kind())
}

func (s *NumberSuite) TestNewNaturalNegativePanics() {
	s.Panics(func() {
		NewNatural(-5)
	})
}

func (s *NumberSuite) TestNewInteger() {
	// Positive integer becomes Natural
	n := NewInteger(42)
	s.Equal("42", n.Inspect())
	s.Equal(KindNatural, n.Kind())

	// Negative integer stays Integer
	n = NewInteger(-42)
	s.Equal("-42", n.Inspect())
	s.Equal(KindInteger, n.Kind())
}

func (s *NumberSuite) TestNewReal() {
	// Fractional value stays Real (1/2 = 0.5)
	n := NewReal(1, 2)
	s.Equal("1/2", n.Inspect())
	s.Equal(KindRational, n.Kind())

	// Whole number becomes Natural (10/2 = 5)
	n = NewReal(10, 2)
	s.Equal("5", n.Inspect())
	s.Equal(KindNatural, n.Kind())

	// Negative whole number becomes Integer (-10/1 = -10)
	n = NewReal(-10, 1)
	s.Equal("-10", n.Inspect())
	s.Equal(KindInteger, n.Kind())
}

func (s *NumberSuite) TestNewFloat() {
	// NewFloat creates a Real (potentially irrational) number
	n := NewFloat(3.14)
	s.Equal(KindReal, n.Kind())
	s.Equal("3.14", n.Inspect())

	// Even whole numbers from float are Real (contaminated)
	n = NewFloat(42.0)
	s.Equal("42", n.Inspect())
	s.Equal(KindReal, n.Kind())

	// Negative floats are also Real
	n = NewFloat(-10.0)
	s.Equal("-10", n.Inspect())
	s.Equal(KindReal, n.Kind())
}

func (s *NumberSuite) TestInspect() {
	tests := []struct {
		name     string
		num      *Number
		expected string
	}{
		{"natural", NewNatural(42), "42"},
		{"zero", NewInteger(0), "0"},
		{"negative", NewInteger(-100), "-100"},
		{"real", NewReal(314, 100), "157/50"},    // RatString format (reduced)
		{"negative real", NewReal(-5, 2), "-5/2"}, // RatString format
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.expected, tt.num.Inspect())
		})
	}
}

func (s *NumberSuite) TestSetValue() {
	n := NewInteger(0)
	source := NewInteger(42)

	err := n.SetValue(source)
	s.NoError(err)
	s.Equal("42", n.Inspect())
}

func (s *NumberSuite) TestSetValueIncompatible() {
	n := NewInteger(0)
	source := NewBoolean(true)

	err := n.SetValue(source)
	s.Error(err)
}

func (s *NumberSuite) TestClone() {
	original := NewInteger(42)
	clone := original.Clone().(*Number)

	s.Equal(original.Inspect(), clone.Inspect())
	s.Equal(original.Type(), clone.Type())

	// Modify clone via SetValue, original unchanged
	clone.SetValue(NewInteger(100))
	s.Equal("42", original.Inspect())
	s.Equal("100", clone.Inspect())
}

func (s *NumberSuite) TestCloneReal() {
	original := NewReal(314, 100) // 3.14
	clone := original.Clone().(*Number)

	s.Equal(original.Inspect(), clone.Inspect())
	s.Equal(KindRational, clone.Kind())

	// Modify clone via SetValue, original unchanged
	clone.SetValue(NewReal(271, 100)) // 2.71
	s.Equal("157/50", original.Inspect())
	s.Equal("271/100", clone.Inspect())
}

func (s *NumberSuite) TestAdd() {
	// Natural + Natural = Natural
	n1 := NewNatural(5)
	n2 := NewNatural(3)
	result := n1.Add(n2)
	s.Equal("8", result.Inspect())
	s.Equal(KindNatural, result.Kind())

	// Integer + Integer = Integer (may become Natural)
	n1 = NewInteger(-5)
	n2 = NewInteger(10)
	result = n1.Add(n2)
	s.Equal("5", result.Inspect())
	s.Equal(KindNatural, result.Kind())

	// Integer + Integer = Integer (negative)
	n1 = NewInteger(-5)
	n2 = NewInteger(-3)
	result = n1.Add(n2)
	s.Equal("-8", result.Inspect())
	s.Equal(KindInteger, result.Kind())

	// Real + Integer = may stay Real
	n1 = NewReal(5, 2) // 2.5
	n2 = NewInteger(2)
	result = n1.Add(n2)
	s.Equal("9/2", result.Inspect())
	s.Equal(KindRational, result.Kind())
}

func (s *NumberSuite) TestSub() {
	// 5 - 3 = 2 (Natural)
	n1 := NewNatural(5)
	n2 := NewNatural(3)
	result := n1.Sub(n2)
	s.Equal("2", result.Inspect())
	s.Equal(KindNatural, result.Kind())

	// 3 - 5 = -2 (Integer)
	n1 = NewNatural(3)
	n2 = NewNatural(5)
	result = n1.Sub(n2)
	s.Equal("-2", result.Inspect())
	s.Equal(KindInteger, result.Kind())
}

func (s *NumberSuite) TestMul() {
	// 5 * 3 = 15
	n1 := NewNatural(5)
	n2 := NewNatural(3)
	result := n1.Mul(n2)
	s.Equal("15", result.Inspect())
	s.Equal(KindNatural, result.Kind())

	// -5 * 3 = -15
	n1 = NewInteger(-5)
	n2 = NewInteger(3)
	result = n1.Mul(n2)
	s.Equal("-15", result.Inspect())
	s.Equal(KindInteger, result.Kind())
}

func (s *NumberSuite) TestDiv() {
	// 10 / 4 = 2.5 (Real)
	n1 := NewNatural(10)
	n2 := NewNatural(4)
	result := n1.Div(n2)
	s.Equal("5/2", result.Inspect())
	s.Equal(KindRational, result.Kind())

	// 10 / 2 = 5 (normalizes to Natural)
	n1 = NewNatural(10)
	n2 = NewNatural(2)
	result = n1.Div(n2)
	s.Equal("5", result.Inspect())
	s.Equal(KindNatural, result.Kind())
}

func (s *NumberSuite) TestIntDiv() {
	// 10 \div 3 = 3
	n1 := NewNatural(10)
	n2 := NewNatural(3)
	result, err := n1.IntDiv(n2)
	s.NoError(err)
	s.Equal("3", result.Inspect())

	// Division by zero
	n2 = NewNatural(0)
	_, err = n1.IntDiv(n2)
	s.Error(err)

	// Real operand not allowed
	n1 = NewReal(21, 2) // 10.5
	n2 = NewNatural(3)
	_, err = n1.IntDiv(n2)
	s.Error(err)
}

func (s *NumberSuite) TestMod() {
	// 10 % 3 = 1
	n1 := NewNatural(10)
	n2 := NewNatural(3)
	result, err := n1.Mod(n2)
	s.NoError(err)
	s.Equal("1", result.Inspect())

	// Division by zero
	n2 = NewNatural(0)
	_, err = n1.Mod(n2)
	s.Error(err)
}

func (s *NumberSuite) TestNeg() {
	// -5 = -5
	n := NewNatural(5)
	result := n.Neg()
	s.Equal("-5", result.Inspect())
	s.Equal(KindInteger, result.Kind())

	// -(-5) = 5
	n = NewInteger(-5)
	result = n.Neg()
	s.Equal("5", result.Inspect())
	s.Equal(KindNatural, result.Kind())

	// -3.14 = -3.14
	n = NewReal(314, 100)
	result = n.Neg()
	s.Equal("-157/50", result.Inspect())
	s.Equal(KindRational, result.Kind())
}

func (s *NumberSuite) TestCmp() {
	n1 := NewInteger(5)
	n2 := NewInteger(10)
	n3 := NewInteger(5)

	s.Equal(-1, n1.Cmp(n2))
	s.Equal(1, n2.Cmp(n1))
	s.Equal(0, n1.Cmp(n3))

	// Compare integer with real
	n4 := NewReal(5, 1) // 5.0
	s.Equal(0, n1.Cmp(n4))
}

func (s *NumberSuite) TestEquals() {
	n1 := NewInteger(5)
	n2 := NewInteger(5)
	n3 := NewInteger(10)
	n4 := NewReal(10, 2) // 5.0

	s.True(n1.Equals(n2))
	s.False(n1.Equals(n3))
	s.True(n1.Equals(n4)) // Integer 5 equals Real 5.0
}

func (s *NumberSuite) TestKindTransitions() {
	// Start with Natural, subtract to become Integer
	n := NewNatural(5)
	s.Equal(KindNatural, n.Kind())

	result := n.Sub(NewNatural(10))
	s.Equal(KindInteger, result.Kind())

	// Start with Integer, divide to become Real
	n1 := NewInteger(5)
	n2 := NewInteger(2)
	result = n1.Div(n2)
	s.Equal(KindRational, result.Kind())

	// Real that happens to be whole number normalizes to Natural
	n = NewReal(10, 1)
	s.Equal(KindNatural, n.Kind())
}

func (s *NumberSuite) TestRealContamination() {
	// Once a number is Real (irrational), any operation with it stays Real
	real := NewFloat(1.4142135623730951) // sqrt(2)
	s.Equal(KindReal, real.Kind())

	// Real + Natural = Real
	result := real.Add(NewNatural(1))
	s.Equal(KindReal, result.Kind())

	// Real + Integer = Real
	result = real.Add(NewInteger(-1))
	s.Equal(KindReal, result.Kind())

	// Real + Rational = Real
	result = real.Add(NewRational(1, 2))
	s.Equal(KindReal, result.Kind())

	// Natural + Real = Real
	result = NewNatural(1).Add(real)
	s.Equal(KindReal, result.Kind())

	// Real - Natural = Real
	result = real.Sub(NewNatural(1))
	s.Equal(KindReal, result.Kind())

	// Real * Natural = Real
	result = real.Mul(NewNatural(2))
	s.Equal(KindReal, result.Kind())

	// Real / Natural = Real
	result = real.Div(NewNatural(2))
	s.Equal(KindReal, result.Kind())

	// Real negation = Real
	result = real.Neg()
	s.Equal(KindReal, result.Kind())

	// Real absolute value = Real
	result = NewFloat(-3.14).Abs()
	s.Equal(KindReal, result.Kind())
}

func (s *NumberSuite) TestRealFromPow() {
	// 2^(1/2) = sqrt(2) is irrational, returns Real
	base := NewNatural(2)
	exp := NewRational(1, 2)
	result, err := base.Pow(exp)
	s.NoError(err)
	s.Equal(KindReal, result.Kind())
	s.InDelta(1.4142135623730951, result.Float64(), 0.0000001)

	// 4^(1/2) = 2 is rational, returns Natural
	base = NewNatural(4)
	result, err = base.Pow(exp)
	s.NoError(err)
	s.Equal(KindNatural, result.Kind())
	s.Equal("2", result.Inspect())

	// Real base with any exponent = Real
	realBase := NewFloat(2.5)
	result, err = realBase.Pow(NewNatural(2))
	s.NoError(err)
	s.Equal(KindReal, result.Kind())

	// Any base with Real exponent = Real
	result, err = NewNatural(2).Pow(NewFloat(0.5))
	s.NoError(err)
	s.Equal(KindReal, result.Kind())
}

func (s *NumberSuite) TestRealComparison() {
	// Compare Real with exact rationals
	sqrt2 := NewFloat(1.4142135623730951)
	onePointFour := NewRational(14, 10)

	// sqrt(2) > 1.4
	s.Equal(1, sqrt2.Cmp(onePointFour))

	// Real == Real
	sqrt2Copy := NewFloat(1.4142135623730951)
	s.True(sqrt2.Equals(sqrt2Copy))

	// Real with same value as rational
	twoReal := NewFloat(2.0)
	twoNatural := NewNatural(2)
	s.True(twoReal.Equals(twoNatural))
}
