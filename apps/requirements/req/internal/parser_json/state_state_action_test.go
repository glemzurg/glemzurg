package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStateActionInOutRoundTrip(t *testing.T) {
	domainKey, err := identity.NewDomainKey("domain1")
	require.NoError(t, err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, "sub1")
	require.NoError(t, err)
	classKey, err := identity.NewClassKey(subdomainKey, "class1")
	require.NoError(t, err)
	stateKey, err := identity.NewStateKey(classKey, "state1")
	require.NoError(t, err)
	actionKey, err := identity.NewActionKey(classKey, "action1")
	require.NoError(t, err)
	stateActionKey, err := identity.NewStateActionKey(stateKey, "entry", "action1")
	require.NoError(t, err)

	original := model_state.StateAction{
		Key:       stateActionKey,
		ActionKey: actionKey,
		When:      "entry",
	}

	inOut := FromRequirementsStateAction(original)
	back, err := inOut.ToRequirements()
	require.NoError(t, err)
	assert.Equal(t, original, back)
}
