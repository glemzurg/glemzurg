package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_data_type"
	"github.com/stretchr/testify/assert"
)

func TestAtomicInOutRoundTrip(t *testing.T) {

	original := model_data_type.Atomic{
		ConstraintType: "enumeration",
		Span: &model_data_type.AtomicSpan{
			LowerType:         "closed",
			LowerValue:        t_IntPtr(1),
			LowerDenominator:  t_IntPtr(2),
			HigherType:        "open",
			HigherValue:       t_IntPtr(10),
			HigherDenominator: t_IntPtr(20),
			Units:             "kg",
			Precision:         0.1,
		},
		Reference:   t_StrPtr("ref1"),
		EnumOrdered: t_BoolPtr(true),
		Enums: []model_data_type.AtomicEnum{
			{Value: "val1", SortOrder: 1},
			{Value: "val2", SortOrder: 2},
		},
		ObjectClassKey: t_StrPtr("class1"),
	}

	inOut := FromRequirementsAtomic(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}
