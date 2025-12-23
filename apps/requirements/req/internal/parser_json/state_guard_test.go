package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
)

func TestGuardInOutRoundTrip(t *testing.T) {
	originalReq := requirements.Guard{
		Key:     "guard1",
		Name:    "Authenticated",
		Details: "User must be authenticated",
	}

	// Convert to InOut
	inOut := FromRequirementsGuard(originalReq)

	// Convert back to requirements
	convertedBack := inOut.ToRequirements()

	// Check fields
	if convertedBack.Key != originalReq.Key ||
		convertedBack.Name != originalReq.Name ||
		convertedBack.Details != originalReq.Details {
		t.Errorf("Round trip failed: got %+v, want %+v", convertedBack, originalReq)
	}
}
