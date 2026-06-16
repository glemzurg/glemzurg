package generate

import (
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/generate/req_flat"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/test_helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateScenarioMermaidContents_HappyPath(t *testing.T) {
	reqs := req_flat.NewRequirements(test_helper.GetTestModel())
	reqs.PrepLookups()

	scenarios := reqs.ScenarioLookup()
	var happyScenarioKey string
	for key, scenario := range scenarios {
		if scenario.Name == "Happy Path" {
			happyScenarioKey = key
			break
		}
	}
	require.NotEmpty(t, happyScenarioKey)

	contents, err := generateScenarioMermaidContents(reqs, scenarios[happyScenarioKey])
	require.NoError(t, err)

	assert.Contains(t, contents, `participant `)
	aliceIdx := strings.Index(contents, ` as Alice<br/>Customer`)
	orderIdx := strings.Index(contents, ` as Order 42`)
	productIdx := strings.Index(contents, ` as *<br/>Product`)
	assert.Greater(t, orderIdx, aliceIdx, "objects should follow YAML order: customer before order")
	assert.Greater(t, productIdx, orderIdx, "objects should follow YAML order: order before product")
	assert.Contains(t, contents, `Customer submits orderSubmit(`)
	assert.Contains(t, contents, "loop while items remain")
	assert.Contains(t, contents, "alt order is valid")
	assert.Contains(t, contents, "else order is invalid")
	assert.NotContains(t, contents, `"`)
	assert.Contains(t, contents, "(delete)")
	assert.Contains(t, contents, "Scenario: View Details")
	assert.Contains(t, contents, "Scenario: Error Path")
}

func TestGenerateScenarioMermaidContents_EmptyScenario(t *testing.T) {
	reqs := req_flat.NewRequirements(test_helper.GetTestModel())
	reqs.PrepLookups()

	scenarios := reqs.ScenarioLookup()
	var errorScenarioKey string
	for key, scenario := range scenarios {
		if scenario.Name == "Error Path" {
			errorScenarioKey = key
			break
		}
	}
	require.NotEmpty(t, errorScenarioKey)

	contents, err := generateScenarioMermaidContents(reqs, scenarios[errorScenarioKey])
	require.NoError(t, err)
	assert.Contains(t, contents, "No actors defined")
}

func TestGenerateUseCaseMdContents_EmbedsMermaidSequence(t *testing.T) {
	reqs := req_flat.NewRequirements(test_helper.GetTestModel())
	reqs.PrepLookups()

	useCases := reqs.UseCaseLookup()
	var placeOrderKey string
	for key, useCase := range useCases {
		if useCase.Name == "Place Order" {
			placeOrderKey = key
			break
		}
	}
	require.NotEmpty(t, placeOrderKey)

	contents, err := generateUseCaseMdContents(reqs, useCases[placeOrderKey])
	require.NoError(t, err)

	assert.Contains(t, contents, "```mermaid")
	assert.Contains(t, contents, "sequenceDiagram")
	assert.NotContains(t, contents, ".scenario.", "use case markdown should not link external scenario SVG files")
}
