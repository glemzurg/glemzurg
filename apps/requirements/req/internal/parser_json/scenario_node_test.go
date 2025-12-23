package parser_json

import (
	"encoding/json"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
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
	if unmarshaled.Loop != original.Loop || unmarshaled.Description != original.Description ||
		unmarshaled.FromObjectKey != original.FromObjectKey || unmarshaled.ToObjectKey != original.ToObjectKey ||
		unmarshaled.EventKey != original.EventKey || unmarshaled.ScenarioKey != original.ScenarioKey ||
		unmarshaled.AttributeKey != original.AttributeKey || unmarshaled.IsDelete != original.IsDelete {
		t.Errorf("Basic fields mismatch: got %+v, want %+v", unmarshaled, original)
	}

	// Check statements
	if len(unmarshaled.Statements) != len(original.Statements) {
		t.Errorf("Statements length mismatch: got %d, want %d", len(unmarshaled.Statements), len(original.Statements))
	}

	// Check cases
	if len(unmarshaled.Cases) != len(original.Cases) {
		t.Errorf("Cases length mismatch: got %d, want %d", len(unmarshaled.Cases), len(original.Cases))
	}
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
	if convertedBack.Loop != originalReq.Loop || convertedBack.Description != originalReq.Description ||
		convertedBack.FromObjectKey != originalReq.FromObjectKey || convertedBack.ToObjectKey != originalReq.ToObjectKey ||
		convertedBack.EventKey != originalReq.EventKey || convertedBack.ScenarioKey != originalReq.ScenarioKey ||
		convertedBack.AttributeKey != originalReq.AttributeKey || convertedBack.IsDelete != originalReq.IsDelete {
		t.Errorf("Basic fields mismatch: got %+v, want %+v", convertedBack, originalReq)
	}

	// Check statements
	if len(convertedBack.Statements) != len(originalReq.Statements) {
		t.Errorf("Statements length mismatch: got %d, want %d", len(convertedBack.Statements), len(originalReq.Statements))
	}

	// Check cases
	if len(convertedBack.Cases) != len(originalReq.Cases) {
		t.Errorf("Cases length mismatch: got %d, want %d", len(convertedBack.Cases), len(originalReq.Cases))
	}
}
