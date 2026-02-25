package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestBooleanLiteralSuite(t *testing.T) {
	suite.Run(t, new(BooleanLiteralSuite))
}

type BooleanLiteralSuite struct {
	suite.Suite
}

func (suite *BooleanLiteralSuite) TestString() {
	tests := []struct {
		testName string
		value    bool
		expected string
	}{
		{
			testName: `true value`,
			value:    true,
			expected: `TRUE`,
		},
		{
			testName: `false value`,
			value:    false,
			expected: `FALSE`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			b := &BooleanLiteral{Value: tt.value}
			assert.Equal(t, tt.expected, b.String())
		})
	}
}

func (suite *BooleanLiteralSuite) TestAscii() {
	tests := []struct {
		testName string
		value    bool
		expected string
	}{
		{
			testName: `true value`,
			value:    true,
			expected: `TRUE`,
		},
		{
			testName: `false value`,
			value:    false,
			expected: `FALSE`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			b := &BooleanLiteral{Value: tt.value}
			assert.Equal(t, tt.expected, b.Ascii())
		})
	}
}

func (suite *BooleanLiteralSuite) TestValidate() {
	tests := []struct {
		testName string
		b        *BooleanLiteral
		errstr   string
	}{
		// OK.
		{
			testName: `valid true`,
			b:        &BooleanLiteral{Value: true},
		},
		{
			testName: `valid false`,
			b:        &BooleanLiteral{Value: false},
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.b.Validate()
			if tt.errstr == `` {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

func (suite *BooleanLiteralSuite) TestExpressionNode() {
	// Verify that BooleanLiteral implements the expressionNode interface method.
	b := &BooleanLiteral{Value: true}
	// This should compile and not panic.
	b.expressionNode()
}
