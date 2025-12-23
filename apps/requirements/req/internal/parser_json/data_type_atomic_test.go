package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/data_type"
	"github.com/stretchr/testify/assert"
)

func TestAtomicInOutRoundTrip(t *testing.T) {

	lowerVal := 1
	lowerDenom := 2
	higherVal := 10
	higherDenom := 20
	originalSpan := data_type.AtomicSpan{
		LowerType:         "closed",
		LowerValue:        &lowerVal,
		LowerDenominator:  &lowerDenom,
		HigherType:        "open",
		HigherValue:       &higherVal,
		HigherDenominator: &higherDenom,
		Units:             "kg",
		Precision:         0.1,
	}

	ref := "ref1"
	objKey := "class1"
	ordered := true
	original := data_type.Atomic{
		ConstraintType: "enumeration",
		Span:           &originalSpan,
		Reference:      &ref,
		EnumOrdered:    &ordered,
		Enums: []data_type.AtomicEnum{
			{Value: "val1", SortOrder: 1},
			{Value: "val2", SortOrder: 2},
		},
		ObjectClassKey: &objKey,
	}

	inOut := FromRequirementsAtomic(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}
