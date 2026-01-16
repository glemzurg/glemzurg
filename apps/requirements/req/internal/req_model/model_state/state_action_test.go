package model_state

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestStateActionSuite(t *testing.T) {
	suite.Run(t, new(StateActionSuite))
}

type StateActionSuite struct {
	suite.Suite
}

// TestValidate tests all validation rules for StateAction.
func (suite *StateActionSuite) TestValidate() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	stateKey := helper.Must(identity.NewStateKey(classKey, "state1"))
	actionKey := helper.Must(identity.NewActionKey(classKey, "action1"))
	validKey := helper.Must(identity.NewStateActionKey(stateKey, "stateaction1"))

	tests := []struct {
		testName    string
		stateAction StateAction
		errstr      string
	}{
		{
			testName: "valid state action with entry",
			stateAction: StateAction{
				Key:       validKey,
				ActionKey: actionKey,
				When:      "entry",
			},
		},
		{
			testName: "valid state action with exit",
			stateAction: StateAction{
				Key:       validKey,
				ActionKey: actionKey,
				When:      "exit",
			},
		},
		{
			testName: "valid state action with do",
			stateAction: StateAction{
				Key:       validKey,
				ActionKey: actionKey,
				When:      "do",
			},
		},
		{
			testName: "error empty key",
			stateAction: StateAction{
				Key:       identity.Key{},
				ActionKey: actionKey,
				When:      "entry",
			},
			errstr: "keyType: cannot be blank",
		},
		{
			testName: "error wrong key type",
			stateAction: StateAction{
				Key:       domainKey,
				ActionKey: actionKey,
				When:      "entry",
			},
			errstr: "Key: invalid key type 'domain' for state action",
		},
		{
			testName: "error empty action key",
			stateAction: StateAction{
				Key:       validKey,
				ActionKey: identity.Key{},
				When:      "entry",
			},
			errstr: "keyType: cannot be blank",
		},
		{
			testName: "error wrong action key type",
			stateAction: StateAction{
				Key:       validKey,
				ActionKey: domainKey,
				When:      "entry",
			},
			errstr: "ActionKey: invalid key type 'domain' for action",
		},
		{
			testName: "error blank when",
			stateAction: StateAction{
				Key:       validKey,
				ActionKey: actionKey,
				When:      "",
			},
			errstr: "When: cannot be blank",
		},
		{
			testName: "error unknown when",
			stateAction: StateAction{
				Key:       validKey,
				ActionKey: actionKey,
				When:      "unknown",
			},
			errstr: "When: must be a valid value",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.stateAction.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewStateAction maps parameters correctly and calls Validate.
func (suite *StateActionSuite) TestNew() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	stateKey := helper.Must(identity.NewStateKey(classKey, "state1"))
	actionKey := helper.Must(identity.NewActionKey(classKey, "action1"))
	key := helper.Must(identity.NewStateActionKey(stateKey, "stateaction1"))

	// Test parameters are mapped correctly.
	stateAction, err := NewStateAction(key, actionKey, "entry")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), key, stateAction.Key)
	assert.Equal(suite.T(), actionKey, stateAction.ActionKey)
	assert.Equal(suite.T(), "entry", stateAction.When)

	// Test that Validate is called (invalid data should fail).
	_, err = NewStateAction(key, actionKey, "")
	assert.ErrorContains(suite.T(), err, "When: cannot be blank")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *StateActionSuite) TestValidateWithParent() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	stateKey := helper.Must(identity.NewStateKey(classKey, "state1"))
	actionKey := helper.Must(identity.NewActionKey(classKey, "action1"))
	validKey := helper.Must(identity.NewStateActionKey(stateKey, "stateaction1"))
	otherStateKey := helper.Must(identity.NewStateKey(classKey, "other_state"))

	// Test that Validate is called.
	stateAction := StateAction{
		Key:       validKey,
		ActionKey: actionKey,
		When:      "", // Invalid
	}
	err := stateAction.ValidateWithParent(&stateKey)
	assert.ErrorContains(suite.T(), err, "When: cannot be blank", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - stateAction key has state1 as parent, but we pass other_state.
	stateAction = StateAction{
		Key:       validKey,
		ActionKey: actionKey,
		When:      "entry",
	}
	err = stateAction.ValidateWithParent(&otherStateKey)
	assert.ErrorContains(suite.T(), err, "does not match expected parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = stateAction.ValidateWithParent(&stateKey)
	assert.NoError(suite.T(), err)
}
