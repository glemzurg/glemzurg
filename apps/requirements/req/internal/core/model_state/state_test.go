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
			errstr: "'KeyType' failed on the 'required' tag",
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
			errstr: "Name",
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
	assert.Equal(suite.T(), State{
		Key:        key,
		Name:       "Name",
		Details:    "Details",
		UmlComment: "UmlComment",
	}, state)

	// Test that Validate is called (invalid data should fail).
	_, err = NewState(key, "", "Details", "UmlComment")
	assert.ErrorContains(suite.T(), err, "Name")
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
	assert.ErrorContains(suite.T(), err, "Name", "ValidateWithParent should call Validate()")

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

// TestValidateWithParentAndActions tests that ValidateWithParentAndActions validates child StateActions.
func (suite *StateSuite) TestValidateWithParentAndActions() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	stateKey := helper.Must(identity.NewStateKey(classKey, "state1"))
	actionKey := helper.Must(identity.NewActionKey(classKey, "action1"))
	stateActionKey := helper.Must(identity.NewStateActionKey(stateKey, "entry", "sa1"))

	actionKeys := map[identity.Key]bool{
		actionKey: true,
	}

	// Test valid state with valid StateAction child.
	state := State{
		Key:  stateKey,
		Name: "Name",
		Actions: []StateAction{
			{Key: stateActionKey, ActionKey: actionKey, When: "entry"},
		},
	}
	err := state.ValidateWithParentAndActions(&classKey, actionKeys)
	assert.NoError(suite.T(), err)

	// Test invalid child StateAction (empty action key) propagates error.
	state = State{
		Key:  stateKey,
		Name: "Name",
		Actions: []StateAction{
			{Key: stateActionKey, ActionKey: identity.Key{}, When: "entry"},
		},
	}
	err = state.ValidateWithParentAndActions(&classKey, actionKeys)
	assert.Error(suite.T(), err, "Invalid child StateAction should propagate error")

	// Test action reference validation - reference non-existent action.
	nonExistentActionKey := helper.Must(identity.NewActionKey(classKey, "nonexistent"))
	state = State{
		Key:  stateKey,
		Name: "Name",
		Actions: []StateAction{
			{Key: stateActionKey, ActionKey: nonExistentActionKey, When: "entry"},
		},
	}
	err = state.ValidateWithParentAndActions(&classKey, actionKeys)
	assert.ErrorContains(suite.T(), err, "references non-existent action", "Should validate action references")
}

// TestSetActions tests that SetActions sets and sorts actions.
func (suite *StateSuite) TestSetActions() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	stateKey := helper.Must(identity.NewStateKey(classKey, "state1"))
	actionKey := helper.Must(identity.NewActionKey(classKey, "action1"))
	saKeyExit := helper.Must(identity.NewStateActionKey(stateKey, "exit", "sa_exit"))
	saKeyEntry := helper.Must(identity.NewStateActionKey(stateKey, "entry", "sa_entry"))

	state := State{Key: stateKey, Name: "Name"}

	// Add actions in reverse order (exit before entry).
	actions := []StateAction{
		{Key: saKeyExit, ActionKey: actionKey, When: "exit"},
		{Key: saKeyEntry, ActionKey: actionKey, When: "entry"},
	}
	state.SetActions(actions)

	// Verify actions are set.
	assert.Equal(suite.T(), 2, len(state.Actions))
	// Verify sorted: entry should come before exit.
	assert.Equal(suite.T(), "entry", state.Actions[0].When)
	assert.Equal(suite.T(), "exit", state.Actions[1].When)
}
