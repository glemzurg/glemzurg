package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"
	"github.com/stretchr/testify/assert"
)

func TestUseCaseActorInOutRoundTrip(t *testing.T) {
	original := model_use_case.Actor{
		UmlComment: "comment",
	}

	inOut := FromRequirementsUseCaseActor(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}
