package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_scenario"
	"github.com/stretchr/testify/assert"
)

func TestNodeInOutConversionRoundTrip(t *testing.T) {
	original := model_scenario.Node{
		Statements: []model_scenario.Node{
			{
				Description: "First step",
			},
			{
				Description: "Second step",
			},
		},
		Cases: []model_scenario.Case{
			{
				Condition: "success",
			},
			{
				Condition: "alternative",
			},
		},
		Loop:          "while condition",
		Description:   "Main scenario",
		FromObjectKey: "client",
		ToObjectKey:   "server",
		EventKey:      "request",
		ScenarioKey:   "scenario1",
		AttributeKey:  "status",
		IsDelete:      true,
	}

	inOut := FromRequirementsNode(original)
	back := inOut.ToRequirements()
	assert.Equal(t, original, back)

}
