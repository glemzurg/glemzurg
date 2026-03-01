package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestLogicBoundQuantifierSuite(t *testing.T) {
	suite.Run(t, new(LogicBoundQuantifierSuite))
}

type LogicBoundQuantifierSuite struct {
	suite.Suite
}

func (suite *LogicBoundQuantifierSuite) TestString() {
	tests := []struct {
		testName   string
		quantifier string
		membership Expression
		predicate  Expression
		expected   string
	}{
		{
			testName:   `forall quantifier`,
			quantifier: LogicQuantifierForAll,
			membership: &LogicMembership{
				Operator: MembershipOperatorIn,
				Left:     &Identifier{Value: `x`},
				Right:    &SetConstant{Value: SetConstantBoolean},
			},
			predicate: &BooleanLiteral{Value: true},
			expected:  `(∀x ∈ BOOLEAN : TRUE)`,
		},
		{
			testName:   `exists quantifier`,
			quantifier: LogicQuantifierExists,
			membership: &LogicMembership{
				Operator: MembershipOperatorIn,
				Left:     &Identifier{Value: `y`},
				Right:    &SetConstant{Value: SetConstantBoolean},
			},
			predicate: &BooleanLiteral{Value: false},
			expected:  `(∃y ∈ BOOLEAN : FALSE)`,
		},
		{
			testName:   `nested predicate`,
			quantifier: LogicQuantifierForAll,
			membership: &LogicMembership{
				Operator: MembershipOperatorIn,
				Left:     &Identifier{Value: `x`},
				Right:    &SetConstant{Value: SetConstantBoolean},
			},
			predicate: &LogicInfixExpression{
				Operator: LogicOperatorImplies,
				Left:     &BooleanLiteral{Value: true},
				Right:    &BooleanLiteral{Value: false},
			},
			expected: `(∀x ∈ BOOLEAN : TRUE ⇒ FALSE)`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			q := &LogicBoundQuantifier{
				Quantifier: tt.quantifier,
				Membership: tt.membership,
				Predicate:  tt.predicate,
			}
			assert.Equal(t, tt.expected, q.String())
		})
	}
}

func (suite *LogicBoundQuantifierSuite) TestAscii() {
	tests := []struct {
		testName   string
		quantifier string
		membership Expression
		predicate  Expression
		expected   string
	}{
		{
			testName:   `forall quantifier`,
			quantifier: LogicQuantifierForAll,
			membership: &LogicMembership{
				Operator: MembershipOperatorIn,
				Left:     &Identifier{Value: `x`},
				Right:    &SetConstant{Value: SetConstantBoolean},
			},
			predicate: &BooleanLiteral{Value: true},
			expected:  `(\A x \in BOOLEAN : TRUE)`,
		},
		{
			testName:   `exists quantifier`,
			quantifier: LogicQuantifierExists,
			membership: &LogicMembership{
				Operator: MembershipOperatorIn,
				Left:     &Identifier{Value: `y`},
				Right:    &SetConstant{Value: SetConstantBoolean},
			},
			predicate: &BooleanLiteral{Value: false},
			expected:  `(\E y \in BOOLEAN : FALSE)`,
		},
		{
			testName:   `nested predicate`,
			quantifier: LogicQuantifierForAll,
			membership: &LogicMembership{
				Operator: MembershipOperatorIn,
				Left:     &Identifier{Value: `x`},
				Right:    &SetConstant{Value: SetConstantBoolean},
			},
			predicate: &LogicInfixExpression{
				Operator: LogicOperatorImplies,
				Left:     &BooleanLiteral{Value: true},
				Right:    &BooleanLiteral{Value: false},
			},
			expected: `(\A x \in BOOLEAN : TRUE => FALSE)`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			q := &LogicBoundQuantifier{
				Quantifier: tt.quantifier,
				Membership: tt.membership,
				Predicate:  tt.predicate,
			}
			assert.Equal(t, tt.expected, q.Ascii())
		})
	}
}

func (suite *LogicBoundQuantifierSuite) TestValidate() {
	tests := []struct {
		testName string
		q        *LogicBoundQuantifier
		errstr   string
	}{
		// OK.
		{
			testName: `valid forall quantifier`,
			q: &LogicBoundQuantifier{
				Quantifier: LogicQuantifierForAll,
				Membership: &LogicMembership{
					Operator: MembershipOperatorIn,
					Left:     &Identifier{Value: `x`},
					Right:    &SetConstant{Value: SetConstantBoolean},
				},
				Predicate: &BooleanLiteral{Value: true},
			},
		},
		{
			testName: `valid exists quantifier`,
			q: &LogicBoundQuantifier{
				Quantifier: LogicQuantifierExists,
				Membership: &LogicMembership{
					Operator: MembershipOperatorIn,
					Left:     &Identifier{Value: `x`},
					Right:    &SetConstant{Value: SetConstantBoolean},
				},
				Predicate: &BooleanLiteral{Value: true},
			},
		},
		{
			testName: `valid nested predicate`,
			q: &LogicBoundQuantifier{
				Quantifier: LogicQuantifierForAll,
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
			testName: `error missing quantifier`,
			q: &LogicBoundQuantifier{
				Membership: &LogicMembership{
					Operator: MembershipOperatorIn,
					Left:     &Identifier{Value: `x`},
					Right:    &SetConstant{Value: SetConstantBoolean},
				},
				Predicate: &BooleanLiteral{Value: true},
			},
			errstr: `Quantifier`,
		},
		{
			testName: `error invalid quantifier`,
			q: &LogicBoundQuantifier{
				Quantifier: `invalid`,
				Membership: &LogicMembership{
					Operator: MembershipOperatorIn,
					Left:     &Identifier{Value: `x`},
					Right:    &SetConstant{Value: SetConstantBoolean},
				},
				Predicate: &BooleanLiteral{Value: true},
			},
			errstr: `Quantifier`,
		},
		{
			testName: `error missing membership`,
			q: &LogicBoundQuantifier{
				Quantifier: LogicQuantifierForAll,
				Predicate:  &BooleanLiteral{Value: true},
			},
			errstr: `Membership`,
		},
		{
			testName: `error missing predicate`,
			q: &LogicBoundQuantifier{
				Quantifier: LogicQuantifierForAll,
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
			q: &LogicBoundQuantifier{
				Quantifier: LogicQuantifierForAll,
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
			q: &LogicBoundQuantifier{
				Quantifier: LogicQuantifierForAll,
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
			err := tt.q.Validate()
			if tt.errstr == `` {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

func (suite *LogicBoundQuantifierSuite) TestExpressionNode() {
	// Verify that LogicBoundQuantifier implements the expressionNode interface method.
	q := &LogicBoundQuantifier{
		Quantifier: LogicQuantifierForAll,
		Membership: &LogicMembership{
			Operator: MembershipOperatorIn,
			Left:     &Identifier{Value: `x`},
			Right:    &SetConstant{Value: SetConstantBoolean},
		},
		Predicate: &BooleanLiteral{Value: true},
	}
	// This should compile and not panic.
	q.expressionNode()
}
