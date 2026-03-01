package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestStringLiteralSuite(t *testing.T) {
	suite.Run(t, new(StringLiteralSuite))
}

type StringLiteralSuite struct {
	suite.Suite
}

func (suite *StringLiteralSuite) TestString() {
	tests := []struct {
		testName string
		value    string
		expected string
	}{
		{
			testName: `simple string`,
			value:    `hello`,
			expected: `"hello"`,
		},
		{
			testName: `empty string`,
			value:    ``,
			expected: `""`,
		},
		{
			testName: `string with spaces`,
			value:    `hello world`,
			expected: `"hello world"`,
		},
		// Escaped characters: \\ \" \n \r \t \f
		{
			testName: `escaped backslash`,
			value:    `hello\\world`,
			expected: `"hello\\world"`,
		},
		{
			testName: `escaped quote`,
			value:    `hello\"world`,
			expected: `"hello\"world"`,
		},
		{
			testName: `escaped newline`,
			value:    "hello\nworld",
			expected: "\"hello\nworld\"",
		},
		{
			testName: `escaped carriage return`,
			value:    "hello\rworld",
			expected: "\"hello\rworld\"",
		},
		{
			testName: `escaped tab`,
			value:    "hello\tworld",
			expected: "\"hello\tworld\"",
		},
		{
			testName: `escaped form feed`,
			value:    "hello\fworld",
			expected: "\"hello\fworld\"",
		},
		{
			testName: `multiple escaped characters`,
			value:    "line1\nline2\ttab\\slash",
			expected: "\"line1\nline2\ttab\\slash\"",
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			s := &StringLiteral{Value: tt.value}
			assert.Equal(t, tt.expected, s.String())
		})
	}
}

func (suite *StringLiteralSuite) TestAscii() {
	tests := []struct {
		testName string
		value    string
		expected string
	}{
		{
			testName: `simple string`,
			value:    `hello`,
			expected: `"hello"`,
		},
		{
			testName: `empty string`,
			value:    ``,
			expected: `""`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			s := &StringLiteral{Value: tt.value}
			assert.Equal(t, tt.expected, s.Ascii())
		})
	}
}

func (suite *StringLiteralSuite) TestValidate() {
	tests := []struct {
		testName string
		s        *StringLiteral
		errstr   string
	}{
		// OK.
		{
			testName: `valid string`,
			s:        &StringLiteral{Value: `hello`},
		},
		{
			testName: `valid empty string`,
			s:        &StringLiteral{Value: ``},
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.s.Validate()
			if tt.errstr == `` {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

func (suite *StringLiteralSuite) TestExpressionNode() {
	// Verify that StringLiteral implements the expressionNode interface method.
	s := &StringLiteral{Value: `hello`}
	// This should compile and not panic.
	s.expressionNode()
}

