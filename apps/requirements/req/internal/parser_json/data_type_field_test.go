package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/data_type"
)

func TestFieldInOutRoundTrip(t *testing.T) {
	original := data_type.Field{
		Name:          "field1",
		FieldDataType: nil, // TODO: test with data type
	}

	inOut := FromRequirementsField(original)
	back := inOut.ToRequirements()

	if back.Name != original.Name || back.FieldDataType != original.FieldDataType {
		t.Errorf("Round trip failed: got %+v, want %+v", back, original)
	}
}
