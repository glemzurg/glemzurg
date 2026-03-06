package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
			value:    NewIntLiteral(0),
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
					NewIntLiteral(1),
					NewIntLiteral(2),
					NewIntLiteral(3),
				},
			},
			expected: `items' = ⟨1, 2, 3⟩`,
		},
		{
			testName: `assign if-else expression to state`,
			target:   &Identifier{Value: `result`},
			value: &ExpressionIfElse{
				Condition: &BooleanLiteral{Value: true},
				Then:      NewIntLiteral(1),
				Else:      NewIntLiteral(0),
			},
			expected: `result' = IF TRUE THEN 1 ELSE 0`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			stmt := &Assignment{
				Target: tt.target,
				Value:  tt.value,
			}
			assert.Equal(t, tt.expected, stmt.String())
		})
	}
}

func (suite *AssignmentSuite) TestAscii() {
	tests := []struct {
		testName string
		target   *Identifier
		value    Expression
		expected string
	}{
		{
			testName: `assign integer to state`,
			target:   &Identifier{Value: `count`},
			value:    NewIntLiteral(42),
			expected: `count' = 42`,
		},
		{
			testName: `assign tuple to state`,
			target:   &Identifier{Value: `items`},
			value: &TupleLiteral{
				Elements: []Expression{
					NewIntLiteral(1),
					NewIntLiteral(2),
				},
			},
			expected: `items' = <<1, 2>>`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			stmt := &Assignment{
				Target: tt.target,
				Value:  tt.value,
			}
			assert.Equal(t, tt.expected, stmt.Ascii())
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
				Value:  NewIntLiteral(1),
			},
		},

		// Errors.
		{
			testName: `error missing target`,
			a: &Assignment{
				Value: NewIntLiteral(1),
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
				Value:  NewIntLiteral(1),
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
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.a.Validate()
			if tt.errstr == `` {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

func (suite *AssignmentSuite) TestStatementNode() {
	a := &Assignment{
		Target: &Identifier{Value: `x`},
		Value:  NewIntLiteral(1),
	}
	a.statementNode()
}
