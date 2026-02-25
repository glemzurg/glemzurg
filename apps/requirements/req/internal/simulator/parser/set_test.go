package parser

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/ast"
	"github.com/stretchr/testify/suite"
)

func TestSetSuite(t *testing.T) {
	suite.Run(t, new(SetSuite))
}

type SetSuite struct {
	suite.Suite
}

// =============================================================================
// Set Literals
// =============================================================================

func (s *SetSuite) TestSetLiteral_Empty() {
	expr, err := ParseExpression("{}")
	s.NoError(err)

	set, ok := expr.(*ast.SetLiteral)
	s.True(ok, "expected *ast.SetLiteral, got %T", expr)
	s.Len(set.Elements, 0)
}

func (s *SetSuite) TestSetLiteral_SingleElement() {
	expr, err := ParseExpression("{1}")
	s.NoError(err)

	set, ok := expr.(*ast.SetLiteral)
	s.True(ok, "expected *ast.SetLiteral, got %T", expr)
	s.Len(set.Elements, 1)
}

func (s *SetSuite) TestSetLiteral_MultipleElements() {
	expr, err := ParseExpression("{1, 2, 3}")
	s.NoError(err)

	set, ok := expr.(*ast.SetLiteral)
	s.True(ok, "expected *ast.SetLiteral, got %T", expr)
	s.Len(set.Elements, 3)
}

func (s *SetSuite) TestSetLiteral_WithExpressions() {
	expr, err := ParseExpression("{x, y + 1, z * 2}")
	s.NoError(err)

	set, ok := expr.(*ast.SetLiteral)
	s.True(ok, "expected *ast.SetLiteral, got %T", expr)
	s.Len(set.Elements, 3)

	// First element should be identifier
	_, ok = set.Elements[0].(*ast.Identifier)
	s.True(ok, "expected first element to be identifier")

	// Second element should be arithmetic expression
	_, ok = set.Elements[1].(*ast.BinaryArithmetic)
	s.True(ok, "expected second element to be arithmetic")
}

func (s *SetSuite) TestSetLiteral_Nested() {
	expr, err := ParseExpression("{{1, 2}, {3, 4}}")
	s.NoError(err)

	set, ok := expr.(*ast.SetLiteral)
	s.True(ok, "expected *ast.SetLiteral, got %T", expr)
	s.Len(set.Elements, 2)

	// Each element should be a set
	_, ok = set.Elements[0].(*ast.SetLiteral)
	s.True(ok, "expected first element to be SetLiteral")
}

func (s *SetSuite) TestSetLiteral_String() {
	expr, err := ParseExpression("{1, 2, 3}")
	s.NoError(err)
	s.Equal("{1, 2, 3}", expr.String())
}

// =============================================================================
// Set Range
// =============================================================================

func (s *SetSuite) TestSetRange_Simple() {
	expr, err := ParseExpression("1..10")
	s.NoError(err)

	rng, ok := expr.(*ast.SetRangeExpr)
	s.True(ok, "expected *ast.SetRangeExpr, got %T", expr)

	start, ok := rng.Start.(*ast.NumberLiteral)
	s.True(ok, "expected start to be NumberLiteral")
	s.Equal("1", start.String())

	end, ok := rng.End.(*ast.NumberLiteral)
	s.True(ok, "expected end to be NumberLiteral")
	s.Equal("10", end.String())
}

func (s *SetSuite) TestSetRange_WithVariables() {
	expr, err := ParseExpression("x..y")
	s.NoError(err)

	rng, ok := expr.(*ast.SetRangeExpr)
	s.True(ok, "expected *ast.SetRangeExpr, got %T", expr)

	_, ok = rng.Start.(*ast.Identifier)
	s.True(ok, "expected start to be Identifier")

	_, ok = rng.End.(*ast.Identifier)
	s.True(ok, "expected end to be Identifier")
}

func (s *SetSuite) TestSetRange_WithExpressions() {
	expr, err := ParseExpression("(n-1)..(n+1)")
	s.NoError(err)

	_, ok := expr.(*ast.SetRangeExpr)
	s.True(ok, "expected *ast.SetRangeExpr, got %T", expr)
}

func (s *SetSuite) TestSetRange_String() {
	expr, err := ParseExpression("1..10")
	s.NoError(err)
	s.Equal("1 .. 10", expr.String())
}

// =============================================================================
// Set Membership
// =============================================================================

func (s *SetSuite) TestSetMembership_In_Unicode() {
	expr, err := ParseExpression("x ∈ S")
	s.NoError(err)

	mem, ok := expr.(*ast.Membership)
	s.True(ok, "expected *ast.Membership, got %T", expr)
	s.Equal("∈", mem.Operator)
}

func (s *SetSuite) TestSetMembership_In_ASCII() {
	expr, err := ParseExpression("x \\in S")
	s.NoError(err)

	mem, ok := expr.(*ast.Membership)
	s.True(ok, "expected *ast.Membership, got %T", expr)
	s.Equal("∈", mem.Operator) // Normalized to Unicode
}

func (s *SetSuite) TestSetMembership_NotIn_Unicode() {
	expr, err := ParseExpression("x ∉ S")
	s.NoError(err)

	mem, ok := expr.(*ast.Membership)
	s.True(ok, "expected *ast.Membership, got %T", expr)
	s.Equal("∉", mem.Operator)
}

func (s *SetSuite) TestSetMembership_NotIn_ASCII() {
	expr, err := ParseExpression("x \\notin S")
	s.NoError(err)

	mem, ok := expr.(*ast.Membership)
	s.True(ok, "expected *ast.Membership, got %T", expr)
	s.Equal("∉", mem.Operator)
}

func (s *SetSuite) TestSetMembership_WithLiterals() {
	expr, err := ParseExpression("1 ∈ {1, 2, 3}")
	s.NoError(err)

	mem, ok := expr.(*ast.Membership)
	s.True(ok, "expected *ast.Membership, got %T", expr)
	s.Equal("∈", mem.Operator)

	_, ok = mem.Right.(*ast.SetLiteral)
	s.True(ok, "expected right to be SetLiteral")
}

// =============================================================================
// Set Operations
// =============================================================================

func (s *SetSuite) TestSetUnion_Unicode() {
	expr, err := ParseExpression("A ∪ B")
	s.NoError(err)

	op, ok := expr.(*ast.BinarySetOperation)
	s.True(ok, "expected *ast.BinarySetOperation, got %T", expr)
	s.Equal("∪", op.Operator)
}

func (s *SetSuite) TestSetUnion_ASCII_union() {
	expr, err := ParseExpression("A \\union B")
	s.NoError(err)

	op, ok := expr.(*ast.BinarySetOperation)
	s.True(ok, "expected *ast.BinarySetOperation, got %T", expr)
	s.Equal("∪", op.Operator)
}

func (s *SetSuite) TestSetUnion_ASCII_cup() {
	expr, err := ParseExpression("A \\cup B")
	s.NoError(err)

	op, ok := expr.(*ast.BinarySetOperation)
	s.True(ok, "expected *ast.BinarySetOperation, got %T", expr)
	s.Equal("∪", op.Operator)
}

func (s *SetSuite) TestSetIntersection_Unicode() {
	expr, err := ParseExpression("A ∩ B")
	s.NoError(err)

	op, ok := expr.(*ast.BinarySetOperation)
	s.True(ok, "expected *ast.BinarySetOperation, got %T", expr)
	s.Equal("∩", op.Operator)
}

func (s *SetSuite) TestSetIntersection_ASCII_intersect() {
	expr, err := ParseExpression("A \\intersect B")
	s.NoError(err)

	op, ok := expr.(*ast.BinarySetOperation)
	s.True(ok, "expected *ast.BinarySetOperation, got %T", expr)
	s.Equal("∩", op.Operator)
}

func (s *SetSuite) TestSetIntersection_ASCII_cap() {
	expr, err := ParseExpression("A \\cap B")
	s.NoError(err)

	op, ok := expr.(*ast.BinarySetOperation)
	s.True(ok, "expected *ast.BinarySetOperation, got %T", expr)
	s.Equal("∩", op.Operator)
}

func (s *SetSuite) TestSetDifference() {
	expr, err := ParseExpression("A \\ B")
	s.NoError(err)

	op, ok := expr.(*ast.BinarySetOperation)
	s.True(ok, "expected *ast.BinarySetOperation, got %T", expr)
	s.Equal(`\`, op.Operator)
}

func (s *SetSuite) TestSetOperations_Chained() {
	// A ∪ B ∪ C should parse as (A ∪ B) ∪ C
	expr, err := ParseExpression("A ∪ B ∪ C")
	s.NoError(err)

	outer, ok := expr.(*ast.BinarySetOperation)
	s.True(ok, "expected outer *ast.BinarySetOperation, got %T", expr)
	s.Equal("∪", outer.Operator)

	inner, ok := outer.Left.(*ast.BinarySetOperation)
	s.True(ok, "expected inner *ast.BinarySetOperation, got %T", outer.Left)
	s.Equal("∪", inner.Operator)
}

// =============================================================================
// Set Comparisons
// =============================================================================

func (s *SetSuite) TestSetComparison_SubsetEq_Unicode() {
	expr, err := ParseExpression("A ⊆ B")
	s.NoError(err)

	cmp, ok := expr.(*ast.BinarySetComparison)
	s.True(ok, "expected *ast.BinarySetComparison, got %T", expr)
	s.Equal("⊆", cmp.Operator)
}

func (s *SetSuite) TestSetComparison_SubsetEq_ASCII() {
	expr, err := ParseExpression("A \\subseteq B")
	s.NoError(err)

	cmp, ok := expr.(*ast.BinarySetComparison)
	s.True(ok, "expected *ast.BinarySetComparison, got %T", expr)
	s.Equal("⊆", cmp.Operator)
}

func (s *SetSuite) TestSetComparison_SupersetEq_Unicode() {
	expr, err := ParseExpression("A ⊇ B")
	s.NoError(err)

	cmp, ok := expr.(*ast.BinarySetComparison)
	s.True(ok, "expected *ast.BinarySetComparison, got %T", expr)
	s.Equal("⊇", cmp.Operator)
}

func (s *SetSuite) TestSetComparison_SupersetEq_ASCII() {
	expr, err := ParseExpression("A \\supseteq B")
	s.NoError(err)

	cmp, ok := expr.(*ast.BinarySetComparison)
	s.True(ok, "expected *ast.BinarySetComparison, got %T", expr)
	s.Equal("⊇", cmp.Operator)
}

func (s *SetSuite) TestSetComparison_Subset_Unicode() {
	expr, err := ParseExpression("A ⊂ B")
	s.NoError(err)

	cmp, ok := expr.(*ast.BinarySetComparison)
	s.True(ok, "expected *ast.BinarySetComparison, got %T", expr)
	s.Equal("⊂", cmp.Operator)
}

func (s *SetSuite) TestSetComparison_Subset_ASCII() {
	expr, err := ParseExpression("A \\subset B")
	s.NoError(err)

	cmp, ok := expr.(*ast.BinarySetComparison)
	s.True(ok, "expected *ast.BinarySetComparison, got %T", expr)
	s.Equal("⊂", cmp.Operator)
}

func (s *SetSuite) TestSetComparison_Superset_Unicode() {
	expr, err := ParseExpression("A ⊃ B")
	s.NoError(err)

	cmp, ok := expr.(*ast.BinarySetComparison)
	s.True(ok, "expected *ast.BinarySetComparison, got %T", expr)
	s.Equal("⊃", cmp.Operator)
}

func (s *SetSuite) TestSetComparison_Superset_ASCII() {
	expr, err := ParseExpression("A \\supset B")
	s.NoError(err)

	cmp, ok := expr.(*ast.BinarySetComparison)
	s.True(ok, "expected *ast.BinarySetComparison, got %T", expr)
	s.Equal("⊃", cmp.Operator)
}

// =============================================================================
// Precedence Tests
// =============================================================================

func (s *SetSuite) TestPrecedence_RangeHigherThanUnion() {
	// 1..5 ∪ 6..10 should parse as (1..5) ∪ (6..10)
	expr, err := ParseExpression("1..5 ∪ 6..10")
	s.NoError(err)

	op, ok := expr.(*ast.BinarySetOperation)
	s.True(ok, "expected *ast.BinarySetOperation, got %T", expr)
	s.Equal("∪", op.Operator)

	_, ok = op.Left.(*ast.SetRangeExpr)
	s.True(ok, "expected left to be SetRangeExpr")

	_, ok = op.Right.(*ast.SetRangeExpr)
	s.True(ok, "expected right to be SetRangeExpr")
}

func (s *SetSuite) TestPrecedence_UnionLowerThanMembership() {
	// x ∈ A ∪ B should parse as x ∈ (A ∪ B)
	expr, err := ParseExpression("x ∈ A ∪ B")
	s.NoError(err)

	mem, ok := expr.(*ast.Membership)
	s.True(ok, "expected *ast.Membership, got %T", expr)

	_, ok = mem.Right.(*ast.BinarySetOperation)
	s.True(ok, "expected right to be BinarySetOperation")
}

func (s *SetSuite) TestPrecedence_SetComparisonLowerThanMembership() {
	// x ∈ A ⊆ B should parse as (x ∈ A) ⊆ B - but this is actually weird
	// Let's test A ⊆ B ∪ C which should be A ⊆ (B ∪ C)
	expr, err := ParseExpression("A ⊆ B ∪ C")
	s.NoError(err)

	cmp, ok := expr.(*ast.BinarySetComparison)
	s.True(ok, "expected *ast.BinarySetComparison, got %T", expr)

	_, ok = cmp.Right.(*ast.BinarySetOperation)
	s.True(ok, "expected right to be BinarySetOperation")
}

func (s *SetSuite) TestPrecedence_ArithmeticHigherThanRange() {
	// 1+2..3+4 should parse as (1+2)..(3+4)
	expr, err := ParseExpression("1+2..3+4")
	s.NoError(err)

	rng, ok := expr.(*ast.SetRangeExpr)
	s.True(ok, "expected *ast.SetRangeExpr, got %T", expr)

	_, ok = rng.Start.(*ast.BinaryArithmetic)
	s.True(ok, "expected start to be BinaryArithmetic")

	_, ok = rng.End.(*ast.BinaryArithmetic)
	s.True(ok, "expected end to be BinaryArithmetic")
}
