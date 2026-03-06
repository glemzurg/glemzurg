package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestRecordInstanceSuite(t *testing.T) {
	suite.Run(t, new(RecordInstanceSuite))
}

type RecordInstanceSuite struct {
	suite.Suite
}

func (suite *RecordInstanceSuite) TestString() {
	tests := []struct {
		testName string
		bindings []*FieldBinding
		expected string
	}{
		{
			testName: `single field`,
			bindings: []*FieldBinding{
				{
					Field:      &Identifier{Value: `a`},
					Expression: NewIntLiteral(1),
				},
			},
			expected: `[a ↦ 1]`,
		},
		{
			testName: `two fields`,
			bindings: []*FieldBinding{
				{
					Field:      &Identifier{Value: `a`},
					Expression: NewIntLiteral(1),
				},
				{
					Field:      &Identifier{Value: `b`},
					Expression: NewIntLiteral(2),
				},
			},
			expected: `[a ↦ 1, b ↦ 2]`,
		},
		{
			testName: `three fields`,
			bindings: []*FieldBinding{
				{
					Field:      &Identifier{Value: `a`},
					Expression: NewIntLiteral(1),
				},
				{
					Field:      &Identifier{Value: `b`},
					Expression: NewIntLiteral(2),
				},
				{
					Field:      &Identifier{Value: `c`},
					Expression: NewIntLiteral(3),
				},
			},
			expected: `[a ↦ 1, b ↦ 2, c ↦ 3]`,
		},
		{
			testName: `with string values`,
			bindings: []*FieldBinding{
				{
					Field:      &Identifier{Value: `name`},
					Expression: &StringLiteral{Value: `Alice`},
				},
				{
					Field:      &Identifier{Value: `city`},
					Expression: &StringLiteral{Value: `Boston`},
				},
			},
			expected: `[name ↦ "Alice", city ↦ "Boston"]`,
		},
		{
			testName: `with identifier values`,
			bindings: []*FieldBinding{
				{
					Field:      &Identifier{Value: `val`},
					Expression: &Identifier{Value: `x`},
				},
				{
					Field:      &Identifier{Value: `rdy`},
					Expression: &Identifier{Value: `y`},
				},
			},
			expected: `[val ↦ x, rdy ↦ y]`,
		},
		{
			testName: `with expression values`,
			bindings: []*FieldBinding{
				{
					Field: &Identifier{Value: `sum`},
					Expression: &RealInfixExpression{
						Left:     NewIntLiteral(1),
						Operator: RealOperatorAdd,
						Right:    NewIntLiteral(2),
					},
				},
			},
			expected: `[sum ↦ 1 + 2]`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			r := &RecordInstance{
				Bindings: tt.bindings,
			}
			assert.Equal(t, tt.expected, r.String())
		})
	}
}

func (suite *RecordInstanceSuite) TestAscii() {
	tests := []struct {
		testName string
		bindings []*FieldBinding
		expected string
	}{
		{
			testName: `single field`,
			bindings: []*FieldBinding{
				{
					Field:      &Identifier{Value: `a`},
					Expression: NewIntLiteral(1),
				},
			},
			expected: `[a |-> 1]`,
		},
		{
			testName: `three fields`,
			bindings: []*FieldBinding{
				{
					Field:      &Identifier{Value: `a`},
					Expression: NewIntLiteral(1),
				},
				{
					Field:      &Identifier{Value: `b`},
					Expression: NewIntLiteral(2),
				},
				{
					Field:      &Identifier{Value: `c`},
					Expression: NewIntLiteral(3),
				},
			},
			expected: `[a |-> 1, b |-> 2, c |-> 3]`,
		},
		{
			testName: `with division operator`,
			bindings: []*FieldBinding{
				{
					Field: &Identifier{Value: `ratio`},
					Expression: &RealInfixExpression{
						Left:     NewIntLiteral(10),
						Operator: RealOperatorDivide,
						Right:    NewIntLiteral(2),
					},
				},
			},
			expected: `[ratio |-> 10 \div 2]`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			r := &RecordInstance{
				Bindings: tt.bindings,
			}
			assert.Equal(t, tt.expected, r.Ascii())
		})
	}
}

func (suite *RecordInstanceSuite) TestValidate() {
	tests := []struct {
		testName string
		r        *RecordInstance
		errstr   string
	}{
		// OK.
		{
			testName: `valid single binding`,
			r: &RecordInstance{
				Bindings: []*FieldBinding{
					{
						Field:      &Identifier{Value: `a`},
						Expression: NewIntLiteral(1),
					},
				},
			},
		},
		{
			testName: `valid multiple bindings`,
			r: &RecordInstance{
				Bindings: []*FieldBinding{
					{
						Field:      &Identifier{Value: `a`},
						Expression: NewIntLiteral(1),
					},
					{
						Field:      &Identifier{Value: `b`},
						Expression: NewIntLiteral(2),
					},
					{
						Field:      &Identifier{Value: `c`},
						Expression: NewIntLiteral(3),
					},
				},
			},
		},

		// Errors.
		{
			testName: `error missing bindings`,
			r:        &RecordInstance{},
			errstr:   `Bindings`,
		},
		{
			testName: `error empty bindings`,
			r: &RecordInstance{
				Bindings: []*FieldBinding{},
			},
			errstr: `Bindings`,
		},
		{
			testName: `error nil binding`,
			r: &RecordInstance{
				Bindings: []*FieldBinding{nil},
			},
			errstr: `Bindings[0]`,
		},
		{
			testName: `error nil field in binding`,
			r: &RecordInstance{
				Bindings: []*FieldBinding{
					{
						Field:      nil,
						Expression: NewIntLiteral(1),
					},
				},
			},
			errstr: `Bindings[0].Field`,
		},
		{
			testName: `error empty field name`,
			r: &RecordInstance{
				Bindings: []*FieldBinding{
					{
						Field:      &Identifier{Value: ``},
						Expression: NewIntLiteral(1),
					},
				},
			},
			errstr: `Value`,
		},
		{
			testName: `error nil expression in binding`,
			r: &RecordInstance{
				Bindings: []*FieldBinding{
					{
						Field:      &Identifier{Value: `a`},
						Expression: nil,
					},
				},
			},
			errstr: `Bindings[0].Expression`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.r.Validate()
			if tt.errstr == `` {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

func (suite *RecordInstanceSuite) TestExpressionNode() {
	// Verify that RecordInstance implements the expressionNode interface method.
	r := &RecordInstance{
		Bindings: []*FieldBinding{
			{
				Field:      &Identifier{Value: `a`},
				Expression: NewIntLiteral(1),
			},
		},
	}
	// This should compile and not panic.
	r.expressionNode()
}

