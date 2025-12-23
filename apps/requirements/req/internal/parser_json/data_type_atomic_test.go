package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/data_type"
)

func TestAtomicInOutRoundTrip(t *testing.T) {
	ref := "ref1"
	objKey := "class1"
	ordered := true
	original := data_type.Atomic{
		ConstraintType: "enumeration",
		Span:           nil,
		Reference:      &ref,
		EnumOrdered:    &ordered,
		Enums: []data_type.AtomicEnum{
			{Value: "val1", SortOrder: 1},
		},
		ObjectClassKey: &objKey,
	}

	inOut := FromRequirementsAtomic(original)
	back := inOut.ToRequirements()

	if back.ConstraintType != original.ConstraintType || (back.Reference == nil && original.Reference != nil) || (back.Reference != nil && *back.Reference != *original.Reference) ||
		(back.EnumOrdered == nil && original.EnumOrdered != nil) || (back.EnumOrdered != nil && *back.EnumOrdered != *original.EnumOrdered) ||
		len(back.Enums) != len(original.Enums) || (back.ObjectClassKey == nil && original.ObjectClassKey != nil) || (back.ObjectClassKey != nil && *back.ObjectClassKey != *original.ObjectClassKey) {
		t.Errorf("Round trip failed: got %+v, want %+v", back, original)
	}
}