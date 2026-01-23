package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActorInOutRoundTrip(t *testing.T) {
	key, err := identity.NewActorKey("user")
	require.NoError(t, err)

	original := model_actor.Actor{
		Key:        key,
		Name:       "User",
		Details:    "A user",
		Type:       "person",
		UmlComment: "comment",
	}

	inOut := FromRequirementsActor(original)
	back, err := inOut.ToRequirements()
	require.NoError(t, err)
	assert.Equal(t, original, back)
}
