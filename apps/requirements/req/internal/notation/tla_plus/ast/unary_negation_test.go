package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestNumericPrefixExpressionSuite(t *testing.T) {
	suite.Run(t, new(NumericPrefixExpressionSuite))
}

type NumericPrefixExpressionSuite struct {
	suite.Suite
}

func (suite *NumericPrefixExpressionSuite) TestString() {
	tests := []struct {
		testName string
		n        *NumericPrefixExpression
		expected string
	}{
		{
			testName: "negation of integer",
			n:        &NumericPrefixExpression{Operator: "-", Right: &NumberLiteral{Base: BaseDecimal, IntegerPart: "1"}},
			expected: "-1",
		},
		{
			testName: "negation of decimal",
			n:        &NumericPrefixExpression{Operator: "-", Right: &NumberLiteral{Base: BaseDecimal, IntegerPart: "3", HasDecimalPoint: true, FractionalPart: "14"}},
			expected: "-3.14",
		},
		{
			testName: "double negation",
			n: &NumericPrefixExpression{
				Operator: "-",
				Right: &NumericPrefixExpression{
					Operator: "-",
					Right:    &NumberLiteral{Base: BaseDecimal, IntegerPart: "5"},
				},
			},
			expected: "--5",
		},
		{
			testName: "negation of parenthesized",
			n: &NumericPrefixExpression{
				Operator: "-",
				Right:    &ParenExpr{Inner: &NumberLiteral{Base: BaseDecimal, IntegerPart: "42"}},
			},
			expected: "-(42)",
		},
	}
	for _, tt := range tests {
		_ = suite.Run(tt.testName, func() {
			t := suite.T()
			assert.Equal(t, tt.expected, tt.n.String())
		})
	}
}

func (suite *NumericPrefixExpressionSuite) TestASCII() {
	// ASCII should be same as String for NumericPrefixExpression
	n := &NumericPrefixExpression{Operator: "-", Right: &NumberLiteral{Base: BaseDecimal, IntegerPart: "1"}}
	suite.Equal(n.String(), n.ASCII())
}

func (suite *NumericPrefixExpressionSuite) TestValidate() {
	tests := []struct {
		testName string
		n        *NumericPrefixExpression
		errstr   string
	}{
		// OK.
		{
			testName: "valid negation",
			n:        &NumericPrefixExpression{Operator: "-", Right: &NumberLiteral{Base: BaseDecimal, IntegerPart: "1"}},
		},
		// Errors.
		{
			testName: "error nil right",
			n:        &NumericPrefixExpression{Operator: "-", Right: nil},
			errstr:   "Right",
		},
		{
			testName: "error empty operator",
			n:        &NumericPrefixExpression{Operator: "", Right: &NumberLiteral{Base: BaseDecimal, IntegerPart: "1"}},
			errstr:   "Operator",
		},
	}
	for _, tt := range tests {
		_ = suite.Run(tt.testName, func() {
			t := suite.T()
			err := tt.n.Validate()
			if tt.errstr == "" {
				require.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

func (suite *NumericPrefixExpressionSuite) TestNewNegation() {
	inner := &NumberLiteral{Base: BaseDecimal, IntegerPart: "42"}
	n := NewNegation(inner)
	suite.Equal("-", n.Operator)
	suite.Equal(inner, n.Right)
}

func (suite *NumericPrefixExpressionSuite) TestExpressionNode() {
	n := &NumericPrefixExpression{Operator: "-", Right: &NumberLiteral{Base: BaseDecimal, IntegerPart: "1"}}
	n.expressionNode()
}
