package parser_json

import (
	"encoding/json"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, original.Condition, unmarshaled.Condition)

	assert.Len(t, unmarshaled.Statements, len(original.Statements))

	if len(unmarshaled.Statements) > 0 {
		assert.Equal(t, original.Statements[0].Description, unmarshaled.Statements[0].Description)
		assert.Equal(t, original.Statements[0].FromObjectKey, unmarshaled.Statements[0].FromObjectKey)
		assert.Equal(t, original.Statements[0].ToObjectKey, unmarshaled.Statements[0].ToObjectKey)
		assert.Equal(t, original.Statements[0].EventKey, unmarshaled.Statements[0].EventKey)
		assert.Equal(t, original.Statements[0].IsDelete, unmarshaled.Statements[0].IsDelete)
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
	assert.Equal(t, originalReq.Condition, convertedBack.Condition)

	assert.Len(t, convertedBack.Statements, len(originalReq.Statements))

	if len(convertedBack.Statements) > 0 {
		assert.Equal(t, originalReq.Statements[0].Description, convertedBack.Statements[0].Description)
		assert.Equal(t, originalReq.Statements[0].FromObjectKey, convertedBack.Statements[0].FromObjectKey)
		assert.Equal(t, originalReq.Statements[0].ToObjectKey, convertedBack.Statements[0].ToObjectKey)
		assert.Equal(t, originalReq.Statements[0].EventKey, convertedBack.Statements[0].EventKey)
		assert.Equal(t, originalReq.Statements[0].IsDelete, convertedBack.Statements[0].IsDelete)
	}
}
