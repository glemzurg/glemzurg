package parser_json

import (
	"encoding/json"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
)

func TestUseCaseInOutJSONRoundTrip(t *testing.T) {
	original := useCaseInOut{
		Key:        "usecase1",
		Name:       "Login Use Case",
		Details:    "User logs into the system",
		Level:      "sea",
		ReadOnly:   false,
		UmlComment: "Login flow",
		Actors: map[string]useCaseActorInOut{
			"user": {
				UmlComment: "The user",
			},
		},
		Scenarios: []scenarioInOut{
			{
				Key:     "scenario1",
				Name:    "Happy Path",
				Details: "Successful login",
				Steps: nodeInOut{
					Description: "User enters credentials",
				},
				Objects: []scenarioObjectInOut{},
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal to JSON: %v", err)
	}

	// Unmarshal back
	var unmarshaled useCaseInOut
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal from JSON: %v", err)
	}

	// Check fields
	if unmarshaled.Key != original.Key ||
		unmarshaled.Name != original.Name ||
		unmarshaled.Details != original.Details ||
		unmarshaled.Level != original.Level ||
		unmarshaled.ReadOnly != original.ReadOnly ||
		unmarshaled.UmlComment != original.UmlComment {
		t.Errorf("Basic fields mismatch: got %+v, want %+v", unmarshaled, original)
	}

	// Check actors
	if len(unmarshaled.Actors) != len(original.Actors) {
		t.Errorf("Actors length mismatch: got %d, want %d", len(unmarshaled.Actors), len(original.Actors))
	}

	// Check scenarios
	if len(unmarshaled.Scenarios) != len(original.Scenarios) {
		t.Errorf("Scenarios length mismatch: got %d, want %d", len(unmarshaled.Scenarios), len(original.Scenarios))
	}
}

func TestUseCaseInOutConversionRoundTrip(t *testing.T) {
	originalReq := requirements.UseCase{
		Key:        "usecase1",
		Name:       "Login Use Case",
		Details:    "User logs into the system",
		Level:      "sea",
		ReadOnly:   false,
		UmlComment: "Login flow",
		Actors: map[string]requirements.UseCaseActor{
			"user": {
				UmlComment: "The user",
			},
		},
		Scenarios: []requirements.Scenario{
			{
				Key:     "scenario1",
				Name:    "Happy Path",
				Details: "Successful login",
				Steps: requirements.Node{
					Description: "User enters credentials",
				},
				Objects: []requirements.ScenarioObject{},
			},
		},
	}

	// Convert to InOut
	inOut := FromRequirementsUseCase(originalReq)

	// Convert back to requirements
	convertedBack := inOut.ToRequirements()

	// Check fields
	if convertedBack.Key != originalReq.Key ||
		convertedBack.Name != originalReq.Name ||
		convertedBack.Details != originalReq.Details ||
		convertedBack.Level != originalReq.Level ||
		convertedBack.ReadOnly != originalReq.ReadOnly ||
		convertedBack.UmlComment != originalReq.UmlComment {
		t.Errorf("Basic fields mismatch: got %+v, want %+v", convertedBack, originalReq)
	}

	// Check actors
	if len(convertedBack.Actors) != len(originalReq.Actors) {
		t.Errorf("Actors length mismatch: got %d, want %d", len(convertedBack.Actors), len(originalReq.Actors))
	}

	// Check scenarios
	if len(convertedBack.Scenarios) != len(originalReq.Scenarios) {
		t.Errorf("Scenarios length mismatch: got %d, want %d", len(convertedBack.Scenarios), len(originalReq.Scenarios))
	}
}
