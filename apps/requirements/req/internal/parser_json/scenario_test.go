package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/scenario"
	"github.com/stretchr/testify/assert"
)

func TestScenarioInOutConversionRoundTrip(t *testing.T) {
	original := scenario.Scenario{
		Key:     "scenario1",
		Name:    "Login Scenario",
		Details: "User logs into the system",
		Steps: scenario.Node{
			Description: "User enters credentials",
			EventKey:    "login",
		},
		Objects: []scenario.ScenarioObject{
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
