package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestSetConditionalSuite(t *testing.T) {
	suite.Run(t, new(SetConditionalSuite))
}

type SetConditionalSuite struct {
	suite.Suite
}

func (suite *SetConditionalSuite) TestString() {
	tests := []struct {
		testName   string
		membership Expression
		predicate  Expression
		expected   string
	}{
		{
			testName: `simple conditional`,
			membership: &LogicMembership{
				Operator: MembershipOperatorIn,
				Left:     &Identifier{Value: `x`},
				Right:    &SetConstant{Value: SetConstantBoolean},
			},
			predicate: &BooleanLiteral{Value: true},
			expected:  `{x ∈ BOOLEAN : TRUE}`,
		},
		{
			testName: `conditional with complex predicate`,
			membership: &LogicMembership{
				Operator: MembershipOperatorIn,
				Left:     &Identifier{Value: `y`},
				Right:    &SetConstant{Value: SetConstantBoolean},
			},
			predicate: &LogicInfixExpression{
				Operator: LogicOperatorAnd,
				Left:     &BooleanLiteral{Value: true},
				Right:    &BooleanLiteral{Value: false},
			},
			expected: `{y ∈ BOOLEAN : TRUE ∧ FALSE}`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			sc := &SetConditional{
				Membership: tt.membership,
				Predicate:  tt.predicate,
			}
			assert.Equal(t, tt.expected, sc.String())
		})
	}
}

func (suite *SetConditionalSuite) TestAscii() {
	tests := []struct {
		testName   string
		membership Expression
		predicate  Expression
		expected   string
	}{
		{
			testName: `simple conditional`,
			membership: &LogicMembership{
				Operator: MembershipOperatorIn,
				Left:     &Identifier{Value: `x`},
				Right:    &SetConstant{Value: SetConstantBoolean},
			},
			predicate: &BooleanLiteral{Value: true},
			expected:  `{x \in BOOLEAN : TRUE}`,
		},
		{
			testName: `conditional with complex predicate`,
			membership: &LogicMembership{
				Operator: MembershipOperatorIn,
				Left:     &Identifier{Value: `y`},
				Right:    &SetConstant{Value: SetConstantBoolean},
			},
			predicate: &LogicInfixExpression{
				Operator: LogicOperatorAnd,
				Left:     &BooleanLiteral{Value: true},
				Right:    &BooleanLiteral{Value: false},
			},
			expected: `{y \in BOOLEAN : TRUE /\ FALSE}`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			sc := &SetConditional{
				Membership: tt.membership,
				Predicate:  tt.predicate,
			}
			assert.Equal(t, tt.expected, sc.Ascii())
		})
	}
}

func (suite *SetConditionalSuite) TestValidate() {
	tests := []struct {
		testName string
		sc       *SetConditional
		errstr   string
	}{
		// OK.
		{
			testName: `valid simple conditional`,
			sc: &SetConditional{
				Membership: &LogicMembership{
					Operator: MembershipOperatorIn,
					Left:     &Identifier{Value: `x`},
					Right:    &SetConstant{Value: SetConstantBoolean},
				},
				Predicate: &BooleanLiteral{Value: true},
			},
		},
		{
			testName: `valid conditional with complex predicate`,
			sc: &SetConditional{
				Membership: &LogicMembership{
					Operator: MembershipOperatorIn,
					Left:     &Identifier{Value: `x`},
					Right:    &SetConstant{Value: SetConstantBoolean},
				},
				Predicate: &LogicInfixExpression{
					Operator: LogicOperatorAnd,
					Left:     &BooleanLiteral{Value: true},
					Right:    &BooleanLiteral{Value: false},
				},
			},
		},

		// Errors.
		{
			testName: `error missing membership`,
			sc: &SetConditional{
				Predicate: &BooleanLiteral{Value: true},
			},
			errstr: `Membership`,
		},
		{
			testName: `error missing predicate`,
			sc: &SetConditional{
				Membership: &LogicMembership{
					Operator: MembershipOperatorIn,
					Left:     &Identifier{Value: `x`},
					Right:    &SetConstant{Value: SetConstantBoolean},
				},
			},
			errstr: `Predicate`,
		},
		{
			testName: `error invalid membership`,
			sc: &SetConditional{
				Membership: &LogicMembership{
					Operator: `invalid`,
					Left:     &Identifier{Value: `x`},
					Right:    &SetConstant{Value: SetConstantBoolean},
				},
				Predicate: &BooleanLiteral{Value: true},
			},
			errstr: `Operator`,
		},
		{
			testName: `error invalid predicate`,
			sc: &SetConditional{
				Membership: &LogicMembership{
					Operator: MembershipOperatorIn,
					Left:     &Identifier{Value: `x`},
					Right:    &SetConstant{Value: SetConstantBoolean},
				},
				Predicate: &LogicInfixExpression{
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
			err := tt.sc.Validate()
			if tt.errstr == `` {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

func (suite *SetConditionalSuite) TestExpressionNode() {
	// Verify that SetConditional implements the expressionNode interface method.
	sc := &SetConditional{
		Membership: &LogicMembership{
			Operator: MembershipOperatorIn,
			Left:     &Identifier{Value: `x`},
			Right:    &SetConstant{Value: SetConstantBoolean},
		},
		Predicate: &BooleanLiteral{Value: true},
	}
	// This should compile and not panic.
	sc.expressionNode()
}
