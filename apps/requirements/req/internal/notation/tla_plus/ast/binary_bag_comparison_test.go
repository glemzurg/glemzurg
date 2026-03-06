package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestLogicInfixBagSuite(t *testing.T) {
	suite.Run(t, new(LogicInfixBagSuite))
}

type LogicInfixBagSuite struct {
	suite.Suite
}

// Helper to create a BuiltinCall that returns a Bag (simulating _Bags!SetToBag)
// Uses Identifier as a placeholder argument since SetLiteralInt doesn't implement Expression
func makeLogicBagCall(name string) *BuiltinCall {
	return &BuiltinCall{
		Name: "_Bags!SetToBag",
		Args: []Expression{
			&Identifier{Value: name},
		},
	}
}

func (suite *LogicInfixBagSuite) TestString() {
	tests := []struct {
		testName string
		operator string
		left     Expression
		right    Expression
		expected string
	}{
		{
			testName: `subbag operator`,
			operator: LogicBagOperatorSubBag,
			left:     makeLogicBagCall("set1"),
			right:    makeLogicBagCall("set2"),
			expected: `_Bags!SetToBag(set1) ⊑ _Bags!SetToBag(set2)`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			ib := &LogicInfixBag{
				Operator: tt.operator,
				Left:     tt.left,
				Right:    tt.right,
			}
			assert.Equal(t, tt.expected, ib.String())
		})
	}
}

func (suite *LogicInfixBagSuite) TestAscii() {
	tests := []struct {
		testName string
		operator string
		left     Expression
		right    Expression
		expected string
	}{
		{
			testName: `subbag operator`,
			operator: LogicBagOperatorSubBag,
			left:     makeLogicBagCall("set1"),
			right:    makeLogicBagCall("set2"),
			expected: `_Bags!SetToBag(set1) \sqsubseteq _Bags!SetToBag(set2)`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			ib := &LogicInfixBag{
				Operator: tt.operator,
				Left:     tt.left,
				Right:    tt.right,
			}
			assert.Equal(t, tt.expected, ib.Ascii())
		})
	}
}

func (suite *LogicInfixBagSuite) TestValidate() {
	tests := []struct {
		testName string
		ib       *LogicInfixBag
		errstr   string
	}{
		// OK.
		{
			testName: `valid subbag operator`,
			ib: &LogicInfixBag{
				Operator: LogicBagOperatorSubBag,
				Left:     makeLogicBagCall("set1"),
				Right:    makeLogicBagCall("set2"),
			},
		},

		// Errors.
		{
			testName: `error missing operator`,
			ib: &LogicInfixBag{
				Left:  makeLogicBagCall("set1"),
				Right: makeLogicBagCall("set2"),
			},
			errstr: `Operator`,
		},
		{
			testName: `error invalid operator`,
			ib: &LogicInfixBag{
				Operator: `invalid`,
				Left:     makeLogicBagCall("set1"),
				Right:    makeLogicBagCall("set2"),
			},
			errstr: `Operator`,
		},
		{
			testName: `error missing left`,
			ib: &LogicInfixBag{
				Operator: LogicBagOperatorSubBag,
				Right:    makeLogicBagCall("set2"),
			},
			errstr: `Left`,
		},
		{
			testName: `error missing right`,
			ib: &LogicInfixBag{
				Operator: LogicBagOperatorSubBag,
				Left:     makeLogicBagCall("set1"),
			},
			errstr: `Right`,
		},
		{
			testName: `error invalid left bag`,
			ib: &LogicInfixBag{
				Operator: LogicBagOperatorSubBag,
				Left:     &BuiltinCall{Name: "_Bags!SetToBag", Args: nil},
				Right:    makeLogicBagCall("set2"),
			},
			errstr: `Args`,
		},
		{
			testName: `error invalid right bag`,
			ib: &LogicInfixBag{
				Operator: LogicBagOperatorSubBag,
				Left:     makeLogicBagCall("set1"),
				Right:    &BuiltinCall{Name: "_Bags!SetToBag", Args: nil},
			},
			errstr: `Args`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.ib.Validate()
			if tt.errstr == `` {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

func (suite *LogicInfixBagSuite) TestExpressionNode() {
	// Verify that LogicInfixBag implements the expressionNode interface method.
	ib := &LogicInfixBag{
		Operator: LogicBagOperatorSubBag,
		Left:     makeLogicBagCall("set1"),
		Right:    makeLogicBagCall("set2"),
	}
	// This should compile and not panic.
	ib.expressionNode()
}
