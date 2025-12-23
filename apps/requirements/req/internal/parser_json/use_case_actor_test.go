package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
)

func TestUseCaseActorInOutRoundTrip(t *testing.T) {
	original := requirements.UseCaseActor{
		UmlComment: "comment",
	}

	inOut := FromRequirementsUseCaseActor(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}
