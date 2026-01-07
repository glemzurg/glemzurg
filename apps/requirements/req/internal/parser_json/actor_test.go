package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_actor"
	"github.com/stretchr/testify/assert"
)

func TestActorInOutRoundTrip(t *testing.T) {
	original := model_actor.Actor{
		Key:        "actor1",
		Name:       "User",
		Details:    "A user",
		Type:       "person",
		UmlComment: "comment",
	}

	inOut := FromRequirementsActor(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}
