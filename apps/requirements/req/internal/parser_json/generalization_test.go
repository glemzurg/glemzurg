package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
)

func TestGeneralizationInOutRoundTrip(t *testing.T) {
	original := requirements.Generalization{
		Key:           "gen1",
		Name:          "Gen1",
		Details:       "Details",
		IsComplete:    true,
		IsStatic:      false,
		UmlComment:    "comment",
		SuperclassKey: "class1",
		SubclassKeys:  []string{"class2"},
	}

	inOut := FromRequirementsGeneralization(original)
	back := inOut.ToRequirements()

	// Check fields that are preserved
	assert.Equal(t, original.Key, back.Key)
	assert.Equal(t, original.Name, back.Name)
	assert.Equal(t, original.Details, back.Details)
	assert.Equal(t, original.IsComplete, back.IsComplete)
	assert.Equal(t, original.IsStatic, back.IsStatic)
	assert.Equal(t, original.UmlComment, back.UmlComment)
}
