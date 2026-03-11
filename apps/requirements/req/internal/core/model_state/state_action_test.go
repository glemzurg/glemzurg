package model_state

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
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
	validKey := helper.Must(identity.NewStateActionKey(stateKey, "entry", "stateaction1"))

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
			errstr: "key type is required",
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
			errstr: "key type is required",
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
			errstr: "When",
		},
		{
			testName: "error unknown when",
			stateAction: StateAction{
				Key:       validKey,
				ActionKey: actionKey,
				When:      "unknown",
			},
			errstr: "When",
		},
	}
	for _, tt := range tests {
		suite.Run(tt.testName, func() {
			ctx := coreerr.NewContext("test", "")
			err := tt.stateAction.Validate(ctx)
			if tt.errstr == "" {
				suite.Require().NoError(err)
			} else {
				suite.Require().ErrorContains(err, tt.errstr)
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
	key := helper.Must(identity.NewStateActionKey(stateKey, "entry", "stateaction1"))

	// Test parameters are mapped correctly.
	stateAction := NewStateAction(key, actionKey, "entry")
	suite.Equal(StateAction{
		Key:       key,
		ActionKey: actionKey,
		When:      "entry",
	}, stateAction)
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *StateActionSuite) TestValidateWithParent() {
	ctx := coreerr.NewContext("test", "")
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	stateKey := helper.Must(identity.NewStateKey(classKey, "state1"))
	actionKey := helper.Must(identity.NewActionKey(classKey, "action1"))
	validKey := helper.Must(identity.NewStateActionKey(stateKey, "entry", "stateaction1"))
	otherStateKey := helper.Must(identity.NewStateKey(classKey, "other_state"))

	// Test that Validate is called.
	stateAction := StateAction{
		Key:       validKey,
		ActionKey: actionKey,
		When:      "", // Invalid
	}
	err := stateAction.ValidateWithParent(ctx, &stateKey)
	suite.Require().ErrorContains(err, "When", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - stateAction key has state1 as parent, but we pass other_state.
	stateAction = StateAction{
		Key:       validKey,
		ActionKey: actionKey,
		When:      "entry",
	}
	err = stateAction.ValidateWithParent(ctx, &otherStateKey)
	suite.Require().ErrorContains(err, "does not match expected parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = stateAction.ValidateWithParent(ctx, &stateKey)
	suite.Require().NoError(err)
}

// TestValidateReferences tests that ValidateReferences validates action references correctly.
func (suite *StateActionSuite) TestValidateReferences() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	stateKey := helper.Must(identity.NewStateKey(classKey, "state1"))
	actionKey := helper.Must(identity.NewActionKey(classKey, "action1"))
	nonExistentActionKey := helper.Must(identity.NewActionKey(classKey, "nonexistent"))
	validKey := helper.Must(identity.NewStateActionKey(stateKey, "entry", "stateaction1"))

	// Build lookup map with valid actions.
	actions := map[identity.Key]bool{
		actionKey: true,
	}

	tests := []struct {
		testName    string
		stateAction StateAction
		actions     map[identity.Key]bool
		errstr      string
	}{
		{
			testName: "valid state action with existing action",
			stateAction: StateAction{
				Key:       validKey,
				ActionKey: actionKey,
				When:      "entry",
			},
			actions: actions,
		},
		{
			testName: "error ActionKey references non-existent action",
			stateAction: StateAction{
				Key:       validKey,
				ActionKey: nonExistentActionKey,
				When:      "entry",
			},
			actions: actions,
			errstr:  "references non-existent action",
		},
	}
	for _, tt := range tests {
		suite.Run(tt.testName, func() {
			ctx := coreerr.NewContext("test", "")
			err := tt.stateAction.ValidateReferences(ctx, tt.actions)
			if tt.errstr == "" {
				suite.Require().NoError(err)
			} else {
				suite.Require().ErrorContains(err, tt.errstr)
			}
		})
	}
}
