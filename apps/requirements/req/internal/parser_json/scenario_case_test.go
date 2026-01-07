package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_scenario"
	"github.com/stretchr/testify/assert"
)

func TestCaseInOutConversionRoundTrip(t *testing.T) {
	original := model_scenario.Case{
		Condition: "x > 5",
		Statements: []model_scenario.Node{
			{
				Description: "Do something",
			},
			{
				Description: "Do something2",
			},
		},
	}

	inOut := FromRequirementsCase(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)
}
