package parser

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/stretchr/testify/suite"
)

// StressQuantifierTestSuite stress-tests quantifier parsing edge cases.
// Quantifier predicates consume full Expression (recursive), which enables
// deep nesting, complex predicates, and variable shadowing.
type StressQuantifierTestSuite struct {
	suite.Suite
}

func TestStressQuantifierSuite(t *testing.T) {
	suite.Run(t, new(StressQuantifierTestSuite))
}

// TestNestedQuantifiers tests deeply nested quantifier expressions.
func (s *StressQuantifierTestSuite) TestNestedQuantifiers() {
	tests := []struct {
		input string
		desc  string
	}{
		{`\A x \in S : \A y \in T : x = y`, "nested universal quantifiers"},
		{`\E x \in S : \E y \in T : x + y > 10`, "nested existential quantifiers"},
		{`\A x \in S : \E y \in T : x < y`, "mixed nesting (forall-exists)"},
		{`\E x \in S : \A y \in T : x < y`, "mixed nesting (exists-forall)"},
		{`\A x \in {1, 2} : \A y \in {3, 4} : \A z \in {5, 6} : x + y + z > 8`, "triple nesting"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := ParseExpression(tt.input)
			s.NoError(err, "should parse: %q (%s)", tt.input, tt.desc)

			// Verify it's a quantifier
			quant, ok := expr.(*ast.Quantifier)
			s.True(ok, "expected Quantifier, got %T", expr)

			// The predicate of the outer quantifier should also be a quantifier
			_, isInnerQuant := quant.Predicate.(*ast.Quantifier)
			s.True(isInnerQuant, "inner expression should be Quantifier for nested case")
		})
	}
}

// TestVariableShadowing tests that a quantifier can use the same variable
// name as an outer quantifier (variable shadowing).
func (s *StressQuantifierTestSuite) TestVariableShadowing() {
	expr, err := ParseExpression(`\A x \in S : \A x \in T : x > 0`)
	s.NoError(err, "variable shadowing should parse")

	outer, ok := expr.(*ast.Quantifier)
	s.True(ok, "expected Quantifier")

	// The variable is inside the Membership node's Left field
	outerMem, ok := outer.Membership.(*ast.Membership)
	s.True(ok, "outer membership should be Membership")
	outerVar, ok := outerMem.Left.(*ast.Identifier)
	s.True(ok, "outer variable should be Identifier")
	s.Equal("x", outerVar.Value)

	inner, ok := outer.Predicate.(*ast.Quantifier)
	s.True(ok, "predicate should be Quantifier")
	innerMem, ok := inner.Membership.(*ast.Membership)
	s.True(ok, "inner membership should be Membership")
	innerVar, ok := innerMem.Left.(*ast.Identifier)
	s.True(ok, "inner variable should be Identifier")
	s.Equal("x", innerVar.Value, "inner quantifier should shadow outer variable")
}

// TestComplexPredicates tests quantifiers with complex predicate expressions.
func (s *StressQuantifierTestSuite) TestComplexPredicates() {
	tests := []struct {
		input string
		desc  string
	}{
		{`\A x \in S : x > 0 /\ x < 100`, "AND in predicate"},
		{`\A x \in S : x > 0 \/ x = 0`, "OR in predicate"},
		{`\A x \in S : x > 0 => x * x > 0`, "implies in predicate"},
		{`\A x \in S : IF x > 0 THEN x > -1 ELSE FALSE`, "IF in predicate"},
		{`\E x \in {1, 2, 3} : x \in {2, 3, 4}`, "membership in predicate"},
		{`\A x \in S : ~(x = 0)`, "negation in predicate"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := ParseExpression(tt.input)
			s.NoError(err, "should parse: %q (%s)", tt.input, tt.desc)

			_, ok := expr.(*ast.Quantifier)
			s.True(ok, "expected Quantifier, got %T", expr)
		})
	}
}

// TestQuantifierDomainExpressions tests various domain expressions in quantifiers.
func (s *StressQuantifierTestSuite) TestQuantifierDomainExpressions() {
	tests := []struct {
		input string
		desc  string
	}{
		{`\A x \in {1, 2, 3} : x > 0`, "set literal domain"},
		{`\A x \in 1..10 : x > 0`, "range domain"},
		{`\A x \in S \cup T : x > 0`, "union domain"},
		{`\A x \in S \cap T : x > 0`, "intersection domain"},
		{`\A x \in BOOLEAN : x \/ ~x`, "BOOLEAN set constant"},
		{`\A x \in {} : FALSE`, "empty set domain (vacuously true)"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := ParseExpression(tt.input)
			s.NoError(err, "should parse: %q (%s)", tt.input, tt.desc)

			quant, ok := expr.(*ast.Quantifier)
			s.True(ok, "expected Quantifier, got %T", expr)
			// Domain is inside the Membership node's Right field
			mem, ok := quant.Membership.(*ast.Membership)
			s.True(ok, "membership should be Membership node")
			s.NotNil(mem.Right, "domain (membership right) should not be nil")
		})
	}
}

// TestQuantifierUnicodeVariants tests Unicode (∀, ∃, ∈) vs ASCII (\A, \E, \in).
func (s *StressQuantifierTestSuite) TestQuantifierUnicodeVariants() {
	tests := []struct {
		input    string
		desc     string
		operator string // expected quantifier operator
	}{
		{`\A x \in S : x > 0`, "ASCII universal", "∀"},
		{`∀ x ∈ S : x > 0`, "Unicode universal", "∀"},
		{`\E x \in S : x > 0`, "ASCII existential", "∃"},
		{`∃ x ∈ S : x > 0`, "Unicode existential", "∃"},
		// Mixed
		{`\A x ∈ S : x > 0`, "ASCII quantifier + Unicode membership", "∀"},
		{`∀ x \in S : x > 0`, "Unicode quantifier + ASCII membership", "∀"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			expr, err := ParseExpression(tt.input)
			s.NoError(err, "should parse: %q (%s)", tt.input, tt.desc)

			quant, ok := expr.(*ast.Quantifier)
			s.True(ok, "expected Quantifier, got %T", expr)
			s.Equal(tt.operator, quant.Quantifier)
		})
	}
}

// TestQuantifierMalformed tests that malformed quantifiers are rejected.
func (s *StressQuantifierTestSuite) TestQuantifierMalformed() {
	tests := []struct {
		input string
		desc  string
	}{
		{`\A x \in S`, "missing colon and predicate"},
		{`\A x \in S :`, "missing predicate after colon"},
		{`\A x`, "missing domain and predicate"},
		{`\A \in S : TRUE`, "missing variable name"},
	}

	for _, tt := range tests {
		s.Run(tt.desc, func() {
			_, err := ParseExpression(tt.input)
			s.Error(err, "should fail: %q (%s)", tt.input, tt.desc)
		})
	}
}
