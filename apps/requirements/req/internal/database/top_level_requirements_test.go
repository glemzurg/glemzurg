package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/test_helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestRequirementsSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(RequirementsSuite))
}

type RequirementsSuite struct {
	suite.Suite
	db *sql.DB
}

func (suite *RequirementsSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

}

func (suite *RequirementsSuite) TestWriteRead() {

	input := test_helper.GetTestModel()

	// Validate the model tree before testing.
	err := input.Validate()
	assert.Nil(suite.T(), err, "input model should be valid")

	// Nothing in database yet.
	output, err := ReadModel(suite.db, input.Key)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), output)

	// Write model to the database.
	err = WriteModel(suite.db, input)
	assert.Nil(suite.T(), err)

	// Write model to the database a second time, should be safe (idempotent).
	err = WriteModel(suite.db, input)
	assert.Nil(suite.T(), err)

	// Read model from the database.
	output, err = ReadModel(suite.db, input.Key)
	assert.Nil(suite.T(), err)

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
