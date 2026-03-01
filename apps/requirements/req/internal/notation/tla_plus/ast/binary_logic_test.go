package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestLogicInfixExpressionSuite(t *testing.T) {
	suite.Run(t, new(LogicInfixExpressionSuite))
}

type LogicInfixExpressionSuite struct {
	suite.Suite
}

func (suite *LogicInfixExpressionSuite) TestString() {
	tests := []struct {
		testName string
		left     Expression
		operator string
		right    Expression
		expected string
	}{
		{
			testName: `and operator`,
			left:     &BooleanLiteral{Value: true},
			operator: LogicOperatorAnd,
			right:    &BooleanLiteral{Value: false},
			expected: `TRUE ∧ FALSE`,
		},
		{
			testName: `or operator`,
			left:     &BooleanLiteral{Value: false},
			operator: LogicOperatorOr,
			right:    &BooleanLiteral{Value: true},
			expected: `FALSE ∨ TRUE`,
		},
		{
			testName: `nested expression`,
			left:     &LogicInfixExpression{Left: &BooleanLiteral{Value: true}, Operator: LogicOperatorAnd, Right: &BooleanLiteral{Value: true}},
			operator: LogicOperatorOr,
			right:    &BooleanLiteral{Value: false},
			expected: `TRUE ∧ TRUE ∨ FALSE`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			ie := &LogicInfixExpression{
				Left:     tt.left,
				Operator: tt.operator,
				Right:    tt.right,
			}
			assert.Equal(t, tt.expected, ie.String())
		})
	}
}

func (suite *LogicInfixExpressionSuite) TestAscii() {
	tests := []struct {
		testName string
		left     Expression
		operator string
		right    Expression
		expected string
	}{
		{
			testName: `and operator`,
			left:     &BooleanLiteral{Value: true},
			operator: LogicOperatorAnd,
			right:    &BooleanLiteral{Value: false},
			expected: `TRUE /\ FALSE`,
		},
		{
			testName: `or operator`,
			left:     &BooleanLiteral{Value: false},
			operator: LogicOperatorOr,
			right:    &BooleanLiteral{Value: true},
			expected: `FALSE \/ TRUE`,
		},
		{
			testName: `nested expression`,
			left:     &LogicInfixExpression{Left: &BooleanLiteral{Value: true}, Operator: LogicOperatorAnd, Right: &BooleanLiteral{Value: true}},
			operator: LogicOperatorOr,
			right:    &BooleanLiteral{Value: false},
			expected: `TRUE /\ TRUE \/ FALSE`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			ie := &LogicInfixExpression{
				Left:     tt.left,
				Operator: tt.operator,
				Right:    tt.right,
			}
			assert.Equal(t, tt.expected, ie.Ascii())
		})
	}
}

func (suite *LogicInfixExpressionSuite) TestValidate() {
	tests := []struct {
		testName string
		ie       *LogicInfixExpression
		errstr   string
	}{
		// OK.
		{
			testName: `valid and operator`,
			ie: &LogicInfixExpression{
				Operator: LogicOperatorAnd,
				Left:     &BooleanLiteral{Value: true},
				Right:    &BooleanLiteral{Value: false},
			},
		},
		{
			testName: `valid or operator`,
			ie: &LogicInfixExpression{
				Operator: LogicOperatorOr,
				Left:     &BooleanLiteral{Value: true},
				Right:    &BooleanLiteral{Value: false},
			},
		},
		{
			testName: `valid implies operator`,
			ie: &LogicInfixExpression{
				Operator: LogicOperatorImplies,
				Left:     &BooleanLiteral{Value: true},
				Right:    &BooleanLiteral{Value: false},
			},
		},
		{
			testName: `valid equiv operator`,
			ie: &LogicInfixExpression{
				Operator: LogicOperatorEquiv,
				Left:     &BooleanLiteral{Value: true},
				Right:    &BooleanLiteral{Value: false},
			},
		},
		{
			testName: `valid nested expression`,
			ie: &LogicInfixExpression{
				Operator: LogicOperatorOr,
				Left: &LogicInfixExpression{
					Operator: LogicOperatorAnd,
					Left:     &BooleanLiteral{Value: true},
					Right:    &BooleanLiteral{Value: true},
				},
				Right: &BooleanLiteral{Value: false},
			},
		},

		// Errors.
		{
			testName: `error missing operator`,
			ie: &LogicInfixExpression{
				Left:  &BooleanLiteral{Value: true},
				Right: &BooleanLiteral{Value: false},
			},
			errstr: `Operator`,
		},
		{
			testName: `error invalid operator`,
			ie: &LogicInfixExpression{
				Operator: `invalid`,
				Left:     &BooleanLiteral{Value: true},
				Right:    &BooleanLiteral{Value: false},
			},
			errstr: `Operator`,
		},
		{
			testName: `error missing left`,
			ie: &LogicInfixExpression{
				Operator: LogicOperatorAnd,
				Right:    &BooleanLiteral{Value: false},
			},
			errstr: `Left`,
		},
		{
			testName: `error missing right`,
			ie: &LogicInfixExpression{
				Operator: LogicOperatorAnd,
				Left:     &BooleanLiteral{Value: true},
			},
			errstr: `Right`,
		},
		{
			testName: `error invalid nested operator`,
			ie: &LogicInfixExpression{
				Operator: LogicOperatorOr,
				Left: &LogicInfixExpression{
					Operator: `invalid`,
					Left:     &BooleanLiteral{Value: true},
					Right:    &BooleanLiteral{Value: true},
				},
				Right: &BooleanLiteral{Value: false},
			},
			errstr: `Operator`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.ie.Validate()
			if tt.errstr == `` {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

func (suite *LogicInfixExpressionSuite) TestExpressionNode() {
	// Verify that LogicInfixExpression implements the expressionNode interface method.
	ie := &LogicInfixExpression{
		Left:     &BooleanLiteral{Value: true},
		Operator: LogicOperatorAnd,
		Right:    &BooleanLiteral{Value: false},
	}
	// This should compile and not panic.
	ie.expressionNode()
}
