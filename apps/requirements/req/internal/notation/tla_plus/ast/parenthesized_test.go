package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestParenExprSuite(t *testing.T) {
	suite.Run(t, new(ParenExprSuite))
}

type ParenExprSuite struct {
	suite.Suite
}

func (suite *ParenExprSuite) TestString() {
	tests := []struct {
		testName string
		p        *ParenExpr
		expected string
	}{
		{
			testName: "simple number",
			p:        &ParenExpr{Inner: &NumberLiteral{Base: BaseDecimal, IntegerPart: "42"}},
			expected: "(42)",
		},
		{
			testName: "nested parentheses",
			p: &ParenExpr{
				Inner: &ParenExpr{
					Inner: &NumberLiteral{Base: BaseDecimal, IntegerPart: "1"},
				},
			},
			expected: "((1))",
		},
		{
			testName: "boolean literal",
			p:        &ParenExpr{Inner: &BooleanLiteral{Value: true}},
			expected: "(TRUE)",
		},
		{
			testName: "nil inner",
			p:        &ParenExpr{Inner: nil},
			expected: "()",
		},
	}
	for _, tt := range tests {
		_ = suite.Run(tt.testName, func() {
			t := suite.T()
			assert.Equal(t, tt.expected, tt.p.String())
		})
	}
}

func (suite *ParenExprSuite) TestASCII() {
	// ASCII should be same as String for ParenExpr
	p := &ParenExpr{Inner: &NumberLiteral{Base: BaseDecimal, IntegerPart: "42"}}
	suite.Equal(p.String(), p.ASCII())
}

func (suite *ParenExprSuite) TestValidate() {
	tests := []struct {
		testName string
		p        *ParenExpr
		errstr   string
	}{
		// OK.
		{
			testName: "valid with number",
			p:        &ParenExpr{Inner: &NumberLiteral{Base: BaseDecimal, IntegerPart: "42"}},
		},
		{
			testName: "valid with boolean",
			p:        &ParenExpr{Inner: &BooleanLiteral{Value: true}},
		},
		// Errors.
		{
			testName: "error nil inner",
			p:        &ParenExpr{Inner: nil},
			errstr:   "inner expression cannot be nil",
		},
	}
	for _, tt := range tests {
		_ = suite.Run(tt.testName, func() {
			t := suite.T()
			err := tt.p.Validate()
			if tt.errstr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

func (suite *ParenExprSuite) TestNewParenExpr() {
	inner := &NumberLiteral{Base: BaseDecimal, IntegerPart: "42"}
	p := NewParenExpr(inner)
	suite.Equal(inner, p.Inner)
}

func (suite *ParenExprSuite) TestExpressionNode() {
	p := &ParenExpr{Inner: &NumberLiteral{Base: BaseDecimal, IntegerPart: "42"}}
	p.expressionNode()
}
