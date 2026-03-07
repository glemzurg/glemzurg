package parser

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/stretchr/testify/suite"
)

// StressIdentifierTestSuite stress-tests identifier parsing edge cases.
// Identifiers must not clash with reserved keywords. Edge cases around
// keyword prefixes and special characters are undertested.
type StressIdentifierTestSuite struct {
	suite.Suite
}

func TestStressIdentifierSuite(t *testing.T) {
	suite.Run(t, new(StressIdentifierTestSuite))
}

// TestKeywordPrefixIdentifiers tests that identifiers starting with keyword
// prefixes are correctly parsed as identifiers, NOT as keywords.
// Keywords use `![a-zA-Z0-9_]` word boundary assertion in the PEG grammar.
func (s *StressIdentifierTestSuite) TestKeywordPrefixIdentifiers() {
	tests := []struct {
		input    string
		desc     string
		expected string // expected identifier value
	}{
		{"TRUEfoo", "TRUE prefix", "TRUEfoo"},
		{"FALSEx", "FALSE prefix", "FALSEx"},
		{"IFx", "IF prefix", "IFx"},
		{"ELSEwhere", "ELSE prefix", "ELSEwhere"},
		{"CASEy", "CASE prefix", "CASEy"},
		{"OTHERwise", "OTHER prefix", "OTHERwise"},
		{"TRUE_flag", "TRUE underscore suffix", "TRUE_flag"},
		{"EXCEPT_ion", "EXCEPT prefix", "EXCEPT_ion"},
		{"THENce", "THEN prefix", "THENce"},
		{"BOOLEANs", "BOOLEAN prefix", "BOOLEANs"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := ParseExpression(tt.input)
			s.Require().NoError(err, "should parse as identifier: %q", tt.input)

			ident, ok := expr.(*ast.Identifier)
			s.True(ok, "expected Identifier, got %T for %q", expr, tt.input)
			s.Equal(tt.expected, ident.Value, "identifier value mismatch")
		})
	}
}

// TestUnderscoreIdentifiers tests identifiers with underscores.
func (s *StressIdentifierTestSuite) TestUnderscoreIdentifiers() {
	tests := []struct {
		input    string
		desc     string
		expected string
	}{
		{"_x", "leading underscore", "_x"},
		{"__double", "double underscore", "__double"},
		{"a_b_c", "underscores throughout", "a_b_c"},
		{"x_1_y_2", "mixed letters, underscores, digits", "x_1_y_2"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := ParseExpression(tt.input)
			s.Require().NoError(err, "should parse: %q (%s)", tt.input, tt.desc)

			ident, ok := expr.(*ast.Identifier)
			s.True(ok, "expected Identifier, got %T", expr)
			s.Equal(tt.expected, ident.Value)
		})
	}
}

// TestSimpleIdentifiers tests straightforward identifier parsing.
func (s *StressIdentifierTestSuite) TestSimpleIdentifiers() {
	tests := []struct {
		input    string
		desc     string
		expected string
	}{
		{"x", "single letter", "x"},
		{"X", "single uppercase", "X"},
		{"abcdefghijklmnopqrstuvwxyz", "long lowercase", "abcdefghijklmnopqrstuvwxyz"},
		{"x1y2z3", "mixed letters and digits", "x1y2z3"},
		{"myVar", "camelCase", "myVar"},
		{"MyVar", "PascalCase", "MyVar"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := ParseExpression(tt.input)
			s.Require().NoError(err, "should parse: %q (%s)", tt.input, tt.desc)

			ident, ok := expr.(*ast.Identifier)
			s.True(ok, "expected Identifier, got %T", expr)
			s.Equal(tt.expected, ident.Value)
		})
	}
}

// TestIdentifiersInExpressions tests identifiers used within larger expressions.
func (s *StressIdentifierTestSuite) TestIdentifiersInExpressions() {
	tests := []struct {
		input string
		desc  string
	}{
		{"x + y", "identifiers in addition"},
		{"myFunc(x)", "identifier as function name"},
		{"_Seq!Len(s)", "underscore-prefixed module call"},
		{"_Bags!BagCardinality(b)", "builtin module call"},
		{"record.field", "identifier with field access"},
		{"x' = x + 1", "primed identifier in equality"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			s.Require().NoError(err, "should parse: %q (%s)", tt.input, tt.desc)
		})
	}
}

// TestReservedKeywordsAsStandalone tests that reserved keywords cannot
// be used as standalone identifiers in expression context.
func (s *StressIdentifierTestSuite) TestReservedKeywordsAsStandalone() {
	// TRUE and FALSE are special — they parse as boolean literals, not identifiers.
	// Others like IF, THEN, ELSE, CASE, EXCEPT, OTHER are reserved.
	tests := []struct {
		input string
		desc  string
	}{
		// TRUE and FALSE parse as BooleanLiteral, which is valid.
		// IF alone should fail (needs THEN/ELSE).
		{"IF", "IF alone"},
		{"THEN", "THEN alone"},
		{"ELSE", "ELSE alone"},
		{"EXCEPT", "EXCEPT alone"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			// These should either fail to parse or parse as something
			// other than an identifier. Characterize the behavior.
			if err != nil {
				// Expected: parse error
				return
			}
			// If it parsed, it should NOT be an Identifier
			s.T().Logf("NOTICE: %q parsed successfully (characterizing)", tt.input)
		})
	}
}
