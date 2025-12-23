package parser_json

import (
	"encoding/json"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
)

func TestNodeInOutJSONRoundTrip(t *testing.T) {
	original := nodeInOut{
		Statements: []nodeInOut{
			{
				Description:   "First step",
				FromObjectKey: "user",
				ToObjectKey:   "system",
				EventKey:      "login",
				IsDelete:      false,
			},
		},
		Cases: []caseInOut{
			{
				Condition: "success",
				Statements: []nodeInOut{
					{
						Description: "Success case",
						EventKey:    "success",
					},
				},
			},
		},
		Loop:          "while condition",
		Description:   "Main scenario",
		FromObjectKey: "client",
		ToObjectKey:   "server",
		EventKey:      "request",
		ScenarioKey:   "scenario1",
		AttributeKey:  "status",
		IsDelete:      false,
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal to JSON: %v", err)
	}

	// Unmarshal back
	var unmarshaled nodeInOut
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal from JSON: %v", err)
	}

	// Check basic fields
	assert.Equal(t, original.Loop, unmarshaled.Loop)
	assert.Equal(t, original.Description, unmarshaled.Description)
	assert.Equal(t, original.FromObjectKey, unmarshaled.FromObjectKey)
	assert.Equal(t, original.ToObjectKey, unmarshaled.ToObjectKey)
	assert.Equal(t, original.EventKey, unmarshaled.EventKey)
	assert.Equal(t, original.ScenarioKey, unmarshaled.ScenarioKey)
	assert.Equal(t, original.AttributeKey, unmarshaled.AttributeKey)
	assert.Equal(t, original.IsDelete, unmarshaled.IsDelete)

	// Check statements
	assert.Len(t, unmarshaled.Statements, len(original.Statements))

	// Check cases
	assert.Len(t, unmarshaled.Cases, len(original.Cases))
}

func TestNodeInOutConversionRoundTrip(t *testing.T) {
	originalReq := requirements.Node{
		Statements: []requirements.Node{
			{
				Description:   "First step",
				FromObjectKey: "user",
				ToObjectKey:   "system",
				EventKey:      "login",
				IsDelete:      false,
			},
		},
		Cases: []requirements.Case{
			{
				Condition: "success",
				Statements: []requirements.Node{
					{
						Description: "Success case",
						EventKey:    "success",
					},
				},
			},
		},
		Loop:          "while condition",
		Description:   "Main scenario",
		FromObjectKey: "client",
		ToObjectKey:   "server",
		EventKey:      "request",
		ScenarioKey:   "scenario1",
		AttributeKey:  "status",
		IsDelete:      false,
	}

	// Convert to InOut
	inOut := FromRequirementsNode(originalReq)

	// Convert back to requirements
	convertedBack := inOut.ToRequirements()

	// Check basic fields
	assert.Equal(t, originalReq.Loop, convertedBack.Loop)
	assert.Equal(t, originalReq.Description, convertedBack.Description)
	assert.Equal(t, originalReq.FromObjectKey, convertedBack.FromObjectKey)
	assert.Equal(t, originalReq.ToObjectKey, convertedBack.ToObjectKey)
	assert.Equal(t, originalReq.EventKey, convertedBack.EventKey)
	assert.Equal(t, originalReq.ScenarioKey, convertedBack.ScenarioKey)
	assert.Equal(t, originalReq.AttributeKey, convertedBack.AttributeKey)
	assert.Equal(t, originalReq.IsDelete, convertedBack.IsDelete)

	// Check statements
	assert.Len(t, convertedBack.Statements, len(originalReq.Statements))

	// Check cases
	assert.Len(t, convertedBack.Cases, len(originalReq.Cases))
}
