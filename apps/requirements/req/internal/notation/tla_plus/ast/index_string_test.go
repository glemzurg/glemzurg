package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestStringIndexSuite(t *testing.T) {
	suite.Run(t, new(StringIndexSuite))
}

type StringIndexSuite struct {
	suite.Suite
}

func (suite *StringIndexSuite) TestString() {
	tests := []struct {
		testName string
		str      Expression
		index    Expression
		expected string
	}{
		{
			testName: `string literal with index`,
			str:      &StringLiteral{Value: `hello`},
			index:    NewIntLiteral(0),
			expected: `"hello"[0]`,
		},
		{
			testName: `string literal with index 3`,
			str:      &StringLiteral{Value: `world`},
			index:    NewIntLiteral(3),
			expected: `"world"[3]`,
		},
		{
			testName: `concatenated string with index`,
			str: &StringInfixExpression{
				Operator: StringOperatorConcat,
				Operands: []Expression{
					&StringLiteral{Value: `hello`},
					&StringLiteral{Value: `world`},
				},
			},
			index:    NewIntLiteral(5),
			expected: `"hello" ∘ "world"[5]`,
		},
	}
	for _, tt := range tests {
		_ = suite.Run(tt.testName, func() {
			t := suite.T()
			expr := &StringIndex{
				Str:   tt.str,
				Index: tt.index,
			}
			assert.Equal(t, tt.expected, expr.String())
		})
	}
}

func (suite *StringIndexSuite) TestASCII() {
	tests := []struct {
		testName string
		str      Expression
		index    Expression
		expected string
	}{
		{
			testName: `string literal with index`,
			str:      &StringLiteral{Value: `hello`},
			index:    NewIntLiteral(0),
			expected: `"hello"[0]`,
		},
		{
			testName: `concatenated string with index`,
			str: &StringInfixExpression{
				Operator: StringOperatorConcat,
				Operands: []Expression{
					&StringLiteral{Value: `hello`},
					&StringLiteral{Value: `world`},
				},
			},
			index:    NewIntLiteral(5),
			expected: `"hello" \o "world"[5]`,
		},
	}
	for _, tt := range tests {
		_ = suite.Run(tt.testName, func() {
			t := suite.T()
			expr := &StringIndex{
				Str:   tt.str,
				Index: tt.index,
			}
			assert.Equal(t, tt.expected, expr.ASCII())
		})
	}
}

func (suite *StringIndexSuite) TestValidate() {
	tests := []struct {
		testName string
		s        *StringIndex
		errstr   string
	}{
		// OK.
		{
			testName: `valid index`,
			s: &StringIndex{
				Str:   &StringLiteral{Value: `hello`},
				Index: NewIntLiteral(0),
			},
		},

		// Errors.
		{
			testName: `error missing string`,
			s: &StringIndex{
				Index: NewIntLiteral(0),
			},
			errstr: `Str`,
		},
		{
			testName: `error missing index`,
			s: &StringIndex{
				Str: &StringLiteral{Value: `hello`},
			},
			errstr: `Index`,
		},
	}
	for _, tt := range tests {
		_ = suite.Run(tt.testName, func() {
			t := suite.T()
			err := tt.s.Validate()
			if tt.errstr == `` {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

func (suite *StringIndexSuite) TestExpressionNode() {
	s := &StringIndex{
		Str:   &StringLiteral{Value: `hello`},
		Index: NewIntLiteral(0),
	}
	s.expressionNode()
}
