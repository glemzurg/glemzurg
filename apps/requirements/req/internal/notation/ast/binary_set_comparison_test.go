package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestLogicInfixSetSuite(t *testing.T) {
	suite.Run(t, new(LogicInfixSetSuite))
}

type LogicInfixSetSuite struct {
	suite.Suite
}

func (suite *LogicInfixSetSuite) TestString() {
	tests := []struct {
		testName string
		operator string
		left     Expression
		right    Expression
		expected string
	}{
		{
			testName: `equal operator`,
			operator: LogicSetOperatorEqual,
			left:     &SetConstant{Value: SetConstantBoolean},
			right:    &SetConstant{Value: SetConstantBoolean},
			expected: `BOOLEAN = BOOLEAN`,
		},
		{
			testName: `not equal operator`,
			operator: LogicSetOperatorNotEqual,
			left:     &SetConstant{Value: SetConstantBoolean},
			right:    &SetConstant{Value: SetConstantBoolean},
			expected: `BOOLEAN ≠ BOOLEAN`,
		},
		{
			testName: `subset or equal operator`,
			operator: LogicSetOperatorSubsetEq,
			left:     &SetConstant{Value: SetConstantBoolean},
			right:    &SetConstant{Value: SetConstantBoolean},
			expected: `BOOLEAN ⊆ BOOLEAN`,
		},
		{
			testName: `subset operator`,
			operator: LogicSetOperatorSubset,
			left:     &SetConstant{Value: SetConstantBoolean},
			right:    &SetConstant{Value: SetConstantBoolean},
			expected: `BOOLEAN ⊂ BOOLEAN`,
		},
		{
			testName: `superset or equal operator`,
			operator: LogicSetOperatorSupersetEq,
			left:     &SetConstant{Value: SetConstantBoolean},
			right:    &SetConstant{Value: SetConstantBoolean},
			expected: `BOOLEAN ⊇ BOOLEAN`,
		},
		{
			testName: `superset operator`,
			operator: LogicSetOperatorSuperset,
			left:     &SetConstant{Value: SetConstantBoolean},
			right:    &SetConstant{Value: SetConstantBoolean},
			expected: `BOOLEAN ⊃ BOOLEAN`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			is := &LogicInfixSet{
				Operator: tt.operator,
				Left:     tt.left,
				Right:    tt.right,
			}
			assert.Equal(t, tt.expected, is.String())
		})
	}
}

func (suite *LogicInfixSetSuite) TestAscii() {
	tests := []struct {
		testName string
		operator string
		left     Expression
		right    Expression
		expected string
	}{
		{
			testName: `equal operator`,
			operator: LogicSetOperatorEqual,
			left:     &SetConstant{Value: SetConstantBoolean},
			right:    &SetConstant{Value: SetConstantBoolean},
			expected: `BOOLEAN = BOOLEAN`,
		},
		{
			testName: `not equal operator`,
			operator: LogicSetOperatorNotEqual,
			left:     &SetConstant{Value: SetConstantBoolean},
			right:    &SetConstant{Value: SetConstantBoolean},
			expected: `BOOLEAN /= BOOLEAN`,
		},
		{
			testName: `subset or equal operator`,
			operator: LogicSetOperatorSubsetEq,
			left:     &SetConstant{Value: SetConstantBoolean},
			right:    &SetConstant{Value: SetConstantBoolean},
			expected: `BOOLEAN \subseteq BOOLEAN`,
		},
		{
			testName: `subset operator`,
			operator: LogicSetOperatorSubset,
			left:     &SetConstant{Value: SetConstantBoolean},
			right:    &SetConstant{Value: SetConstantBoolean},
			expected: `BOOLEAN \subset BOOLEAN`,
		},
		{
			testName: `superset or equal operator`,
			operator: LogicSetOperatorSupersetEq,
			left:     &SetConstant{Value: SetConstantBoolean},
			right:    &SetConstant{Value: SetConstantBoolean},
			expected: `BOOLEAN \supseteq BOOLEAN`,
		},
		{
			testName: `superset operator`,
			operator: LogicSetOperatorSuperset,
			left:     &SetConstant{Value: SetConstantBoolean},
			right:    &SetConstant{Value: SetConstantBoolean},
			expected: `BOOLEAN \supset BOOLEAN`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			is := &LogicInfixSet{
				Operator: tt.operator,
				Left:     tt.left,
				Right:    tt.right,
			}
			assert.Equal(t, tt.expected, is.Ascii())
		})
	}
}

func (suite *LogicInfixSetSuite) TestValidate() {
	tests := []struct {
		testName string
		is       *LogicInfixSet
		errstr   string
	}{
		// OK.
		{
			testName: `valid equal operator`,
			is: &LogicInfixSet{
				Operator: LogicSetOperatorEqual,
				Left:     &SetConstant{Value: SetConstantBoolean},
				Right:    &SetConstant{Value: SetConstantBoolean},
			},
		},
		{
			testName: `valid not equal operator`,
			is: &LogicInfixSet{
				Operator: LogicSetOperatorNotEqual,
				Left:     &SetConstant{Value: SetConstantBoolean},
				Right:    &SetConstant{Value: SetConstantBoolean},
			},
		},
		{
			testName: `valid subset or equal operator`,
			is: &LogicInfixSet{
				Operator: LogicSetOperatorSubsetEq,
				Left:     &SetConstant{Value: SetConstantBoolean},
				Right:    &SetConstant{Value: SetConstantBoolean},
			},
		},
		{
			testName: `valid subset operator`,
			is: &LogicInfixSet{
				Operator: LogicSetOperatorSubset,
				Left:     &SetConstant{Value: SetConstantBoolean},
				Right:    &SetConstant{Value: SetConstantBoolean},
			},
		},
		{
			testName: `valid superset or equal operator`,
			is: &LogicInfixSet{
				Operator: LogicSetOperatorSupersetEq,
				Left:     &SetConstant{Value: SetConstantBoolean},
				Right:    &SetConstant{Value: SetConstantBoolean},
			},
		},
		{
			testName: `valid superset operator`,
			is: &LogicInfixSet{
				Operator: LogicSetOperatorSuperset,
				Left:     &SetConstant{Value: SetConstantBoolean},
				Right:    &SetConstant{Value: SetConstantBoolean},
			},
		},

		// Errors.
		{
			testName: `error missing operator`,
			is: &LogicInfixSet{
				Left:  &SetConstant{Value: SetConstantBoolean},
				Right: &SetConstant{Value: SetConstantBoolean},
			},
			errstr: `Operator`,
		},
		{
			testName: `error invalid operator`,
			is: &LogicInfixSet{
				Operator: `invalid`,
				Left:     &SetConstant{Value: SetConstantBoolean},
				Right:    &SetConstant{Value: SetConstantBoolean},
			},
			errstr: `Operator`,
		},
		{
			testName: `error missing left`,
			is: &LogicInfixSet{
				Operator: LogicSetOperatorEqual,
				Right:    &SetConstant{Value: SetConstantBoolean},
			},
			errstr: `Left`,
		},
		{
			testName: `error missing right`,
			is: &LogicInfixSet{
				Operator: LogicSetOperatorEqual,
				Left:     &SetConstant{Value: SetConstantBoolean},
			},
			errstr: `Right`,
		},
		{
			testName: `error invalid left set`,
			is: &LogicInfixSet{
				Operator: LogicSetOperatorEqual,
				Left:     &SetConstant{Value: `INVALID`},
				Right:    &SetConstant{Value: SetConstantBoolean},
			},
			errstr: `Value`,
		},
		{
			testName: `error invalid right set`,
			is: &LogicInfixSet{
				Operator: LogicSetOperatorEqual,
				Left:     &SetConstant{Value: SetConstantBoolean},
				Right:    &SetConstant{Value: `INVALID`},
			},
			errstr: `Value`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.is.Validate()
			if tt.errstr == `` {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

func (suite *LogicInfixSetSuite) TestExpressionNode() {
	// Verify that LogicInfixSet implements the expressionNode interface method.
	is := &LogicInfixSet{
		Operator: LogicSetOperatorEqual,
		Left:     &SetConstant{Value: SetConstantBoolean},
		Right:    &SetConstant{Value: SetConstantBoolean},
	}
	// This should compile and not panic.
	is.expressionNode()
}
