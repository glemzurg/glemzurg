package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestBagInfixSuite(t *testing.T) {
	suite.Run(t, new(BagInfixSuite))
}

type BagInfixSuite struct {
	suite.Suite
}

// Helper to create a BuiltinCall that returns a Bag (simulating _Bags!SetToBag)
// Uses Identifier as a placeholder argument since SetLiteralInt doesn't implement Expression
func makeBagCall(name string) *BuiltinCall {
	return &BuiltinCall{
		Name: "_Bags!SetToBag",
		Args: []Expression{
			&Identifier{Value: name},
		},
	}
}

func (suite *BagInfixSuite) TestString() {
	tests := []struct {
		testName string
		operator string
		left     Expression
		right    Expression
		expected string
	}{
		{
			testName: `union operator`,
			operator: BagOperatorUnion,
			left:     makeBagCall("set1"),
			right:    makeBagCall("set2"),
			expected: `_Bags!SetToBag(set1) ⊕ _Bags!SetToBag(set2)`,
		},
		{
			testName: `subtraction operator`,
			operator: BagOperatorSubtraction,
			left:     makeBagCall("set1"),
			right:    makeBagCall("set2"),
			expected: `_Bags!SetToBag(set1) ⊖ _Bags!SetToBag(set2)`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			bi := &BagInfix{
				Operator: tt.operator,
				Left:     tt.left,
				Right:    tt.right,
			}
			assert.Equal(t, tt.expected, bi.String())
		})
	}
}

func (suite *BagInfixSuite) TestAscii() {
	tests := []struct {
		testName string
		operator string
		left     Expression
		right    Expression
		expected string
	}{
		{
			testName: `union operator`,
			operator: BagOperatorUnion,
			left:     makeBagCall("set1"),
			right:    makeBagCall("set2"),
			expected: `_Bags!SetToBag(set1) (+) _Bags!SetToBag(set2)`,
		},
		{
			testName: `subtraction operator`,
			operator: BagOperatorSubtraction,
			left:     makeBagCall("set1"),
			right:    makeBagCall("set2"),
			expected: `_Bags!SetToBag(set1) (-) _Bags!SetToBag(set2)`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			bi := &BagInfix{
				Operator: tt.operator,
				Left:     tt.left,
				Right:    tt.right,
			}
			assert.Equal(t, tt.expected, bi.Ascii())
		})
	}
}

func (suite *BagInfixSuite) TestValidate() {
	tests := []struct {
		testName string
		bi       *BagInfix
		errstr   string
	}{
		// OK.
		{
			testName: `valid union operator`,
			bi: &BagInfix{
				Operator: BagOperatorUnion,
				Left:     makeBagCall("set1"),
				Right:    makeBagCall("set2"),
			},
		},
		{
			testName: `valid subtraction operator`,
			bi: &BagInfix{
				Operator: BagOperatorSubtraction,
				Left:     makeBagCall("set1"),
				Right:    makeBagCall("set2"),
			},
		},

		// Errors.
		{
			testName: `error missing operator`,
			bi: &BagInfix{
				Left:  makeBagCall("set1"),
				Right: makeBagCall("set2"),
			},
			errstr: `Operator`,
		},
		{
			testName: `error invalid operator`,
			bi: &BagInfix{
				Operator: `invalid`,
				Left:     makeBagCall("set1"),
				Right:    makeBagCall("set2"),
			},
			errstr: `Operator`,
		},
		{
			testName: `error missing left`,
			bi: &BagInfix{
				Operator: BagOperatorUnion,
				Right:    makeBagCall("set2"),
			},
			errstr: `Left`,
		},
		{
			testName: `error missing right`,
			bi: &BagInfix{
				Operator: BagOperatorUnion,
				Left:     makeBagCall("set1"),
			},
			errstr: `Right`,
		},
		{
			testName: `error invalid left bag`,
			bi: &BagInfix{
				Operator: BagOperatorUnion,
				Left:     &BuiltinCall{Name: "_Bags!SetToBag", Args: nil},
				Right:    makeBagCall("set2"),
			},
			errstr: `Args`,
		},
		{
			testName: `error invalid right bag`,
			bi: &BagInfix{
				Operator: BagOperatorUnion,
				Left:     makeBagCall("set1"),
				Right:    &BuiltinCall{Name: "_Bags!SetToBag", Args: nil},
			},
			errstr: `Args`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.bi.Validate()
			if tt.errstr == `` {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

func (suite *BagInfixSuite) TestExpressionNode() {
	// Verify that BagInfix implements the expressionNode interface method.
	bi := &BagInfix{
		Operator: BagOperatorUnion,
		Left:     makeBagCall("set1"),
		Right:    makeBagCall("set2"),
	}
	// This should compile and not panic.
	bi.expressionNode()
}
