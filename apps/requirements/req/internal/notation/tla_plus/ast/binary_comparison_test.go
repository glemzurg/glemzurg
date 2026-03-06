package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
			left:     NewIntLiteral(1),
			operator: RealComparisonLessThan,
			right:    NewIntLiteral(2),
			expected: `1 < 2`,
		},
		{
			testName: `greater than`,
			left:     NewIntLiteral(5),
			operator: RealComparisonGreaterThan,
			right:    NewIntLiteral(3),
			expected: `5 > 3`,
		},
		{
			testName: `less than or equal`,
			left:     NewIntLiteral(4),
			operator: RealComparisonLessThanOrEqual,
			right:    NewIntLiteral(4),
			expected: `4 ≤ 4`,
		},
		{
			testName: `greater than or equal`,
			left:     NewIntLiteral(10),
			operator: RealComparisonGreaterThanOrEqual,
			right:    NewIntLiteral(5),
			expected: `10 ≥ 5`,
		},
		{
			testName: `with natural literals`,
			left:     NewIntLiteral(0),
			operator: RealComparisonLessThanOrEqual,
			right:    NewIntLiteral(100),
			expected: `0 ≤ 100`,
		},
		{
			testName: `with real literal`,
			left:     NewDecimalNumberLiteral("3", "14"),
			operator: RealComparisonLessThan,
			right:    NewIntLiteral(4),
			expected: `3.14 < 4`,
		},
		{
			testName: `with arithmetic expression`,
			left: &RealInfixExpression{
				Left:     NewIntLiteral(1),
				Operator: RealOperatorAdd,
				Right:    NewIntLiteral(2),
			},
			operator: RealComparisonLessThan,
			right:    NewIntLiteral(5),
			expected: `1 + 2 < 5`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			l := &LogicRealComparison{
				Left:     tt.left,
				Operator: tt.operator,
				Right:    tt.right,
			}
			assert.Equal(t, tt.expected, l.String())
		})
	}
}

func (suite *LogicRealComparisonSuite) TestAscii() {
	tests := []struct {
		testName string
		left     Expression
		operator string
		right    Expression
		expected string
	}{
		{
			testName: `less than`,
			left:     NewIntLiteral(1),
			operator: RealComparisonLessThan,
			right:    NewIntLiteral(2),
			expected: `1 < 2`,
		},
		{
			testName: `greater than`,
			left:     NewIntLiteral(5),
			operator: RealComparisonGreaterThan,
			right:    NewIntLiteral(3),
			expected: `5 > 3`,
		},
		{
			testName: `less than or equal unicode to ascii`,
			left:     NewIntLiteral(4),
			operator: RealComparisonLessThanOrEqual,
			right:    NewIntLiteral(4),
			expected: `4 =< 4`,
		},
		{
			testName: `greater than or equal unicode to ascii`,
			left:     NewIntLiteral(10),
			operator: RealComparisonGreaterThanOrEqual,
			right:    NewIntLiteral(5),
			expected: `10 >= 5`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			l := &LogicRealComparison{
				Left:     tt.left,
				Operator: tt.operator,
				Right:    tt.right,
			}
			assert.Equal(t, tt.expected, l.Ascii())
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
				Left:     NewIntLiteral(1),
				Right:    NewIntLiteral(2),
			},
		},
		{
			testName: `valid greater than`,
			l: &LogicRealComparison{
				Operator: RealComparisonGreaterThan,
				Left:     NewIntLiteral(5),
				Right:    NewIntLiteral(3),
			},
		},
		{
			testName: `valid less than or equal`,
			l: &LogicRealComparison{
				Operator: RealComparisonLessThanOrEqual,
				Left:     NewIntLiteral(4),
				Right:    NewIntLiteral(4),
			},
		},
		{
			testName: `valid greater than or equal`,
			l: &LogicRealComparison{
				Operator: RealComparisonGreaterThanOrEqual,
				Left:     NewIntLiteral(10),
				Right:    NewIntLiteral(5),
			},
		},
		{
			testName: `valid with nested arithmetic`,
			l: &LogicRealComparison{
				Operator: RealComparisonLessThan,
				Left: &RealInfixExpression{
					Operator: RealOperatorAdd,
					Left:     NewIntLiteral(1),
					Right:    NewIntLiteral(2),
				},
				Right: NewIntLiteral(5),
			},
		},

		// Errors.
		{
			testName: `error missing operator`,
			l: &LogicRealComparison{
				Left:  NewIntLiteral(1),
				Right: NewIntLiteral(2),
			},
			errstr: `Operator`,
		},
		{
			testName: `error invalid operator`,
			l: &LogicRealComparison{
				Operator: `invalid`,
				Left:     NewIntLiteral(1),
				Right:    NewIntLiteral(2),
			},
			errstr: `Operator`,
		},
		{
			testName: `error missing left`,
			l: &LogicRealComparison{
				Operator: RealComparisonLessThan,
				Right:    NewIntLiteral(2),
			},
			errstr: `Left`,
		},
		{
			testName: `error missing right`,
			l: &LogicRealComparison{
				Operator: RealComparisonLessThan,
				Left:     NewIntLiteral(1),
			},
			errstr: `Right`,
		},
		{
			testName: `error invalid nested operator`,
			l: &LogicRealComparison{
				Operator: RealComparisonLessThan,
				Left: &RealInfixExpression{
					Operator: `invalid`,
					Left:     NewIntLiteral(1),
					Right:    NewIntLiteral(2),
				},
				Right: NewIntLiteral(5),
			},
			errstr: `Operator`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.l.Validate()
			if tt.errstr == `` {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

func (suite *LogicRealComparisonSuite) TestExpressionNode() {
	// Verify that LogicRealComparison implements the expressionNode interface method.
	l := &LogicRealComparison{
		Left:     NewIntLiteral(1),
		Operator: RealComparisonLessThan,
		Right:    NewIntLiteral(2),
	}
	// This should compile and not panic.
	l.expressionNode()
}
