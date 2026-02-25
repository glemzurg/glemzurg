package parser

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/ast"
	"github.com/stretchr/testify/suite"
)

type OperatorTestSuite struct {
	suite.Suite
}

func TestOperatorSuite(t *testing.T) {
	suite.Run(t, new(OperatorTestSuite))
}

// =============================================================================
// Negation Operator
// =============================================================================

func (s *OperatorTestSuite) TestParseNegation() {
	expr, err := ParseExpression("-1")
	s.NoError(err)
	s.Equal(&ast.NumericPrefixExpression{
		Operator: "-",
		Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "1"},
	}, expr)
}

func (s *OperatorTestSuite) TestParseNegationDecimal() {
	expr, err := ParseExpression("-.5")
	s.NoError(err)
	s.Equal(&ast.NumericPrefixExpression{
		Operator: "-",
		Right: &ast.NumberLiteral{
			Base:            ast.BaseDecimal,
			IntegerPart:     "",
			HasDecimalPoint: true,
			FractionalPart:  "5",
		},
	}, expr)
}

func (s *OperatorTestSuite) TestParseDoubleNegation() {
	expr, err := ParseExpression("--1")
	s.NoError(err)
	s.Equal(&ast.NumericPrefixExpression{
		Operator: "-",
		Right: &ast.NumericPrefixExpression{
			Operator: "-",
			Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "1"},
		},
	}, expr)
}

// =============================================================================
// Fraction Operator
// =============================================================================

func (s *OperatorTestSuite) TestParseFraction() {
	expr, err := ParseExpression("3/4")
	s.NoError(err)
	s.Equal(&ast.FractionExpr{
		Numerator:   &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "3"},
		Denominator: &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "4"},
	}, expr)
}

func (s *OperatorTestSuite) TestParseFractionWithDecimals() {
	expr, err := ParseExpression("1.5/2.5")
	s.NoError(err)
	s.Equal(&ast.FractionExpr{
		Numerator: &ast.NumberLiteral{
			Base:            ast.BaseDecimal,
			IntegerPart:     "1",
			HasDecimalPoint: true,
			FractionalPart:  "5",
		},
		Denominator: &ast.NumberLiteral{
			Base:            ast.BaseDecimal,
			IntegerPart:     "2",
			HasDecimalPoint: true,
			FractionalPart:  "5",
		},
	}, expr)
}

func (s *OperatorTestSuite) TestParseDivisionByZero() {
	expr, err := ParseExpression("1/0")
	s.NoError(err)
	s.Equal(&ast.FractionExpr{
		Numerator:   &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "1"},
		Denominator: &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "0"},
	}, expr)
}

// =============================================================================
// Combined Negation and Fraction
// =============================================================================

func (s *OperatorTestSuite) TestParseNegatedFraction() {
	// -3/4 parses as: operator - { operator / { number 3, number 4 } }
	expr, err := ParseExpression("-3/4")
	s.NoError(err)
	s.Equal(&ast.NumericPrefixExpression{
		Operator: "-",
		Right: &ast.FractionExpr{
			Numerator:   &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "3"},
			Denominator: &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "4"},
		},
	}, expr)
}

func (s *OperatorTestSuite) TestParseFractionWithNegatedDenominator() {
	// 3/(-1.4) parses as: operator / { number 3, paren { operator - { number 1.4 } } }
	expr, err := ParseExpression("3/(-1.4)")
	s.NoError(err)
	s.Equal(&ast.FractionExpr{
		Numerator: &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "3"},
		Denominator: &ast.ParenExpr{
			Inner: &ast.NumericPrefixExpression{
				Operator: "-",
				Right: &ast.NumberLiteral{
					Base:            ast.BaseDecimal,
					IntegerPart:     "1",
					HasDecimalPoint: true,
					FractionalPart:  "4",
				},
			},
		},
	}, expr)
}

func (s *OperatorTestSuite) TestParseComplexNested() {
	// -3/-(1.4/.2) parses as complex nested structure
	expr, err := ParseExpression("-3/-(1.4/.2)")
	s.NoError(err)
	s.Equal(&ast.NumericPrefixExpression{
		Operator: "-",
		Right: &ast.FractionExpr{
			Numerator: &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "3"},
			Denominator: &ast.NumericPrefixExpression{
				Operator: "-",
				Right: &ast.ParenExpr{
					Inner: &ast.FractionExpr{
						Numerator: &ast.NumberLiteral{
							Base:            ast.BaseDecimal,
							IntegerPart:     "1",
							HasDecimalPoint: true,
							FractionalPart:  "4",
						},
						Denominator: &ast.NumberLiteral{
							Base:            ast.BaseDecimal,
							IntegerPart:     "",
							HasDecimalPoint: true,
							FractionalPart:  "2",
						},
					},
				},
			},
		},
	}, expr)
}

func (s *OperatorTestSuite) TestParseParenthesizedNumerator() {
	// (-3)/4 parses as: operator / { paren { operator - { number 3 } }, number 4 }
	expr, err := ParseExpression("(-3)/4")
	s.NoError(err)
	s.Equal(&ast.FractionExpr{
		Numerator: &ast.ParenExpr{
			Inner: &ast.NumericPrefixExpression{
				Operator: "-",
				Right:    &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "3"},
			},
		},
		Denominator: &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "4"},
	}, expr)
}

// =============================================================================
// Parenthesized Expressions
// =============================================================================

func (s *OperatorTestSuite) TestParseParenthesizedNumber() {
	expr, err := ParseExpression("(42)")
	s.NoError(err)
	s.Equal(&ast.ParenExpr{
		Inner: &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "42"},
	}, expr)
}

func (s *OperatorTestSuite) TestParseNestedParentheses() {
	expr, err := ParseExpression("((123))")
	s.NoError(err)
	s.Equal(&ast.ParenExpr{
		Inner: &ast.ParenExpr{
			Inner: &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "123"},
		},
	}, expr)
}

func (s *OperatorTestSuite) TestParseParenthesesWithWhitespace() {
	expr, err := ParseExpression("( 42 )")
	s.NoError(err)
	s.Equal(&ast.ParenExpr{
		Inner: &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "42"},
	}, expr)
}
