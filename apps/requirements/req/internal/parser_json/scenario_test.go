package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
)

func TestScenarioInOutConversionRoundTrip(t *testing.T) {
	original := requirements.Scenario{
		Key:     "scenario1",
		Name:    "Login Scenario",
		Details: "User logs into the system",
		Steps: requirements.Node{
			Description: "User enters credentials",
			EventKey:    "login",
		},
		Objects: []requirements.ScenarioObject{
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
