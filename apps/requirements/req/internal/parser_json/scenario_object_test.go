package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
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
	if back.Key != original.Key || back.ObjectNumber != original.ObjectNumber || back.Name != original.Name ||
		back.NameStyle != original.NameStyle || back.ClassKey != original.ClassKey || back.Multi != original.Multi || back.UmlComment != original.UmlComment {
		t.Errorf("Round trip failed: got %+v, want %+v", back, original)
	}
}
