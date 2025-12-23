package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/data_type"
	"github.com/stretchr/testify/assert"
)

func TestFieldInOutRoundTrip(t *testing.T) {
	original := data_type.Field{
		Name:          "field1",
		FieldDataType: nil, // TODO: test with data type
	}

	inOut := FromRequirementsField(original)
	back := inOut.ToRequirements()

	assert.Equal(t, original.Name, back.Name)
	assert.Equal(t, original.FieldDataType, back.FieldDataType)
}
