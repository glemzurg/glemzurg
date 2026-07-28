package ast

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestLogicRealComparisonSuite(t *testing.T) {
	suite.Run(t, new(LogicRealComparisonSuite))
}

type LogicRealComparisonSuite struct {
	suite.Suite
}

func (suite *LogicRealComparisonSuite) TestString() {
	tests := []struct {
		testName string
		left     Expression
		operator string
		right    Expression
		expected string
	}{
		{
			testName: `less than`,
			left:     newIntLiteral(1),
			operator: RealComparisonLessThan,
			right:    newIntLiteral(2),
			expected: `1 < 2`,
		},
		{
			testName: `greater than`,
			left:     newIntLiteral(5),
			operator: RealComparisonGreaterThan,
			right:    newIntLiteral(3),
			expected: `5 > 3`,
		},
		{
			testName: `less than or equal`,
			left:     newIntLiteral(4),
			operator: RealComparisonLessThanOrEqual,
			right:    newIntLiteral(4),
			expected: `4 ≤ 4`,
		},
		{
			testName: `greater than or equal`,
			left:     newIntLiteral(10),
			operator: RealComparisonGreaterThanOrEqual,
			right:    newIntLiteral(5),
			expected: `10 ≥ 5`,
		},
		{
			testName: `with natural literals`,
			left:     newIntLiteral(0),
			operator: RealComparisonLessThanOrEqual,
			right:    newIntLiteral(100),
			expected: `0 ≤ 100`,
		},
		{
			testName: `with real literal`,
			left:     NewDecimalNumberLiteral("3", "14"),
			operator: RealComparisonLessThan,
			right:    newIntLiteral(4),
			expected: `3.14 < 4`,
		},
		{
			testName: `with arithmetic expression`,
			left: &RealInfixExpression{
				Left:     newIntLiteral(1),
				Operator: RealOperatorAdd,
				Right:    newIntLiteral(2),
			},
			operator: RealComparisonLessThan,
			right:    newIntLiteral(5),
			expected: `1 + 2 < 5`,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.testName, func() {
			l := &LogicRealComparison{
				Left:     tt.left,
				Operator: tt.operator,
				Right:    tt.right,
			}
			suite.Equal(tt.expected, l.String())
		})
	}
}

func (suite *LogicRealComparisonSuite) TestASCII() {
	tests := []struct {
		testName string
		left     Expression
		operator string
		right    Expression
		expected string
	}{
		{
			testName: `less than`,
			left:     newIntLiteral(1),
			operator: RealComparisonLessThan,
			right:    newIntLiteral(2),
			expected: `1 < 2`,
		},
		{
			testName: `greater than`,
			left:     newIntLiteral(5),
			operator: RealComparisonGreaterThan,
			right:    newIntLiteral(3),
			expected: `5 > 3`,
		},
		{
			testName: `less than or equal unicode to ascii`,
			left:     newIntLiteral(4),
			operator: RealComparisonLessThanOrEqual,
			right:    newIntLiteral(4),
			expected: `4 =< 4`,
		},
		{
			testName: `greater than or equal unicode to ascii`,
			left:     newIntLiteral(10),
			operator: RealComparisonGreaterThanOrEqual,
			right:    newIntLiteral(5),
			expected: `10 >= 5`,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.testName, func() {
			l := &LogicRealComparison{
				Left:     tt.left,
				Operator: tt.operator,
				Right:    tt.right,
			}
			suite.Equal(tt.expected, l.ASCII())
		})
	}
}

func (suite *LogicRealComparisonSuite) TestValidate() {
	tests := []struct {
		testName string
		l        *LogicRealComparison
		errstr   string
	}{
		// OK.
		{
			testName: `valid less than`,
			l: &LogicRealComparison{
				Operator: RealComparisonLessThan,
				Left:     newIntLiteral(1),
				Right:    newIntLiteral(2),
			},
		},
		{
			testName: `valid greater than`,
			l: &LogicRealComparison{
				Operator: RealComparisonGreaterThan,
				Left:     newIntLiteral(5),
				Right:    newIntLiteral(3),
			},
		},
		{
			testName: `valid less than or equal`,
			l: &LogicRealComparison{
				Operator: RealComparisonLessThanOrEqual,
				Left:     newIntLiteral(4),
				Right:    newIntLiteral(4),
			},
		},
		{
			testName: `valid greater than or equal`,
			l: &LogicRealComparison{
				Operator: RealComparisonGreaterThanOrEqual,
				Left:     newIntLiteral(10),
				Right:    newIntLiteral(5),
			},
		},
		{
			testName: `valid with nested arithmetic`,
			l: &LogicRealComparison{
				Operator: RealComparisonLessThan,
				Left: &RealInfixExpression{
					Operator: RealOperatorAdd,
					Left:     newIntLiteral(1),
					Right:    newIntLiteral(2),
				},
				Right: newIntLiteral(5),
			},
		},

		// Errors.
		{
			testName: `error missing operator`,
			l: &LogicRealComparison{
				Left:  newIntLiteral(1),
				Right: newIntLiteral(2),
			},
			errstr: `Operator`,
		},
		{
			testName: `error invalid operator`,
			l: &LogicRealComparison{
				Operator: `invalid`,
				Left:     newIntLiteral(1),
				Right:    newIntLiteral(2),
			},
			errstr: `Operator`,
		},
		{
			testName: `error missing left`,
			l: &LogicRealComparison{
				Operator: RealComparisonLessThan,
				Right:    newIntLiteral(2),
			},
			errstr: `Left`,
		},
		{
			testName: `error missing right`,
			l: &LogicRealComparison{
				Operator: RealComparisonLessThan,
				Left:     newIntLiteral(1),
			},
			errstr: `Right`,
		},
		{
			testName: `error invalid nested operator`,
			l: &LogicRealComparison{
				Operator: RealComparisonLessThan,
				Left: &RealInfixExpression{
					Operator: `invalid`,
					Left:     newIntLiteral(1),
					Right:    newIntLiteral(2),
				},
				Right: newIntLiteral(5),
			},
			errstr: `Operator`,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.testName, func() {
			err := tt.l.Validate()
			if tt.errstr == `` {
				suite.Require().NoError(err)
			} else {
				suite.Require().ErrorContains(err, tt.errstr)
			}
		})
	}
}

func (suite *LogicRealComparisonSuite) TestExpressionNode() {
	// Verify that LogicRealComparison implements the expressionNode interface method.
	l := &LogicRealComparison{
		Left:     newIntLiteral(1),
		Operator: RealComparisonLessThan,
		Right:    newIntLiteral(2),
	}
	// This should compile and not panic.
	l.expressionNode()
}
