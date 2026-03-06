package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			i := &Identifier{Value: tt.value}
			assert.Equal(t, tt.expected, i.String())
		})
	}
}

func (suite *IdentifierSuite) TestAscii() {
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
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			i := &Identifier{Value: tt.value}
			assert.Equal(t, tt.expected, i.Ascii())
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
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.i.Validate()
			if tt.errstr == `` {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
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
