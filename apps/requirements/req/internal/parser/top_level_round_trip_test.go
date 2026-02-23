package parser

import (
	"fmt"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"
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

func (suite *RoundTripSuite) TestRoundTrip() {

	input := test_helper.GetTestModel()

	// Validate the model before writing.
	err := input.Validate()
	assert.Nil(suite.T(), err, "input model should be valid")

	// Write to a temporary folder.
	tempDir := suite.T().TempDir()
	err = Write(input, tempDir)
	assert.Nil(suite.T(), err, "writing model should succeed")

	// Read from the temporary folder.
	output, err := Parse(tempDir)
	assert.Nil(suite.T(), err, "parsing model should succeed")

	// The parsed model's Key will be the tempDir path, not our original key.
	// Overwrite it for comparison since the parser uses the modelPath as the key.
	output.Key = input.Key

	// Compare scenarios individually for better error diagnostics.
	compareScenarios(suite.T(), input, output)

	// Compare the entire model tree.
	// If input and output model do not match, to explore this very deep tree of data, prune back to just the model and then iterate by layering in children one tier at a time  until the full tree exists again.
	assert.Equal(suite.T(), input, output, `Input and output model do not match, to explore this very deep tree of data, prune back to just the model and then iterate by layering in children one tier at a time until the full tree exists again`)
}
