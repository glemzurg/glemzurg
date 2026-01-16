package model_class

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestClassSuite(t *testing.T) {
	suite.Run(t, new(ClassSuite))
}

type ClassSuite struct {
	suite.Suite
}

// TestValidate tests all validation rules for Class.
func (suite *ClassSuite) TestValidate() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	validKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))

	tests := []struct {
		testName string
		class    Class
		errstr   string
	}{
		{
			testName: "valid class",
			class: Class{
				Key:  validKey,
				Name: "Name",
			},
		},
		{
			testName: "error empty key",
			class: Class{
				Key:  identity.Key{},
				Name: "Name",
			},
			errstr: "keyType: cannot be blank",
		},
		{
			testName: "error wrong key type",
			class: Class{
				Key:  domainKey,
				Name: "Name",
			},
			errstr: "Key: invalid key type 'domain' for class.",
		},
		{
			testName: "error blank name",
			class: Class{
				Key:  validKey,
				Name: "",
			},
			errstr: "Name: cannot be blank",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.class.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewClass maps parameters correctly and calls Validate.
func (suite *ClassSuite) TestNew() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	key := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	actorKey := helper.Must(identity.NewActorKey("actor1"))
	generalizationKey := helper.Must(identity.NewGeneralizationKey(subdomainKey, "gen1"))

	// Test parameters are mapped correctly.
	class, err := NewClass(key, "Name", "Details", &actorKey, &generalizationKey, &generalizationKey, "UmlComment")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), Class{
		Key:             key,
		Name:            "Name",
		Details:         "Details",
		ActorKey:        &actorKey,
		SuperclassOfKey: &generalizationKey,
		SubclassOfKey:   &generalizationKey,
		UmlComment:      "UmlComment",
	}, class)

	// Test that Validate is called (invalid data should fail).
	_, err = NewClass(key, "", "Details", nil, nil, nil, "UmlComment")
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *ClassSuite) TestValidateWithParent() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	validKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	otherSubdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "other_subdomain"))

	// Test that Validate is called.
	class := Class{
		Key:  validKey,
		Name: "", // Invalid
	}
	err := class.ValidateWithParent(&subdomainKey)
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - class key has subdomain1 as parent, but we pass other_subdomain.
	class = Class{
		Key:  validKey,
		Name: "Name",
	}
	err = class.ValidateWithParent(&otherSubdomainKey)
	assert.ErrorContains(suite.T(), err, "does not match expected parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = class.ValidateWithParent(&subdomainKey)
	assert.NoError(suite.T(), err)
}
