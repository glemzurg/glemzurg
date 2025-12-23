package parser_json

import (
	"encoding/json"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
)

func TestScenarioInOutJSONRoundTrip(t *testing.T) {
	original := scenarioInOut{
		Key:     "scenario1",
		Name:    "Login Scenario",
		Details: "User logs into the system",
		Steps: nodeInOut{
			Description: "User enters credentials",
			EventKey:    "login",
		},
		Objects: []scenarioObjectInOut{
			{
				Key:          "user",
				ObjectNumber: 1,
				Name:         "User",
				ClassKey:     "user_class",
				Multi:        false,
			},
			{
				Key:          "system",
				ObjectNumber: 2,
				Name:         "System",
				ClassKey:     "system_class",
				Multi:        false,
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal to JSON: %v", err)
	}

	// Unmarshal back
	var unmarshaled scenarioInOut
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal from JSON: %v", err)
	}

	// Check fields
	if unmarshaled.Key != original.Key ||
		unmarshaled.Name != original.Name ||
		unmarshaled.Details != original.Details {
		t.Errorf("Basic fields mismatch: got %+v, want %+v", unmarshaled, original)
	}

	// Check steps
	if unmarshaled.Steps.Description != original.Steps.Description ||
		unmarshaled.Steps.EventKey != original.Steps.EventKey {
		t.Errorf("Steps mismatch: got %+v, want %+v", unmarshaled.Steps, original.Steps)
	}

	// Check objects
	if len(unmarshaled.Objects) != len(original.Objects) {
		t.Errorf("Objects length mismatch: got %d, want %d", len(unmarshaled.Objects), len(original.Objects))
	}
}

func TestScenarioInOutConversionRoundTrip(t *testing.T) {
	originalReq := requirements.Scenario{
		Key:     "scenario1",
		Name:    "Login Scenario",
		Details: "User logs into the system",
		Steps: requirements.Node{
			Description: "User enters credentials",
			EventKey:    "login",
		},
		Objects: []requirements.ScenarioObject{
			{
				Key:          "user",
				ObjectNumber: 1,
				Name:         "User",
				ClassKey:     "user_class",
				Multi:        false,
			},
			{
				Key:          "system",
				ObjectNumber: 2,
				Name:         "System",
				ClassKey:     "system_class",
				Multi:        false,
			},
		},
	}

	// Convert to InOut
	inOut := FromRequirementsScenario(originalReq)

	// Convert back to requirements
	convertedBack := inOut.ToRequirements()

	// Check fields
	if convertedBack.Key != originalReq.Key ||
		convertedBack.Name != originalReq.Name ||
		convertedBack.Details != originalReq.Details {
		t.Errorf("Basic fields mismatch: got %+v, want %+v", convertedBack, originalReq)
	}

	// Check steps
	if convertedBack.Steps.Description != originalReq.Steps.Description ||
		convertedBack.Steps.EventKey != originalReq.Steps.EventKey {
		t.Errorf("Steps mismatch: got %+v, want %+v", convertedBack.Steps, originalReq.Steps)
	}

	// Check objects
	if len(convertedBack.Objects) != len(originalReq.Objects) {
		t.Errorf("Objects length mismatch: got %d, want %d", len(convertedBack.Objects), len(originalReq.Objects))
	}
}
