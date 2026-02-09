package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestStringInfixExpressionSuite(t *testing.T) {
	suite.Run(t, new(StringInfixExpressionSuite))
}

type StringInfixExpressionSuite struct {
	suite.Suite
}

func (suite *StringInfixExpressionSuite) TestString() {
	tests := []struct {
		testName string
		operands []Expression
		operator string
		expected string
	}{
		{
			testName: `two operands`,
			operator: StringOperatorConcat,
			operands: []Expression{
				&StringLiteral{Value: "hello"},
				&StringLiteral{Value: "world"},
			},
			expected: `"hello" ∘ "world"`,
		},
		{
			testName: `three operands`,
			operator: StringOperatorConcat,
			operands: []Expression{
				&StringLiteral{Value: "a"},
				&StringLiteral{Value: "b"},
				&StringLiteral{Value: "c"},
			},
			expected: `"a" ∘ "b" ∘ "c"`,
		},
		{
			testName: `four operands`,
			operator: StringOperatorConcat,
			operands: []Expression{
				&StringLiteral{Value: "one"},
				&StringLiteral{Value: "two"},
				&StringLiteral{Value: "three"},
				&StringLiteral{Value: "four"},
			},
			expected: `"one" ∘ "two" ∘ "three" ∘ "four"`,
		},
		{
			testName: `nested concatenation`,
			operator: StringOperatorConcat,
			operands: []Expression{
				&StringInfixExpression{
					Operator: StringOperatorConcat,
					Operands: []Expression{
						&StringLiteral{Value: "a"},
						&StringLiteral{Value: "b"},
					},
				},
				&StringLiteral{Value: "c"},
			},
			expected: `"a" ∘ "b" ∘ "c"`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			s := &StringInfixExpression{
				Operator: tt.operator,
				Operands: tt.operands,
			}
			assert.Equal(t, tt.expected, s.String())
		})
	}
}

func (suite *StringInfixExpressionSuite) TestAscii() {
	tests := []struct {
		testName string
		operands []Expression
		operator string
		expected string
	}{
		{
			testName: `two operands`,
			operator: StringOperatorConcat,
			operands: []Expression{
				&StringLiteral{Value: "hello"},
				&StringLiteral{Value: "world"},
			},
			expected: `"hello" \o "world"`,
		},
		{
			testName: `three operands`,
			operator: StringOperatorConcat,
			operands: []Expression{
				&StringLiteral{Value: "a"},
				&StringLiteral{Value: "b"},
				&StringLiteral{Value: "c"},
			},
			expected: `"a" \o "b" \o "c"`,
		},
		{
			testName: `nested concatenation`,
			operator: StringOperatorConcat,
			operands: []Expression{
				&StringInfixExpression{
					Operator: StringOperatorConcat,
					Operands: []Expression{
						&StringLiteral{Value: "a"},
						&StringLiteral{Value: "b"},
					},
				},
				&StringLiteral{Value: "c"},
			},
			expected: `"a" \o "b" \o "c"`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			s := &StringInfixExpression{
				Operator: tt.operator,
				Operands: tt.operands,
			}
			assert.Equal(t, tt.expected, s.Ascii())
		})
	}
}

func (suite *StringInfixExpressionSuite) TestValidate() {
	tests := []struct {
		testName string
		s        *StringInfixExpression
		errstr   string
	}{
		// OK.
		{
			testName: `valid two operands`,
			s: &StringInfixExpression{
				Operator: StringOperatorConcat,
				Operands: []Expression{
					&StringLiteral{Value: "hello"},
					&StringLiteral{Value: "world"},
				},
			},
		},
		{
			testName: `valid three operands`,
			s: &StringInfixExpression{
				Operator: StringOperatorConcat,
				Operands: []Expression{
					&StringLiteral{Value: "a"},
					&StringLiteral{Value: "b"},
					&StringLiteral{Value: "c"},
				},
			},
		},
		{
			testName: `valid nested`,
			s: &StringInfixExpression{
				Operator: StringOperatorConcat,
				Operands: []Expression{
					&StringInfixExpression{
						Operator: StringOperatorConcat,
						Operands: []Expression{
							&StringLiteral{Value: "a"},
							&StringLiteral{Value: "b"},
						},
					},
					&StringLiteral{Value: "c"},
				},
			},
		},

		// Errors.
		{
			testName: `error missing operator`,
			s: &StringInfixExpression{
				Operands: []Expression{
					&StringLiteral{Value: "hello"},
					&StringLiteral{Value: "world"},
				},
			},
			errstr: `Operator`,
		},
		{
			testName: `error invalid operator`,
			s: &StringInfixExpression{
				Operator: `invalid`,
				Operands: []Expression{
					&StringLiteral{Value: "hello"},
					&StringLiteral{Value: "world"},
				},
			},
			errstr: `Operator`,
		},
		{
			testName: `error missing operands`,
			s: &StringInfixExpression{
				Operator: StringOperatorConcat,
			},
			errstr: `Operands`,
		},
		{
			testName: `error single operand`,
			s: &StringInfixExpression{
				Operator: StringOperatorConcat,
				Operands: []Expression{
					&StringLiteral{Value: "hello"},
				},
			},
			errstr: `Operands`,
		},
		{
			testName: `error nil operand`,
			s: &StringInfixExpression{
				Operator: StringOperatorConcat,
				Operands: []Expression{
					&StringLiteral{Value: "hello"},
					nil,
				},
			},
			errstr: `Operands[1]`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.s.Validate()
			if tt.errstr == `` {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

func (suite *StringInfixExpressionSuite) TestExpressionNode() {
	// Verify that StringInfixExpression implements the expressionNode interface method.
	s := &StringInfixExpression{
		Operator: StringOperatorConcat,
		Operands: []Expression{
			&StringLiteral{Value: "hello"},
			&StringLiteral{Value: "world"},
		},
	}
	// This should compile and not panic.
	s.expressionNode()
}
