package parser_json

import (
	"encoding/json"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, original.Key, unmarshaled.Key)
	assert.Equal(t, original.Name, unmarshaled.Name)
	assert.Equal(t, original.Details, unmarshaled.Details)
	assert.Equal(t, original.Level, unmarshaled.Level)
	assert.Equal(t, original.ReadOnly, unmarshaled.ReadOnly)
	assert.Equal(t, original.UmlComment, unmarshaled.UmlComment)

	// Check actors
	assert.Len(t, unmarshaled.Actors, len(original.Actors))

	// Check scenarios
	assert.Len(t, unmarshaled.Scenarios, len(original.Scenarios))
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
	assert.Equal(t, originalReq.Key, convertedBack.Key)
	assert.Equal(t, originalReq.Name, convertedBack.Name)
	assert.Equal(t, originalReq.Details, convertedBack.Details)
	assert.Equal(t, originalReq.Level, convertedBack.Level)
	assert.Equal(t, originalReq.ReadOnly, convertedBack.ReadOnly)
	assert.Equal(t, originalReq.UmlComment, convertedBack.UmlComment)

	// Check actors
	assert.Len(t, convertedBack.Actors, len(originalReq.Actors))

	// Check scenarios
	assert.Len(t, convertedBack.Scenarios, len(originalReq.Scenarios))
}
