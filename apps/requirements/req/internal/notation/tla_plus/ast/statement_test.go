package ast

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestAssignmentSuite(t *testing.T) {
	suite.Run(t, new(AssignmentSuite))
}

type AssignmentSuite struct {
	suite.Suite
}

func (suite *AssignmentSuite) TestString() {
	tests := []struct {
		testName string
		target   *Identifier
		value    Expression
		expected string
	}{
		{
			testName: `assign integer to state`,
			target:   &Identifier{Value: `count`},
			value:    newIntLiteral(0),
			expected: `count' = 0`,
		},
		{
			testName: `assign string to state`,
			target:   &Identifier{Value: `name`},
			value:    &StringLiteral{Value: `hello`},
			expected: `name' = "hello"`,
		},
		{
			testName: `assign identifier to state`,
			target:   &Identifier{Value: `x`},
			value:    &Identifier{Value: `y`},
			expected: `x' = y`,
		},
		{
			testName: `assign tuple to state`,
			target:   &Identifier{Value: `items`},
			value: &TupleLiteral{
				Elements: []Expression{
					newIntLiteral(1),
					newIntLiteral(2),
					newIntLiteral(3),
				},
			},
			expected: `items' = ⟨1, 2, 3⟩`,
		},
		{
			testName: `assign if-else expression to state`,
			target:   &Identifier{Value: `result`},
			value: &ExpressionIfElse{
				Condition: &BooleanLiteral{Value: true},
				Then:      newIntLiteral(1),
				Else:      newIntLiteral(0),
			},
			expected: `result' = IF TRUE THEN 1 ELSE 0`,
		},
	}
	for _, tt := range tests {
		_ = suite.Run(tt.testName, func() {
			stmt := &Assignment{
				Target: tt.target,
				Value:  tt.value,
			}
			suite.Equal(tt.expected, stmt.String())
		})
	}
}

func (suite *AssignmentSuite) TestASCII() {
	tests := []struct {
		testName string
		target   *Identifier
		value    Expression
		expected string
	}{
		{
			testName: `assign integer to state`,
			target:   &Identifier{Value: `count`},
			value:    newIntLiteral(42),
			expected: `count' = 42`,
		},
		{
			testName: `assign tuple to state`,
			target:   &Identifier{Value: `items`},
			value: &TupleLiteral{
				Elements: []Expression{
					newIntLiteral(1),
					newIntLiteral(2),
				},
			},
			expected: `items' = <<1, 2>>`,
		},
	}
	for _, tt := range tests {
		_ = suite.Run(tt.testName, func() {
			stmt := &Assignment{
				Target: tt.target,
				Value:  tt.value,
			}
			suite.Equal(tt.expected, stmt.ASCII())
		})
	}
}

func (suite *AssignmentSuite) TestValidate() {
	tests := []struct {
		testName string
		a        *Assignment
		errstr   string
	}{
		// OK.
		{
			testName: `valid assignment`,
			a: &Assignment{
				Target: &Identifier{Value: `x`},
				Value:  newIntLiteral(1),
			},
		},

		// Errors.
		{
			testName: `error missing target`,
			a: &Assignment{
				Value: newIntLiteral(1),
			},
			errstr: `Target`,
		},
		{
			testName: `error missing value`,
			a: &Assignment{
				Target: &Identifier{Value: `x`},
			},
			errstr: `Value`,
		},
		{
			testName: `error invalid target`,
			a: &Assignment{
				Target: &Identifier{Value: ``},
				Value:  newIntLiteral(1),
			},
			errstr: `Value`,
		},
		{
			testName: `error invalid value`,
			a: &Assignment{
				Target: &Identifier{Value: `x`},
				Value:  &Identifier{Value: ``},
			},
			errstr: `Value`,
		},
	}
	for _, tt := range tests {
		_ = suite.Run(tt.testName, func() {
			err := tt.a.Validate()
			if tt.errstr == `` {
				suite.Require().NoError(err)
			} else {
				suite.Require().ErrorContains(err, tt.errstr)
			}
		})
	}
}

func (suite *AssignmentSuite) TestStatementNode() {
	a := &Assignment{
		Target: &Identifier{Value: `x`},
		Value:  newIntLiteral(1),
	}
	a.statementNode()
}
