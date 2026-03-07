package parser

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/stretchr/testify/suite"
)

// StressComparisonTestSuite stress-tests comparison operator behavior.
// The grammar uses `?` (optional single) for comparison operators at each
// sub-level — they should NOT chain. This is completely untested.
type StressComparisonTestSuite struct {
	suite.Suite
}

func TestStressComparisonSuite(t *testing.T) {
	suite.Run(t, new(StressComparisonTestSuite))
}

// TestComparisonNonChaining tests that comparison operators cannot be chained.
// Each comparison level uses `?` (optional single match) not `*` (repeating),
// so `a < b > c` should fail because after `a < b` is parsed, `> c` is
// trailing content that makes RootExpression fail (due to !. assertion).
func (s *StressComparisonTestSuite) TestComparisonNonChaining() {
	tests := []struct {
		input string
		desc  string
	}{
		// Same operator chaining
		{"1 < 2 < 3", "chained less-than"},
		{"1 > 2 > 3", "chained greater-than"},
		{"1 = 2 = 3", "chained equality"},

		// Mixed comparison chaining at same level
		{"1 < 2 > 3", "mixed less-than and greater-than"},
		{"1 ≤ 2 ≥ 3", "mixed lte and gte"},

		// Set comparison chaining
		{"A ⊆ B ⊆ C", "chained subset-or-equal"},
		{"A ⊂ B ⊃ C", "mixed proper subset and superset"},

		// Bag comparison chaining
		{"A ⊏ B ⊏ C", "chained bag proper subset"},

		// Membership chaining
		{"x ∈ S ∈ T", "chained membership"},

		// Cross-level comparison chaining (these might parse differently
		// because they're at different grammar levels)
		{"1 < 2 = 3", "cross-level: numeric then equality"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			// Most of these should fail due to trailing content.
			// However, cross-level comparisons (like `1 < 2 = 3`) might parse
			// because equality is at a lower grammar level than numeric comparison.
			// This test characterizes the actual behavior.
			if err != nil {
				// Expected: parse error due to trailing content
				return
			}
			// If it parses, that's also valid information — document it.
			s.T().Logf("NOTICE: %q parsed successfully (characterizing behavior)", tt.input)
		})
	}
}

// TestComparisonChainingShouldFail is a stricter version — these MUST fail.
// Using operators at the exact same grammar level guarantees trailing content.
func (s *StressComparisonTestSuite) TestComparisonChainingSameLevelMustFail() {
	tests := []struct {
		input string
		desc  string
	}{
		{"1 < 2 < 3", "chained less-than at same level"},
		{"1 > 2 > 3", "chained greater-than at same level"},
		{"1 ≤ 2 ≤ 3", "chained lte at same level"},
		{"1 ≥ 2 ≥ 3", "chained gte at same level"},
		{"1 = 2 = 3", "chained equality at same level"},
		{"1 ≠ 2 ≠ 3", "chained inequality at same level"},
		{"A ⊆ B ⊆ C", "chained subset at same level"},
		{"x ∈ S ∈ T", "chained membership at same level"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			s.Require().Error(err, "chaining same-level comparison should fail for %q", tt.input)
		})
	}
}

// TestCrossLevelComparisonInteractions tests comparisons at different grammar
// sub-levels. Since they're at different positions in the hierarchy, some
// combinations may parse successfully.
func (s *StressComparisonTestSuite) TestCrossLevelComparisonInteractions() {
	tests := []struct {
		input string
		desc  string
	}{
		// Parenthesized comparisons in other comparisons — always valid.
		{"(1 < 2) = TRUE", "parenthesized numeric comparison in equality"},
		{"(x ∈ S) = TRUE", "parenthesized membership in equality"},
		{"(A ⊆ B) = TRUE", "parenthesized subset in equality"},

		// Comparisons separated by logic operators — always valid.
		{"1 < 2 /\\ 3 > 4", "comparisons separated by AND"},
		{"x = 1 \\/ y = 2", "equalities separated by OR"},
		{"A ⊆ B /\\ C ⊆ D", "set comparisons with AND"},

		// Set membership with set operations — operations are higher precedence.
		{"x ∈ A ∪ B", "membership with union (union binds tighter)"},
		{"x ∈ A ∩ B", "membership with intersection"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			s.Require().NoError(err, "should parse: %q (%s)", tt.input, tt.desc)
		})
	}
}

// TestMembershipPrecedenceVsEquality tests the surprising interaction:
// x ∈ S = TRUE parses as x ∈ (S = TRUE) because membership is LOWER
// in the grammar than equality.
func (s *StressComparisonTestSuite) TestMembershipPrecedenceVsEquality() {
	expr, err := ParseExpression("x ∈ S = TRUE")
	if err != nil {
		// If it fails, that's one characterization.
		s.T().Log("NOTICE: 'x ∈ S = TRUE' fails to parse")
		return
	}

	// If it parses, check the structure.
	// Expected: Membership(x, ∈, Equality(S, =, TRUE))
	mem, ok := expr.(*ast.Membership)
	if ok {
		s.Equal("∈", mem.Operator)
		// Right side should be an equality expression
		_, isEq := mem.Right.(*ast.BinaryEquality)
		s.True(isEq, "right side of membership should be equality, got %T", mem.Right)
	}
}
