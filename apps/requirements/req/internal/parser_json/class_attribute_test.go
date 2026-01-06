package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/data_type"
	"github.com/stretchr/testify/assert"
)

func TestAttributeInOutRoundTrip(t *testing.T) {
	original := requirements.Attribute{
		Key:              "attr1",
		Name:             "Name",
		Details:          "Details",
		DataTypeRules:    "string",
		DerivationPolicy: "derived",
		Nullable:         true,
		UmlComment:       "comment",
		IndexNums:        []uint{1},
		DataType: &data_type.DataType{
			Key: "dt1",
		},
	}

	inOut := FromRequirementsAttribute(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}
