package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestLogicMembershipSuite(t *testing.T) {
	suite.Run(t, new(LogicMembershipSuite))
}

type LogicMembershipSuite struct {
	suite.Suite
}

func (suite *LogicMembershipSuite) TestString() {
	tests := []struct {
		testName string
		operator string
		left     Expression
		right    Expression
		expected string
	}{
		{
			testName: `in operator`,
			operator: MembershipOperatorIn,
			left:     &Identifier{Value: `x`},
			right:    &SetConstant{Value: SetConstantBoolean},
			expected: `x ∈ BOOLEAN`,
		},
		{
			testName: `not in operator`,
			operator: MembershipOperatorNotIn,
			left:     &Identifier{Value: `y`},
			right:    &SetConstant{Value: SetConstantBoolean},
			expected: `y ∉ BOOLEAN`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			m := &LogicMembership{
				Operator: tt.operator,
				Left:     tt.left,
				Right:    tt.right,
			}
			assert.Equal(t, tt.expected, m.String())
		})
	}
}

func (suite *LogicMembershipSuite) TestAscii() {
	tests := []struct {
		testName string
		operator string
		left     Expression
		right    Expression
		expected string
	}{
		{
			testName: `in operator`,
			operator: MembershipOperatorIn,
			left:     &Identifier{Value: `x`},
			right:    &SetConstant{Value: SetConstantBoolean},
			expected: `x \in BOOLEAN`,
		},
		{
			testName: `not in operator`,
			operator: MembershipOperatorNotIn,
			left:     &Identifier{Value: `y`},
			right:    &SetConstant{Value: SetConstantBoolean},
			expected: `y \notin BOOLEAN`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			m := &LogicMembership{
				Operator: tt.operator,
				Left:     tt.left,
				Right:    tt.right,
			}
			assert.Equal(t, tt.expected, m.Ascii())
		})
	}
}

func (suite *LogicMembershipSuite) TestValidate() {
	tests := []struct {
		testName string
		m        *LogicMembership
		errstr   string
	}{
		// OK.
		{
			testName: `valid in operator`,
			m: &LogicMembership{
				Operator: MembershipOperatorIn,
				Left:     &Identifier{Value: `x`},
				Right:    &SetConstant{Value: SetConstantBoolean},
			},
		},
		{
			testName: `valid not in operator`,
			m: &LogicMembership{
				Operator: MembershipOperatorNotIn,
				Left:     &Identifier{Value: `x`},
				Right:    &SetConstant{Value: SetConstantBoolean},
			},
		},

		// Errors.
		{
			testName: `error missing operator`,
			m: &LogicMembership{
				Left:  &Identifier{Value: `x`},
				Right: &SetConstant{Value: SetConstantBoolean},
			},
			errstr: `Operator`,
		},
		{
			testName: `error invalid operator`,
			m: &LogicMembership{
				Operator: `invalid`,
				Left:     &Identifier{Value: `x`},
				Right:    &SetConstant{Value: SetConstantBoolean},
			},
			errstr: `Operator`,
		},
		{
			testName: `error missing left`,
			m: &LogicMembership{
				Operator: MembershipOperatorIn,
				Right:    &SetConstant{Value: SetConstantBoolean},
			},
			errstr: `Left`,
		},
		{
			testName: `error missing right`,
			m: &LogicMembership{
				Operator: MembershipOperatorIn,
				Left:     &Identifier{Value: `x`},
			},
			errstr: `Right`,
		},
		{
			testName: `error invalid left`,
			m: &LogicMembership{
				Operator: MembershipOperatorIn,
				Left:     &Identifier{Value: ``},
				Right:    &SetConstant{Value: SetConstantBoolean},
			},
			errstr: `Value`,
		},
		{
			testName: `error invalid right set`,
			m: &LogicMembership{
				Operator: MembershipOperatorIn,
				Left:     &Identifier{Value: `x`},
				Right:    &SetConstant{Value: `INVALID`},
			},
			errstr: `Value`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.m.Validate()
			if tt.errstr == `` {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

func (suite *LogicMembershipSuite) TestExpressionNode() {
	// Verify that LogicMembership implements the expressionNode interface method.
	m := &LogicMembership{
		Operator: MembershipOperatorIn,
		Left:     &Identifier{Value: `x`},
		Right:    &SetConstant{Value: SetConstantBoolean},
	}
	// This should compile and not panic.
	m.expressionNode()
}
