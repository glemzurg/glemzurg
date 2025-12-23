package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/data_type"
	"github.com/stretchr/testify/assert"
)

func TestDataTypeInOutRoundTrip(t *testing.T) {

	original := data_type.DataType{
		Key:              "dt1",
		CollectionType:   "ordered",
		CollectionUnique: t_BoolPtr(true),
		CollectionMin:    t_IntPtr(1),
		CollectionMax:    t_IntPtr(10),
		Atomic:           &data_type.Atomic{ConstraintType: "span"},
		RecordFields: []data_type.Field{
			{
				Name: "field1",
			},
			{
				Name: "field2",
			},
		},
	}

	inOut := FromRequirementsDataType(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}
