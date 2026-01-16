package model_domain

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestSubdomainSuite(t *testing.T) {
	suite.Run(t, new(SubdomainSuite))
}

type SubdomainSuite struct {
	suite.Suite
	domainKey identity.Key
}

func (suite *SubdomainSuite) SetupTest() {
	suite.domainKey = helper.Must(identity.NewDomainKey("domain1"))
}

// TestValidate tests all validation rules for Subdomain.
func (suite *SubdomainSuite) TestValidate() {
	validKey := helper.Must(identity.NewSubdomainKey(suite.domainKey, "subdomain1"))

	tests := []struct {
		testName  string
		subdomain Subdomain
		errstr    string
	}{
		{
			testName: "valid subdomain",
			subdomain: Subdomain{
				Key:  validKey,
				Name: "Name",
			},
		},
		{
			testName: "error empty key",
			subdomain: Subdomain{
				Key:  identity.Key{},
				Name: "Name",
			},
			errstr: "keyType: cannot be blank",
		},
		{
			testName: "error wrong key type",
			subdomain: Subdomain{
				Key:  helper.Must(identity.NewActorKey("actor1")),
				Name: "Name",
			},
			errstr: "Key: invalid key type 'actor' for subdomain.",
		},
		{
			testName: "error blank name",
			subdomain: Subdomain{
				Key:  validKey,
				Name: "",
			},
			errstr: "Name: cannot be blank",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.subdomain.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewSubdomain maps parameters correctly and calls Validate.
func (suite *SubdomainSuite) TestNew() {
	key := helper.Must(identity.NewSubdomainKey(suite.domainKey, "subdomain1"))

	// Test parameters are mapped correctly.
	subdomain, err := NewSubdomain(key, "Name", "Details", "UmlComment")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), key, subdomain.Key)
	assert.Equal(suite.T(), "Name", subdomain.Name)
	assert.Equal(suite.T(), "Details", subdomain.Details)
	assert.Equal(suite.T(), "UmlComment", subdomain.UmlComment)

	// Test that Validate is called (invalid data should fail).
	_, err = NewSubdomain(key, "", "Details", "UmlComment")
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *SubdomainSuite) TestValidateWithParent() {
	validKey := helper.Must(identity.NewSubdomainKey(suite.domainKey, "subdomain1"))
	otherDomainKey := helper.Must(identity.NewDomainKey("other_domain"))

	// Test that Validate is called.
	subdomain := Subdomain{
		Key:  validKey,
		Name: "", // Invalid
	}
	err := subdomain.ValidateWithParent(&suite.domainKey)
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - subdomain key has domain1 as parent, but we pass other_domain.
	subdomain = Subdomain{
		Key:  validKey,
		Name: "Name",
	}
	err = subdomain.ValidateWithParent(&otherDomainKey)
	assert.ErrorContains(suite.T(), err, "does not match expected parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = subdomain.ValidateWithParent(&suite.domainKey)
	assert.NoError(suite.T(), err)
}
