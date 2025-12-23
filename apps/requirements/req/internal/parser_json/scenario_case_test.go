package parser_json

import (
	"encoding/json"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
)

func TestCaseInOutJSONRoundTrip(t *testing.T) {
	original := caseInOut{
		Condition: "x > 5",
		Statements: []nodeInOut{
			{
				Description:   "Do something",
				FromObjectKey: "obj1",
				ToObjectKey:   "obj2",
				EventKey:      "event1",
				IsDelete:      false,
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal to JSON: %v", err)
	}

	// Unmarshal back
	var unmarshaled caseInOut
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal from JSON: %v", err)
	}

	// Check fields
	if unmarshaled.Condition != original.Condition {
		t.Errorf("Condition mismatch: got %q, want %q", unmarshaled.Condition, original.Condition)
	}

	if len(unmarshaled.Statements) != len(original.Statements) {
		t.Errorf("Statements length mismatch: got %d, want %d", len(unmarshaled.Statements), len(original.Statements))
	}

	if len(unmarshaled.Statements) > 0 {
		if unmarshaled.Statements[0].Description != original.Statements[0].Description ||
			unmarshaled.Statements[0].FromObjectKey != original.Statements[0].FromObjectKey ||
			unmarshaled.Statements[0].ToObjectKey != original.Statements[0].ToObjectKey ||
			unmarshaled.Statements[0].EventKey != original.Statements[0].EventKey ||
			unmarshaled.Statements[0].IsDelete != original.Statements[0].IsDelete {
			t.Errorf("Statement mismatch: got %+v, want %+v", unmarshaled.Statements[0], original.Statements[0])
		}
	}
}

func TestCaseInOutConversionRoundTrip(t *testing.T) {
	originalReq := requirements.Case{
		Condition: "x > 5",
		Statements: []requirements.Node{
			{
				Description:   "Do something",
				FromObjectKey: "obj1",
				ToObjectKey:   "obj2",
				EventKey:      "event1",
				IsDelete:      false,
			},
		},
	}

	// Convert to InOut
	inOut := FromRequirementsCase(originalReq)

	// Convert back to requirements
	convertedBack := inOut.ToRequirements()

	// Check fields
	if convertedBack.Condition != originalReq.Condition {
		t.Errorf("Condition mismatch: got %q, want %q", convertedBack.Condition, originalReq.Condition)
	}

	if len(convertedBack.Statements) != len(originalReq.Statements) {
		t.Errorf("Statements length mismatch: got %d, want %d", len(convertedBack.Statements), len(originalReq.Statements))
	}

	if len(convertedBack.Statements) > 0 {
		if convertedBack.Statements[0].Description != originalReq.Statements[0].Description ||
			convertedBack.Statements[0].FromObjectKey != originalReq.Statements[0].FromObjectKey ||
			convertedBack.Statements[0].ToObjectKey != originalReq.Statements[0].ToObjectKey ||
			convertedBack.Statements[0].EventKey != originalReq.Statements[0].EventKey ||
			convertedBack.Statements[0].IsDelete != originalReq.Statements[0].IsDelete {
			t.Errorf("Statement mismatch: got %+v, want %+v", convertedBack.Statements[0], originalReq.Statements[0])
		}
	}
}
