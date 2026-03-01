package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.p.String())
		})
	}
}

func (suite *ParenExprSuite) TestAscii() {
	// Ascii should be same as String for ParenExpr
	p := &ParenExpr{Inner: &NumberLiteral{Base: BaseDecimal, IntegerPart: "42"}}
	assert.Equal(suite.T(), p.String(), p.Ascii())
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
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.p.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

func (suite *ParenExprSuite) TestNewParenExpr() {
	inner := &NumberLiteral{Base: BaseDecimal, IntegerPart: "42"}
	p := NewParenExpr(inner)
	assert.Equal(suite.T(), inner, p.Inner)
}

func (suite *ParenExprSuite) TestExpressionNode() {
	p := &ParenExpr{Inner: &NumberLiteral{Base: BaseDecimal, IntegerPart: "42"}}
	p.expressionNode()
}
