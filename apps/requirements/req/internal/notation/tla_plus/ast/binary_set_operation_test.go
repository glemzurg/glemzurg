package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestSetInfixSuite(t *testing.T) {
	suite.Run(t, new(SetInfixSuite))
}

type SetInfixSuite struct {
	suite.Suite
}

func (suite *SetInfixSuite) TestString() {
	tests := []struct {
		testName string
		operator string
		left     Expression
		right    Expression
		expected string
	}{
		{
			testName: `union operator`,
			operator: SetOperatorUnion,
			left:     &SetConstant{Value: SetConstantBoolean},
			right:    &SetConstant{Value: SetConstantBoolean},
			expected: `BOOLEAN ∪ BOOLEAN`,
		},
		{
			testName: `intersection operator`,
			operator: SetOperatorIntersection,
			left:     &SetConstant{Value: SetConstantBoolean},
			right:    &SetConstant{Value: SetConstantBoolean},
			expected: `BOOLEAN ∩ BOOLEAN`,
		},
		{
			testName: `difference operator`,
			operator: SetOperatorDifference,
			left:     &SetConstant{Value: SetConstantBoolean},
			right:    &SetConstant{Value: SetConstantBoolean},
			expected: `BOOLEAN \ BOOLEAN`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			si := &SetInfix{
				Operator: tt.operator,
				Left:     tt.left,
				Right:    tt.right,
			}
			assert.Equal(t, tt.expected, si.String())
		})
	}
}

func (suite *SetInfixSuite) TestAscii() {
	tests := []struct {
		testName string
		operator string
		left     Expression
		right    Expression
		expected string
	}{
		{
			testName: `union operator`,
			operator: SetOperatorUnion,
			left:     &SetConstant{Value: SetConstantBoolean},
			right:    &SetConstant{Value: SetConstantBoolean},
			expected: `BOOLEAN \union BOOLEAN`,
		},
		{
			testName: `intersection operator`,
			operator: SetOperatorIntersection,
			left:     &SetConstant{Value: SetConstantBoolean},
			right:    &SetConstant{Value: SetConstantBoolean},
			expected: `BOOLEAN \intersect BOOLEAN`,
		},
		{
			testName: `difference operator`,
			operator: SetOperatorDifference,
			left:     &SetConstant{Value: SetConstantBoolean},
			right:    &SetConstant{Value: SetConstantBoolean},
			expected: `BOOLEAN \ BOOLEAN`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			si := &SetInfix{
				Operator: tt.operator,
				Left:     tt.left,
				Right:    tt.right,
			}
			assert.Equal(t, tt.expected, si.Ascii())
		})
	}
}

func (suite *SetInfixSuite) TestValidate() {
	tests := []struct {
		testName string
		si       *SetInfix
		errstr   string
	}{
		// OK.
		{
			testName: `valid union operator`,
			si: &SetInfix{
				Operator: SetOperatorUnion,
				Left:     &SetConstant{Value: SetConstantBoolean},
				Right:    &SetConstant{Value: SetConstantBoolean},
			},
		},
		{
			testName: `valid intersection operator`,
			si: &SetInfix{
				Operator: SetOperatorIntersection,
				Left:     &SetConstant{Value: SetConstantBoolean},
				Right:    &SetConstant{Value: SetConstantBoolean},
			},
		},
		{
			testName: `valid difference operator`,
			si: &SetInfix{
				Operator: SetOperatorDifference,
				Left:     &SetConstant{Value: SetConstantBoolean},
				Right:    &SetConstant{Value: SetConstantBoolean},
			},
		},

		// Errors.
		{
			testName: `error missing operator`,
			si: &SetInfix{
				Left:  &SetConstant{Value: SetConstantBoolean},
				Right: &SetConstant{Value: SetConstantBoolean},
			},
			errstr: `Operator`,
		},
		{
			testName: `error invalid operator`,
			si: &SetInfix{
				Operator: `invalid`,
				Left:     &SetConstant{Value: SetConstantBoolean},
				Right:    &SetConstant{Value: SetConstantBoolean},
			},
			errstr: `Operator`,
		},
		{
			testName: `error missing left`,
			si: &SetInfix{
				Operator: SetOperatorUnion,
				Right:    &SetConstant{Value: SetConstantBoolean},
			},
			errstr: `Left`,
		},
		{
			testName: `error missing right`,
			si: &SetInfix{
				Operator: SetOperatorUnion,
				Left:     &SetConstant{Value: SetConstantBoolean},
			},
			errstr: `Right`,
		},
		{
			testName: `error invalid left set`,
			si: &SetInfix{
				Operator: SetOperatorUnion,
				Left:     &SetConstant{Value: `INVALID`},
				Right:    &SetConstant{Value: SetConstantBoolean},
			},
			errstr: `Value`,
		},
		{
			testName: `error invalid right set`,
			si: &SetInfix{
				Operator: SetOperatorUnion,
				Left:     &SetConstant{Value: SetConstantBoolean},
				Right:    &SetConstant{Value: `INVALID`},
			},
			errstr: `Value`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.si.Validate()
			if tt.errstr == `` {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

func (suite *SetInfixSuite) TestExpressionNode() {
	// Verify that SetInfix implements the expressionNode interface method.
	si := &SetInfix{
		Operator: SetOperatorUnion,
		Left:     &SetConstant{Value: SetConstantBoolean},
		Right:    &SetConstant{Value: SetConstantBoolean},
	}
	// This should compile and not panic.
	si.expressionNode()
}
