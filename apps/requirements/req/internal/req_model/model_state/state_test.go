package model_state

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestStateSuite(t *testing.T) {
	suite.Run(t, new(StateSuite))
}

type StateSuite struct {
	suite.Suite
}

// TestValidate tests all validation rules for State.
func (suite *StateSuite) TestValidate() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	validKey := helper.Must(identity.NewStateKey(classKey, "state1"))

	tests := []struct {
		testName string
		state    State
		errstr   string
	}{
		{
			testName: "valid state",
			state: State{
				Key:  validKey,
				Name: "Name",
			},
		},
		{
			testName: "error empty key",
			state: State{
				Key:  identity.Key{},
				Name: "Name",
			},
			errstr: "keyType: cannot be blank",
		},
		{
			testName: "error wrong key type",
			state: State{
				Key:  domainKey,
				Name: "Name",
			},
			errstr: "Key: invalid key type 'domain' for state",
		},
		{
			testName: "error blank name",
			state: State{
				Key:  validKey,
				Name: "",
			},
			errstr: "Name: cannot be blank",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.state.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewState maps parameters correctly and calls Validate.
func (suite *StateSuite) TestNew() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	key := helper.Must(identity.NewStateKey(classKey, "state1"))

	// Test parameters are mapped correctly.
	state, err := NewState(key, "Name", "Details", "UmlComment")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), key, state.Key)
	assert.Equal(suite.T(), "Name", state.Name)
	assert.Equal(suite.T(), "Details", state.Details)
	assert.Equal(suite.T(), "UmlComment", state.UmlComment)

	// Test that Validate is called (invalid data should fail).
	_, err = NewState(key, "", "Details", "UmlComment")
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *StateSuite) TestValidateWithParent() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	validKey := helper.Must(identity.NewStateKey(classKey, "state1"))
	otherClassKey := helper.Must(identity.NewClassKey(subdomainKey, "other_class"))

	// Test that Validate is called.
	state := State{
		Key:  validKey,
		Name: "", // Invalid
	}
	err := state.ValidateWithParent(&classKey)
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - state key has class1 as parent, but we pass other_class.
	state = State{
		Key:  validKey,
		Name: "Name",
	}
	err = state.ValidateWithParent(&otherClassKey)
	assert.ErrorContains(suite.T(), err, "does not match expected parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = state.ValidateWithParent(&classKey)
	assert.NoError(suite.T(), err)
}
