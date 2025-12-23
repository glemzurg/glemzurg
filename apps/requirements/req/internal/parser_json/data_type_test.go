package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/data_type"
	"github.com/stretchr/testify/assert"
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

	assert.Equal(t, original.Key, back.Key)
	assert.Equal(t, original.CollectionType, back.CollectionType)
	assert.Equal(t, original.CollectionUnique, back.CollectionUnique)
	assert.NotNil(t, back.CollectionMin)
	assert.Equal(t, *original.CollectionMin, *back.CollectionMin)
	assert.NotNil(t, back.CollectionMax)
	assert.Equal(t, *original.CollectionMax, *back.CollectionMax)
	assert.Equal(t, original.Atomic, back.Atomic)
	assert.Equal(t, original.RecordFields, back.RecordFields)
}
