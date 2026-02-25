package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestSetLiteralIntSuite(t *testing.T) {
	suite.Run(t, new(SetLiteralIntSuite))
}

type SetLiteralIntSuite struct {
	suite.Suite
}

func (suite *SetLiteralIntSuite) TestString() {
	tests := []struct {
		testName string
		values   []int
		expected string
	}{
		{
			testName: `single value`,
			values:   []int{42},
			expected: `{42}`,
		},
		{
			testName: `multiple values`,
			values:   []int{1, 2, 4, 6},
			expected: `{1, 2, 4, 6}`,
		},
		{
			testName: `negative values`,
			values:   []int{-5, 0, 5},
			expected: `{-5, 0, 5}`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			s := &SetLiteralInt{
				Values: tt.values,
			}
			assert.Equal(t, tt.expected, s.String())
		})
	}
}

func (suite *SetLiteralIntSuite) TestAscii() {
	tests := []struct {
		testName string
		values   []int
		expected string
	}{
		{
			testName: `single value`,
			values:   []int{42},
			expected: `{42}`,
		},
		{
			testName: `multiple values`,
			values:   []int{1, 2, 4, 6},
			expected: `{1, 2, 4, 6}`,
		},
		{
			testName: `negative values`,
			values:   []int{-5, 0, 5},
			expected: `{-5, 0, 5}`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			s := &SetLiteralInt{
				Values: tt.values,
			}
			assert.Equal(t, tt.expected, s.Ascii())
		})
	}
}

func (suite *SetLiteralIntSuite) TestValidate() {
	tests := []struct {
		testName string
		s        *SetLiteralInt
		errstr   string
	}{
		// OK.
		{
			testName: `valid single value`,
			s: &SetLiteralInt{
				Values: []int{42},
			},
		},
		{
			testName: `valid multiple values`,
			s: &SetLiteralInt{
				Values: []int{1, 2, 4, 6},
			},
		},

		// Errors.
		{
			testName: `error missing values`,
			s:        &SetLiteralInt{},
			errstr:   `Values`,
		},
		{
			testName: `error empty values`,
			s: &SetLiteralInt{
				Values: []int{},
			},
			errstr: `Values`,
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

func (suite *SetLiteralIntSuite) TestExpressionNode() {
	// Verify that SetLiteralInt implements the expressionNode interface method.
	s := &SetLiteralInt{
		Values: []int{1, 2, 4, 6},
	}
	// This should compile and not panic.
	s.expressionNode()
}
