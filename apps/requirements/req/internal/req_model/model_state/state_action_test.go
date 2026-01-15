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

func (suite *StateActionSuite) TestNew() {

	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	stateKey := helper.Must(identity.NewStateKey(classKey, "state1"))
	actionKey := helper.Must(identity.NewActionKey(classKey, "action1"))

	tests := []struct {
		testName  string
		key       identity.Key
		actionKey identity.Key
		when      string
		obj       StateAction
		errstr    string
	}{
		// OK.
		{
			testName:  "ok with entry",
			key:       helper.Must(identity.NewStateActionKey(stateKey, "stateaction1")),
			actionKey: actionKey,
			when:      "entry",
			obj: StateAction{
				Key:       helper.Must(identity.NewStateActionKey(stateKey, "stateaction1")),
				ActionKey: actionKey,
				When:      "entry",
			},
		},
		{
			testName:  "ok with exit",
			key:       helper.Must(identity.NewStateActionKey(stateKey, "stateaction2")),
			actionKey: actionKey,
			when:      "exit",
			obj: StateAction{
				Key:       helper.Must(identity.NewStateActionKey(stateKey, "stateaction2")),
				ActionKey: actionKey,
				When:      "exit",
			},
		},
		{
			testName:  "ok with do",
			key:       helper.Must(identity.NewStateActionKey(stateKey, "stateaction3")),
			actionKey: actionKey,
			when:      "do",
			obj: StateAction{
				Key:       helper.Must(identity.NewStateActionKey(stateKey, "stateaction3")),
				ActionKey: actionKey,
				When:      "do",
			},
		},

		// Error states.
		{
			testName:  "error empty key",
			key:       identity.Key{},
			actionKey: actionKey,
			when:      "entry",
			errstr:    "keyType: cannot be blank",
		},
		{
			testName:  "error wrong key type",
			key:       helper.Must(identity.NewDomainKey("domain1")),
			actionKey: actionKey,
			when:      "entry",
			errstr:    "Key: invalid key type 'domain' for state action",
		},
		{
			testName:  "error empty action key",
			key:       helper.Must(identity.NewStateActionKey(stateKey, "stateaction4")),
			actionKey: identity.Key{},
			when:      "entry",
			errstr:    "keyType: cannot be blank",
		},
		{
			testName:  "error wrong action key type",
			key:       helper.Must(identity.NewStateActionKey(stateKey, "stateaction5")),
			actionKey: helper.Must(identity.NewDomainKey("domain1")),
			when:      "entry",
			errstr:    "ActionKey: invalid key type 'domain' for action",
		},
		{
			testName:  "error with blank when",
			key:       helper.Must(identity.NewStateActionKey(stateKey, "stateaction6")),
			actionKey: actionKey,
			when:      "",
			errstr:    `When: cannot be blank`,
		},
		{
			testName:  "error with unknown when",
			key:       helper.Must(identity.NewStateActionKey(stateKey, "stateaction7")),
			actionKey: actionKey,
			when:      "unknown",
			errstr:    `When: must be a valid value`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			obj, err := NewStateAction(tt.key, tt.actionKey, tt.when)
			if tt.errstr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.obj, obj)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
				assert.Empty(t, obj)
			}
		})
	}
}
