package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
)

func TestTransitionInOutRoundTrip(t *testing.T) {
	originalReq := requirements.Transition{
		Key:          "transition1",
		FromStateKey: "state1",
		EventKey:     "event1",
		GuardKey:     "guard1",
		ActionKey:    "action1",
		ToStateKey:   "state2",
		UmlComment:   "Transition comment",
	}

	// Convert to InOut
	inOut := FromRequirementsTransition(originalReq)

	// Convert back to requirements
	convertedBack := inOut.ToRequirements()

	// Check fields
	if convertedBack.Key != originalReq.Key ||
		convertedBack.FromStateKey != originalReq.FromStateKey ||
		convertedBack.EventKey != originalReq.EventKey ||
		convertedBack.GuardKey != originalReq.GuardKey ||
		convertedBack.ActionKey != originalReq.ActionKey ||
		convertedBack.ToStateKey != originalReq.ToStateKey ||
		convertedBack.UmlComment != originalReq.UmlComment {
		t.Errorf("Round trip failed: got %+v, want %+v", convertedBack, originalReq)
	}
}
