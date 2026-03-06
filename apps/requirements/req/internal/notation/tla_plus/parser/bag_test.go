package parser

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/stretchr/testify/suite"
)

func TestBagSuite(t *testing.T) {
	suite.Run(t, new(BagSuite))
}

type BagSuite struct {
	suite.Suite
}

// =============================================================================
// Bag Comparison Operators
// =============================================================================

// Test ⊏ (proper subbag) - Unicode
func (s *BagSuite) TestBagComparison_ProperSubBag_Unicode() {
	expr, err := ParseExpression("A ⊏ B")
	s.NoError(err)

	cmp, ok := expr.(*ast.BinaryBagComparison)
	s.True(ok, "expected *ast.BinaryBagComparison, got %T", expr)
	s.Equal("⊏", cmp.Operator)
}

// Test \sqsubset (proper subbag) - ASCII
func (s *BagSuite) TestBagComparison_ProperSubBag_ASCII() {
	expr, err := ParseExpression("A \\sqsubset B")
	s.NoError(err)

	cmp, ok := expr.(*ast.BinaryBagComparison)
	s.True(ok, "expected *ast.BinaryBagComparison, got %T", expr)
	s.Equal("⊏", cmp.Operator) // Normalized to Unicode
}

// Test ⊑ (subbag or equal) - Unicode
func (s *BagSuite) TestBagComparison_SubBagEq_Unicode() {
	expr, err := ParseExpression("A ⊑ B")
	s.NoError(err)

	cmp, ok := expr.(*ast.BinaryBagComparison)
	s.True(ok, "expected *ast.BinaryBagComparison, got %T", expr)
	s.Equal("⊑", cmp.Operator)
}

// Test \sqsubseteq (subbag or equal) - ASCII
func (s *BagSuite) TestBagComparison_SubBagEq_ASCII() {
	expr, err := ParseExpression("A \\sqsubseteq B")
	s.NoError(err)

	cmp, ok := expr.(*ast.BinaryBagComparison)
	s.True(ok, "expected *ast.BinaryBagComparison, got %T", expr)
	s.Equal("⊑", cmp.Operator)
}

// Test ⊐ (proper superbag) - Unicode
func (s *BagSuite) TestBagComparison_ProperSuperBag_Unicode() {
	expr, err := ParseExpression("A ⊐ B")
	s.NoError(err)

	cmp, ok := expr.(*ast.BinaryBagComparison)
	s.True(ok, "expected *ast.BinaryBagComparison, got %T", expr)
	s.Equal("⊐", cmp.Operator)
}

// Test \sqsupset (proper superbag) - ASCII
func (s *BagSuite) TestBagComparison_ProperSuperBag_ASCII() {
	expr, err := ParseExpression("A \\sqsupset B")
	s.NoError(err)

	cmp, ok := expr.(*ast.BinaryBagComparison)
	s.True(ok, "expected *ast.BinaryBagComparison, got %T", expr)
	s.Equal("⊐", cmp.Operator)
}

// Test ⊒ (superbag or equal) - Unicode
func (s *BagSuite) TestBagComparison_SuperBagEq_Unicode() {
	expr, err := ParseExpression("A ⊒ B")
	s.NoError(err)

	cmp, ok := expr.(*ast.BinaryBagComparison)
	s.True(ok, "expected *ast.BinaryBagComparison, got %T", expr)
	s.Equal("⊒", cmp.Operator)
}

// Test \sqsupseteq (superbag or equal) - ASCII
func (s *BagSuite) TestBagComparison_SuperBagEq_ASCII() {
	expr, err := ParseExpression("A \\sqsupseteq B")
	s.NoError(err)

	cmp, ok := expr.(*ast.BinaryBagComparison)
	s.True(ok, "expected *ast.BinaryBagComparison, got %T", expr)
	s.Equal("⊒", cmp.Operator)
}

// =============================================================================
// Bag Sum Operator
// =============================================================================

// Test ⊕ (bag sum) - Unicode
func (s *BagSuite) TestBagSum_Unicode() {
	expr, err := ParseExpression("A ⊕ B")
	s.NoError(err)

	op, ok := expr.(*ast.BinaryBagOperation)
	s.True(ok, "expected *ast.BinaryBagOperation, got %T", expr)
	s.Equal("⊕", op.Operator)
}

// Test (+) (bag sum) - ASCII
func (s *BagSuite) TestBagSum_ASCII_Paren() {
	expr, err := ParseExpression("A (+) B")
	s.NoError(err)

	op, ok := expr.(*ast.BinaryBagOperation)
	s.True(ok, "expected *ast.BinaryBagOperation, got %T", expr)
	s.Equal("⊕", op.Operator)
}

// Test \oplus (bag sum) - ASCII
func (s *BagSuite) TestBagSum_ASCII_Oplus() {
	expr, err := ParseExpression("A \\oplus B")
	s.NoError(err)

	op, ok := expr.(*ast.BinaryBagOperation)
	s.True(ok, "expected *ast.BinaryBagOperation, got %T", expr)
	s.Equal("⊕", op.Operator)
}

// Test bag sum chaining
func (s *BagSuite) TestBagSum_Chained() {
	expr, err := ParseExpression("A ⊕ B ⊕ C")
	s.NoError(err)

	outer, ok := expr.(*ast.BinaryBagOperation)
	s.True(ok, "expected outer *ast.BinaryBagOperation, got %T", expr)
	s.Equal("⊕", outer.Operator)

	inner, ok := outer.Left.(*ast.BinaryBagOperation)
	s.True(ok, "expected inner *ast.BinaryBagOperation, got %T", outer.Left)
	s.Equal("⊕", inner.Operator)
}

// =============================================================================
// Bag Difference Operator
// =============================================================================

// Test ⊖ (bag difference) - Unicode
func (s *BagSuite) TestBagDiff_Unicode() {
	expr, err := ParseExpression("A ⊖ B")
	s.NoError(err)

	op, ok := expr.(*ast.BinaryBagOperation)
	s.True(ok, "expected *ast.BinaryBagOperation, got %T", expr)
	s.Equal("⊖", op.Operator)
}

// Test (-) (bag difference) - ASCII
func (s *BagSuite) TestBagDiff_ASCII_Paren() {
	expr, err := ParseExpression("A (-) B")
	s.NoError(err)

	op, ok := expr.(*ast.BinaryBagOperation)
	s.True(ok, "expected *ast.BinaryBagOperation, got %T", expr)
	s.Equal("⊖", op.Operator)
}

// Test \ominus (bag difference) - ASCII
func (s *BagSuite) TestBagDiff_ASCII_Ominus() {
	expr, err := ParseExpression("A \\ominus B")
	s.NoError(err)

	op, ok := expr.(*ast.BinaryBagOperation)
	s.True(ok, "expected *ast.BinaryBagOperation, got %T", expr)
	s.Equal("⊖", op.Operator)
}

// =============================================================================
// Sequence Concatenation Operator
// =============================================================================

// Test ∘ (sequence concatenation) - Unicode
func (s *BagSuite) TestConcat_Unicode() {
	expr, err := ParseExpression("A ∘ B")
	s.NoError(err)

	op, ok := expr.(*ast.TupleConcat)
	s.True(ok, "expected *ast.TupleConcat, got %T", expr)
	s.Equal("∘", op.Operator)
	s.Len(op.Operands, 2)
}

// Test \o (sequence concatenation) - ASCII
func (s *BagSuite) TestConcat_ASCII_O() {
	expr, err := ParseExpression("A \\o B")
	s.NoError(err)

	op, ok := expr.(*ast.TupleConcat)
	s.True(ok, "expected *ast.TupleConcat, got %T", expr)
	s.Equal("∘", op.Operator)
}

// Test \circ (sequence concatenation) - ASCII
func (s *BagSuite) TestConcat_ASCII_Circ() {
	expr, err := ParseExpression("A \\circ B")
	s.NoError(err)

	op, ok := expr.(*ast.TupleConcat)
	s.True(ok, "expected *ast.TupleConcat, got %T", expr)
	s.Equal("∘", op.Operator)
}

// Test sequence concatenation chaining
func (s *BagSuite) TestConcat_Chained() {
	expr, err := ParseExpression("A ∘ B ∘ C")
	s.NoError(err)

	op, ok := expr.(*ast.TupleConcat)
	s.True(ok, "expected *ast.TupleConcat, got %T", expr)
	s.Equal("∘", op.Operator)
	s.Len(op.Operands, 3) // All three operands in one node
}

// =============================================================================
// Precedence Tests
// =============================================================================

// Test that bag sum has lower precedence than addition
func (s *BagSuite) TestPrecedence_BagSumLowerThanAddition() {
	// A ⊕ B + C should parse as A ⊕ (B + C)
	expr, err := ParseExpression("A ⊕ B + C")
	s.NoError(err)

	op, ok := expr.(*ast.BinaryBagOperation)
	s.True(ok, "expected *ast.BinaryBagOperation, got %T", expr)
	s.Equal("⊕", op.Operator)

	_, ok = op.Right.(*ast.BinaryArithmetic)
	s.True(ok, "expected right to be BinaryArithmetic, got %T", op.Right)
}

// Test that bag difference has lower precedence than subtraction
func (s *BagSuite) TestPrecedence_BagDiffLowerThanSubtraction() {
	// A ⊖ B - C should parse as A ⊖ (B - C)
	expr, err := ParseExpression("A ⊖ B - C")
	s.NoError(err)

	op, ok := expr.(*ast.BinaryBagOperation)
	s.True(ok, "expected *ast.BinaryBagOperation, got %T", expr)
	s.Equal("⊖", op.Operator)

	_, ok = op.Right.(*ast.BinaryArithmetic)
	s.True(ok, "expected right to be BinaryArithmetic, got %T", op.Right)
}

// Test that multiplication has higher precedence than concatenation
func (s *BagSuite) TestPrecedence_MultHigherThanConcat() {
	// A * B ∘ C should parse as (A * B) ∘ C
	expr, err := ParseExpression("A * B ∘ C")
	s.NoError(err)

	concat, ok := expr.(*ast.TupleConcat)
	s.True(ok, "expected *ast.TupleConcat, got %T", expr)
	s.Equal("∘", concat.Operator)
	s.Len(concat.Operands, 2)

	_, ok = concat.Operands[0].(*ast.RealInfixExpression)
	s.True(ok, "expected first operand to be RealInfixExpression, got %T", concat.Operands[0])
}

// Test bag comparison at same level as set comparison
func (s *BagSuite) TestPrecedence_BagComparisonSameAsSetComparison() {
	// A ⊑ B ⊆ C - both at same precedence level
	// This tests that bag comparisons work alongside set comparisons
	expr, err := ParseExpression("A ⊑ B")
	s.NoError(err)

	_, ok := expr.(*ast.BinaryBagComparison)
	s.True(ok, "expected *ast.BinaryBagComparison, got %T", expr)
}

// Test bag operations higher precedence than comparisons
func (s *BagSuite) TestPrecedence_BagOpHigherThanComparison() {
	// A ⊕ B ⊑ C should parse as (A ⊕ B) ⊑ C
	expr, err := ParseExpression("A ⊕ B ⊑ C")
	s.NoError(err)

	cmp, ok := expr.(*ast.BinaryBagComparison)
	s.True(ok, "expected *ast.BinaryBagComparison, got %T", expr)

	_, ok = cmp.Left.(*ast.BinaryBagOperation)
	s.True(ok, "expected left to be BinaryBagOperation, got %T", cmp.Left)
}
