package model_spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_expression_type"
)

type TypeSpecTestSuite struct {
	suite.Suite
}

func TestTypeSpecSuite(t *testing.T) {
	suite.Run(t, new(TypeSpecTestSuite))
}

func (s *TypeSpecTestSuite) TestValidate() {
	tests := []struct {
		testName string
		spec     TypeSpec
		errstr   string
	}{
		{
			testName: "valid minimal",
			spec: TypeSpec{
				Notation: "tla_plus",
			},
		},
		{
			testName: "valid with specification",
			spec: TypeSpec{
				Notation:      "tla_plus",
				Specification: "Nat",
			},
		},
		{
			testName: "valid with expression type",
			spec: TypeSpec{
				Notation:       "tla_plus",
				ExpressionType: &model_expression_type.IntegerType{},
			},
		},
		{
			testName: "valid with both specification and expression type",
			spec: TypeSpec{
				Notation:       "tla_plus",
				Specification:  "Nat",
				ExpressionType: &model_expression_type.IntegerType{},
			},
		},
		{
			testName: "error missing notation",
			spec: TypeSpec{
				Notation: "",
			},
			errstr: "Notation",
		},
		{
			testName: "error invalid notation",
			spec: TypeSpec{
				Notation: "Alloy",
			},
			errstr: "Notation",
		},
		{
			testName: "error invalid expression type",
			spec: TypeSpec{
				Notation:       "tla_plus",
				ExpressionType: &model_expression_type.SetType{}, // Missing ElementType.
			},
			errstr: "TypeSpec.ExpressionType",
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

func (s *TypeSpecTestSuite) TestNew() {
	// Valid construction with nil parseFunc.
	spec, err := NewTypeSpec("tla_plus", "Nat", nil)
	s.NoError(err)
	s.Equal("tla_plus", spec.Notation)
	s.Equal("Nat", spec.Specification)
	s.Nil(spec.ExpressionType)
	s.False(spec.ParseOk())

	// With parseFunc that succeeds.
	parseFunc := func(spec string) (model_expression_type.ExpressionType, string) {
		return &model_expression_type.IntegerType{}, "Int"
	}
	spec, err = NewTypeSpec("tla_plus", "Nat", parseFunc)
	s.NoError(err)
	s.NotNil(spec.ExpressionType)
	s.True(spec.ParseOk())
	s.Equal("Int", spec.Specification) // Normalized.

	// With parseFunc that fails (returns nil).
	failFunc := func(spec string) (model_expression_type.ExpressionType, string) {
		return nil, ""
	}
	spec, err = NewTypeSpec("tla_plus", "Nat", failFunc)
	s.NoError(err)
	s.Nil(spec.ExpressionType)
	s.False(spec.ParseOk())
	s.Equal("Nat", spec.Specification) // Unchanged.

	// Invalid notation.
	_, err = NewTypeSpec("", "Nat", nil)
	s.Error(err)
	s.Contains(err.Error(), "Notation")

	// Empty specification skips parseFunc.
	called := false
	trackFunc := func(spec string) (model_expression_type.ExpressionType, string) {
		called = true
		return nil, ""
	}
	spec, err = NewTypeSpec("tla_plus", "", trackFunc)
	s.NoError(err)
	s.False(called)
	s.False(spec.ParseOk())
}
