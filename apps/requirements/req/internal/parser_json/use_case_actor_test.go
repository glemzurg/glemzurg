package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
)

func TestUseCaseActorInOutRoundTrip(t *testing.T) {
	original := requirements.UseCaseActor{
		UmlComment: "comment",
	}

	inOut := FromRequirementsUseCaseActor(original)
	back := inOut.ToRequirements()

	if back != original {
		t.Errorf("Round trip failed: got %+v, want %+v", back, original)
	}
}
