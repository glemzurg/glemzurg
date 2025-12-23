package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/data_type"
)

func TestDataTypeInOutRoundTrip(t *testing.T) {
	min := 1
	max := 10
	original := data_type.DataType{
		Key:              "dt1",
		CollectionType:   "ordered",
		CollectionUnique: nil,
		CollectionMin:    &min,
		CollectionMax:    &max,
		Atomic:           nil,
		RecordFields:     nil,
	}

	inOut := FromRequirementsDataType(original)
	back := inOut.ToRequirements()

	if back.Key != original.Key || back.CollectionType != original.CollectionType ||
		back.CollectionUnique != original.CollectionUnique ||
		(back.CollectionMin == nil && original.CollectionMin != nil) || (back.CollectionMin != nil && *back.CollectionMin != *original.CollectionMin) ||
		(back.CollectionMax == nil && original.CollectionMax != nil) || (back.CollectionMax != nil && *back.CollectionMax != *original.CollectionMax) {
		t.Errorf("Round trip failed: got %+v, want %+v", back, original)
	}
}