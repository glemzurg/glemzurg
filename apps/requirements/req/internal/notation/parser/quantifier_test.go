package parser

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/ast"
	"github.com/stretchr/testify/suite"
)

func TestQuantifierSuite(t *testing.T) {
	suite.Run(t, new(QuantifierSuite))
}

type QuantifierSuite struct {
	suite.Suite
}

// =============================================================================
// Universal Quantifier (ForAll)
// =============================================================================

func (s *QuantifierSuite) TestForAll_Unicode() {
	expr, err := ParseExpression("∀ x ∈ S : x > 0")
	s.NoError(err)

	q, ok := expr.(*ast.Quantifier)
	s.True(ok, "expected *ast.Quantifier, got %T", expr)
	s.Equal("∀", q.Quantifier)

	// Membership should bind x to S
	mem, ok := q.Membership.(*ast.Membership)
	s.True(ok, "expected *ast.Membership, got %T", q.Membership)
	s.Equal("∈", mem.Operator)

	ident, ok := mem.Left.(*ast.Identifier)
	s.True(ok, "expected *ast.Identifier, got %T", mem.Left)
	s.Equal("x", ident.Value)
}

func (s *QuantifierSuite) TestForAll_ASCII() {
	expr, err := ParseExpression("\\A x \\in S : x > 0")
	s.NoError(err)

	q, ok := expr.(*ast.Quantifier)
	s.True(ok, "expected *ast.Quantifier, got %T", expr)
	s.Equal("∀", q.Quantifier) // Normalized to Unicode
}

func (s *QuantifierSuite) TestForAll_WithSetLiteral() {
	expr, err := ParseExpression("∀ n ∈ {1, 2, 3} : n > 0")
	s.NoError(err)

	q, ok := expr.(*ast.Quantifier)
	s.True(ok, "expected *ast.Quantifier, got %T", expr)

	mem, ok := q.Membership.(*ast.Membership)
	s.True(ok, "expected *ast.Membership, got %T", q.Membership)

	_, ok = mem.Right.(*ast.SetLiteral)
	s.True(ok, "expected right to be SetLiteral")
}

func (s *QuantifierSuite) TestForAll_WithRange() {
	expr, err := ParseExpression("∀ i ∈ 1..10 : i >= 1")
	s.NoError(err)

	q, ok := expr.(*ast.Quantifier)
	s.True(ok, "expected *ast.Quantifier, got %T", expr)

	mem, ok := q.Membership.(*ast.Membership)
	s.True(ok, "expected *ast.Membership, got %T", q.Membership)

	_, ok = mem.Right.(*ast.SetRangeExpr)
	s.True(ok, "expected right to be SetRangeExpr")
}

// =============================================================================
// Existential Quantifier (Exists)
// =============================================================================

func (s *QuantifierSuite) TestExists_Unicode() {
	expr, err := ParseExpression("∃ x ∈ S : x = target")
	s.NoError(err)

	q, ok := expr.(*ast.Quantifier)
	s.True(ok, "expected *ast.Quantifier, got %T", expr)
	s.Equal("∃", q.Quantifier)
}

func (s *QuantifierSuite) TestExists_ASCII() {
	expr, err := ParseExpression("\\E y \\in T : y = value")
	s.NoError(err)

	q, ok := expr.(*ast.Quantifier)
	s.True(ok, "expected *ast.Quantifier, got %T", expr)
	s.Equal("∃", q.Quantifier) // Normalized to Unicode
}

func (s *QuantifierSuite) TestExists_WithSetLiteral() {
	expr, err := ParseExpression("∃ x ∈ {1, 2, 3} : x = 2")
	s.NoError(err)

	q, ok := expr.(*ast.Quantifier)
	s.True(ok, "expected *ast.Quantifier, got %T", expr)

	mem, ok := q.Membership.(*ast.Membership)
	s.True(ok, "expected *ast.Membership, got %T", q.Membership)

	_, ok = mem.Right.(*ast.SetLiteral)
	s.True(ok, "expected right to be SetLiteral")
}

// =============================================================================
// Complex Predicates
// =============================================================================

func (s *QuantifierSuite) TestForAll_ComplexPredicate() {
	// Predicate with AND
	expr, err := ParseExpression("∀ x ∈ S : x > 0 ∧ x < 100")
	s.NoError(err)

	q, ok := expr.(*ast.Quantifier)
	s.True(ok, "expected *ast.Quantifier, got %T", expr)

	_, ok = q.Predicate.(*ast.LogicInfixExpression)
	s.True(ok, "expected predicate to be LogicInfixExpression")
}

func (s *QuantifierSuite) TestForAll_ImpliesPredicate() {
	// Predicate with implies
	expr, err := ParseExpression("∀ x ∈ S : x > 0 => x >= 1")
	s.NoError(err)

	q, ok := expr.(*ast.Quantifier)
	s.True(ok, "expected *ast.Quantifier, got %T", expr)

	logic, ok := q.Predicate.(*ast.LogicInfixExpression)
	s.True(ok, "expected predicate to be LogicInfixExpression")
	s.Equal("⇒", logic.Operator)
}

// =============================================================================
// Nested Quantifiers
// =============================================================================

func (s *QuantifierSuite) TestNested_ForAllExists() {
	// ∀ x ∈ A : ∃ y ∈ B : x < y
	expr, err := ParseExpression("∀ x ∈ A : ∃ y ∈ B : x < y")
	s.NoError(err)

	outer, ok := expr.(*ast.Quantifier)
	s.True(ok, "expected outer *ast.Quantifier, got %T", expr)
	s.Equal("∀", outer.Quantifier)

	inner, ok := outer.Predicate.(*ast.Quantifier)
	s.True(ok, "expected inner *ast.Quantifier, got %T", outer.Predicate)
	s.Equal("∃", inner.Quantifier)
}

func (s *QuantifierSuite) TestNested_ExistsForAll() {
	// ∃ x ∈ A : ∀ y ∈ B : x >= y
	expr, err := ParseExpression("∃ x ∈ A : ∀ y ∈ B : x >= y")
	s.NoError(err)

	outer, ok := expr.(*ast.Quantifier)
	s.True(ok, "expected outer *ast.Quantifier, got %T", expr)
	s.Equal("∃", outer.Quantifier)

	inner, ok := outer.Predicate.(*ast.Quantifier)
	s.True(ok, "expected inner *ast.Quantifier, got %T", outer.Predicate)
	s.Equal("∀", inner.Quantifier)
}

// =============================================================================
// Precedence Tests
// =============================================================================

func (s *QuantifierSuite) TestPrecedence_QuantifierWithLogic() {
	// ∀ x ∈ S : P(x) /\ Q(x) should have /\ inside the predicate
	expr, err := ParseExpression("∀ x ∈ S : P ∧ Q")
	s.NoError(err)

	q, ok := expr.(*ast.Quantifier)
	s.True(ok, "expected *ast.Quantifier, got %T", expr)

	// The predicate should be the entire P /\ Q expression
	_, ok = q.Predicate.(*ast.LogicInfixExpression)
	s.True(ok, "expected predicate to be LogicInfixExpression")
}

func (s *QuantifierSuite) TestPrecedence_QuantifierWithImplies() {
	// ∀ x ∈ S : a => b should have => inside the predicate
	expr, err := ParseExpression("∀ x ∈ S : a => b")
	s.NoError(err)

	q, ok := expr.(*ast.Quantifier)
	s.True(ok, "expected *ast.Quantifier, got %T", expr)

	logic, ok := q.Predicate.(*ast.LogicInfixExpression)
	s.True(ok, "expected predicate to be LogicInfixExpression")
	s.Equal("⇒", logic.Operator)
}

// =============================================================================
// Mixed ASCII and Unicode
// =============================================================================

func (s *QuantifierSuite) TestMixed_ASCIIQuantifierUnicodeMembership() {
	expr, err := ParseExpression("\\A x ∈ S : x > 0")
	s.NoError(err)

	q, ok := expr.(*ast.Quantifier)
	s.True(ok, "expected *ast.Quantifier, got %T", expr)
	s.Equal("∀", q.Quantifier) // Normalized to Unicode

	mem, ok := q.Membership.(*ast.Membership)
	s.True(ok, "expected *ast.Membership, got %T", q.Membership)
	s.Equal("∈", mem.Operator)
}

func (s *QuantifierSuite) TestMixed_UnicodeQuantifierASCIIMembership() {
	expr, err := ParseExpression("∃ x \\in S : x > 0")
	s.NoError(err)

	q, ok := expr.(*ast.Quantifier)
	s.True(ok, "expected *ast.Quantifier, got %T", expr)
	s.Equal("∃", q.Quantifier)

	mem, ok := q.Membership.(*ast.Membership)
	s.True(ok, "expected *ast.Membership, got %T", q.Membership)
	s.Equal("∈", mem.Operator) // Normalized to Unicode
}
