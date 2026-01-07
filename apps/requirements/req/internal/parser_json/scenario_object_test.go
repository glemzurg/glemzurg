package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_scenario"
	"github.com/stretchr/testify/assert"
)

func TestScenarioObjectInOutRoundTrip(t *testing.T) {
	original := model_scenario.ScenarioObject{
		Key:          "obj1",
		ObjectNumber: 1,
		Name:         "Object1",
		NameStyle:    "name",
		ClassKey:     "class1",
		Multi:        true,
		UmlComment:   "comment",
	}

	inOut := FromRequirementsScenarioObject(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}
