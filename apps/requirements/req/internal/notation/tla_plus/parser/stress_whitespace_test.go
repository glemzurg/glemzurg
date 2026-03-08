package parser

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// StressWhitespaceTestSuite stress-tests whitespace handling.
// The PEG grammar uses `ws <- [ \t\n\r]+`. Various whitespace
// patterns are untested elsewhere.
type StressWhitespaceTestSuite struct {
	suite.Suite
}

func TestStressWhitespaceSuite(t *testing.T) {
	suite.Run(t, new(StressWhitespaceTestSuite))
}

// TestWhitespaceVariations tests parsing with different whitespace patterns.
func (s *StressWhitespaceTestSuite) TestWhitespaceVariations() {
	tests := []struct {
		input string
		desc  string
	}{
		{"1+2", "no whitespace around operator"},
		{"1  +  2", "extra whitespace around operator"},
		{"1\t+\t2", "tab whitespace around operator"},
		{"1 +\n2", "newline between tokens"},
		{"  42  ", "leading and trailing whitespace"},
		{" 1 + 2 ", "whitespace around entire expression"},
		{"1   +   2   *   3", "excessive whitespace"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			s.Require().NoError(err, "should parse: %q (%s)", tt.input, tt.desc)
		})
	}
}

// TestWhitespaceInCollections tests whitespace inside set/tuple/record literals.
func (s *StressWhitespaceTestSuite) TestWhitespaceInCollections() {
	tests := []struct {
		input string
		desc  string
	}{
		{"{  1 , 2 , 3  }", "whitespace in set"},
		{"<<  1 , 2  >>", "whitespace in tuple"},
		{"[ x |-> 1 ]", "whitespace in record"},
		{"{ 1,2,3 }", "minimal whitespace in set"},
		{"<<1,2,3>>", "no whitespace in tuple"},
		{"{1 ,2 ,3}", "whitespace before commas only"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			s.Require().NoError(err, "should parse: %q (%s)", tt.input, tt.desc)
		})
	}
}

// TestWhitespaceInControlFlow tests whitespace in IF/CASE constructs.
func (s *StressWhitespaceTestSuite) TestWhitespaceInControlFlow() {
	tests := []struct {
		input string
		desc  string
	}{
		{"IF  TRUE  THEN  1  ELSE  2", "extra whitespace in IF"},
		{"IF TRUE THEN 1 ELSE 2", "minimal whitespace in IF"},
		{"IF\tTRUE\tTHEN\t1\tELSE\t2", "tab whitespace in IF"},
		{"IF\nTRUE\nTHEN\n1\nELSE\n2", "newline whitespace in IF"},
		{"CASE  TRUE  ->  1  []  OTHER  ->  0", "extra whitespace in CASE"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			s.Require().NoError(err, "should parse: %q (%s)", tt.input, tt.desc)
		})
	}
}

// TestWhitespaceInQuantifiers tests whitespace in quantifier expressions.
func (s *StressWhitespaceTestSuite) TestWhitespaceInQuantifiers() {
	tests := []struct {
		input string
		desc  string
	}{
		{`∀  x  ∈  S  :  x > 0`, "extra whitespace in quantifier"},
		{`∀ x ∈ S : x > 0`, "minimal whitespace in quantifier"},
		{"∀\tx\t∈\tS\t:\tx > 0", "tab whitespace in quantifier"},
		{"∀\nx\n∈\nS\n:\nx > 0", "newline whitespace in quantifier"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			s.Require().NoError(err, "should parse: %q (%s)", tt.input, tt.desc)
		})
	}
}

// TestMultilineExpressions tests expressions spanning multiple lines.
func (s *StressWhitespaceTestSuite) TestMultilineExpressions() {
	tests := []struct {
		input string
		desc  string
	}{
		{"1 +\n2 +\n3", "addition across lines"},
		{"IF TRUE\nTHEN 1\nELSE 2", "IF across lines"},
		{"{1,\n2,\n3}", "set across lines"},
		{"<<1,\n2,\n3>>", "tuple across lines"},
		{"[a |-> 1,\nb |-> 2]", "record across lines"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			s.Require().NoError(err, "should parse: %q (%s)", tt.input, tt.desc)
		})
	}
}
