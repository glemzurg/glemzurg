package ast

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSetConstantSuite(t *testing.T) {
	suite.Run(t, new(SetConstantSuite))
}

type SetConstantSuite struct {
	suite.Suite
}

func (suite *SetConstantSuite) TestString() {
	tests := []struct {
		testName string
		value    string
		expected string
	}{
		{
			testName: `boolean set`,
			value:    SetConstantBoolean,
			expected: `BOOLEAN`,
		},
		{
			testName: `nat set`,
			value:    SetConstantNat,
			expected: `Nat`,
		},
		{
			testName: `int set`,
			value:    SetConstantInt,
			expected: `Int`,
		},
		{
			testName: `real set`,
			value:    SetConstantReal,
			expected: `Real`,
		},
	}
	for _, tt := range tests {
		_ = suite.Run(tt.testName, func() {
			s := &SetConstant{Value: tt.value}
			suite.Equal(tt.expected, s.String())
		})
	}
}

func (suite *SetConstantSuite) TestASCII() {
	tests := []struct {
		testName string
		value    string
		expected string
	}{
		{
			testName: `boolean set`,
			value:    SetConstantBoolean,
			expected: `BOOLEAN`,
		},
		{
			testName: `nat set`,
			value:    SetConstantNat,
			expected: `Nat`,
		},
		{
			testName: `int set`,
			value:    SetConstantInt,
			expected: `Int`,
		},
		{
			testName: `real set`,
			value:    SetConstantReal,
			expected: `Real`,
		},
	}
	for _, tt := range tests {
		_ = suite.Run(tt.testName, func() {
			s := &SetConstant{Value: tt.value}
			suite.Equal(tt.expected, s.ASCII())
		})
	}
}

func (suite *SetConstantSuite) TestValidate() {
	tests := []struct {
		testName string
		s        *SetConstant
		errstr   string
	}{
		// OK.
		{
			testName: `valid boolean set`,
			s: &SetConstant{
				Value: SetConstantBoolean,
			},
		},
		{
			testName: `valid nat set`,
			s: &SetConstant{
				Value: SetConstantNat,
			},
		},
		{
			testName: `valid int set`,
			s: &SetConstant{
				Value: SetConstantInt,
			},
		},
		{
			testName: `valid real set`,
			s: &SetConstant{
				Value: SetConstantReal,
			},
		},

		// Errors.
		{
			testName: `error missing value`,
			s:        &SetConstant{},
			errstr:   `Value`,
		},
		{
			testName: `error invalid value`,
			s: &SetConstant{
				Value: `INVALID`,
			},
			errstr: `Value`,
		},
	}
	for _, tt := range tests {
		_ = suite.Run(tt.testName, func() {
			err := tt.s.Validate()
			if tt.errstr == `` {
				suite.Require().NoError(err)
			} else {
				suite.Require().ErrorContains(err, tt.errstr)
			}
		})
	}
}

func (suite *SetConstantSuite) TestExpressionNode() {
	// Verify that SetConstant implements the expressionNode interface method.
	s := &SetConstant{Value: SetConstantBoolean}
	// This should compile and not panic.
	s.expressionNode()
}
