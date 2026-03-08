package model_use_case

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
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
	validKey := helper.Must(identity.NewUseCaseGeneralizationKey(subdomainKey, "gen1"))

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
			errstr: "'KeyType' failed on the 'required' tag",
		},
		{
			testName: "error wrong key type",
			generalization: Generalization{
				Key:  domainKey,
				Name: "Name",
			},
			errstr: "key: invalid key type 'domain' for use case generalization",
		},
		{
			testName: "error blank name",
			generalization: Generalization{
				Key:  validKey,
				Name: "",
			},
			errstr: "Name",
		},
	}
	for _, tt := range tests {
		suite.Run(tt.testName, func() {
			err := tt.generalization.Validate()
			if tt.errstr == "" {
				suite.Require().NoError(err)
			} else {
				suite.Require().ErrorContains(err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewGeneralization maps parameters correctly and calls Validate.
func (suite *GeneralizationSuite) TestNew() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	key := helper.Must(identity.NewUseCaseGeneralizationKey(subdomainKey, "gen1"))

	// Test parameters are mapped correctly.
	gen, err := NewGeneralization(key, "Name", "Details", true, false, "UmlComment")
	suite.Require().NoError(err)
	suite.Equal(Generalization{
		Key:        key,
		Name:       "Name",
		Details:    "Details",
		IsComplete: true,
		IsStatic:   false,
		UmlComment: "UmlComment",
	}, gen)

	// Test that Validate is called (invalid data should fail).
	_, err = NewGeneralization(key, "", "Details", true, false, "UmlComment")
	suite.Require().ErrorContains(err, "Name")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *GeneralizationSuite) TestValidateWithParent() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	validKey := helper.Must(identity.NewUseCaseGeneralizationKey(subdomainKey, "gen1"))
	otherSubdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "other_subdomain"))

	// Test that Validate is called.
	gen := Generalization{
		Key:  validKey,
		Name: "", // Invalid
	}
	err := gen.ValidateWithParent(&subdomainKey)
	suite.Require().ErrorContains(err, "Name", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - generalization key has subdomain1 as parent, but we pass other_subdomain.
	gen = Generalization{
		Key:  validKey,
		Name: "Name",
	}
	err = gen.ValidateWithParent(&otherSubdomainKey)
	suite.Require().ErrorContains(err, "does not match expected parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = gen.ValidateWithParent(&subdomainKey)
	suite.Require().NoError(err)
}
