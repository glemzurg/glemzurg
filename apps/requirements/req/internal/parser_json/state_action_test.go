package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
)

func TestActionInOutRoundTrip(t *testing.T) {
	originalReq := requirements.Action{
		Key:        "action1",
		Name:       "Login Action",
		Details:    "User logs in",
		Requires:   []string{"user_authenticated"},
		Guarantees: []string{"session_created"},
	}

	// Convert to InOut
	inOut := FromRequirementsAction(originalReq)

	// Convert back to requirements
	convertedBack := inOut.ToRequirements()

	// Check fields
	if convertedBack.Key != originalReq.Key ||
		convertedBack.Name != originalReq.Name ||
		convertedBack.Details != originalReq.Details ||
		len(convertedBack.Requires) != len(originalReq.Requires) ||
		len(convertedBack.Guarantees) != len(originalReq.Guarantees) {
		t.Errorf("Round trip failed: got %+v, want %+v", convertedBack, originalReq)
	}

	for i, req := range originalReq.Requires {
		if convertedBack.Requires[i] != req {
			t.Errorf("Requires[%d] mismatch: got %q, want %q", i, convertedBack.Requires[i], req)
		}
	}

	for i, gua := range originalReq.Guarantees {
		if convertedBack.Guarantees[i] != gua {
			t.Errorf("Guarantees[%d] mismatch: got %q, want %q", i, convertedBack.Guarantees[i], gua)
		}
	}
}
