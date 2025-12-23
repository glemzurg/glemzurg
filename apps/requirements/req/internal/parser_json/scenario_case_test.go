package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/stretchr/testify/assert"
)

func TestCaseInOutConversionRoundTrip(t *testing.T) {
	original := requirements.Case{
		Condition: "x > 5",
		Statements: []requirements.Node{
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
