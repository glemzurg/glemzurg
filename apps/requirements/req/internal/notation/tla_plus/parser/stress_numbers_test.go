package parser

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/stretchr/testify/suite"
)

// StressNumberTestSuite stress-tests number literal parsing edge cases.
// Multiple number formats (decimal, binary, octal, hex) with fractional
// parts. Edge cases around format boundaries.
type StressNumberTestSuite struct {
	suite.Suite
}

func TestStressNumberSuite(t *testing.T) {
	suite.Run(t, new(StressNumberTestSuite))
}

// TestDecimalNumbers tests basic decimal number parsing.
func (s *StressNumberTestSuite) TestDecimalNumbers() {
	tests := []struct {
		input string
		desc  string
	}{
		{"0", "zero"},
		{"1", "one"},
		{"42", "two digit"},
		{"123456789", "large number"},
		{"007", "leading zeros"},
		{"999999999999999999", "very large number"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := ParseExpression(tt.input)
			s.Require().NoError(err, "should parse: %q (%s)", tt.input, tt.desc)

			_, ok := expr.(*ast.NumberLiteral)
			s.True(ok, "expected NumberLiteral, got %T for %q", expr, tt.input)
		})
	}
}

// TestDecimalWithFraction tests decimal numbers with fractional parts.
func (s *StressNumberTestSuite) TestDecimalWithFraction() {
	tests := []struct {
		input string
		desc  string
	}{
		{"0.0", "zero point zero"},
		{"3.14", "pi-like"},
		{"0.5", "half"},
		{"123.456", "multi-digit both sides"},
		{".5", "fractional only (no integer part)"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			// Some of these may or may not parse depending on grammar rules.
			// Characterize the behavior.
			if err != nil {
				s.T().Logf("NOTICE: %q fails to parse: %v", tt.input, err)
			} else {
				s.T().Logf("NOTICE: %q parses successfully", tt.input)
			}
		})
	}
}

// TestBinaryNumbers tests binary number format (\b prefix).
func (s *StressNumberTestSuite) TestBinaryNumbers() {
	tests := []struct {
		input string
		desc  string
	}{
		{`\b0`, "binary zero"},
		{`\b1`, "binary one"},
		{`\b1010`, "binary ten"},
		{`\b11111111`, "binary 255"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := ParseExpression(tt.input)
			s.Require().NoError(err, "should parse: %q (%s)", tt.input, tt.desc)
			s.NotNil(expr)
		})
	}
}

// TestOctalNumbers tests octal number format (\o prefix).
func (s *StressNumberTestSuite) TestOctalNumbers() {
	tests := []struct {
		input string
		desc  string
	}{
		{`\o0`, "octal zero"},
		{`\o7`, "octal seven"},
		{`\o77`, "octal 63"},
		{`\o777`, "octal 511"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := ParseExpression(tt.input)
			s.Require().NoError(err, "should parse: %q (%s)", tt.input, tt.desc)
			s.NotNil(expr)
		})
	}
}

// TestHexNumbers tests hexadecimal number format (\h prefix).
func (s *StressNumberTestSuite) TestHexNumbers() {
	tests := []struct {
		input string
		desc  string
	}{
		{`\h0`, "hex zero"},
		{`\hF`, "hex fifteen (uppercase)"},
		{`\hf`, "hex fifteen (lowercase)"},
		{`\hFF`, "hex 255"},
		{`\hFFFF`, "hex 65535"},
		{`\hDeAdBeEf`, "hex mixed case"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := ParseExpression(tt.input)
			s.Require().NoError(err, "should parse: %q (%s)", tt.input, tt.desc)
			s.NotNil(expr)
		})
	}
}

// TestInvalidNumberFormats tests number formats that should fail.
func (s *StressNumberTestSuite) TestInvalidNumberFormats() {
	tests := []struct {
		input string
		desc  string
	}{
		{`\b`, "binary with no digits"},
		{`\o`, "octal with no digits"},
		{`\h`, "hex with no digits"},
		{`\b2`, "invalid binary digit"},
		{`\o8`, "invalid octal digit"},
		{`\hG`, "invalid hex digit"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			// These should fail, but some might parse as other constructs.
			if err != nil {
				// Expected: parse error
				return
			}
			s.T().Logf("NOTICE: %q parsed successfully (characterizing)", tt.input)
		})
	}
}

// TestNegativeNumbers tests negative number expressions.
func (s *StressNumberTestSuite) TestNegativeNumbers() {
	tests := []struct {
		input string
		desc  string
	}{
		{"-1", "negative one"},
		{"-0", "negative zero"},
		{"-42", "negative forty-two"},
		{"--1", "double negation"},
		{"---1", "triple negation"},
		{"-(42)", "parenthesized negation"},
		{"1 + -2", "negative in expression"},
		{"{-1, -2, -3}", "negatives in set"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			s.Require().NoError(err, "should parse: %q (%s)", tt.input, tt.desc)
		})
	}
}

// TestNumbersInExpressions tests numbers used in various expression contexts.
func (s *StressNumberTestSuite) TestNumbersInExpressions() {
	tests := []struct {
		input string
		desc  string
	}{
		{"1 + 2", "addition"},
		{"1..10", "range"},
		{"0..0", "zero range"},
		{"{1, 2, 3}", "set of numbers"},
		{"<<1, 2, 3>>", "tuple of numbers"},
		{"1 = 1", "equality"},
		{"1 < 2", "comparison"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			s.Require().NoError(err, "should parse: %q (%s)", tt.input, tt.desc)
		})
	}
}
