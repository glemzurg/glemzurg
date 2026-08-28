package parser

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/stretchr/testify/suite"
)

type APITestSuite struct {
	suite.Suite
}

func TestAPISuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}

// =============================================================================
// parseExpressionList
// =============================================================================

func (s *APITestSuite) TestParseExpressionList() {
	inputs := []string{"TRUE", "42", `"hello"`, "-3/4"}
	exprs, err := parseExpressionList(inputs)
	s.Require().NoError(err)
	s.Len(exprs, 4)

	s.Equal(&ast.BooleanLiteral{Value: true}, exprs[0])
	s.Equal(&ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "42"}, exprs[1])
	s.Equal(&ast.StringLiteral{Value: "hello"}, exprs[2])
	s.Equal(&ast.UnaryNegation{
		Operator: "-",
		Right: &ast.Fraction{
			Numerator:   &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "3"},
			Denominator: &ast.NumberLiteral{Base: ast.BaseDecimal, IntegerPart: "4"},
		},
	}, exprs[3])
}

func (s *APITestSuite) TestParseExpressionListError() {
	inputs := []string{"TRUE", "invalid@#$", "42"}
	_, err := parseExpressionList(inputs)
	s.Require().Error(err)
	s.Contains(err.Error(), "expression 1")
}

// =============================================================================
// mustParseExpression
// =============================================================================

func (s *APITestSuite) TestMustParseExpressionSuccess() {
	s.NotPanics(func() {
		expr := mustParseExpression("TRUE")
		s.NotNil(expr)
	})
}

func (s *APITestSuite) TestMustParseExpressionPanics() {
	s.Panics(func() {
		mustParseExpression("invalid@#$")
	})
}
