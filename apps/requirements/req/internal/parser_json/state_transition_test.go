package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransitionInOutRoundTrip(t *testing.T) {
	domainKey, err := identity.NewDomainKey("domain1")
	require.NoError(t, err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, "sub1")
	require.NoError(t, err)
	classKey, err := identity.NewClassKey(subdomainKey, "class1")
	require.NoError(t, err)
	state1Key, err := identity.NewStateKey(classKey, "state1")
	require.NoError(t, err)
	state2Key, err := identity.NewStateKey(classKey, "state2")
	require.NoError(t, err)
	eventKey, err := identity.NewEventKey(classKey, "event1")
	require.NoError(t, err)
	guardKey, err := identity.NewGuardKey(classKey, "guard1")
	require.NoError(t, err)
	actionKey, err := identity.NewActionKey(classKey, "action1")
	require.NoError(t, err)
	transitionKey, err := identity.NewTransitionKey(classKey, "state1", "event1", "guard1", "action1", "state2")
	require.NoError(t, err)

	original := model_state.Transition{
		Key:          transitionKey,
		FromStateKey: &state1Key,
		EventKey:     eventKey,
		GuardKey:     &guardKey,
		ActionKey:    &actionKey,
		ToStateKey:   &state2Key,
		UmlComment:   "Transition comment",
	}

	inOut := FromRequirementsTransition(original)
	back, err := inOut.ToRequirements()
	require.NoError(t, err)
	assert.Equal(t, original, back)
}
