package parser

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/ast"
	"github.com/stretchr/testify/suite"
)

type ArithmeticTestSuite struct {
	suite.Suite
}

func TestArithmeticSuite(t *testing.T) {
	suite.Run(t, new(ArithmeticTestSuite))
}

// =============================================================================
// Addition
// =============================================================================

func (s *ArithmeticTestSuite) TestParseAddition() {
	expr, err := ParseExpression("1 + 2")
	s.NoError(err)
	s.Equal(&ast.RealInfixExpression{
		Left:     &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "1"},
		Operator: "+",
		Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "2"},
	}, expr)
}

func (s *ArithmeticTestSuite) TestParseAdditionNoSpaces() {
	expr, err := ParseExpression("1+2")
	s.NoError(err)
	s.Equal(&ast.RealInfixExpression{
		Left:     &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "1"},
		Operator: "+",
		Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "2"},
	}, expr)
}

func (s *ArithmeticTestSuite) TestParseAdditionChain() {
	// 1 + 2 + 3 = (1 + 2) + 3 (left-associative)
	expr, err := ParseExpression("1 + 2 + 3")
	s.NoError(err)
	s.Equal(&ast.RealInfixExpression{
		Left: &ast.RealInfixExpression{
			Left:     &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "1"},
			Operator: "+",
			Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "2"},
		},
		Operator: "+",
		Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "3"},
	}, expr)
}

// =============================================================================
// Subtraction
// =============================================================================

func (s *ArithmeticTestSuite) TestParseSubtraction() {
	expr, err := ParseExpression("5 - 3")
	s.NoError(err)
	s.Equal(&ast.RealInfixExpression{
		Left:     &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "5"},
		Operator: "-",
		Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "3"},
	}, expr)
}

func (s *ArithmeticTestSuite) TestParseSubtractionChain() {
	// 5 - 3 - 1 = (5 - 3) - 1 (left-associative)
	expr, err := ParseExpression("5 - 3 - 1")
	s.NoError(err)
	s.Equal(&ast.RealInfixExpression{
		Left: &ast.RealInfixExpression{
			Left:     &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "5"},
			Operator: "-",
			Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "3"},
		},
		Operator: "-",
		Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "1"},
	}, expr)
}

func (s *ArithmeticTestSuite) TestParseMixedAddSub() {
	// 1 + 2 - 3 = 1 + (2 - 3) because - has higher precedence than + in TLA+
	expr, err := ParseExpression("1 + 2 - 3")
	s.NoError(err)
	s.Equal(&ast.RealInfixExpression{
		Left:     &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "1"},
		Operator: "+",
		Right: &ast.RealInfixExpression{
			Left:     &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "2"},
			Operator: "-",
			Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "3"},
		},
	}, expr)
}

// =============================================================================
// Multiplication
// =============================================================================

func (s *ArithmeticTestSuite) TestParseMultiplication() {
	expr, err := ParseExpression("2 * 3")
	s.NoError(err)
	s.Equal(&ast.RealInfixExpression{
		Left:     &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "2"},
		Operator: "*",
		Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "3"},
	}, expr)
}

func (s *ArithmeticTestSuite) TestParseMultiplicationChain() {
	// 2 * 3 * 4 = (2 * 3) * 4 (left-associative)
	expr, err := ParseExpression("2 * 3 * 4")
	s.NoError(err)
	s.Equal(&ast.RealInfixExpression{
		Left: &ast.RealInfixExpression{
			Left:     &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "2"},
			Operator: "*",
			Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "3"},
		},
		Operator: "*",
		Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "4"},
	}, expr)
}

// =============================================================================
// Division (using ÷ or \div)
// =============================================================================

func (s *ArithmeticTestSuite) TestParseDivision() {
	expr, err := ParseExpression("6 ÷ 2")
	s.NoError(err)
	s.Equal(&ast.RealInfixExpression{
		Left:     &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "6"},
		Operator: "÷",
		Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "2"},
	}, expr)
}

func (s *ArithmeticTestSuite) TestParseDivisionAscii() {
	// \div is the ASCII alternative for ÷
	expr, err := ParseExpression(`6 \div 2`)
	s.NoError(err)
	s.Equal(&ast.RealInfixExpression{
		Left:     &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "6"},
		Operator: "÷",
		Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "2"},
	}, expr)
}

// =============================================================================
// Modulo
// =============================================================================

func (s *ArithmeticTestSuite) TestParseModulo() {
	expr, err := ParseExpression("7 % 3")
	s.NoError(err)
	s.Equal(&ast.RealInfixExpression{
		Left:     &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "7"},
		Operator: "%",
		Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "3"},
	}, expr)
}

func (s *ArithmeticTestSuite) TestModuloLowestPrecedence() {
	// 1 + 2 % 3 - 4 = (1 + 2) % (3 - 4) because % has lowest binary precedence
	expr, err := ParseExpression("1 + 2 % 3 - 4")
	s.NoError(err)
	s.Equal(&ast.RealInfixExpression{
		Left: &ast.RealInfixExpression{
			Left:     &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "1"},
			Operator: "+",
			Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "2"},
		},
		Operator: "%",
		Right: &ast.RealInfixExpression{
			Left:     &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "3"},
			Operator: "-",
			Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "4"},
		},
	}, expr)
}

// =============================================================================
// Exponentiation
// =============================================================================

func (s *ArithmeticTestSuite) TestParsePower() {
	expr, err := ParseExpression("2 ^ 3")
	s.NoError(err)
	s.Equal(&ast.RealInfixExpression{
		Left:     &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "2"},
		Operator: "^",
		Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "3"},
	}, expr)
}

// =============================================================================
// Operator Precedence
// =============================================================================

func (s *ArithmeticTestSuite) TestPrecedenceMultOverAdd() {
	// 1 + 2 * 3 = 1 + (2 * 3) because * has higher precedence
	expr, err := ParseExpression("1 + 2 * 3")
	s.NoError(err)
	s.Equal(&ast.RealInfixExpression{
		Left:     &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "1"},
		Operator: "+",
		Right: &ast.RealInfixExpression{
			Left:     &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "2"},
			Operator: "*",
			Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "3"},
		},
	}, expr)
}

func (s *ArithmeticTestSuite) TestPrecedenceMultOverSub() {
	// 5 - 2 * 3 = 5 - (2 * 3)
	expr, err := ParseExpression("5 - 2 * 3")
	s.NoError(err)
	s.Equal(&ast.RealInfixExpression{
		Left:     &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "5"},
		Operator: "-",
		Right: &ast.RealInfixExpression{
			Left:     &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "2"},
			Operator: "*",
			Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "3"},
		},
	}, expr)
}

func (s *ArithmeticTestSuite) TestPrecedenceDivOverAdd() {
	// 1 + 6 ÷ 2 = 1 + (6 ÷ 2)
	expr, err := ParseExpression("1 + 6 ÷ 2")
	s.NoError(err)
	s.Equal(&ast.RealInfixExpression{
		Left:     &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "1"},
		Operator: "+",
		Right: &ast.RealInfixExpression{
			Left:     &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "6"},
			Operator: "÷",
			Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "2"},
		},
	}, expr)
}

func (s *ArithmeticTestSuite) TestPrecedenceComplexExpression() {
	// 1 + 2 * 3 - 4 = 1 + ((2 * 3) - 4) because - > + and * > -
	expr, err := ParseExpression("1 + 2 * 3 - 4")
	s.NoError(err)
	s.Equal(&ast.RealInfixExpression{
		Left:     &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "1"},
		Operator: "+",
		Right: &ast.RealInfixExpression{
			Left: &ast.RealInfixExpression{
				Left:     &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "2"},
				Operator: "*",
				Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "3"},
			},
			Operator: "-",
			Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "4"},
		},
	}, expr)
}

// =============================================================================
// Parentheses Override Precedence
// =============================================================================

func (s *ArithmeticTestSuite) TestParenthesesOverridePrecedence() {
	// (1 + 2) * 3 = explicit grouping
	expr, err := ParseExpression("(1 + 2) * 3")
	s.NoError(err)
	s.Equal(&ast.RealInfixExpression{
		Left: &ast.ParenExpr{
			Inner: &ast.RealInfixExpression{
				Left:     &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "1"},
				Operator: "+",
				Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "2"},
			},
		},
		Operator: "*",
		Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "3"},
	}, expr)
}

func (s *ArithmeticTestSuite) TestParenthesesOnRight() {
	// 2 * (3 + 4) = explicit grouping
	expr, err := ParseExpression("2 * (3 + 4)")
	s.NoError(err)
	s.Equal(&ast.RealInfixExpression{
		Left:     &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "2"},
		Operator: "*",
		Right: &ast.ParenExpr{
			Inner: &ast.RealInfixExpression{
				Left:     &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "3"},
				Operator: "+",
				Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "4"},
			},
		},
	}, expr)
}

// =============================================================================
// Unary Negation with Arithmetic
// =============================================================================

func (s *ArithmeticTestSuite) TestNegationWithAddition() {
	// -1 + 2 = (-1) + 2 because prefix - (at 12) has higher precedence than + (at 10)
	expr, err := ParseExpression("-1 + 2")
	s.NoError(err)
	s.Equal(&ast.RealInfixExpression{
		Left: &ast.NumericPrefixExpression{
			Operator: "-",
			Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "1"},
		},
		Operator: "+",
		Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "2"},
	}, expr)
}

func (s *ArithmeticTestSuite) TestNegationWithMultiplication() {
	// -2 * 3 = -(2 * 3) because prefix - (at 12) has LOWER precedence than * (at 13)
	// Lower precedence = binds looser = captures more
	// To get (-2) * 3, use parentheses explicitly
	expr, err := ParseExpression("-2 * 3")
	s.NoError(err)
	s.Equal(&ast.NumericPrefixExpression{
		Operator: "-",
		Right: &ast.RealInfixExpression{
			Left:     &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "2"},
			Operator: "*",
			Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "3"},
		},
	}, expr)
}

func (s *ArithmeticTestSuite) TestParenthesizedNegationWithMultiplication() {
	// (-2) * 3 = explicit negation on operand
	expr, err := ParseExpression("(-2) * 3")
	s.NoError(err)
	s.Equal(&ast.RealInfixExpression{
		Left: &ast.ParenExpr{
			Inner: &ast.NumericPrefixExpression{
				Operator: "-",
				Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "2"},
			},
		},
		Operator: "*",
		Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "3"},
	}, expr)
}

func (s *ArithmeticTestSuite) TestSubtractionVsNegation() {
	// 5 - -3 = 5 - (-3) (binary minus followed by unary minus)
	expr, err := ParseExpression("5 - -3")
	s.NoError(err)
	s.Equal(&ast.RealInfixExpression{
		Left:     &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "5"},
		Operator: "-",
		Right: &ast.NumericPrefixExpression{
			Operator: "-",
			Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "3"},
		},
	}, expr)
}

// =============================================================================
// Fraction vs Division
// =============================================================================

func (s *ArithmeticTestSuite) TestFractionWithMultiplication() {
	// 2 * 3/4 = 2 * (3/4) because / (at 13.7) has higher precedence than * (at 13.5)
	expr, err := ParseExpression("2 * 3/4")
	s.NoError(err)
	s.Equal(&ast.RealInfixExpression{
		Left:     &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "2"},
		Operator: "*",
		Right: &ast.FractionExpr{
			Numerator:   &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "3"},
			Denominator: &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "4"},
		},
	}, expr)
}

func (s *ArithmeticTestSuite) TestFractionWithAddition() {
	// 1/2 + 3/4 = (1/2) + (3/4)
	expr, err := ParseExpression("1/2 + 3/4")
	s.NoError(err)
	s.Equal(&ast.RealInfixExpression{
		Left: &ast.FractionExpr{
			Numerator:   &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "1"},
			Denominator: &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "2"},
		},
		Operator: "+",
		Right: &ast.FractionExpr{
			Numerator:   &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "3"},
			Denominator: &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "4"},
		},
	}, expr)
}

// =============================================================================
// Decimal Numbers in Arithmetic
// =============================================================================

func (s *ArithmeticTestSuite) TestDecimalArithmetic() {
	expr, err := ParseExpression("3.14 + 2.86")
	s.NoError(err)
	s.Equal(&ast.RealInfixExpression{
		Left: &ast.NumberLiteral{
			Base:            ast.BaseDecimal,
			IntegerPart:     "3",
			HasDecimalPoint: true,
			FractionalPart:  "14",
		},
		Operator: "+",
		Right: &ast.NumberLiteral{
			Base:            ast.BaseDecimal,
			IntegerPart:     "2",
			HasDecimalPoint: true,
			FractionalPart:  "86",
		},
	}, expr)
}
