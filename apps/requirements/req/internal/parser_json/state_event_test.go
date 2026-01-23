package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventInOutRoundTrip(t *testing.T) {
	domainKey, err := identity.NewDomainKey("domain1")
	require.NoError(t, err)
	subdomainKey, err := identity.NewSubdomainKey(domainKey, "sub1")
	require.NoError(t, err)
	classKey, err := identity.NewClassKey(subdomainKey, "class1")
	require.NoError(t, err)
	eventKey, err := identity.NewEventKey(classKey, "event1")
	require.NoError(t, err)

	original := model_state.Event{
		Key:     eventKey,
		Name:    "Login Event",
		Details: "User attempts to log in",
		Parameters: []model_state.EventParameter{
			{
				Name: "username",
			},
			{
				Name: "password",
			},
		},
	}

	inOut := FromRequirementsEvent(original)
	back, err := inOut.ToRequirements()
	require.NoError(t, err)
	assert.Equal(t, original, back)
}
