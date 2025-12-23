package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
)

func TestEventInOutRoundTrip(t *testing.T) {
	originalReq := requirements.Event{
		Key:     "event1",
		Name:    "Login Event",
		Details: "User attempts to log in",
		Parameters: []requirements.EventParameter{
			{
				Name:   "username",
				Source: "user_input",
			},
			{
				Name:   "password",
				Source: "user_input",
			},
		},
	}

	// Convert to InOut
	inOut := FromRequirementsEvent(originalReq)

	// Convert back to requirements
	convertedBack := inOut.ToRequirements()

	// Check fields
	if convertedBack.Key != originalReq.Key ||
		convertedBack.Name != originalReq.Name ||
		convertedBack.Details != originalReq.Details ||
		len(convertedBack.Parameters) != len(originalReq.Parameters) {
		t.Errorf("Round trip failed: got %+v, want %+v", convertedBack, originalReq)
	}

	for i, param := range originalReq.Parameters {
		if convertedBack.Parameters[i].Name != param.Name ||
			convertedBack.Parameters[i].Source != param.Source {
			t.Errorf("Parameter[%d] mismatch: got %+v, want %+v", i, convertedBack.Parameters[i], param)
		}
	}
}
