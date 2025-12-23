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

	assert.Equal(t, original.Key, back.Key)
	assert.Equal(t, original.Name, back.Name)
	assert.Equal(t, original.Details, back.Details)
	assert.Equal(t, original.DataTypeRules, back.DataTypeRules)
	assert.Equal(t, original.DerivationPolicy, back.DerivationPolicy)
	assert.Equal(t, original.Nullable, back.Nullable)
	assert.Equal(t, original.UmlComment, back.UmlComment)
	assert.Equal(t, original.IndexNums, back.IndexNums)
	assert.Equal(t, original.DataType, back.DataType)
}
