package model_domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

func TestDomainSuite(t *testing.T) {
	suite.Run(t, new(DomainSuite))
}

type DomainSuite struct {
	suite.Suite
}

// TestValidate tests all validation rules for Domain.
func (suite *DomainSuite) TestValidate() {
	validKey := helper.Must(identity.NewDomainKey("domain1"))

	tests := []struct {
		testName string
		domain   Domain
		errstr   string
	}{
		{
			testName: "valid domain",
			domain: Domain{
				Key:  validKey,
				Name: "Name",
			},
		},
		{
			testName: "error empty key",
			domain: Domain{
				Key:  identity.Key{},
				Name: "Name",
			},
			errstr: "keyType: cannot be blank",
		},
		{
			testName: "error wrong key type",
			domain: Domain{
				Key:  helper.Must(identity.NewActorKey("actor1")),
				Name: "Name",
			},
			errstr: "Key: invalid key type 'actor' for domain.",
		},
		{
			testName: "error blank name",
			domain: Domain{
				Key:  validKey,
				Name: "",
			},
			errstr: "Name: cannot be blank",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.domain.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewDomain maps parameters correctly and calls Validate.
func (suite *DomainSuite) TestNew() {
	key := helper.Must(identity.NewDomainKey("domain1"))

	// Test parameters are mapped correctly.
	domain, err := NewDomain(key, "Name", "Details", true, "UmlComment")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), key, domain.Key)
	assert.Equal(suite.T(), "Name", domain.Name)
	assert.Equal(suite.T(), "Details", domain.Details)
	assert.Equal(suite.T(), true, domain.Realized)
	assert.Equal(suite.T(), "UmlComment", domain.UmlComment)

	// Test that Validate is called (invalid data should fail).
	_, err = NewDomain(key, "", "Details", true, "UmlComment")
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *DomainSuite) TestValidateWithParent() {
	validKey := helper.Must(identity.NewDomainKey("domain1"))
	otherDomainKey := helper.Must(identity.NewDomainKey("other_domain"))

	// Test that Validate is called.
	domain := Domain{
		Key:  validKey,
		Name: "", // Invalid
	}
	err := domain.ValidateWithParent(nil)
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - domains should have nil parent.
	domain = Domain{
		Key:  validKey,
		Name: "Name",
	}
	err = domain.ValidateWithParent(&otherDomainKey)
	assert.ErrorContains(suite.T(), err, "should not have a parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = domain.ValidateWithParent(nil)
	assert.NoError(suite.T(), err)
}
