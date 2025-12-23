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
	assert.Equal(t, original, back)
}
