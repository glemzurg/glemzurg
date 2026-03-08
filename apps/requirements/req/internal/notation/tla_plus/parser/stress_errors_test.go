package parser

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// StressErrorTestSuite stress-tests parse error handling.
// The parser should reject malformed input gracefully.
type StressErrorTestSuite struct {
	suite.Suite
}

func TestStressErrorSuite(t *testing.T) {
	suite.Run(t, new(StressErrorTestSuite))
}

// TestIncompleteExpressions tests expressions that are cut off mid-parse.
func (s *StressErrorTestSuite) TestIncompleteExpressions() {
	tests := []struct {
		input string
		desc  string
	}{
		{"1 +", "trailing addition operator"},
		{"1 *", "trailing multiplication operator"},
		{"1 =", "trailing equality operator"},
		{"1 <", "trailing less-than"},
		{"1 >", "trailing greater-than"},
		{`1 \in`, "trailing membership operator"},
		{"IF TRUE THEN", "IF missing THEN-expr and ELSE"},
		{"IF TRUE THEN 1", "IF missing ELSE clause"},
		{"IF TRUE", "IF missing THEN and ELSE"},
		{`\A x \in S :`, "quantifier missing predicate"},
		{`\A x \in S`, "quantifier missing colon and predicate"},
		{`\A x`, "quantifier missing set and predicate"},
		{"CASE", "empty CASE expression"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			s.Require().Error(err, "expected parse error for %q (%s)", tt.input, tt.desc)
		})
	}
}

// TestUnmatchedDelimiters tests expressions with missing closing delimiters.
func (s *StressErrorTestSuite) TestUnmatchedDelimiters() {
	tests := []struct {
		input string
		desc  string
	}{
		{"(1 + 2", "unmatched open parenthesis"},
		{"1 + 2)", "unmatched close parenthesis"},
		{"{1, 2", "unmatched open brace (set)"},
		{"<<1, 2", "unmatched open angle-angle (tuple)"},
		{"[x |-> 1", "unmatched open bracket (record)"},
		{`"hello`, "unterminated string literal"},
		{"((1)", "nested unmatched parenthesis"},
		{"{1, {2, 3}", "nested unmatched brace"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			s.Require().Error(err, "expected parse error for %q (%s)", tt.input, tt.desc)
		})
	}
}

// TestInvalidOperatorCombinations tests sequences of operators that don't form valid expressions.
func (s *StressErrorTestSuite) TestInvalidOperatorCombinations() {
	tests := []struct {
		input string
		desc  string
	}{
		{"1 + * 2", "adjacent binary operators (+ *)"},
		{"1 2", "missing operator between integer operands"},
		{"/ 2", "leading division operator"},
		{"* 2", "leading multiplication operator"},
		{"1 .. .. 5", "double range operator"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			s.Require().Error(err, "expected parse error for %q (%s)", tt.input, tt.desc)
		})
	}
}

// TestMalformedConstructs tests structurally broken expressions.
func (s *StressErrorTestSuite) TestMalformedConstructs() {
	tests := []struct {
		input string
		desc  string
	}{
		{"[x |->]", "record missing value after |->"},
		{"[|-> 1]", "record missing field name before |->"},
		{"CASE -> 1", "CASE missing condition before arrow"},
		{"CASE x > 0", "CASE missing arrow and result"},
		{"Func(,)", "function call with empty argument before comma"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			s.Require().Error(err, "expected parse error for %q (%s)", tt.input, tt.desc)
		})
	}
}

// TestTrailingContent tests that the parser rejects input with extra content after a valid expression.
// The PEG grammar's RootExpression rule uses !. (end-of-input assertion).
func (s *StressErrorTestSuite) TestTrailingContent() {
	tests := []struct {
		input string
		desc  string
	}{
		{"1 + 2 xyz", "trailing identifier after arithmetic"},
		{"TRUE FALSE", "two boolean literals"},
		{"42 42", "two integer literals"},
		{"{1} {2}", "two set literals"},
		{"(1 + 2)(3)", "parenthesized expr followed by parenthesized expr"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			s.Require().Error(err, "expected parse error for %q (%s)", tt.input, tt.desc)
		})
	}
}

// TestReservedKeywordMisuse tests that reserved keywords cannot be used as standalone identifiers.
func (s *StressErrorTestSuite) TestReservedKeywordMisuse() {
	tests := []struct {
		input string
		desc  string
	}{
		// These keywords should not parse as identifiers when used alone.
		// Note: IF/THEN/ELSE/CASE/EXCEPT/OTHER are reserved.
		{"CASE = 1", "CASE used as operand in equality"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			s.Require().Error(err, "expected parse error for %q (%s)", tt.input, tt.desc)
		})
	}
}

// TestValidEdgeCasesNotErrors tests inputs that look suspicious but should parse successfully.
func (s *StressErrorTestSuite) TestValidEdgeCasesNotErrors() {
	tests := []struct {
		input string
		desc  string
	}{
		// TRUE + FALSE is syntactically valid (two booleans with +); fails only at eval.
		{"TRUE + FALSE", "boolean arithmetic is syntactically valid"},
		// + is not a prefix operator, but - is. Negation chains are valid.
		{"--1", "double negation is valid"},
		{"---1", "triple negation is valid"},
		// Empty set and tuple are valid.
		{"{}", "empty set literal"},
		{"<<>>", "empty tuple literal"},
		// @ is a valid atomic expression (though only meaningful in EXCEPT context).
		{"@", "existing value reference"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			s.Require().NoError(err, "expected no parse error for %q (%s)", tt.input, tt.desc)
		})
	}
}
