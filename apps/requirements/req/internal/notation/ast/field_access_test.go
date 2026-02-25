package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestFieldIdentifierSuite(t *testing.T) {
	suite.Run(t, new(FieldIdentifierSuite))
}

type FieldIdentifierSuite struct {
	suite.Suite
}

func (suite *FieldIdentifierSuite) TestString() {
	tests := []struct {
		testName   string
		identifier *Identifier
		member     string
		expected   string
	}{
		{
			testName:   `simple field access`,
			identifier: &Identifier{Value: `foo`},
			member:     `bar`,
			expected:   `foo.bar`,
		},
		{
			testName:   `record field`,
			identifier: &Identifier{Value: `record`},
			member:     `field`,
			expected:   `record.field`,
		},
		{
			testName:   `state field`,
			identifier: &Identifier{Value: `state`},
			member:     `value`,
			expected:   `state.value`,
		},
		{
			testName:   `nil identifier outputs !`,
			identifier: nil,
			member:     `bar`,
			expected:   `!.bar`,
		},
		{
			testName:   `nil identifier with longer member`,
			identifier: nil,
			member:     `fieldName`,
			expected:   `!.fieldName`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			f := &FieldIdentifier{
				Identifier: tt.identifier,
				Member:     tt.member,
			}
			assert.Equal(t, tt.expected, f.String())
		})
	}
}

func (suite *FieldIdentifierSuite) TestAscii() {
	tests := []struct {
		testName   string
		identifier *Identifier
		member     string
		expected   string
	}{
		{
			testName:   `simple field access`,
			identifier: &Identifier{Value: `foo`},
			member:     `bar`,
			expected:   `foo.bar`,
		},
		{
			testName:   `nil identifier outputs !`,
			identifier: nil,
			member:     `bar`,
			expected:   `!.bar`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			f := &FieldIdentifier{
				Identifier: tt.identifier,
				Member:     tt.member,
			}
			assert.Equal(t, tt.expected, f.Ascii())
		})
	}
}

func (suite *FieldIdentifierSuite) TestValidate() {
	tests := []struct {
		testName string
		f        *FieldIdentifier
		errstr   string
	}{
		// OK.
		{
			testName: `valid field access`,
			f: &FieldIdentifier{
				Identifier: &Identifier{Value: `foo`},
				Member:     `bar`,
			},
		},
		{
			testName: `valid single char names`,
			f: &FieldIdentifier{
				Identifier: &Identifier{Value: `x`},
				Member:     `y`,
			},
		},
		{
			testName: `valid nil identifier`,
			f: &FieldIdentifier{
				Identifier: nil,
				Member:     `bar`,
			},
		},

		// Errors.
		{
			testName: `error missing member`,
			f: &FieldIdentifier{
				Identifier: &Identifier{Value: `foo`},
			},
			errstr: `Member`,
		},
		{
			testName: `error empty identifier value`,
			f: &FieldIdentifier{
				Identifier: &Identifier{Value: ``},
				Member:     `bar`,
			},
			errstr: `Value`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.f.Validate()
			if tt.errstr == `` {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

func (suite *FieldIdentifierSuite) TestExpressionNode() {
	// Verify that FieldIdentifier implements the expressionNode interface method.
	f := &FieldIdentifier{
		Identifier: &Identifier{Value: `foo`},
		Member:     `bar`,
	}
	// This should compile and not panic.
	f.expressionNode()
}
