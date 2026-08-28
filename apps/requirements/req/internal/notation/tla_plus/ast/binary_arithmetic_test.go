package ast

import (
	"testing"

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
			left:     newIntLiteral(1),
			operator: RealOperatorAdd,
			right:    newIntLiteral(2),
			expected: `1 + 2`,
		},
		{
			testName: `subtraction`,
			left:     newIntLiteral(5),
			operator: RealOperatorSubtract,
			right:    newIntLiteral(3),
			expected: `5 - 3`,
		},
		{
			testName: `multiplication`,
			left:     newIntLiteral(4),
			operator: RealOperatorMultiply,
			right:    newIntLiteral(6),
			expected: `4 * 6`,
		},
		{
			testName: `power`,
			left:     newIntLiteral(2),
			operator: RealOperatorPower,
			right:    newIntLiteral(8),
			expected: `2 ^ 8`,
		},
		{
			testName: `division`,
			left:     newIntLiteral(10),
			operator: RealOperatorDivide,
			right:    newIntLiteral(2),
			expected: `10 ÷ 2`,
		},
		{
			testName: `modulo`,
			left:     newIntLiteral(10),
			operator: RealOperatorModulo,
			right:    newIntLiteral(3),
			expected: `10 % 3`,
		},
		{
			testName: `with natural literals`,
			left:     newIntLiteral(5),
			operator: RealOperatorAdd,
			right:    newIntLiteral(3),
			expected: `5 + 3`,
		},
		{
			testName: `with real literal`,
			left:     NewDecimalNumberLiteral("3", "14"),
			operator: RealOperatorMultiply,
			right:    newIntLiteral(2),
			expected: `3.14 * 2`,
		},
		{
			testName: `nested expression`,
			left: &RealInfixExpression{
				Left:     newIntLiteral(1),
				Operator: RealOperatorAdd,
				Right:    newIntLiteral(2),
			},
			operator: RealOperatorMultiply,
			right:    newIntLiteral(3),
			expected: `1 + 2 * 3`,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.testName, func() {
			r := &RealInfixExpression{
				Left:     tt.left,
				Operator: tt.operator,
				Right:    tt.right,
			}
			suite.Equal(tt.expected, r.String())
		})
	}
}

func (suite *RealInfixExpressionSuite) TestASCII() {
	tests := []struct {
		testName string
		left     Expression
		operator string
		right    Expression
		expected string
	}{
		{
			testName: `addition`,
			left:     newIntLiteral(1),
			operator: RealOperatorAdd,
			right:    newIntLiteral(2),
			expected: `1 + 2`,
		},
		{
			testName: `division unicode to ascii`,
			left:     newIntLiteral(10),
			operator: RealOperatorDivide,
			right:    newIntLiteral(2),
			expected: `10 \div 2`,
		},
		{
			testName: `nested expression`,
			left: &RealInfixExpression{
				Left:     newIntLiteral(1),
				Operator: RealOperatorAdd,
				Right:    newIntLiteral(2),
			},
			operator: RealOperatorDivide,
			right:    newIntLiteral(3),
			expected: `1 + 2 \div 3`,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.testName, func() {
			r := &RealInfixExpression{
				Left:     tt.left,
				Operator: tt.operator,
				Right:    tt.right,
			}
			suite.Equal(tt.expected, r.ASCII())
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
				Left:     newIntLiteral(1),
				Right:    newIntLiteral(2),
			},
		},
		{
			testName: `valid subtraction`,
			r: &RealInfixExpression{
				Operator: RealOperatorSubtract,
				Left:     newIntLiteral(5),
				Right:    newIntLiteral(3),
			},
		},
		{
			testName: `valid multiplication`,
			r: &RealInfixExpression{
				Operator: RealOperatorMultiply,
				Left:     newIntLiteral(4),
				Right:    newIntLiteral(6),
			},
		},
		{
			testName: `valid power`,
			r: &RealInfixExpression{
				Operator: RealOperatorPower,
				Left:     newIntLiteral(2),
				Right:    newIntLiteral(8),
			},
		},
		{
			testName: `valid division`,
			r: &RealInfixExpression{
				Operator: RealOperatorDivide,
				Left:     newIntLiteral(10),
				Right:    newIntLiteral(2),
			},
		},
		{
			testName: `valid modulo`,
			r: &RealInfixExpression{
				Operator: RealOperatorModulo,
				Left:     newIntLiteral(10),
				Right:    newIntLiteral(3),
			},
		},
		{
			testName: `valid nested expression`,
			r: &RealInfixExpression{
				Operator: RealOperatorMultiply,
				Left: &RealInfixExpression{
					Operator: RealOperatorAdd,
					Left:     newIntLiteral(1),
					Right:    newIntLiteral(2),
				},
				Right: newIntLiteral(3),
			},
		},

		// Errors.
		{
			testName: `error missing operator`,
			r: &RealInfixExpression{
				Left:  newIntLiteral(1),
				Right: newIntLiteral(2),
			},
			errstr: `Operator`,
		},
		{
			testName: `error invalid operator`,
			r: &RealInfixExpression{
				Operator: `invalid`,
				Left:     newIntLiteral(1),
				Right:    newIntLiteral(2),
			},
			errstr: `Operator`,
		},
		{
			testName: `error missing left`,
			r: &RealInfixExpression{
				Operator: RealOperatorAdd,
				Right:    newIntLiteral(2),
			},
			errstr: `Left`,
		},
		{
			testName: `error missing right`,
			r: &RealInfixExpression{
				Operator: RealOperatorAdd,
				Left:     newIntLiteral(1),
			},
			errstr: `Right`,
		},
		{
			testName: `error invalid nested operator`,
			r: &RealInfixExpression{
				Operator: RealOperatorMultiply,
				Left: &RealInfixExpression{
					Operator: `invalid`,
					Left:     newIntLiteral(1),
					Right:    newIntLiteral(2),
				},
				Right: newIntLiteral(3),
			},
			errstr: `Operator`,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.testName, func() {
			err := tt.r.Validate()
			if tt.errstr == `` {
				suite.Require().NoError(err)
			} else {
				suite.Require().ErrorContains(err, tt.errstr)
			}
		})
	}
}

func (suite *RealInfixExpressionSuite) TestExpressionNode() {
	// Verify that RealInfixExpression implements the expressionNode interface method.
	r := &RealInfixExpression{
		Left:     newIntLiteral(1),
		Operator: RealOperatorAdd,
		Right:    newIntLiteral(2),
	}
	// This should compile and not panic.
	r.expressionNode()
}
