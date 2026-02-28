package model_spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_expression_type"
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
	// Valid construction.
	spec, err := NewTypeSpec("tla_plus", "Nat", nil)
	s.NoError(err)
	s.Equal("tla_plus", spec.Notation)
	s.Equal("Nat", spec.Specification)
	s.Nil(spec.ExpressionType)

	// With expression type.
	spec, err = NewTypeSpec("tla_plus", "Nat", &model_expression_type.IntegerType{})
	s.NoError(err)
	s.NotNil(spec.ExpressionType)

	// Invalid notation.
	_, err = NewTypeSpec("", "Nat", nil)
	s.Error(err)
	s.Contains(err.Error(), "Notation")
}
