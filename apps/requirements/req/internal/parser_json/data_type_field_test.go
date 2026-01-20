package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_data_type"
	"github.com/stretchr/testify/assert"
)

func TestFieldInOutRoundTrip(t *testing.T) {
	original := model_data_type.Field{
		Name: "field1",
		FieldDataType: &model_data_type.DataType{
			Key:              "dt1",
			CollectionType:   "ordered",
			CollectionUnique: t_BoolPtr(true),
			CollectionMin:    t_IntPtr(1),
			CollectionMax:    t_IntPtr(10),
			Atomic:           &model_data_type.Atomic{ConstraintType: "span"},
			RecordFields:     nil,
		},
	}

	inOut := FromRequirementsField(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}
