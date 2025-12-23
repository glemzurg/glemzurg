package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
)

func TestStateActionInOutRoundTrip(t *testing.T) {
	originalReq := requirements.StateAction{
		Key:       "state_action1",
		ActionKey: "action1",
		When:      "entry",
		StateKey:  "state1", // This won't be preserved in round trip
	}

	// Convert to InOut
	inOut := FromRequirementsStateAction(originalReq)

	// Convert back to requirements
	convertedBack := inOut.ToRequirements()

	// Check fields
	if convertedBack.Key != originalReq.Key ||
		convertedBack.ActionKey != originalReq.ActionKey ||
		convertedBack.When != originalReq.When {
		t.Errorf("Round trip failed: got %+v, want %+v", convertedBack, originalReq)
	}

	// Note: StateKey is not stored in JSON, so it's empty in convertedBack
}
