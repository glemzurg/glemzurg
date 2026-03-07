package ast

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestTupleLiteralSuite(t *testing.T) {
	suite.Run(t, new(TupleLiteralSuite))
}

type TupleLiteralSuite struct {
	suite.Suite
}

func (suite *TupleLiteralSuite) TestString() {
	tests := []struct {
		testName string
		elements []Expression
		expected string
	}{
		{
			testName: `empty tuple`,
			elements: []Expression{},
			expected: `⟨⟩`,
		},
		{
			testName: `single element`,
			elements: []Expression{
				NewIntLiteral(3),
			},
			expected: `⟨3⟩`,
		},
		{
			testName: `two elements`,
			elements: []Expression{
				NewIntLiteral(3),
				NewIntLiteral(7),
			},
			expected: `⟨3, 7⟩`,
		},
		{
			testName: `three elements`,
			elements: []Expression{
				NewIntLiteral(3),
				NewIntLiteral(7),
				NewIntLiteral(3),
			},
			expected: `⟨3, 7, 3⟩`,
		},
		{
			testName: `with string values`,
			elements: []Expression{
				&StringLiteral{Value: `hello`},
				&StringLiteral{Value: `world`},
			},
			expected: `⟨"hello", "world"⟩`,
		},
		{
			testName: `with identifier values`,
			elements: []Expression{
				&Identifier{Value: `x`},
				&Identifier{Value: `y`},
				&Identifier{Value: `z`},
			},
			expected: `⟨x, y, z⟩`,
		},
		{
			testName: `with mixed values`,
			elements: []Expression{
				NewIntLiteral(1),
				&StringLiteral{Value: `two`},
				&Identifier{Value: `three`},
			},
			expected: `⟨1, "two", three⟩`,
		},
		{
			testName: `nested tuple`,
			elements: []Expression{
				&TupleLiteral{
					Elements: []Expression{
						NewIntLiteral(1),
						NewIntLiteral(2),
					},
				},
				NewIntLiteral(3),
			},
			expected: `⟨⟨1, 2⟩, 3⟩`,
		},
	}
	for _, tt := range tests {
		_ = suite.Run(tt.testName, func() {
			tup := &TupleLiteral{
				Elements: tt.elements,
			}
			suite.Equal(tt.expected, tup.String())
		})
	}
}

func (suite *TupleLiteralSuite) TestASCII() {
	tests := []struct {
		testName string
		elements []Expression
		expected string
	}{
		{
			testName: `empty tuple`,
			elements: []Expression{},
			expected: `<<>>`,
		},
		{
			testName: `three elements`,
			elements: []Expression{
				NewIntLiteral(3),
				NewIntLiteral(7),
				NewIntLiteral(3),
			},
			expected: `<<3, 7, 3>>`,
		},
		{
			testName: `with real infix`,
			elements: []Expression{
				&RealInfixExpression{
					Left:     NewIntLiteral(1),
					Operator: RealOperatorAdd,
					Right:    NewIntLiteral(2),
				},
			},
			expected: `<<1 + 2>>`,
		},
	}
	for _, tt := range tests {
		_ = suite.Run(tt.testName, func() {
			tup := &TupleLiteral{
				Elements: tt.elements,
			}
			suite.Equal(tt.expected, tup.ASCII())
		})
	}
}

func (suite *TupleLiteralSuite) TestValidate() {
	tests := []struct {
		testName string
		t        *TupleLiteral
		errstr   string
	}{
		// OK.
		{
			testName: `valid empty tuple`,
			t:        &TupleLiteral{Elements: []Expression{}},
		},
		{
			testName: `valid nil elements slice`,
			t:        &TupleLiteral{Elements: nil},
		},
		{
			testName: `valid single element`,
			t: &TupleLiteral{
				Elements: []Expression{
					NewIntLiteral(1),
				},
			},
		},
		{
			testName: `valid multiple elements`,
			t: &TupleLiteral{
				Elements: []Expression{
					NewIntLiteral(1),
					NewIntLiteral(2),
					NewIntLiteral(3),
				},
			},
		},

		// Errors.
		{
			testName: `error nil element`,
			t: &TupleLiteral{
				Elements: []Expression{nil},
			},
			errstr: `elements[0]`,
		},
		{
			testName: `error nil element in middle`,
			t: &TupleLiteral{
				Elements: []Expression{
					NewIntLiteral(1),
					nil,
					NewIntLiteral(3),
				},
			},
			errstr: `elements[1]`,
		},
		{
			testName: `error invalid element`,
			t: &TupleLiteral{
				Elements: []Expression{
					&Identifier{Value: ``},
				},
			},
			errstr: `Value`,
		},
	}
	for _, tt := range tests {
		_ = suite.Run(tt.testName, func() {
			err := tt.t.Validate()
			if tt.errstr == `` {
				suite.Require().NoError(err)
			} else {
				suite.Require().ErrorContains(err, tt.errstr)
			}
		})
	}
}

func (suite *TupleLiteralSuite) TestExpressionNode() {
	// Verify that TupleLiteral implements the expressionNode interface method.
	t := &TupleLiteral{
		Elements: []Expression{
			NewIntLiteral(1),
		},
	}
	// This should compile and not panic.
	t.expressionNode()
}
