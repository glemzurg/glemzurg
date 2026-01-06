package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/use_case"
	"github.com/stretchr/testify/assert"
)

func TestUseCaseActorInOutRoundTrip(t *testing.T) {
	original := use_case.UseCaseActor{
		UmlComment: "comment",
	}

	inOut := FromRequirementsUseCaseActor(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}
