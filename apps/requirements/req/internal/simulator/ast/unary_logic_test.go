package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestLogicPrefixExpressionSuite(t *testing.T) {
	suite.Run(t, new(LogicPrefixExpressionSuite))
}

type LogicPrefixExpressionSuite struct {
	suite.Suite
}

func (suite *LogicPrefixExpressionSuite) TestString() {
	tests := []struct {
		testName string
		operator string
		right    Expression
		expected string
	}{
		{
			testName: `not operator`,
			operator: LogicOperatorNot,
			right:    &BooleanLiteral{Value: true},
			expected: `¬TRUE`,
		},
		{
			testName: `not with nested expression`,
			operator: LogicOperatorNot,
			right:    &LogicInfixExpression{Left: &BooleanLiteral{Value: true}, Operator: LogicOperatorAnd, Right: &BooleanLiteral{Value: false}},
			expected: `¬TRUE ∧ FALSE`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			pe := &LogicPrefixExpression{
				Operator: tt.operator,
				Right:    tt.right,
			}
			assert.Equal(t, tt.expected, pe.String())
		})
	}
}

func (suite *LogicPrefixExpressionSuite) TestAscii() {
	tests := []struct {
		testName string
		operator string
		right    Expression
		expected string
	}{
		{
			testName: `not operator`,
			operator: LogicOperatorNot,
			right:    &BooleanLiteral{Value: true},
			expected: `~TRUE`,
		},
		{
			testName: `not with nested expression`,
			operator: LogicOperatorNot,
			right:    &LogicInfixExpression{Left: &BooleanLiteral{Value: true}, Operator: LogicOperatorAnd, Right: &BooleanLiteral{Value: false}},
			expected: `~TRUE /\ FALSE`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			pe := &LogicPrefixExpression{
				Operator: tt.operator,
				Right:    tt.right,
			}
			assert.Equal(t, tt.expected, pe.Ascii())
		})
	}
}

func (suite *LogicPrefixExpressionSuite) TestValidate() {
	tests := []struct {
		testName string
		pe       *LogicPrefixExpression
		errstr   string
	}{
		// OK.
		{
			testName: `valid not operator`,
			pe: &LogicPrefixExpression{
				Operator: LogicOperatorNot,
				Right:    &BooleanLiteral{Value: true},
			},
		},
		{
			testName: `valid nested expression`,
			pe: &LogicPrefixExpression{
				Operator: LogicOperatorNot,
				Right: &LogicInfixExpression{
					Operator: LogicOperatorAnd,
					Left:     &BooleanLiteral{Value: true},
					Right:    &BooleanLiteral{Value: false},
				},
			},
		},

		// Errors.
		{
			testName: `error missing operator`,
			pe: &LogicPrefixExpression{
				Right: &BooleanLiteral{Value: true},
			},
			errstr: `Operator`,
		},
		{
			testName: `error invalid operator`,
			pe: &LogicPrefixExpression{
				Operator: `invalid`,
				Right:    &BooleanLiteral{Value: true},
			},
			errstr: `Operator`,
		},
		{
			testName: `error missing right`,
			pe: &LogicPrefixExpression{
				Operator: LogicOperatorNot,
			},
			errstr: `Right`,
		},
		{
			testName: `error invalid nested operator`,
			pe: &LogicPrefixExpression{
				Operator: LogicOperatorNot,
				Right: &LogicInfixExpression{
					Operator: `invalid`,
					Left:     &BooleanLiteral{Value: true},
					Right:    &BooleanLiteral{Value: false},
				},
			},
			errstr: `Operator`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.pe.Validate()
			if tt.errstr == `` {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

func (suite *LogicPrefixExpressionSuite) TestExpressionNode() {
	// Verify that LogicPrefixExpression implements the expressionNode interface method.
	pe := &LogicPrefixExpression{
		Operator: LogicOperatorNot,
		Right:    &BooleanLiteral{Value: true},
	}
	// This should compile and not panic.
	pe.expressionNode()
}
