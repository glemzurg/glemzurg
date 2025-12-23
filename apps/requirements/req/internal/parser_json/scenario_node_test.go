package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
)

func TestNodeInOutConversionRoundTrip(t *testing.T) {
	original := requirements.Node{
		Statements: []requirements.Node{
			{
				Description: "First step",
			},
			{
				Description: "Second step",
			},
		},
		Cases: []requirements.Case{
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
