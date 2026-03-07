package ast

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestIdentifierSuite(t *testing.T) {
	suite.Run(t, new(IdentifierSuite))
}

type IdentifierSuite struct {
	suite.Suite
}

func (suite *IdentifierSuite) TestString() {
	tests := []struct {
		testName string
		value    string
		expected string
	}{
		{
			testName: `simple identifier`,
			value:    `x`,
			expected: `x`,
		},
		{
			testName: `longer identifier`,
			value:    `myVariable`,
			expected: `myVariable`,
		},
	}
	for _, tt := range tests {
		_ = suite.Run(tt.testName, func() {
			i := &Identifier{Value: tt.value}
			suite.Equal(tt.expected, i.String())
		})
	}
}

func (suite *IdentifierSuite) TestASCII() {
	tests := []struct {
		testName string
		value    string
		expected string
	}{
		{
			testName: `simple identifier`,
			value:    `x`,
			expected: `x`,
		},
		{
			testName: `longer identifier`,
			value:    `myVariable`,
			expected: `myVariable`,
		},
	}
	for _, tt := range tests {
		_ = suite.Run(tt.testName, func() {
			i := &Identifier{Value: tt.value}
			suite.Equal(tt.expected, i.ASCII())
		})
	}
}

func (suite *IdentifierSuite) TestValidate() {
	tests := []struct {
		testName string
		i        *Identifier
		errstr   string
	}{
		// OK.
		{
			testName: `valid identifier`,
			i:        &Identifier{Value: `x`},
		},
		{
			testName: `valid longer identifier`,
			i:        &Identifier{Value: `myVariable`},
		},

		// Errors.
		{
			testName: `error missing value`,
			i:        &Identifier{},
			errstr:   `Value`,
		},
		{
			testName: `error empty value`,
			i:        &Identifier{Value: ``},
			errstr:   `Value`,
		},
	}
	for _, tt := range tests {
		_ = suite.Run(tt.testName, func() {
			err := tt.i.Validate()
			if tt.errstr == `` {
				suite.Require().NoError(err)
			} else {
				suite.Require().ErrorContains(err, tt.errstr)
			}
		})
	}
}

func (suite *IdentifierSuite) TestExpressionNode() {
	// Verify that Identifier implements the expressionNode interface method.
	i := &Identifier{Value: `x`}
	// This should compile and not panic.
	i.expressionNode()
}
