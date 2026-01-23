package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStateInOutRoundTrip(t *testing.T) {
	domainKey, err := identity.NewDomainKey("domain1")
	require.NoError(t, err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, "sub1")
	require.NoError(t, err)
	classKey, err := identity.NewClassKey(subdomainKey, "class1")
	require.NoError(t, err)
	stateKey, err := identity.NewStateKey(classKey, "state1")
	require.NoError(t, err)
	action1Key, err := identity.NewActionKey(classKey, "action1")
	require.NoError(t, err)
	action2Key, err := identity.NewActionKey(classKey, "action2")
	require.NoError(t, err)
	stateAction1Key, err := identity.NewStateActionKey(stateKey, "entry", "action1")
	require.NoError(t, err)
	stateAction2Key, err := identity.NewStateActionKey(stateKey, "exit", "action2")
	require.NoError(t, err)

	original := model_state.State{
		Key:        stateKey,
		Name:       "Initial State",
		Details:    "The starting state",
		UmlComment: "State comment",
		Actions: []model_state.StateAction{
			{
				Key:       stateAction1Key,
				ActionKey: action1Key,
				When:      "entry",
			},
			{
				Key:       stateAction2Key,
				ActionKey: action2Key,
				When:      "exit",
			},
		},
	}

	inOut := FromRequirementsState(original)
	back, err := inOut.ToRequirements()
	require.NoError(t, err)
	assert.Equal(t, original, back)
}
