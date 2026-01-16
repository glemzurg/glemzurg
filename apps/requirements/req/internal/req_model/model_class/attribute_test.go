package model_class

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_data_type"
	"github.com/stretchr/testify/assert"
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
			errstr: "keyType: cannot be blank",
		},
		{
			testName: "error wrong key type",
			attribute: Attribute{
				Key:  domainKey,
				Name: "Name",
			},
			errstr: "Key: invalid key type 'domain' for attribute.",
		},
		{
			testName: "error blank name",
			attribute: Attribute{
				Key:  validKey,
				Name: "",
			},
			errstr: "Name: cannot be blank",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.attribute.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
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

	// Test parameters are mapped correctly.
	attr, err := NewAttribute(key, "Name", "Details", "DataTypeRules", "DerivationPolicy", true, "UmlComment", []uint{1, 2})
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), key, attr.Key)
	assert.Equal(suite.T(), "Name", attr.Name)
	assert.Equal(suite.T(), "Details", attr.Details)
	assert.Equal(suite.T(), "DataTypeRules", attr.DataTypeRules)
	assert.Equal(suite.T(), "DerivationPolicy", attr.DerivationPolicy)
	assert.Equal(suite.T(), true, attr.Nullable)
	assert.Equal(suite.T(), "UmlComment", attr.UmlComment)
	assert.Equal(suite.T(), []uint{1, 2}, attr.IndexNums)

	// Test parseable data type rules result in DataType being set.
	attrParsedKey := helper.Must(identity.NewAttributeKey(classKey, "attrparsed"))
	attrParsed, err := NewAttribute(attrParsedKey, "NameParsed", "Details", "unconstrained", "DerivationPolicy", true, "UmlComment", []uint{1, 2})
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), attrParsed.DataType)
	assert.Equal(suite.T(), &model_data_type.DataType{
		Key:            attrParsedKey.String(),
		CollectionType: "atomic",
		Atomic: &model_data_type.Atomic{
			ConstraintType: "unconstrained",
		},
	}, attrParsed.DataType)

	// Test that Validate is called (invalid data should fail).
	_, err = NewAttribute(key, "", "Details", "DataTypeRules", "DerivationPolicy", true, "UmlComment", nil)
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *AttributeSuite) TestValidateWithParent() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	validKey := helper.Must(identity.NewAttributeKey(classKey, "attr1"))
	otherClassKey := helper.Must(identity.NewClassKey(subdomainKey, "other_class"))

	// Test that Validate is called.
	attr := Attribute{
		Key:  validKey,
		Name: "", // Invalid
	}
	err := attr.ValidateWithParent(&classKey)
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - attribute key has class1 as parent, but we pass other_class.
	attr = Attribute{
		Key:  validKey,
		Name: "Name",
	}
	err = attr.ValidateWithParent(&otherClassKey)
	assert.ErrorContains(suite.T(), err, "does not match expected parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = attr.ValidateWithParent(&classKey)
	assert.NoError(suite.T(), err)
}
