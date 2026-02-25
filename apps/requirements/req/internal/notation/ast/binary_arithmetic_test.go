package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestRealInfixExpressionSuite(t *testing.T) {
	suite.Run(t, new(RealInfixExpressionSuite))
}

type RealInfixExpressionSuite struct {
	suite.Suite
}

func (suite *RealInfixExpressionSuite) TestString() {
	tests := []struct {
		testName string
		left     Expression
		operator string
		right    Expression
		expected string
	}{
		{
			testName: `addition`,
			left:     NewIntLiteral(1),
			operator: RealOperatorAdd,
			right:    NewIntLiteral(2),
			expected: `1 + 2`,
		},
		{
			testName: `subtraction`,
			left:     NewIntLiteral(5),
			operator: RealOperatorSubtract,
			right:    NewIntLiteral(3),
			expected: `5 - 3`,
		},
		{
			testName: `multiplication`,
			left:     NewIntLiteral(4),
			operator: RealOperatorMultiply,
			right:    NewIntLiteral(6),
			expected: `4 * 6`,
		},
		{
			testName: `power`,
			left:     NewIntLiteral(2),
			operator: RealOperatorPower,
			right:    NewIntLiteral(8),
			expected: `2 ^ 8`,
		},
		{
			testName: `division`,
			left:     NewIntLiteral(10),
			operator: RealOperatorDivide,
			right:    NewIntLiteral(2),
			expected: `10 ÷ 2`,
		},
		{
			testName: `modulo`,
			left:     NewIntLiteral(10),
			operator: RealOperatorModulo,
			right:    NewIntLiteral(3),
			expected: `10 % 3`,
		},
		{
			testName: `with natural literals`,
			left:     NewIntLiteral(5),
			operator: RealOperatorAdd,
			right:    NewIntLiteral(3),
			expected: `5 + 3`,
		},
		{
			testName: `with real literal`,
			left:     NewDecimalNumberLiteral("3", "14"),
			operator: RealOperatorMultiply,
			right:    NewIntLiteral(2),
			expected: `3.14 * 2`,
		},
		{
			testName: `nested expression`,
			left: &RealInfixExpression{
				Left:     NewIntLiteral(1),
				Operator: RealOperatorAdd,
				Right:    NewIntLiteral(2),
			},
			operator: RealOperatorMultiply,
			right:    NewIntLiteral(3),
			expected: `1 + 2 * 3`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			r := &RealInfixExpression{
				Left:     tt.left,
				Operator: tt.operator,
				Right:    tt.right,
			}
			assert.Equal(t, tt.expected, r.String())
		})
	}
}

func (suite *RealInfixExpressionSuite) TestAscii() {
	tests := []struct {
		testName string
		left     Expression
		operator string
		right    Expression
		expected string
	}{
		{
			testName: `addition`,
			left:     NewIntLiteral(1),
			operator: RealOperatorAdd,
			right:    NewIntLiteral(2),
			expected: `1 + 2`,
		},
		{
			testName: `division unicode to ascii`,
			left:     NewIntLiteral(10),
			operator: RealOperatorDivide,
			right:    NewIntLiteral(2),
			expected: `10 \div 2`,
		},
		{
			testName: `nested expression`,
			left: &RealInfixExpression{
				Left:     NewIntLiteral(1),
				Operator: RealOperatorAdd,
				Right:    NewIntLiteral(2),
			},
			operator: RealOperatorDivide,
			right:    NewIntLiteral(3),
			expected: `1 + 2 \div 3`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			r := &RealInfixExpression{
				Left:     tt.left,
				Operator: tt.operator,
				Right:    tt.right,
			}
			assert.Equal(t, tt.expected, r.Ascii())
		})
	}
}

func (suite *RealInfixExpressionSuite) TestValidate() {
	tests := []struct {
		testName string
		r        *RealInfixExpression
		errstr   string
	}{
		// OK.
		{
			testName: `valid addition`,
			r: &RealInfixExpression{
				Operator: RealOperatorAdd,
				Left:     NewIntLiteral(1),
				Right:    NewIntLiteral(2),
			},
		},
		{
			testName: `valid subtraction`,
			r: &RealInfixExpression{
				Operator: RealOperatorSubtract,
				Left:     NewIntLiteral(5),
				Right:    NewIntLiteral(3),
			},
		},
		{
			testName: `valid multiplication`,
			r: &RealInfixExpression{
				Operator: RealOperatorMultiply,
				Left:     NewIntLiteral(4),
				Right:    NewIntLiteral(6),
			},
		},
		{
			testName: `valid power`,
			r: &RealInfixExpression{
				Operator: RealOperatorPower,
				Left:     NewIntLiteral(2),
				Right:    NewIntLiteral(8),
			},
		},
		{
			testName: `valid division`,
			r: &RealInfixExpression{
				Operator: RealOperatorDivide,
				Left:     NewIntLiteral(10),
				Right:    NewIntLiteral(2),
			},
		},
		{
			testName: `valid modulo`,
			r: &RealInfixExpression{
				Operator: RealOperatorModulo,
				Left:     NewIntLiteral(10),
				Right:    NewIntLiteral(3),
			},
		},
		{
			testName: `valid nested expression`,
			r: &RealInfixExpression{
				Operator: RealOperatorMultiply,
				Left: &RealInfixExpression{
					Operator: RealOperatorAdd,
					Left:     NewIntLiteral(1),
					Right:    NewIntLiteral(2),
				},
				Right: NewIntLiteral(3),
			},
		},

		// Errors.
		{
			testName: `error missing operator`,
			r: &RealInfixExpression{
				Left:  NewIntLiteral(1),
				Right: NewIntLiteral(2),
			},
			errstr: `Operator`,
		},
		{
			testName: `error invalid operator`,
			r: &RealInfixExpression{
				Operator: `invalid`,
				Left:     NewIntLiteral(1),
				Right:    NewIntLiteral(2),
			},
			errstr: `Operator`,
		},
		{
			testName: `error missing left`,
			r: &RealInfixExpression{
				Operator: RealOperatorAdd,
				Right:    NewIntLiteral(2),
			},
			errstr: `Left`,
		},
		{
			testName: `error missing right`,
			r: &RealInfixExpression{
				Operator: RealOperatorAdd,
				Left:     NewIntLiteral(1),
			},
			errstr: `Right`,
		},
		{
			testName: `error invalid nested operator`,
			r: &RealInfixExpression{
				Operator: RealOperatorMultiply,
				Left: &RealInfixExpression{
					Operator: `invalid`,
					Left:     NewIntLiteral(1),
					Right:    NewIntLiteral(2),
				},
				Right: NewIntLiteral(3),
			},
			errstr: `Operator`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.r.Validate()
			if tt.errstr == `` {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

func (suite *RealInfixExpressionSuite) TestExpressionNode() {
	// Verify that RealInfixExpression implements the expressionNode interface method.
	r := &RealInfixExpression{
		Left:     NewIntLiteral(1),
		Operator: RealOperatorAdd,
		Right:    NewIntLiteral(2),
	}
	// This should compile and not panic.
	r.expressionNode()
}
