package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
)

func TestAttributeInOutRoundTrip(t *testing.T) {
	original := requirements.Attribute{
		Key:              "attr1",
		Name:             "Name",
		Details:          "Details",
		DataTypeRules:    "string",
		DerivationPolicy: "",
		Nullable:         false,
		UmlComment:       "comment",
		IndexNums:        []uint{1},
		DataType:         nil, // TODO: test with data type
	}

	inOut := FromRequirementsAttribute(original)
	back := inOut.ToRequirements()

	if back.Key != original.Key || back.Name != original.Name || back.Details != original.Details ||
		back.DataTypeRules != original.DataTypeRules || back.DerivationPolicy != original.DerivationPolicy ||
		back.Nullable != original.Nullable || back.UmlComment != original.UmlComment ||
		len(back.IndexNums) != len(original.IndexNums) || back.DataType != original.DataType {
		t.Errorf("Round trip failed: got %+v, want %+v", back, original)
	}
}
