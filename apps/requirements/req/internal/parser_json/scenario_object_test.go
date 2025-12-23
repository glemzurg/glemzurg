package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
)

func TestScenarioObjectInOutRoundTrip(t *testing.T) {
	original := requirements.ScenarioObject{
		Key:          "obj1",
		ObjectNumber: 1,
		Name:         "Object1",
		NameStyle:    "name",
		ClassKey:     "class1",
		Multi:        false,
		UmlComment:   "comment",
		Class:        requirements.Class{Key: "class1"},
	}

	inOut := FromRequirementsScenarioObject(original)
	back := inOut.ToRequirements()

	// Check fields that are preserved
	assert.Equal(t, original.Key, back.Key)
	assert.Equal(t, original.ObjectNumber, back.ObjectNumber)
	assert.Equal(t, original.Name, back.Name)
	assert.Equal(t, original.NameStyle, back.NameStyle)
	assert.Equal(t, original.ClassKey, back.ClassKey)
	assert.Equal(t, original.Multi, back.Multi)
	assert.Equal(t, original.UmlComment, back.UmlComment)
}
