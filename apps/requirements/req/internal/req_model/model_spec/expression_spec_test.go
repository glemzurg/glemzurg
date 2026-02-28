package model_spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_expression"
)

type ExpressionSpecTestSuite struct {
	suite.Suite
}

func TestExpressionSpecSuite(t *testing.T) {
	suite.Run(t, new(ExpressionSpecTestSuite))
}

func (s *ExpressionSpecTestSuite) TestValidate() {
	tests := []struct {
		testName string
		spec     ExpressionSpec
		errstr   string
	}{
		{
			testName: "valid minimal",
			spec: ExpressionSpec{
				Notation: "tla_plus",
			},
		},
		{
			testName: "valid with specification",
			spec: ExpressionSpec{
				Notation:      "tla_plus",
				Specification: "x > 0",
			},
		},
		{
			testName: "valid with expression",
			spec: ExpressionSpec{
				Notation:   "tla_plus",
				Expression: &model_expression.BoolLiteral{Value: true},
			},
		},
		{
			testName: "valid with both specification and expression",
			spec: ExpressionSpec{
				Notation:      "tla_plus",
				Specification: "TRUE",
				Expression:    &model_expression.BoolLiteral{Value: true},
			},
		},
		{
			testName: "error missing notation",
			spec: ExpressionSpec{
				Notation: "",
			},
			errstr: "Notation",
		},
		{
			testName: "error invalid notation",
			spec: ExpressionSpec{
				Notation: "Z",
			},
			errstr: "Notation",
		},
		{
			testName: "error invalid expression",
			spec: ExpressionSpec{
				Notation:   "tla_plus",
				Expression: &model_expression.RationalLiteral{Numerator: 1, Denominator: 0},
			},
			errstr: "ExpressionSpec.Expression",
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.testName, func(t *testing.T) {
			err := tt.spec.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errstr)
			}
		})
	}
}

func (s *ExpressionSpecTestSuite) TestNew() {
	// Valid construction.
	spec, err := NewExpressionSpec("tla_plus", "x > 0", nil)
	s.NoError(err)
	s.Equal("tla_plus", spec.Notation)
	s.Equal("x > 0", spec.Specification)
	s.Nil(spec.Expression)

	// Invalid notation.
	_, err = NewExpressionSpec("", "x > 0", nil)
	s.Error(err)
	s.Contains(err.Error(), "Notation")
}
