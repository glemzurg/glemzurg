package model_class

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestGeneralizationSuite(t *testing.T) {
	suite.Run(t, new(GeneralizationSuite))
}

type GeneralizationSuite struct {
	suite.Suite
}

// TestValidate tests all validation rules for Generalization.
func (suite *GeneralizationSuite) TestValidate() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	validKey := helper.Must(identity.NewGeneralizationKey(subdomainKey, "gen1"))

	tests := []struct {
		testName       string
		generalization Generalization
		errstr         string
	}{
		{
			testName: "valid generalization",
			generalization: Generalization{
				Key:  validKey,
				Name: "Name",
			},
		},
		{
			testName: "error empty key",
			generalization: Generalization{
				Key:  identity.Key{},
				Name: "Name",
			},
			errstr: "keyType: cannot be blank",
		},
		{
			testName: "error wrong key type",
			generalization: Generalization{
				Key:  domainKey,
				Name: "Name",
			},
			errstr: "Key: invalid key type 'domain' for generalization.",
		},
		{
			testName: "error blank name",
			generalization: Generalization{
				Key:  validKey,
				Name: "",
			},
			errstr: "Name: cannot be blank",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.generalization.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewGeneralization maps parameters correctly and calls Validate.
func (suite *GeneralizationSuite) TestNew() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	key := helper.Must(identity.NewGeneralizationKey(subdomainKey, "gen1"))

	// Test parameters are mapped correctly.
	gen, err := NewGeneralization(key, "Name", "Details", true, false, "UmlComment")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), key, gen.Key)
	assert.Equal(suite.T(), "Name", gen.Name)
	assert.Equal(suite.T(), "Details", gen.Details)
	assert.Equal(suite.T(), true, gen.IsComplete)
	assert.Equal(suite.T(), false, gen.IsStatic)
	assert.Equal(suite.T(), "UmlComment", gen.UmlComment)

	// Test that Validate is called (invalid data should fail).
	_, err = NewGeneralization(key, "", "Details", true, false, "UmlComment")
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *GeneralizationSuite) TestValidateWithParent() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	validKey := helper.Must(identity.NewGeneralizationKey(subdomainKey, "gen1"))
	otherSubdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "other_subdomain"))

	// Test that Validate is called.
	gen := Generalization{
		Key:  validKey,
		Name: "", // Invalid
	}
	err := gen.ValidateWithParent(&subdomainKey)
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - generalization key has subdomain1 as parent, but we pass other_subdomain.
	gen = Generalization{
		Key:  validKey,
		Name: "Name",
	}
	err = gen.ValidateWithParent(&otherSubdomainKey)
	assert.ErrorContains(suite.T(), err, "does not match expected parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = gen.ValidateWithParent(&subdomainKey)
	assert.NoError(suite.T(), err)
}
