package parser

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/stretchr/testify/suite"
)

// StressCaseTestSuite stress-tests CASE expression parsing and [] disambiguation.
// The `[]` token is both the CASE branch separator AND the tuple/string indexing
// operator. CASE conditions and results are restricted to OrExpr to avoid
// consuming `[]` as part of the expression.
type StressCaseTestSuite struct {
	suite.Suite
}

func TestStressCaseSuite(t *testing.T) {
	suite.Run(t, new(StressCaseTestSuite))
}

// TestCaseBasicParsing tests standard CASE expressions parse correctly.
func (s *StressCaseTestSuite) TestCaseBasicParsing() {
	tests := []struct {
		input    string
		desc     string
		branches int // expected number of branches (excluding OTHER)
		hasOther bool
	}{
		{"CASE x > 0 -> 1 [] OTHER -> 0", "single branch with OTHER", 1, true},
		{"CASE x > 0 -> 1 [] x < 0 -> -1 [] OTHER -> 0", "two branches with OTHER", 2, true},
		{"CASE x > 0 -> 1", "single branch without OTHER", 1, false},
		{"CASE x > 0 -> 1 [] x < 0 -> -1", "two branches without OTHER", 2, false},
		{"CASE TRUE -> 42 [] OTHER -> 0", "trivial condition", 1, true},
		{"CASE x > 0 -> 1 [] x = 0 -> 0 [] x < 0 -> -1 [] OTHER -> 99", "three branches with OTHER", 3, true},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := ParseExpression(tt.input)
			s.Require().NoError(err, "should parse: %q", tt.input)

			caseExpr, ok := expr.(*ast.CaseExpr)
			s.True(ok, "expected CaseExpr, got %T", expr)
			s.Len(caseExpr.Branches, tt.branches, "branch count for %q", tt.input)
			if tt.hasOther {
				s.NotNil(caseExpr.Other, "expected OTHER clause in %q", tt.input)
			} else {
				s.Nil(caseExpr.Other, "expected no OTHER clause in %q", tt.input)
			}
		})
	}
}

// TestCaseWithIndexingInResults tests that indexing operators in CASE results
// don't conflict with the [] branch separator.
func (s *StressCaseTestSuite) TestCaseWithIndexingInResults() {
	tests := []struct {
		input string
		desc  string
	}{
		// Indexing inside results — the `[1]` is a suffix of FieldAccessExpr,
		// which is inside AtomicExpr inside OrExpr, so it should not conflict
		// with the `[]` CASE separator.
		{"CASE x > 0 -> a[1] [] OTHER -> 0", "array indexing in result"},
		{"CASE TRUE -> <<1, 2>>[1] [] OTHER -> 0", "tuple indexing in result"},
		{"CASE TRUE -> r.field [] OTHER -> 0", "field access in result"},

		// Record EXCEPT in result — starts with `[` but is a complete expression.
		{"CASE TRUE -> [r EXCEPT !.x = 1] [] OTHER -> r", "record EXCEPT in result"},

		// Record literal in result — starts with `[`.
		{"CASE TRUE -> [a |-> 1] [] OTHER -> [b |-> 2]", "record literals in results"},

		// Set in result.
		{"CASE TRUE -> {1, 2, 3} [] OTHER -> {}", "set literals in results"},

		// Tuple in result.
		{"CASE TRUE -> <<1, 2>> [] OTHER -> <<>>", "tuple literals in results"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := ParseExpression(tt.input)
			s.Require().NoError(err, "should parse: %q (%s)", tt.input, tt.desc)

			caseExpr, ok := expr.(*ast.CaseExpr)
			s.True(ok, "expected CaseExpr, got %T", expr)
			s.NotNil(caseExpr.Other, "expected OTHER clause")
		})
	}
}

// TestCaseConditionOrExprRestriction tests that CASE conditions are restricted
// to OrExpr, which excludes implies (⇒) and equivalence (≡) at the top level.
func (s *StressCaseTestSuite) TestCaseConditionOrExprRestriction() {
	tests := []struct {
		input       string
		desc        string
		shouldParse bool
	}{
		// Parenthesized implies in condition — should work because parens
		// make it an AtomicExpr inside OrExpr.
		{"CASE (a => b) -> 1 [] OTHER -> 0", "parenthesized implies in condition", true},

		// Bare implies in condition — `a` becomes the condition (it's a valid
		// OrExpr), then `=> b -> 1` is NOT a valid CaseArrow. This should fail.
		{"CASE a => b -> 1 [] OTHER -> 0", "bare implies in condition", false},

		// Parenthesized implies in result — should work.
		{"CASE TRUE -> (a => b) [] OTHER -> FALSE", "parenthesized implies in result", true},

		// Parenthesized equivalence in condition — should work.
		{"CASE (a <=> b) -> 1 [] OTHER -> 0", "parenthesized equiv in condition", true},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			if tt.shouldParse {
				s.Require().NoError(err, "should parse: %q (%s)", tt.input, tt.desc)
			} else {
				s.Error(err, "should fail: %q (%s)", tt.input, tt.desc)
			}
		})
	}
}

// TestCaseWithUnicodeDelimiters tests Unicode CASE delimiters (→ and □).
func (s *StressCaseTestSuite) TestCaseWithUnicodeDelimiters() {
	tests := []struct {
		input string
		desc  string
	}{
		{"CASE x > 0 → 1 □ x < 0 → -1 □ OTHER → 0", "all Unicode delimiters"},
		{"CASE TRUE → 1 □ OTHER → 0", "minimal Unicode CASE"},
		{"CASE TRUE -> 1 □ OTHER → 0", "mixed: ASCII arrow + Unicode separator"},
		{"CASE TRUE → 1 [] OTHER -> 0", "mixed: Unicode arrow + ASCII separator"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := ParseExpression(tt.input)
			s.Require().NoError(err, "should parse: %q (%s)", tt.input, tt.desc)

			_, ok := expr.(*ast.CaseExpr)
			s.True(ok, "expected CaseExpr, got %T", expr)
		})
	}
}

// TestCaseWithComplexConditions tests CASE with complex conditions that are
// valid OrExpr but exercise various precedence levels.
func (s *StressCaseTestSuite) TestCaseWithComplexConditions() {
	tests := []struct {
		input string
		desc  string
	}{
		{"CASE x > 0 /\\ y > 0 -> 1 [] OTHER -> 0", "AND in condition"},
		{"CASE x > 0 \\/ y > 0 -> 1 [] OTHER -> 0", "OR in condition"},
		{"CASE ~flag -> 1 [] OTHER -> 0", "NOT in condition"},
		{"CASE x + y > 10 -> 1 [] OTHER -> 0", "arithmetic in condition"},
		{`CASE x \in S -> 1 [] OTHER -> 0`, "membership in condition"},
		{"CASE x = 1 /\\ y = 2 -> 1 [] OTHER -> 0", "equality with AND in condition"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := ParseExpression(tt.input)
			s.Require().NoError(err, "should parse: %q (%s)", tt.input, tt.desc)

			caseExpr, ok := expr.(*ast.CaseExpr)
			s.True(ok, "expected CaseExpr, got %T", expr)
			s.NotNil(caseExpr.Other, "expected OTHER clause")
		})
	}
}
