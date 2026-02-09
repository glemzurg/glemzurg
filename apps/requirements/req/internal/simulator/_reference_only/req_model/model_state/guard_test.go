package model_state

import (
	"testing"

	"github.com/glemzurg/go-tlaplus/internal/helper"
	"github.com/glemzurg/go-tlaplus/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestGuardSuite(t *testing.T) {
	suite.Run(t, new(GuardSuite))
}

type GuardSuite struct {
	suite.Suite
}

// TestValidate tests all validation rules for Guard.
func (suite *GuardSuite) TestValidate() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	validKey := helper.Must(identity.NewGuardKey(classKey, "guard1"))

	tests := []struct {
		testName string
		guard    Guard
		errstr   string
	}{
		{
			testName: "valid guard",
			guard: Guard{
				Key:     validKey,
				Name:    "Name",
				Details: "Details",
			},
		},
		{
			testName: "error empty key",
			guard: Guard{
				Key:     identity.Key{},
				Name:    "Name",
				Details: "Details",
			},
			errstr: "keyType: cannot be blank",
		},
		{
			testName: "error wrong key type",
			guard: Guard{
				Key:     domainKey,
				Name:    "Name",
				Details: "Details",
			},
			errstr: "Key: invalid key type 'domain' for guard",
		},
		{
			testName: "error blank name",
			guard: Guard{
				Key:     validKey,
				Name:    "",
				Details: "Details",
			},
			errstr: "Name: cannot be blank",
		},
		{
			testName: "error blank details",
			guard: Guard{
				Key:     validKey,
				Name:    "Name",
				Details: "",
			},
			errstr: "Details: cannot be blank",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.guard.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewGuard maps parameters correctly and calls Validate.
func (suite *GuardSuite) TestNew() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	key := helper.Must(identity.NewGuardKey(classKey, "guard1"))

	// Test parameters are mapped correctly.
	guard, err := NewGuard(key, "Name", "Details")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), Guard{
		Key:     key,
		Name:    "Name",
		Details: "Details",
	}, guard)

	// Test that Validate is called (invalid data should fail).
	_, err = NewGuard(key, "", "Details")
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *GuardSuite) TestValidateWithParent() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	validKey := helper.Must(identity.NewGuardKey(classKey, "guard1"))
	otherClassKey := helper.Must(identity.NewClassKey(subdomainKey, "other_class"))

	// Test that Validate is called.
	guard := Guard{
		Key:     validKey,
		Name:    "", // Invalid
		Details: "Details",
	}
	err := guard.ValidateWithParent(&classKey)
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - guard key has class1 as parent, but we pass other_class.
	guard = Guard{
		Key:     validKey,
		Name:    "Name",
		Details: "Details",
	}
	err = guard.ValidateWithParent(&otherClassKey)
	assert.ErrorContains(suite.T(), err, "does not match expected parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = guard.ValidateWithParent(&classKey)
	assert.NoError(suite.T(), err)
}
