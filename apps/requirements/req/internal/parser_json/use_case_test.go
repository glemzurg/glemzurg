package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_use_case"
	"github.com/stretchr/testify/assert"
)

func TestUseCaseInOutConversionRoundTrip(t *testing.T) {
	original := model_use_case.UseCase{
		Key:        "usecase1",
		Name:       "Login Use Case",
		Details:    "User logs into the system",
		Level:      "sea",
		ReadOnly:   true,
		UmlComment: "Login flow",
		Actors: map[string]model_use_case.UseCaseActor{
			"user": {
				UmlComment: "The user",
			},
		},
		Scenarios: []model_scenario.Scenario{
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
