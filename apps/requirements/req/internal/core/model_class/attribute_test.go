package model_class

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/suite"
)

func TestAttributeSuite(t *testing.T) {
	suite.Run(t, new(AttributeSuite))
}

type AttributeSuite struct {
	suite.Suite
}

// TestValidate tests all validation rules for Attribute.
func (suite *AttributeSuite) TestValidate() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	validKey := helper.Must(identity.NewAttributeKey(classKey, "attr1"))
	derivKey := helper.Must(identity.NewAttributeDerivationKey(validKey, "deriv1"))

	validDerivationPolicy := model_logic.NewLogic(derivKey, model_logic.LogicTypeValue, "Computed from other fields.", "", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil)
	wrongKindDerivationPolicy := model_logic.NewLogic(derivKey, model_logic.LogicTypeAssessment, "Computed from other fields.", "", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil)

	// Invariant keys and logic objects.
	invKey1 := helper.Must(identity.NewAttributeInvariantKey(validKey, "0"))
	invKey2 := helper.Must(identity.NewAttributeInvariantKey(validKey, "1"))
	validInvariant := model_logic.NewLogic(invKey1, model_logic.LogicTypeAssessment, "Must be positive.", "", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil)
	validInvariant2 := model_logic.NewLogic(invKey2, model_logic.LogicTypeAssessment, "Must be less than 100.", "", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil)
	wrongKindInvariant := model_logic.NewLogic(invKey1, model_logic.LogicTypeStateChange, "Should fail.", "target", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil)

	tests := []struct {
		testName  string
		attribute Attribute
		errstr    string
	}{
		{
			testName: "valid attribute",
			attribute: Attribute{
				Key:  validKey,
				Name: "Name",
			},
		},
		{
			testName: "error empty key",
			attribute: Attribute{
				Key:  identity.Key{},
				Name: "Name",
			},
			errstr: "key type is required",
		},
		{
			testName: "error wrong key type",
			attribute: Attribute{
				Key:  domainKey,
				Name: "Name",
			},
			errstr: "key: invalid key type 'domain' for attribute",
		},
		{
			testName: "error blank name",
			attribute: Attribute{
				Key:  validKey,
				Name: "",
			},
			errstr: "Name",
		},
		{
			testName: "valid with DerivationPolicy",
			attribute: Attribute{
				Key:              validKey,
				Name:             "Name",
				DerivationPolicy: &validDerivationPolicy,
			},
		},
		{
			testName: "valid without DerivationPolicy",
			attribute: Attribute{
				Key:  validKey,
				Name: "Name",
			},
		},
		{
			testName: "error invalid DerivationPolicy missing key",
			attribute: Attribute{
				Key:  validKey,
				Name: "Name",
				DerivationPolicy: &model_logic.Logic{
					Key:         identity.Key{},
					Type:        model_logic.LogicTypeValue,
					Description: "Computed from other fields.",
					Spec:        logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus},
				},
			},
			errstr: "DerivationPolicy",
		},
		{
			testName: "error DerivationPolicy wrong kind",
			attribute: Attribute{
				Key:              validKey,
				Name:             "Name",
				DerivationPolicy: &wrongKindDerivationPolicy,
			},
			errstr: "DerivationPolicy logic kind must be 'value'",
		},
		{
			testName: "valid with nil invariants",
			attribute: Attribute{
				Key:        validKey,
				Name:       "Name",
				Invariants: nil,
			},
		},
		{
			testName: "valid with single invariant",
			attribute: Attribute{
				Key:        validKey,
				Name:       "Name",
				Invariants: []model_logic.Logic{validInvariant},
			},
		},
		{
			testName: "valid with multiple invariants",
			attribute: Attribute{
				Key:        validKey,
				Name:       "Name",
				Invariants: []model_logic.Logic{validInvariant, validInvariant2},
			},
		},
		{
			testName: "error invariant wrong logic type",
			attribute: Attribute{
				Key:        validKey,
				Name:       "Name",
				Invariants: []model_logic.Logic{wrongKindInvariant},
			},
			errstr: "logic kind must be 'assessment' or 'let'",
		},
		{
			testName: "error invariant invalid logic missing key",
			attribute: Attribute{
				Key:  validKey,
				Name: "Name",
				Invariants: []model_logic.Logic{
					{
						Key:         identity.Key{},
						Type:        model_logic.LogicTypeAssessment,
						Description: "Missing key.",
						Spec:        logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus},
					},
				},
			},
			errstr: "attribute invariant 0",
		},
		{
			testName: "valid with let in invariants",
			attribute: Attribute{
				Key:  validKey,
				Name: "Name",
				Invariants: []model_logic.Logic{
					model_logic.NewLogic(invKey1, model_logic.LogicTypeLet, "Local total.", "total", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "1 + 2"}, nil),
					validInvariant,
				},
			},
		},
		{
			testName: "error duplicate let target in attribute invariants",
			attribute: Attribute{
				Key:  validKey,
				Name: "Name",
				Invariants: []model_logic.Logic{
					model_logic.NewLogic(invKey1, model_logic.LogicTypeLet, "Local a.", "a", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "1"}, nil),
					model_logic.NewLogic(invKey2, model_logic.LogicTypeLet, "Local a again.", "a", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "2"}, nil),
				},
			},
			errstr: "duplicate let target \"a\"",
		},
	}
	for _, tt := range tests {
		suite.Run(tt.testName, func() {
			err := tt.attribute.Validate()
			if tt.errstr == "" {
				suite.Require().NoError(err)
			} else {
				suite.Require().ErrorContains(err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewAttribute maps parameters correctly and calls Validate.
func (suite *AttributeSuite) TestNew() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	key := helper.Must(identity.NewAttributeKey(classKey, "attr1"))
	derivKey := helper.Must(identity.NewAttributeDerivationKey(key, "deriv1"))

	derivationPolicyVal := model_logic.NewLogic(derivKey, model_logic.LogicTypeValue, "Computed from other fields.", "", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil)
	derivationPolicy := &derivationPolicyVal

	// Test parameters are mapped correctly.
	attr, err := NewAttribute(key, "Name", "Details", "DataTypeRules", derivationPolicy, true,
		AttributeAnnotations{UmlComment: "UmlComment", IndexNums: []uint{1, 2}})
	suite.Require().NoError(err)
	suite.Equal(Attribute{
		Key:              key,
		Name:             "Name",
		Details:          "Details",
		DataTypeRules:    "DataTypeRules",
		DerivationPolicy: derivationPolicy,
		Nullable:         true,
		UmlComment:       "UmlComment",
		IndexNums:        []uint{1, 2},
	}, attr)

	// Test with nil DerivationPolicy (non-derived attribute).
	attrNoDeriv, err := NewAttribute(key, "Name", "Details", "DataTypeRules", nil, true,
		AttributeAnnotations{UmlComment: "UmlComment", IndexNums: []uint{1, 2}})
	suite.Require().NoError(err)
	suite.Nil(attrNoDeriv.DerivationPolicy)

	// Test parseable data type rules result in DataType being set.
	attrParsedKey := helper.Must(identity.NewAttributeKey(classKey, "attrparsed"))
	derivParsedKey := helper.Must(identity.NewAttributeDerivationKey(attrParsedKey, "deriv_parsed"))
	derivParsedPolicyVal := model_logic.NewLogic(derivParsedKey, model_logic.LogicTypeValue, "Computed from other fields.", "", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil)
	derivParsedPolicy := &derivParsedPolicyVal
	attrParsed, err := NewAttribute(attrParsedKey, "NameParsed", "Details", "unconstrained", derivParsedPolicy, true,
		AttributeAnnotations{UmlComment: "UmlComment", IndexNums: []uint{1, 2}})
	suite.Require().NoError(err)
	suite.Equal(Attribute{
		Key:              attrParsedKey,
		Name:             "NameParsed",
		Details:          "Details",
		DataTypeRules:    "unconstrained",
		DerivationPolicy: derivParsedPolicy,
		Nullable:         true,
		UmlComment:       "UmlComment",
		IndexNums:        []uint{1, 2},
		DataType: &model_data_type.DataType{
			Key:            attrParsedKey.String(),
			CollectionType: "atomic",
			Atomic: &model_data_type.Atomic{
				ConstraintType: "unconstrained",
			},
		},
	}, attrParsed)
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *AttributeSuite) TestValidateWithParent() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	validKey := helper.Must(identity.NewAttributeKey(classKey, "attr1"))
	otherClassKey := helper.Must(identity.NewClassKey(subdomainKey, "other_class"))
	derivKey := helper.Must(identity.NewAttributeDerivationKey(validKey, "deriv1"))

	// Test that Validate is called.
	attr := Attribute{
		Key:  validKey,
		Name: "", // Invalid
	}
	err := attr.ValidateWithParent(&classKey)
	suite.Require().ErrorContains(err, "Name", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - attribute key has class1 as parent, but we pass other_class.
	attr = Attribute{
		Key:  validKey,
		Name: "Name",
	}
	err = attr.ValidateWithParent(&otherClassKey)
	suite.Require().ErrorContains(err, "does not match expected parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = attr.ValidateWithParent(&classKey)
	suite.Require().NoError(err)

	// Test valid with derivation policy.
	validDerivPolicy := model_logic.NewLogic(derivKey, model_logic.LogicTypeValue, "Computed from other fields.", "", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil)
	attr = Attribute{
		Key:              validKey,
		Name:             "Name",
		DerivationPolicy: &validDerivPolicy,
	}
	err = attr.ValidateWithParent(&classKey)
	suite.Require().NoError(err)

	// Test derivation policy key validation - wrong parent should fail.
	otherAttrKey := helper.Must(identity.NewAttributeKey(classKey, "other_attr"))
	wrongDerivKey := helper.Must(identity.NewAttributeDerivationKey(otherAttrKey, "deriv1"))
	wrongParentDerivPolicy := model_logic.NewLogic(wrongDerivKey, model_logic.LogicTypeValue, "Computed from other fields.", "", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil)
	attr = Attribute{
		Key:              validKey,
		Name:             "Name",
		DerivationPolicy: &wrongParentDerivPolicy,
	}
	err = attr.ValidateWithParent(&classKey)
	suite.Require().ErrorContains(err, "DerivationPolicy", "ValidateWithParent should validate derivation policy key parent")

	// Test valid with invariants.
	invKey := helper.Must(identity.NewAttributeInvariantKey(validKey, "0"))
	validInvariant := model_logic.NewLogic(invKey, model_logic.LogicTypeAssessment, "Must be positive.", "", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil)
	attr = Attribute{
		Key:        validKey,
		Name:       "Name",
		Invariants: []model_logic.Logic{validInvariant},
	}
	err = attr.ValidateWithParent(&classKey)
	suite.Require().NoError(err)

	// Test invariant with wrong parent key - invariant key has other_attr as parent, but attribute key is validKey.
	wrongInvKey := helper.Must(identity.NewAttributeInvariantKey(otherAttrKey, "0"))
	wrongParentInvariant := model_logic.NewLogic(wrongInvKey, model_logic.LogicTypeAssessment, "Wrong parent.", "", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil)
	attr = Attribute{
		Key:        validKey,
		Name:       "Name",
		Invariants: []model_logic.Logic{wrongParentInvariant},
	}
	err = attr.ValidateWithParent(&classKey)
	suite.Require().ErrorContains(err, "attribute invariant 0", "ValidateWithParent should validate invariant key parent")
}
