package database

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"
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

// compareScenarios compares each scenario individually between input and output models.
// This provides more targeted error messages than comparing the entire model tree at once.
func compareScenarios(t *testing.T, input, output req_model.Model) {
	for domainKey, inputDomain := range input.Domains {
		outputDomain, ok := output.Domains[domainKey]
		if !assert.True(t, ok, fmt.Sprintf("domain %q missing from output", domainKey)) {
			continue
		}
		for subdomainKey, inputSubdomain := range inputDomain.Subdomains {
			outputSubdomain, ok := outputDomain.Subdomains[subdomainKey]
			if !assert.True(t, ok, fmt.Sprintf("subdomain %q missing from output", subdomainKey)) {
				continue
			}
			for useCaseKey, inputUseCase := range inputSubdomain.UseCases {
				outputUseCase, ok := outputSubdomain.UseCases[useCaseKey]
				if !assert.True(t, ok, fmt.Sprintf("use case %q missing from output", useCaseKey)) {
					continue
				}
				compareUseCaseScenarios(t, domainKey, subdomainKey, useCaseKey, inputUseCase.Scenarios, outputUseCase.Scenarios)
			}
		}
	}
}

// compareUseCaseScenarios compares each scenario in a use case between input and output.
func compareUseCaseScenarios(t *testing.T, domainKey, subdomainKey, useCaseKey identity.Key, inputScenarios, outputScenarios map[identity.Key]model_scenario.Scenario) {
	path := fmt.Sprintf("domain %q > subdomain %q > use case %q", domainKey, subdomainKey, useCaseKey)

	assert.Equal(t, len(inputScenarios), len(outputScenarios), fmt.Sprintf("%s: scenario count mismatch", path))

	for scenarioKey, inputScenario := range inputScenarios {
		outputScenario, ok := outputScenarios[scenarioKey]
		if !assert.True(t, ok, fmt.Sprintf("%s: scenario %q missing from output", path, scenarioKey)) {
			continue
		}
		assert.Equal(t, inputScenario, outputScenario, fmt.Sprintf("%s > scenario %q does not match", path, scenarioKey))
	}
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

	// Compare scenarios individually for better error diagnostics.
	compareScenarios(suite.T(), input, output)

	// Compare the entire model tree.
	// If input and output model do not match, to explore this very deep tree of data, prune back to just the model and then iterate by layering in children one tier at a time  until the full tree exists again.
	assert.Equal(suite.T(), input, output, `Input and output model do not match, to explore this very deep tree of data, prune back to just the model and then iterate by layering in children one tier at a time until the full tree exists again`)
}
