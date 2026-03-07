package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestPrimedSuite(t *testing.T) {
	suite.Run(t, new(PrimedSuite))
}

type PrimedSuite struct {
	suite.Suite
}

func (suite *PrimedSuite) TestString() {
	tests := []struct {
		testName string
		base     Expression
		expected string
	}{
		{
			testName: "simple identifier primed",
			base:     &Identifier{Value: "x"},
			expected: "x'",
		},
		{
			testName: "longer identifier primed",
			base:     &Identifier{Value: "counter"},
			expected: "counter'",
		},
		{
			testName: "field access primed",
			base: &FieldAccess{
				Identifier: &Identifier{Value: "record"},
				Member:     "field",
			},
			expected: "record.field'",
		},
	}
	for _, tt := range tests {
		_ = suite.Run(tt.testName, func() {
			t := suite.T()
			p := &Primed{Base: tt.base}
			assert.Equal(t, tt.expected, p.String())
		})
	}
}

func (suite *PrimedSuite) TestASCII() {
	p := &Primed{Base: &Identifier{Value: "x"}}
	suite.Equal("x'", p.ASCII())
}

func (suite *PrimedSuite) TestValidate() {
	tests := []struct {
		testName string
		p        *Primed
		errstr   string
	}{
		// OK.
		{
			testName: "valid primed identifier",
			p:        &Primed{Base: &Identifier{Value: "x"}},
		},
		{
			testName: "valid primed field access",
			p: &Primed{
				Base: &FieldAccess{
					Identifier: &Identifier{Value: "record"},
					Member:     "field",
				},
			},
		},

		// Errors.
		{
			testName: "error nil base",
			p:        &Primed{Base: nil},
			errstr:   "Base",
		},
		{
			testName: "error invalid base identifier",
			p:        &Primed{Base: &Identifier{Value: ""}},
			errstr:   "Value",
		},
	}
	for _, tt := range tests {
		_ = suite.Run(tt.testName, func() {
			err := tt.p.Validate()
			if tt.errstr == "" {
				suite.Require().NoError(err)
			} else {
				suite.Require().ErrorContains(err, tt.errstr)
			}
		})
	}
}

func (suite *PrimedSuite) TestExpressionNode() {
	p := &Primed{Base: &Identifier{Value: "x"}}
	// This should compile and not panic.
	p.expressionNode()
}
