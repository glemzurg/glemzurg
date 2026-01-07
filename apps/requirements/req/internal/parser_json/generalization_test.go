package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_class"
	"github.com/stretchr/testify/assert"
)

func TestGeneralizationInOutRoundTrip(t *testing.T) {
	original := model_class.Generalization{
		Key:        "gen1",
		Name:       "Gen1",
		Details:    "Details",
		IsComplete: true,
		IsStatic:   false,
		UmlComment: "comment",
	}

	inOut := FromRequirementsGeneralization(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}
