package parser_json

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	back, err := inOut.ToRequirements()
	require.NoError(t, err)
	assert.Equal(t, original, back)
}
