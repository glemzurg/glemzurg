package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_scenario"
	"github.com/stretchr/testify/assert"
)

func TestScenarioInOutConversionRoundTrip(t *testing.T) {
	original := model_scenario.Scenario{
		Key:     "scenario1",
		Name:    "Login Scenario",
		Details: "User logs into the system",
		Steps: model_scenario.Node{
			Description: "User enters credentials",
			EventKey:    "login",
		},
		Objects: []model_scenario.Object{
			{
				Key:          "user",
				ObjectNumber: 1,
				Name:         "User",
				ClassKey:     "user_class",
				Multi:        false,
			},
			{
				Key:          "system",
				ObjectNumber: 2,
				Name:         "System",
				ClassKey:     "system_class",
				Multi:        false,
			},
		},
	}

	inOut := FromRequirementsScenario(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}
