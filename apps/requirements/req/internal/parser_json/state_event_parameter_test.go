package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
)

func TestEventParameterInOutRoundTrip(t *testing.T) {
	originalReq := requirements.EventParameter{
		Name:   "username",
		Source: "user_input",
	}

	// Convert to InOut
	inOut := FromRequirementsEventParameter(originalReq)

	// Convert back to requirements
	convertedBack := inOut.ToRequirements()

	// Check fields
	if convertedBack.Name != originalReq.Name ||
		convertedBack.Source != originalReq.Source {
		t.Errorf("Round trip failed: got %+v, want %+v", convertedBack, originalReq)
	}
}
