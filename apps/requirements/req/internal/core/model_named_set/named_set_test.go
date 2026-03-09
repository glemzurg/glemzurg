package model_named_set

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/suite"
)

type NamedSetTestSuite struct {
	suite.Suite
}

func TestNamedSetSuite(t *testing.T) {
	suite.Run(t, new(NamedSetTestSuite))
}

// validSpec returns a valid ExpressionSpec for testing.
func validSpec() model_spec.ExpressionSpec {
	return model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}
}

// validSpecWithBody returns a valid ExpressionSpec with a specification body.
func validSpecWithBody(body string) model_spec.ExpressionSpec {
	return model_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: body}
}

// TestValidate tests all validation rules for NamedSet.
func (s *NamedSetTestSuite) TestValidate() {
	validKey := helper.Must(identity.NewNamedSetKey("valid_statuses"))
	validKey2 := helper.Must(identity.NewNamedSetKey("order_types"))

	tests := []struct {
		testName string
		ns       NamedSet
		errstr   string
	}{
		{
			testName: "valid minimal",
			ns: NamedSet{
				Key:  validKey,
				Name: "Valid Statuses",
				Spec: validSpec(),
			},
		},
		{
			testName: "valid with all fields",
			ns: NamedSet{
				Key:         validKey,
				Name:        "Valid Statuses",
				Description: "The set of all valid order statuses.",
				Spec:        validSpecWithBody(`{"pending", "active", "complete"}`),
				TypeSpec: &model_spec.TypeSpec{
					Notation:      model_logic.NotationTLAPlus,
					Specification: "SUBSET STRING",
				},
			},
		},
		{
			testName: "valid with description no type spec",
			ns: NamedSet{
				Key:         validKey2,
				Name:        "Order Types",
				Description: "Types of orders.",
				Spec:        validSpecWithBody(`{"standard", "express"}`),
			},
		},
		{
			testName: "error empty key",
			ns: NamedSet{
				Key:  identity.Key{},
				Name: "Valid Statuses",
				Spec: validSpec(),
			},
			errstr: "KeyType",
		},
		{
			testName: "error wrong key type",
			ns: NamedSet{
				Key:  helper.Must(identity.NewInvariantKey("0")),
				Name: "Valid Statuses",
				Spec: validSpec(),
			},
			errstr: "invalid key type",
		},
		{
			testName: "error blank name",
			ns: NamedSet{
				Key:  validKey,
				Name: "",
				Spec: validSpec(),
			},
			errstr: "Name",
		},
		{
			testName: "error missing spec notation",
			ns: NamedSet{
				Key:  validKey,
				Name: "Valid Statuses",
				Spec: model_spec.ExpressionSpec{},
			},
			errstr: "Notation",
		},
		{
			testName: "error invalid spec notation",
			ns: NamedSet{
				Key:  validKey,
				Name: "Valid Statuses",
				Spec: model_spec.ExpressionSpec{Notation: "Z"},
			},
			errstr: "Notation",
		},
		{
			testName: "error invalid type spec notation",
			ns: NamedSet{
				Key:  validKey,
				Name: "Valid Statuses",
				Spec: validSpec(),
				TypeSpec: &model_spec.TypeSpec{
					Notation: "invalid",
				},
			},
			errstr: "Notation",
		},
	}
	for _, tt := range tests {
		s.Run(tt.testName, func() {
			err := tt.ns.Validate()
			if tt.errstr == "" {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
				s.Contains(err.Error(), tt.errstr)
			}
		})
	}
}

// TestNew tests that NewNamedSet maps parameters correctly and calls Validate.
func (s *NamedSetTestSuite) TestNew() {
	validKey := helper.Must(identity.NewNamedSetKey("valid_statuses"))

	spec := validSpecWithBody(`{"pending", "active", "complete"}`)
	typeSpec := &model_spec.TypeSpec{
		Notation:      model_logic.NotationTLAPlus,
		Specification: "SUBSET STRING",
	}

	// Test all parameters are mapped correctly.
	ns := NewNamedSet(validKey, "Valid Statuses", "The valid statuses.", spec, typeSpec)
	s.Equal(NamedSet{
		Key:         validKey,
		Name:        "Valid Statuses",
		Description: "The valid statuses.",
		Spec:        spec,
		TypeSpec:    typeSpec,
	}, ns)

	// Test with nil optional fields (Description and TypeSpec are optional).
	ns = NewNamedSet(validKey, "Valid Statuses", "", validSpec(), nil)
	s.Equal("Valid Statuses", ns.Name)
	s.Empty(ns.Description)
	s.Nil(ns.TypeSpec)
}

// TestValidateWithParent tests that ValidateWithParent validates the key's parent relationship.
func (s *NamedSetTestSuite) TestValidateWithParent() {
	validKey := helper.Must(identity.NewNamedSetKey("valid_statuses"))

	// Test valid case - nset key is root-level (nil parent).
	ns := NamedSet{
		Key:  validKey,
		Name: "Valid Statuses",
		Spec: validSpec(),
	}
	err := ns.ValidateWithParent()
	s.Require().NoError(err)

	// Test that Validate is called.
	ns = NamedSet{
		Key:  validKey,
		Name: "", // Invalid: blank name
		Spec: validSpec(),
	}
	err = ns.ValidateWithParent()
	s.Require().Error(err)
	s.Contains(err.Error(), "Name")
}
