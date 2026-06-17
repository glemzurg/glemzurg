package parser_ai

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInputClassMarshalJSON_preservesAttributeOrder(t *testing.T) {
	class := &inputClass{
		Name: "Order",
		Attributes: []inputAttribute{
			{Key: "order_date", Name: "Order Date"},
			{Key: "total", Name: "Total"},
			{Key: "status", Name: "Status"},
		},
	}

	data, err := json.MarshalIndent(class, "", "    ")
	require.NoError(t, err)

	var attrs []inputAttribute
	require.NoError(t, json.Unmarshal(extractAttributesJSON(t, data), &attrs))
	require.Equal(t, []string{"order_date", "total", "status"}, attributeKeysFromSlice(attrs))
}

func TestParseClass_preservesAttributeOrder(t *testing.T) {
	content := []byte(`{
		"name": "Order",
		"attributes": [
			{"key": "order_date", "name": "Order Date"},
			{"key": "total", "name": "Total"},
			{"key": "status", "name": "Status"}
		]
	}`)

	class, err := parseClass(content, "order.class.json")
	require.NoError(t, err)
	require.Equal(t, []string{"order_date", "total", "status"}, class.attributeKeysInOrder())
}

func attributeKeysFromSlice(attrs []inputAttribute) []string {
	keys := make([]string, len(attrs))
	for i, attr := range attrs {
		keys[i] = attr.Key
	}
	return keys
}

func extractAttributesJSON(t *testing.T, classJSON []byte) json.RawMessage {
	t.Helper()
	var raw struct {
		Attributes json.RawMessage `json:"attributes"`
	}
	require.NoError(t, json.Unmarshal(classJSON, &raw))
	return raw.Attributes
}
