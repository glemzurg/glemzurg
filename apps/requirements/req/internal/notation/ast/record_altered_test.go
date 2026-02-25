package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestRecordAlteredSuite(t *testing.T) {
	suite.Run(t, new(RecordAlteredSuite))
}

type RecordAlteredSuite struct {
	suite.Suite
}

func (suite *RecordAlteredSuite) TestString() {
	tests := []struct {
		testName    string
		identifier  string
		alterations []*FieldAlteration
		expected    string
	}{
		{
			testName:   `single field alteration`,
			identifier: `chan`,
			alterations: []*FieldAlteration{
				{
					Field:      &FieldIdentifier{Identifier: nil, Member: `val`},
					Expression: &Identifier{Value: `d`},
				},
			},
			expected: `[chan EXCEPT !.val = d]`,
		},
		{
			testName:   `two field alterations`,
			identifier: `chan`,
			alterations: []*FieldAlteration{
				{
					Field:      &FieldIdentifier{Identifier: nil, Member: `val`},
					Expression: &Identifier{Value: `d`},
				},
				{
					Field:      &FieldIdentifier{Identifier: nil, Member: `rdy`},
					Expression: NewIntLiteral(1),
				},
			},
			expected: `[chan EXCEPT !.val = d, !.rdy = 1]`,
		},
		{
			testName:   `with existing value @`,
			identifier: `chan`,
			alterations: []*FieldAlteration{
				{
					Field:      &FieldIdentifier{Identifier: nil, Member: `val`},
					Expression: &Identifier{Value: `d`},
				},
				{
					Field:      &FieldIdentifier{Identifier: nil, Member: `rdy`},
					Expression: &ExistingValue{},
				},
			},
			expected: `[chan EXCEPT !.val = d, !.rdy = @]`,
		},
		{
			testName:   `three field alterations`,
			identifier: `record`,
			alterations: []*FieldAlteration{
				{
					Field:      &FieldIdentifier{Identifier: nil, Member: `a`},
					Expression: NewIntLiteral(1),
				},
				{
					Field:      &FieldIdentifier{Identifier: nil, Member: `b`},
					Expression: NewIntLiteral(2),
				},
				{
					Field:      &FieldIdentifier{Identifier: nil, Member: `c`},
					Expression: NewIntLiteral(3),
				},
			},
			expected: `[record EXCEPT !.a = 1, !.b = 2, !.c = 3]`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			r := &RecordAltered{
				Identifier:  &Identifier{Value: tt.identifier},
				Alterations: tt.alterations,
			}
			assert.Equal(t, tt.expected, r.String())
		})
	}
}

func (suite *RecordAlteredSuite) TestAscii() {
	tests := []struct {
		testName    string
		identifier  string
		alterations []*FieldAlteration
		expected    string
	}{
		{
			testName:   `single field alteration`,
			identifier: `chan`,
			alterations: []*FieldAlteration{
				{
					Field:      &FieldIdentifier{Identifier: nil, Member: `val`},
					Expression: &Identifier{Value: `d`},
				},
			},
			expected: `[chan EXCEPT !.val = d]`,
		},
		{
			testName:   `with division operator`,
			identifier: `rec`,
			alterations: []*FieldAlteration{
				{
					Field: &FieldIdentifier{Identifier: nil, Member: `x`},
					Expression: &RealInfixExpression{
						Left:     NewIntLiteral(10),
						Operator: RealOperatorDivide,
						Right:    NewIntLiteral(2),
					},
				},
			},
			expected: `[rec EXCEPT !.x = 10 \div 2]`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			r := &RecordAltered{
				Identifier:  &Identifier{Value: tt.identifier},
				Alterations: tt.alterations,
			}
			assert.Equal(t, tt.expected, r.Ascii())
		})
	}
}

func (suite *RecordAlteredSuite) TestValidate() {
	tests := []struct {
		testName string
		r        *RecordAltered
		errstr   string
	}{
		// OK.
		{
			testName: `valid single alteration`,
			r: &RecordAltered{
				Identifier: &Identifier{Value: `chan`},
				Alterations: []*FieldAlteration{
					{
						Field:      &FieldIdentifier{Identifier: nil, Member: `val`},
						Expression: &Identifier{Value: `d`},
					},
				},
			},
		},
		{
			testName: `valid multiple alterations`,
			r: &RecordAltered{
				Identifier: &Identifier{Value: `chan`},
				Alterations: []*FieldAlteration{
					{
						Field:      &FieldIdentifier{Identifier: nil, Member: `val`},
						Expression: &Identifier{Value: `d`},
					},
					{
						Field:      &FieldIdentifier{Identifier: nil, Member: `rdy`},
						Expression: NewIntLiteral(1),
					},
				},
			},
		},

		// Errors.
		{
			testName: `error missing identifier`,
			r: &RecordAltered{
				Alterations: []*FieldAlteration{
					{
						Field:      &FieldIdentifier{Identifier: nil, Member: `val`},
						Expression: &Identifier{Value: `d`},
					},
				},
			},
			errstr: `Identifier`,
		},
		{
			testName: `error empty identifier`,
			r: &RecordAltered{
				Identifier: &Identifier{Value: ``},
				Alterations: []*FieldAlteration{
					{
						Field:      &FieldIdentifier{Identifier: nil, Member: `val`},
						Expression: &Identifier{Value: `d`},
					},
				},
			},
			errstr: `Value`,
		},
		{
			testName: `error missing alterations`,
			r: &RecordAltered{
				Identifier: &Identifier{Value: `chan`},
			},
			errstr: `Alterations`,
		},
		{
			testName: `error empty alterations`,
			r: &RecordAltered{
				Identifier:  &Identifier{Value: `chan`},
				Alterations: []*FieldAlteration{},
			},
			errstr: `Alterations`,
		},
		{
			testName: `error nil alteration`,
			r: &RecordAltered{
				Identifier: &Identifier{Value: `chan`},
				Alterations: []*FieldAlteration{
					nil,
				},
			},
			errstr: `Alterations[0]`,
		},
		{
			testName: `error nil field in alteration`,
			r: &RecordAltered{
				Identifier: &Identifier{Value: `chan`},
				Alterations: []*FieldAlteration{
					{
						Field:      nil,
						Expression: &Identifier{Value: `d`},
					},
				},
			},
			errstr: `Alterations[0].Field`,
		},
		{
			testName: `error nil expression in alteration`,
			r: &RecordAltered{
				Identifier: &Identifier{Value: `chan`},
				Alterations: []*FieldAlteration{
					{
						Field:      &FieldIdentifier{Identifier: nil, Member: `val`},
						Expression: nil,
					},
				},
			},
			errstr: `Alterations[0].Expression`,
		},
		{
			testName: `error missing member in field`,
			r: &RecordAltered{
				Identifier: &Identifier{Value: `chan`},
				Alterations: []*FieldAlteration{
					{
						Field:      &FieldIdentifier{Identifier: nil, Member: ``},
						Expression: &Identifier{Value: `d`},
					},
				},
			},
			errstr: `Member`,
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

func (suite *RecordAlteredSuite) TestExpressionNode() {
	// Verify that RecordAltered implements the expressionNode interface method.
	r := &RecordAltered{
		Identifier: &Identifier{Value: `chan`},
		Alterations: []*FieldAlteration{
			{
				Field:      &FieldIdentifier{Identifier: nil, Member: `val`},
				Expression: &Identifier{Value: `d`},
			},
		},
	}
	// This should compile and not panic.
	r.expressionNode()
}
