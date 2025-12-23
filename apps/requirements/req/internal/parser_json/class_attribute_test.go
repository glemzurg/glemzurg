package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
)

func TestAttributeInOutRoundTrip(t *testing.T) {
	original := requirements.Attribute{
		Key:              "attr1",
		Name:             "Name",
		Details:          "Details",
		DataTypeRules:    "string",
		DerivationPolicy: "",
		Nullable:         false,
		UmlComment:       "comment",
		IndexNums:        []uint{1},
		DataType:         nil, // TODO: test with data type
	}

	inOut := FromRequirementsAttribute(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}
