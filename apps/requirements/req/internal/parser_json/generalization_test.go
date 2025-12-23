package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
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
	if back.Key != original.Key || back.Name != original.Name || back.Details != original.Details ||
		back.IsComplete != original.IsComplete || back.IsStatic != original.IsStatic || back.UmlComment != original.UmlComment {
		t.Errorf("Round trip failed: got %+v, want %+v", back, original)
	}
}
