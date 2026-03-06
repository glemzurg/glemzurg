package parser

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/stretchr/testify/suite"
)

func TestComparisonSuite(t *testing.T) {
	suite.Run(t, new(ComparisonSuite))
}

type ComparisonSuite struct {
	suite.Suite
}

// =============================================================================
// Equality (=)
// =============================================================================

func (s *ComparisonSuite) TestParseEqual() {
	expr, err := ParseExpression("1 = 2")
	s.NoError(err)

	eq := expr.(*ast.LogicEquality)
	s.Equal("=", eq.Operator)

	left := eq.Left.(*ast.NumberLiteral)
	s.Equal("1", left.IntegerPart)

	right := eq.Right.(*ast.NumberLiteral)
	s.Equal("2", right.IntegerPart)
}

func (s *ComparisonSuite) TestParseEqualBooleans() {
	expr, err := ParseExpression("TRUE = FALSE")
	s.NoError(err)

	eq := expr.(*ast.LogicEquality)
	s.Equal("=", eq.Operator)

	left := eq.Left.(*ast.BooleanLiteral)
	s.True(left.Value)

	right := eq.Right.(*ast.BooleanLiteral)
	s.False(right.Value)
}

func (s *ComparisonSuite) TestParseEqualStrings() {
	expr, err := ParseExpression(`"hello" = "world"`)
	s.NoError(err)

	eq := expr.(*ast.LogicEquality)
	s.Equal("=", eq.Operator)

	left := eq.Left.(*ast.StringLiteral)
	s.Equal("hello", left.Value)

	right := eq.Right.(*ast.StringLiteral)
	s.Equal("world", right.Value)
}

// =============================================================================
// Not Equal (≠, /=, #)
// =============================================================================

func (s *ComparisonSuite) TestParseNotEqualUnicode() {
	expr, err := ParseExpression("1 ≠ 2")
	s.NoError(err)

	neq := expr.(*ast.LogicEquality)
	s.Equal("≠", neq.Operator)
}

func (s *ComparisonSuite) TestParseNotEqualSlash() {
	expr, err := ParseExpression("1 /= 2")
	s.NoError(err)

	neq := expr.(*ast.LogicEquality)
	s.Equal("≠", neq.Operator) // Normalized to Unicode
}

func (s *ComparisonSuite) TestParseNotEqualHash() {
	expr, err := ParseExpression("1 # 2")
	s.NoError(err)

	neq := expr.(*ast.LogicEquality)
	s.Equal("≠", neq.Operator) // Normalized to Unicode
}

// =============================================================================
// Less Than (<)
// =============================================================================

func (s *ComparisonSuite) TestParseLessThan() {
	expr, err := ParseExpression("1 < 2")
	s.NoError(err)

	lt := expr.(*ast.LogicRealComparison)
	s.Equal("<", lt.Operator)
}

// =============================================================================
// Greater Than (>)
// =============================================================================

func (s *ComparisonSuite) TestParseGreaterThan() {
	expr, err := ParseExpression("5 > 3")
	s.NoError(err)

	gt := expr.(*ast.LogicRealComparison)
	s.Equal(">", gt.Operator)
}

// =============================================================================
// Less Than or Equal (≤, =<, <=)
// =============================================================================

func (s *ComparisonSuite) TestParseLessOrEqualUnicode() {
	expr, err := ParseExpression("1 ≤ 2")
	s.NoError(err)

	le := expr.(*ast.LogicRealComparison)
	s.Equal("≤", le.Operator)
}

func (s *ComparisonSuite) TestParseLessOrEqualAscii() {
	expr, err := ParseExpression("1 =< 2")
	s.NoError(err)

	le := expr.(*ast.LogicRealComparison)
	s.Equal("≤", le.Operator) // Normalized to Unicode
}

func (s *ComparisonSuite) TestParseLessOrEqualAlternate() {
	expr, err := ParseExpression("1 <= 2")
	s.NoError(err)

	le := expr.(*ast.LogicRealComparison)
	s.Equal("≤", le.Operator) // Normalized to Unicode
}

// =============================================================================
// Greater Than or Equal (≥, >=)
// =============================================================================

func (s *ComparisonSuite) TestParseGreaterOrEqualUnicode() {
	expr, err := ParseExpression("5 ≥ 3")
	s.NoError(err)

	ge := expr.(*ast.LogicRealComparison)
	s.Equal("≥", ge.Operator)
}

func (s *ComparisonSuite) TestParseGreaterOrEqualAscii() {
	expr, err := ParseExpression("5 >= 3")
	s.NoError(err)

	ge := expr.(*ast.LogicRealComparison)
	s.Equal("≥", ge.Operator) // Normalized to Unicode
}

// =============================================================================
// Precedence: Comparisons > Logic
// =============================================================================

func (s *ComparisonSuite) TestPrecedenceComparisonOverAnd() {
	// x > 5 /\ y < 10 = (x > 5) /\ (y < 10)
	expr, err := ParseExpression("1 > 0 /\\ 2 < 3")
	s.NoError(err)

	and := expr.(*ast.LogicInfixExpression)
	s.Equal("∧", and.Operator)

	// Left is (1 > 0)
	left := and.Left.(*ast.LogicRealComparison)
	s.Equal(">", left.Operator)

	// Right is (2 < 3)
	right := and.Right.(*ast.LogicRealComparison)
	s.Equal("<", right.Operator)
}

func (s *ComparisonSuite) TestPrecedenceComparisonOverOr() {
	// x = 1 \/ y = 2 = (x = 1) \/ (y = 2)
	expr, err := ParseExpression("1 = 1 \\/ 2 = 3")
	s.NoError(err)

	or := expr.(*ast.LogicInfixExpression)
	s.Equal("∨", or.Operator)

	// Left is (1 = 1)
	left := or.Left.(*ast.LogicEquality)
	s.Equal("=", left.Operator)

	// Right is (2 = 3)
	right := or.Right.(*ast.LogicEquality)
	s.Equal("=", right.Operator)
}

func (s *ComparisonSuite) TestPrecedenceComparisonOverImplies() {
	// x < y => a > b = (x < y) => (a > b)
	expr, err := ParseExpression("1 < 2 => 3 > 0")
	s.NoError(err)

	implies := expr.(*ast.LogicInfixExpression)
	s.Equal("⇒", implies.Operator)

	// Left is (1 < 2)
	left := implies.Left.(*ast.LogicRealComparison)
	s.Equal("<", left.Operator)

	// Right is (3 > 0)
	right := implies.Right.(*ast.LogicRealComparison)
	s.Equal(">", right.Operator)
}

// =============================================================================
// Precedence: Arithmetic > Comparisons
// =============================================================================

func (s *ComparisonSuite) TestPrecedenceArithmeticOverComparison() {
	// 1 + 2 < 5 = (1 + 2) < 5
	expr, err := ParseExpression("1 + 2 < 5")
	s.NoError(err)

	lt := expr.(*ast.LogicRealComparison)
	s.Equal("<", lt.Operator)

	// Left is (1 + 2)
	left := lt.Left.(*ast.RealInfixExpression)
	s.Equal("+", left.Operator)

	// Right is 5
	right := lt.Right.(*ast.NumberLiteral)
	s.Equal("5", right.IntegerPart)
}

func (s *ComparisonSuite) TestPrecedenceArithmeticBothSides() {
	// 2 * 3 = 3 + 3 = (2 * 3) = (3 + 3)
	expr, err := ParseExpression("2 * 3 = 3 + 3")
	s.NoError(err)

	eq := expr.(*ast.LogicEquality)
	s.Equal("=", eq.Operator)

	// Left is (2 * 3)
	left := eq.Left.(*ast.RealInfixExpression)
	s.Equal("*", left.Operator)

	// Right is (3 + 3)
	right := eq.Right.(*ast.RealInfixExpression)
	s.Equal("+", right.Operator)
}

func (s *ComparisonSuite) TestPrecedencePowerOverComparison() {
	// 2 ^ 3 > 5 = (2 ^ 3) > 5
	expr, err := ParseExpression("2 ^ 3 > 5")
	s.NoError(err)

	gt := expr.(*ast.LogicRealComparison)
	s.Equal(">", gt.Operator)

	// Left is (2 ^ 3)
	left := gt.Left.(*ast.RealInfixExpression)
	s.Equal("^", left.Operator)
}

// =============================================================================
// Complex combined expressions
// =============================================================================

func (s *ComparisonSuite) TestComplexExpression() {
	// 1 + 2 = 3 /\ 4 * 2 > 5 => TRUE
	// = ((1 + 2) = 3) /\ ((4 * 2) > 5) => TRUE
	// = (((1 + 2) = 3) /\ ((4 * 2) > 5)) => TRUE
	expr, err := ParseExpression("1 + 2 = 3 /\\ 4 * 2 > 5 => TRUE")
	s.NoError(err)

	implies := expr.(*ast.LogicInfixExpression)
	s.Equal("⇒", implies.Operator)

	// Right is TRUE
	right := implies.Right.(*ast.BooleanLiteral)
	s.True(right.Value)

	// Left is ((1 + 2 = 3) /\ (4 * 2 > 5))
	and := implies.Left.(*ast.LogicInfixExpression)
	s.Equal("∧", and.Operator)

	// and.Left is (1 + 2 = 3)
	eq := and.Left.(*ast.LogicEquality)
	s.Equal("=", eq.Operator)

	// eq.Left is (1 + 2)
	add := eq.Left.(*ast.RealInfixExpression)
	s.Equal("+", add.Operator)

	// and.Right is (4 * 2 > 5)
	gt := and.Right.(*ast.LogicRealComparison)
	s.Equal(">", gt.Operator)

	// gt.Left is (4 * 2)
	mul := gt.Left.(*ast.RealInfixExpression)
	s.Equal("*", mul.Operator)
}

func (s *ComparisonSuite) TestNegationWithComparison() {
	// ~(1 = 2) - negation of comparison
	expr, err := ParseExpression("~(1 = 2)")
	s.NoError(err)

	not := expr.(*ast.LogicPrefixExpression)
	s.Equal("¬", not.Operator)

	paren := not.Right.(*ast.ParenExpr)
	eq := paren.Inner.(*ast.LogicEquality)
	s.Equal("=", eq.Operator)
}
