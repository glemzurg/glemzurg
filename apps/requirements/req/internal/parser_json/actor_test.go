package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
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
	if back.Key != original.Key || back.Name != original.Name || back.Details != original.Details ||
		back.Type != original.Type || back.UmlComment != original.UmlComment {
		t.Errorf("Round trip failed: got %+v, want %+v", back, original)
	}
}
