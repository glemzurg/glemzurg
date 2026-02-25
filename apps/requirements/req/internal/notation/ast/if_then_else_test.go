package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestExpressionIfElseSuite(t *testing.T) {
	suite.Run(t, new(ExpressionIfElseSuite))
}

type ExpressionIfElseSuite struct {
	suite.Suite
}

func (suite *ExpressionIfElseSuite) TestString() {
	tests := []struct {
		testName  string
		condition Expression
		then      Expression
		elseExpr  Expression
		expected  string
	}{
		{
			testName:  `if then else with literals`,
			condition: &BooleanLiteral{Value: true},
			then:      NewIntLiteral(1),
			elseExpr:  NewIntLiteral(2),
			expected:  `IF TRUE THEN 1 ELSE 2`,
		},
		{
			testName:  `if then else with strings`,
			condition: &BooleanLiteral{Value: false},
			then:      &StringLiteral{Value: `hello`},
			elseExpr:  &StringLiteral{Value: `world`},
			expected:  `IF FALSE THEN "hello" ELSE "world"`,
		},
		{
			testName: `if then else with comparison`,
			condition: &LogicRealComparison{
				Left:     NewIntLiteral(5),
				Operator: RealComparisonGreaterThan,
				Right:    NewIntLiteral(0),
			},
			then:     &Identifier{Value: `positive`},
			elseExpr: &Identifier{Value: `nonpositive`},
			expected: `IF 5 > 0 THEN positive ELSE nonpositive`,
		},
		{
			testName: `nested if then else`,
			condition: &BooleanLiteral{Value: true},
			then: &ExpressionIfElse{
				Condition: &BooleanLiteral{Value: false},
				Then:      NewIntLiteral(1),
				Else:      NewIntLiteral(2),
			},
			elseExpr: NewIntLiteral(3),
			expected: `IF TRUE THEN IF FALSE THEN 1 ELSE 2 ELSE 3`,
		},
		{
			testName: `if with complex condition`,
			condition: &LogicInfixExpression{
				Operator: LogicOperatorAnd,
				Left:     &BooleanLiteral{Value: true},
				Right:    &BooleanLiteral{Value: false},
			},
			then:     NewIntLiteral(1),
			elseExpr: NewIntLiteral(0),
			expected: `IF TRUE ∧ FALSE THEN 1 ELSE 0`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			expr := &ExpressionIfElse{
				Condition: tt.condition,
				Then:      tt.then,
				Else:      tt.elseExpr,
			}
			assert.Equal(t, tt.expected, expr.String())
		})
	}
}

func (suite *ExpressionIfElseSuite) TestAscii() {
	tests := []struct {
		testName  string
		condition Expression
		then      Expression
		elseExpr  Expression
		expected  string
	}{
		{
			testName:  `if then else with literals`,
			condition: &BooleanLiteral{Value: true},
			then:      NewIntLiteral(1),
			elseExpr:  NewIntLiteral(2),
			expected:  `IF TRUE THEN 1 ELSE 2`,
		},
		{
			testName:  `if then else with strings`,
			condition: &BooleanLiteral{Value: false},
			then:      &StringLiteral{Value: `hello`},
			elseExpr:  &StringLiteral{Value: `world`},
			expected:  `IF FALSE THEN "hello" ELSE "world"`,
		},
		{
			testName: `if with and condition`,
			condition: &LogicInfixExpression{
				Operator: LogicOperatorAnd,
				Left:     &BooleanLiteral{Value: true},
				Right:    &BooleanLiteral{Value: false},
			},
			then:     NewIntLiteral(1),
			elseExpr: NewIntLiteral(0),
			expected: `IF TRUE /\ FALSE THEN 1 ELSE 0`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			expr := &ExpressionIfElse{
				Condition: tt.condition,
				Then:      tt.then,
				Else:      tt.elseExpr,
			}
			assert.Equal(t, tt.expected, expr.Ascii())
		})
	}
}

func (suite *ExpressionIfElseSuite) TestValidate() {
	tests := []struct {
		testName string
		e        *ExpressionIfElse
		errstr   string
	}{
		// OK.
		{
			testName: `valid if then else`,
			e: &ExpressionIfElse{
				Condition: &BooleanLiteral{Value: true},
				Then:      NewIntLiteral(1),
				Else:      NewIntLiteral(2),
			},
		},
		{
			testName: `valid if then else with strings`,
			e: &ExpressionIfElse{
				Condition: &BooleanLiteral{Value: true},
				Then:      &StringLiteral{Value: `yes`},
				Else:      &StringLiteral{Value: `no`},
			},
		},
		{
			testName: `valid nested`,
			e: &ExpressionIfElse{
				Condition: &BooleanLiteral{Value: true},
				Then: &ExpressionIfElse{
					Condition: &BooleanLiteral{Value: false},
					Then:      NewIntLiteral(1),
					Else:      NewIntLiteral(2),
				},
				Else: NewIntLiteral(3),
			},
		},

		// Errors.
		{
			testName: `error missing condition`,
			e: &ExpressionIfElse{
				Then: NewIntLiteral(1),
				Else: NewIntLiteral(2),
			},
			errstr: `Condition`,
		},
		{
			testName: `error missing then`,
			e: &ExpressionIfElse{
				Condition: &BooleanLiteral{Value: true},
				Else:      NewIntLiteral(2),
			},
			errstr: `Then`,
		},
		{
			testName: `error missing else`,
			e: &ExpressionIfElse{
				Condition: &BooleanLiteral{Value: true},
				Then:      NewIntLiteral(1),
			},
			errstr: `Else`,
		},
		{
			testName: `error invalid condition`,
			e: &ExpressionIfElse{
				Condition: &LogicInfixExpression{
					Operator: LogicOperatorAnd,
					Left:     &BooleanLiteral{Value: true},
					// Missing Right
				},
				Then: NewIntLiteral(1),
				Else: NewIntLiteral(2),
			},
			errstr: `Right`,
		},
		{
			testName: `error invalid then`,
			e: &ExpressionIfElse{
				Condition: &BooleanLiteral{Value: true},
				Then:      &Identifier{Value: ``},
				Else:      NewIntLiteral(2),
			},
			errstr: `Value`,
		},
		{
			testName: `error invalid else`,
			e: &ExpressionIfElse{
				Condition: &BooleanLiteral{Value: true},
				Then:      NewIntLiteral(1),
				Else:      &Identifier{Value: ``},
			},
			errstr: `Value`,
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

func (suite *ExpressionIfElseSuite) TestExpressionNode() {
	e := &ExpressionIfElse{
		Condition: &BooleanLiteral{Value: true},
		Then:      NewIntLiteral(1),
		Else:      NewIntLiteral(2),
	}
	e.expressionNode()
}
