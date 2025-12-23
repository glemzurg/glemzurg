package parser_json

import (
	"encoding/json"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, original.Key, unmarshaled.Key)
	assert.Equal(t, original.Name, unmarshaled.Name)
	assert.Equal(t, original.Details, unmarshaled.Details)

	// Check steps
	assert.Equal(t, original.Steps.Description, unmarshaled.Steps.Description)
	assert.Equal(t, original.Steps.EventKey, unmarshaled.Steps.EventKey)

	// Check objects
	assert.Len(t, unmarshaled.Objects, len(original.Objects))
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
	assert.Equal(t, originalReq.Key, convertedBack.Key)
	assert.Equal(t, originalReq.Name, convertedBack.Name)
	assert.Equal(t, originalReq.Details, convertedBack.Details)

	// Check steps
	assert.Equal(t, originalReq.Steps.Description, convertedBack.Steps.Description)
	assert.Equal(t, originalReq.Steps.EventKey, convertedBack.Steps.EventKey)

	// Check objects
	assert.Len(t, convertedBack.Objects, len(originalReq.Objects))
}
