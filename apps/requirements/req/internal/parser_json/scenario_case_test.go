package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/scenario"
	"github.com/stretchr/testify/assert"
)

func TestCaseInOutConversionRoundTrip(t *testing.T) {
	original := scenario.Case{
		Condition: "x > 5",
		Statements: []scenario.Node{
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
