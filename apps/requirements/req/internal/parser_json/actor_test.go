package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
)

func TestActorInOutRoundTrip(t *testing.T) {
	original := requirements.Actor{
		Key:        "actor1",
		Name:       "User",
		Details:    "A user",
		Type:       "person",
		UmlComment: "comment",
		ClassKeys:  []string{"class1"}, // This will not round trip
	}

	inOut := FromRequirementsActor(original)
	back := inOut.ToRequirements()

	// Check individual fields, ignoring ClassKeys
	assert.Equal(t, original.Key, back.Key)
	assert.Equal(t, original.Name, back.Name)
	assert.Equal(t, original.Details, back.Details)
	assert.Equal(t, original.Type, back.Type)
	assert.Equal(t, original.UmlComment, back.UmlComment)
}
