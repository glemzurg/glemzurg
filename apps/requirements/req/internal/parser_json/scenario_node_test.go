package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/scenario"
	"github.com/stretchr/testify/assert"
)

func TestNodeInOutConversionRoundTrip(t *testing.T) {
	original := scenario.Node{
		Statements: []scenario.Node{
			{
				Description: "First step",
			},
			{
				Description: "Second step",
			},
		},
		Cases: []scenario.Case{
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
