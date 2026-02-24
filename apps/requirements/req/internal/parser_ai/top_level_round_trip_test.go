package parser_ai

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/test_helper"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestRoundTripSuite(t *testing.T) {
	suite.Run(t, new(RoundTripSuite))
}

type RoundTripSuite struct {
	suite.Suite
}

func (suite *RoundTripSuite) TestRoundTrip() {

	input := test_helper.GetTestModel()

	// Validate the model before writing.
	err := input.Validate()
	assert.Nil(suite.T(), err, "input model should be valid")

	// Write using the top-level parser (original round-trip behavior).
	tempDir := suite.T().TempDir()
	err = WriteModel(input, tempDir)
	assert.Nil(suite.T(), err, "writing model should succeed")

	// Read from the temporary folder.
	output, err := ReadModel(tempDir)
	assert.Nil(suite.T(), err, "parsing model should succeed")

	// The parsed model's Key will be the tempDir path, not our original key.
	// Overwrite it for comparison since the parser uses the modelPath as the key.
	output.Key = input.Key

	// Compare progressively larger slices of the model tree to isolate mismatches.
	// Each check focuses on a specific layer so failures point to the right area.

	// 1. Model + direct children only (no class associations).
	assert.Equal(suite.T(),
		test_helper.PruneToModelOnly(input),
		test_helper.PruneToModelOnly(output),
		"PruneToModelOnly: model with direct children does not match")

	// 2. Add subdomains, classes with attributes, class generalizations (no associations, no states, no use cases).
	assert.Equal(suite.T(),
		test_helper.PruneToClassAttributes(input),
		test_helper.PruneToClassAttributes(output),
		"PruneToClassAttributes: classes and attributes do not match")

	// 3. Add class associations at all levels.
	assert.Equal(suite.T(),
		test_helper.PruneToClassAssociations(input),
		test_helper.PruneToClassAssociations(output),
		"PruneToClassAssociations: class associations do not match")

	// 4. Add states and all state machine sub-parts.
	assert.Equal(suite.T(),
		test_helper.PruneToStateMachine(input),
		test_helper.PruneToStateMachine(output),
		"PruneToStateMachine: state machine does not match")

	// 5. Full model except steps (scenarios with Steps=nil).
	assert.Equal(suite.T(),
		test_helper.PruneToNoSteps(input),
		test_helper.PruneToNoSteps(output),
		"PruneToNoSteps: model without steps does not match")

	// 6. Compare scenarios individually for better error diagnostics on steps.
	inputScenarios := test_helper.ExtractScenarios(input)
	outputScenarios := test_helper.ExtractScenarios(output)
	assert.Equal(suite.T(), len(inputScenarios), len(outputScenarios), "scenario count mismatch")
	for i := range inputScenarios {
		if i >= len(outputScenarios) {
			break
		}
		assert.Equal(suite.T(), inputScenarios[i].Path, outputScenarios[i].Path,
			"scenario path mismatch at index %d", i)
		assert.Equal(suite.T(), inputScenarios[i].Scenario, outputScenarios[i].Scenario,
			"scenario %q does not match", inputScenarios[i].Path)
	}

	// 7. Compare the entire model tree.
	assert.Equal(suite.T(), input, output, "Full model does not match")
}
