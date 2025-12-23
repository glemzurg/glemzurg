package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/data_type"
	"github.com/stretchr/testify/assert"
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

	assert.Equal(t, original.ConstraintType, back.ConstraintType)
	assert.Nil(t, back.Span)
	assert.NotNil(t, back.Reference)
	assert.Equal(t, *original.Reference, *back.Reference)
	assert.NotNil(t, back.EnumOrdered)
	assert.Equal(t, *original.EnumOrdered, *back.EnumOrdered)
	assert.Len(t, back.Enums, len(original.Enums))
	assert.NotNil(t, back.ObjectClassKey)
	assert.Equal(t, *original.ObjectClassKey, *back.ObjectClassKey)
}
