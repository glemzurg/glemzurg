package model_spec

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_expression"
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
				Expression: &model_expression.RationalLiteral{}, // nil Value
			},
			errstr: "ExpressionSpec.Expression",
		},
	}
	for _, tt := range tests {
		s.Run(tt.testName, func() {
			err := tt.spec.Validate()
			if tt.errstr == "" {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
				s.Contains(err.Error(), tt.errstr)
			}
		})
	}
}

func (s *ExpressionSpecTestSuite) TestNew() {
	// Valid construction with nil parseFunc.
	spec, err := NewExpressionSpec("tla_plus", "x > 0", nil)
	s.Require().NoError(err)
	s.Equal("tla_plus", spec.Notation)
	s.Equal("x > 0", spec.Specification)
	s.Nil(spec.Expression)
	s.False(spec.ParseOk())

	// Valid construction with parseFunc that succeeds.
	parseFunc := func(spec string) (model_expression.Expression, string) {
		return &model_expression.BoolLiteral{Value: true}, "TRUE"
	}
	spec, err = NewExpressionSpec("tla_plus", "true", parseFunc)
	s.Require().NoError(err)
	s.Equal("TRUE", spec.Specification) // Normalized.
	s.NotNil(spec.Expression)
	s.True(spec.ParseOk())

	// Valid construction with parseFunc that fails (returns nil).
	failFunc := func(spec string) (model_expression.Expression, string) {
		return nil, ""
	}
	spec, err = NewExpressionSpec("tla_plus", "invalid", failFunc)
	s.Require().NoError(err)
	s.Equal("invalid", spec.Specification) // Unchanged.
	s.Nil(spec.Expression)
	s.False(spec.ParseOk())

	// Empty specification skips parseFunc.
	called := false
	trackFunc := func(spec string) (model_expression.Expression, string) {
		called = true
		return nil, ""
	}
	spec, err = NewExpressionSpec("tla_plus", "", trackFunc)
	s.Require().NoError(err)
	s.False(called)
	s.False(spec.ParseOk())
}
