package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
)

func TestUseCaseInOutConversionRoundTrip(t *testing.T) {
	original := requirements.UseCase{
		Key:        "usecase1",
		Name:       "Login Use Case",
		Details:    "User logs into the system",
		Level:      "sea",
		ReadOnly:   true,
		UmlComment: "Login flow",
		Actors: map[string]requirements.UseCaseActor{
			"user": {
				UmlComment: "The user",
			},
		},
		Scenarios: []requirements.Scenario{
			{
				Key: "scenario1",
			},
			{
				Key: "scenario2",
			},
		},
	}

	inOut := FromRequirementsUseCase(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}
