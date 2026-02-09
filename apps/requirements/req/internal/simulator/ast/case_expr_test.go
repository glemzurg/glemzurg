package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestExpressionCaseSuite(t *testing.T) {
	suite.Run(t, new(ExpressionCaseSuite))
}

type ExpressionCaseSuite struct {
	suite.Suite
}

func (suite *ExpressionCaseSuite) TestString() {
	tests := []struct {
		testName string
		branches []*CaseBranch
		other    Expression
		expected string
	}{
		{
			testName: `single branch`,
			branches: []*CaseBranch{
				{
					Condition: &BooleanLiteral{Value: true},
					Result:    NewIntLiteral(1),
				},
			},
			expected: `CASE TRUE → 1`,
		},
		{
			testName: `two branches`,
			branches: []*CaseBranch{
				{
					Condition: &LogicRealComparison{
						Left:     NewIntLiteral(5),
						Operator: RealComparisonGreaterThanOrEqual,
						Right:    NewIntLiteral(0),
					},
					Result: &StringLiteral{Value: `positive`},
				},
				{
					Condition: &LogicRealComparison{
						Left:     NewIntLiteral(5),
						Operator: RealComparisonLessThan,
						Right:    NewIntLiteral(0),
					},
					Result: &StringLiteral{Value: `negative`},
				},
			},
			expected: `CASE 5 ≥ 0 → "positive" □ 5 < 0 → "negative"`,
		},
		{
			testName: `single branch with other`,
			branches: []*CaseBranch{
				{
					Condition: &BooleanLiteral{Value: true},
					Result:    NewIntLiteral(1),
				},
			},
			other:    NewIntLiteral(0),
			expected: `CASE TRUE → 1 □ OTHER → 0`,
		},
		{
			testName: `three branches with other`,
			branches: []*CaseBranch{
				{
					Condition: &LogicRealComparison{
						Left:     NewIntLiteral(5),
						Operator: RealComparisonGreaterThan,
						Right:    NewIntLiteral(0),
					},
					Result: NewIntLiteral(1),
				},
				{
					Condition: &LogicRealComparison{
						Left:     NewIntLiteral(5),
						Operator: RealComparisonLessThan,
						Right:    NewIntLiteral(0),
					},
					Result: NewIntLiteral(-1),
				},
				{
					Condition: &BooleanLiteral{Value: true},
					Result:    NewIntLiteral(0),
				},
			},
			other:    &StringLiteral{Value: `error`},
			expected: `CASE 5 > 0 → 1 □ 5 < 0 → -1 □ TRUE → 0 □ OTHER → "error"`,
		},
		{
			testName: `with identifier results`,
			branches: []*CaseBranch{
				{
					Condition: &BooleanLiteral{Value: true},
					Result:    &Identifier{Value: `result1`},
				},
				{
					Condition: &BooleanLiteral{Value: false},
					Result:    &Identifier{Value: `result2`},
				},
			},
			other:    &Identifier{Value: `defaultResult`},
			expected: `CASE TRUE → result1 □ FALSE → result2 □ OTHER → defaultResult`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			expr := &ExpressionCase{
				Branches: tt.branches,
				Other:    tt.other,
			}
			assert.Equal(t, tt.expected, expr.String())
		})
	}
}

func (suite *ExpressionCaseSuite) TestAscii() {
	tests := []struct {
		testName string
		branches []*CaseBranch
		other    Expression
		expected string
	}{
		{
			testName: `single branch`,
			branches: []*CaseBranch{
				{
					Condition: &BooleanLiteral{Value: true},
					Result:    NewIntLiteral(1),
				},
			},
			expected: `CASE TRUE -> 1`,
		},
		{
			testName: `two branches`,
			branches: []*CaseBranch{
				{
					Condition: &LogicRealComparison{
						Left:     NewIntLiteral(5),
						Operator: RealComparisonGreaterThanOrEqual,
						Right:    NewIntLiteral(0),
					},
					Result: &StringLiteral{Value: `positive`},
				},
				{
					Condition: &LogicRealComparison{
						Left:     NewIntLiteral(5),
						Operator: RealComparisonLessThan,
						Right:    NewIntLiteral(0),
					},
					Result: &StringLiteral{Value: `negative`},
				},
			},
			expected: `CASE 5 >= 0 -> "positive" [] 5 < 0 -> "negative"`,
		},
		{
			testName: `single branch with other`,
			branches: []*CaseBranch{
				{
					Condition: &BooleanLiteral{Value: true},
					Result:    NewIntLiteral(1),
				},
			},
			other:    NewIntLiteral(0),
			expected: `CASE TRUE -> 1 [] OTHER -> 0`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			expr := &ExpressionCase{
				Branches: tt.branches,
				Other:    tt.other,
			}
			assert.Equal(t, tt.expected, expr.Ascii())
		})
	}
}

func (suite *ExpressionCaseSuite) TestValidate() {
	tests := []struct {
		testName string
		e        *ExpressionCase
		errstr   string
	}{
		// OK.
		{
			testName: `valid single branch`,
			e: &ExpressionCase{
				Branches: []*CaseBranch{
					{
						Condition: &BooleanLiteral{Value: true},
						Result:    NewIntLiteral(1),
					},
				},
			},
		},
		{
			testName: `valid multiple branches`,
			e: &ExpressionCase{
				Branches: []*CaseBranch{
					{
						Condition: &BooleanLiteral{Value: true},
						Result:    NewIntLiteral(1),
					},
					{
						Condition: &BooleanLiteral{Value: false},
						Result:    NewIntLiteral(2),
					},
				},
			},
		},
		{
			testName: `valid with other`,
			e: &ExpressionCase{
				Branches: []*CaseBranch{
					{
						Condition: &BooleanLiteral{Value: true},
						Result:    NewIntLiteral(1),
					},
				},
				Other: NewIntLiteral(0),
			},
		},

		// Errors.
		{
			testName: `error missing branches`,
			e:        &ExpressionCase{},
			errstr:   `Branches`,
		},
		{
			testName: `error empty branches`,
			e: &ExpressionCase{
				Branches: []*CaseBranch{},
			},
			errstr: `Branches`,
		},
		{
			testName: `error nil branch`,
			e: &ExpressionCase{
				Branches: []*CaseBranch{nil},
			},
			errstr: `Branches[0]`,
		},
		{
			testName: `error invalid condition`,
			e: &ExpressionCase{
				Branches: []*CaseBranch{
					{
						Condition: &LogicInfixExpression{
							Operator: LogicOperatorAnd,
							Left:     &BooleanLiteral{Value: true},
							// Missing Right
						},
						Result: NewIntLiteral(1),
					},
				},
			},
			errstr: `Branches[0].Condition`,
		},
		{
			testName: `error invalid result`,
			e: &ExpressionCase{
				Branches: []*CaseBranch{
					{
						Condition: &BooleanLiteral{Value: true},
						Result:    &Identifier{Value: ``},
					},
				},
			},
			errstr: `Branches[0].Result`,
		},
		{
			testName: `error invalid other`,
			e: &ExpressionCase{
				Branches: []*CaseBranch{
					{
						Condition: &BooleanLiteral{Value: true},
						Result:    NewIntLiteral(1),
					},
				},
				Other: &Identifier{Value: ``},
			},
			errstr: `Other`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.e.Validate()
			if tt.errstr == `` {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

func (suite *ExpressionCaseSuite) TestExpressionNode() {
	e := &ExpressionCase{
		Branches: []*CaseBranch{
			{
				Condition: &BooleanLiteral{Value: true},
				Result:    NewIntLiteral(1),
			},
		},
	}
	e.expressionNode()
}
