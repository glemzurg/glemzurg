package parser

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type RoundTripTestSuite struct {
	suite.Suite
}

func TestRoundTripSuite(t *testing.T) {
	suite.Run(t, new(RoundTripTestSuite))
}

func (s *RoundTripTestSuite) TestRoundTrip() {
	tests := []string{
		// Booleans
		"TRUE",
		"FALSE",
		// Decimal integers
		"0",
		"42",
		"007",
		// Decimal with fractions
		"3.14",
		".5",
		"0.123",
		"3.140",
		// Binary
		"\\b1010",
		"\\B0011",
		// Octal
		"\\o17",
		"\\O777",
		// Hexadecimal
		"\\hff",
		"\\H0ABC",
		"\\hDeAdBeEf",
		// Negation
		"-1",
		"-.5",
		"-3.14",
		"--1",
		// Fractions
		"3/4",
		"1/2",
		"1.5/2.5",
		".5/.25",
		// Combined
		"-3/4",
		"(-3)/4",
		"3/(-4)",
		"3/-4",
		"-3/-(1.4/.2)",
		// Parentheses
		"(42)",
		"((123))",
		// Strings
		`"hello"`,
		`"hello world"`,
		`""`,
	}

	for _, input := range tests {
		expr, err := ParseExpression(input)
		s.NoError(err, "parsing %q", input)
		s.Equal(input, expr.String(), "round-trip failed for %q", input)
	}
}

// TestArithmeticStringOutput tests that arithmetic expressions parse and stringify correctly.
// Expressions no longer add auto-parentheses - use ParenExpr for explicit parentheses.
func (s *RoundTripTestSuite) TestArithmeticStringOutput() {
	tests := []struct {
		input    string
		expected string
	}{
		// Basic arithmetic operations
		{"1 + 2", "1 + 2"},
		{"1 - 2", "1 - 2"},
		{"2 * 3", "2 * 3"},
		{"6 ÷ 2", "6 ÷ 2"},
		{"7 % 3", "7 % 3"},
		{"2 ^ 3", "2 ^ 3"},
		// ASCII \div parses to Unicode ÷
		{`6 \div 2`, "6 ÷ 2"},
		// Complex expressions
		{"1 + 2 * 3", "1 + 2 * 3"},
	}

	for _, tt := range tests {
		expr, err := ParseExpression(tt.input)
		s.NoError(err, "parsing %q", tt.input)
		s.Equal(tt.expected, expr.String(), "string output failed for %q", tt.input)
	}
}
