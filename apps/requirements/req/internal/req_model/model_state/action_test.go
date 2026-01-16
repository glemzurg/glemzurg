package model_state

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestActionSuite(t *testing.T) {
	suite.Run(t, new(ActionSuite))
}

type ActionSuite struct {
	suite.Suite
}

// TestValidate tests all validation rules for Action.
func (suite *ActionSuite) TestValidate() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	validKey := helper.Must(identity.NewActionKey(classKey, "action1"))

	tests := []struct {
		testName string
		action   Action
		errstr   string
	}{
		{
			testName: "valid action",
			action: Action{
				Key:  validKey,
				Name: "Name",
			},
		},
		{
			testName: "error empty key",
			action: Action{
				Key:  identity.Key{},
				Name: "Name",
			},
			errstr: "keyType: cannot be blank",
		},
		{
			testName: "error wrong key type",
			action: Action{
				Key:  domainKey,
				Name: "Name",
			},
			errstr: "Key: invalid key type 'domain' for action",
		},
		{
			testName: "error blank name",
			action: Action{
				Key:  validKey,
				Name: "",
			},
			errstr: "Name: cannot be blank",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.action.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewAction maps parameters correctly and calls Validate.
func (suite *ActionSuite) TestNew() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	key := helper.Must(identity.NewActionKey(classKey, "action1"))

	// Test parameters are mapped correctly.
	action, err := NewAction(key, "Name", "Details", []string{"Requires"}, []string{"Guarantees"})
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), key, action.Key)
	assert.Equal(suite.T(), "Name", action.Name)
	assert.Equal(suite.T(), "Details", action.Details)
	assert.Equal(suite.T(), []string{"Requires"}, action.Requires)
	assert.Equal(suite.T(), []string{"Guarantees"}, action.Guarantees)

	// Test that Validate is called (invalid data should fail).
	_, err = NewAction(key, "", "Details", nil, nil)
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *ActionSuite) TestValidateWithParent() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	validKey := helper.Must(identity.NewActionKey(classKey, "action1"))
	otherClassKey := helper.Must(identity.NewClassKey(subdomainKey, "other_class"))

	// Test that Validate is called.
	action := Action{
		Key:  validKey,
		Name: "", // Invalid
	}
	err := action.ValidateWithParent(&classKey)
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - action key has class1 as parent, but we pass other_class.
	action = Action{
		Key:  validKey,
		Name: "Name",
	}
	err = action.ValidateWithParent(&otherClassKey)
	assert.ErrorContains(suite.T(), err, "does not match expected parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = action.ValidateWithParent(&classKey)
	assert.NoError(suite.T(), err)
}
