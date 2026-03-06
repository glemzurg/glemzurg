package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestSetLiteralEnumSuite(t *testing.T) {
	suite.Run(t, new(SetLiteralEnumSuite))
}

type SetLiteralEnumSuite struct {
	suite.Suite
}

func (suite *SetLiteralEnumSuite) TestString() {
	tests := []struct {
		testName string
		values   []string
		expected string
	}{
		{
			testName: `single value`,
			values:   []string{`active`},
			expected: `{"active"}`,
		},
		{
			testName: `multiple values`,
			values:   []string{`pending`, `active`, `completed`},
			expected: `{"pending", "active", "completed"}`,
		},
		{
			testName: `two values`,
			values:   []string{`yes`, `no`},
			expected: `{"yes", "no"}`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			s := &SetLiteralEnum{
				Values: tt.values,
			}
			assert.Equal(t, tt.expected, s.String())
		})
	}
}

func (suite *SetLiteralEnumSuite) TestAscii() {
	tests := []struct {
		testName string
		values   []string
		expected string
	}{
		{
			testName: `single value`,
			values:   []string{`active`},
			expected: `{"active"}`,
		},
		{
			testName: `multiple values`,
			values:   []string{`pending`, `active`, `completed`},
			expected: `{"pending", "active", "completed"}`,
		},
		{
			testName: `two values`,
			values:   []string{`yes`, `no`},
			expected: `{"yes", "no"}`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			s := &SetLiteralEnum{
				Values: tt.values,
			}
			assert.Equal(t, tt.expected, s.Ascii())
		})
	}
}

func (suite *SetLiteralEnumSuite) TestValidate() {
	tests := []struct {
		testName string
		s        *SetLiteralEnum
		errstr   string
	}{
		// OK.
		{
			testName: `valid single value`,
			s: &SetLiteralEnum{
				Values: []string{`active`},
			},
		},
		{
			testName: `valid multiple values`,
			s: &SetLiteralEnum{
				Values: []string{`pending`, `active`, `completed`},
			},
		},

		// Errors.
		{
			testName: `error missing values`,
			s:        &SetLiteralEnum{},
			errstr:   `Values`,
		},
		{
			testName: `error empty values`,
			s: &SetLiteralEnum{
				Values: []string{},
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

func (suite *SetLiteralEnumSuite) TestExpressionNode() {
	// Verify that SetLiteralEnum implements the expressionNode interface method.
	s := &SetLiteralEnum{
		Values: []string{`active`, `inactive`},
	}
	// This should compile and not panic.
	s.expressionNode()
}
