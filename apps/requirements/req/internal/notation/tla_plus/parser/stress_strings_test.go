package parser

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/stretchr/testify/suite"
)

// StressStringTestSuite stress-tests string literal parsing edge cases.
type StressStringTestSuite struct {
	suite.Suite
}

func TestStressStringSuite(t *testing.T) {
	suite.Run(t, new(StressStringTestSuite))
}

// TestBasicStrings tests basic string literal parsing.
func (s *StressStringTestSuite) TestBasicStrings() {
	tests := []struct {
		input string
		desc  string
	}{
		{`""`, "empty string"},
		{`"hello"`, "simple word"},
		{`"hello world"`, "two words with space"},
		{`"Hello, World!"`, "with punctuation"},
		{`"123"`, "numeric string"},
		{`"true"`, "lowercase true (not boolean)"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := ParseExpression(tt.input)
			s.Require().NoError(err, "should parse: %q (%s)", tt.input, tt.desc)

			_, ok := expr.(*ast.StringLiteral)
			s.True(ok, "expected StringLiteral, got %T for %q", expr, tt.input)
		})
	}
}

// TestStringEscapeSequences tests strings with escape characters.
func (s *StressStringTestSuite) TestStringEscapeSequences() {
	tests := []struct {
		input string
		desc  string
	}{
		{`"line1\nline2"`, "newline escape"},
		{`"tab\there"`, "tab escape"},
		{`"quote\"inside"`, "escaped quote"},
		{`"backslash\\"`, "escaped backslash"},
		{`"mixed\t\n"`, "multiple escapes"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			// Some escape sequences may or may not be supported.
			// Characterize the behavior.
			if err != nil {
				s.T().Logf("NOTICE: %q fails to parse: %v", tt.input, err)
			} else {
				s.T().Logf("NOTICE: %q parses successfully", tt.input)
			}
		})
	}
}

// TestUnterminatedStrings tests that unterminated strings are rejected.
func (s *StressStringTestSuite) TestUnterminatedStrings() {
	tests := []struct {
		input string
		desc  string
	}{
		{`"hello`, "missing closing quote"},
		{`"`, "single quote only"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			s.Error(err, "unterminated string should fail: %q", tt.input)
		})
	}
}

// TestStringsInExpressions tests strings used in various expression contexts.
func (s *StressStringTestSuite) TestStringsInExpressions() {
	tests := []struct {
		input string
		desc  string
	}{
		{`"a" = "b"`, "string equality"},
		{`"a" /= "b"`, "string inequality"},
		{`{" hello", "world"}`, "strings in set"},
		{`<<"a", "b", "c">>`, "strings in tuple"},
		{`[name |-> "Alice"]`, "string in record"},
		{`IF TRUE THEN "yes" ELSE "no"`, "strings in IF"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			s.Require().NoError(err, "should parse: %q (%s)", tt.input, tt.desc)
		})
	}
}

// TestStringIndexing tests string indexing operations.
func (s *StressStringTestSuite) TestStringIndexing() {
	tests := []struct {
		input string
		desc  string
	}{
		{`"hello"[1]`, "first character index"},
		{`"hello"[5]`, "last character index"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			// String indexing may or may not be supported at parse level.
			if err != nil {
				s.T().Logf("NOTICE: %q fails to parse: %v", tt.input, err)
			} else {
				s.T().Logf("NOTICE: %q parses successfully", tt.input)
			}
		})
	}
}
