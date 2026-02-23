package database

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/test_helper"
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

	// Prune use case scenarios (steps reference class events via FK).
	for domainKey, domain := range input.Domains {
		for subdomainKey, subdomain := range domain.Subdomains {
			for useCaseKey, useCase := range subdomain.UseCases {
				useCase.Scenarios = nil
				subdomain.UseCases[useCaseKey] = useCase
			}
			domain.Subdomains[subdomainKey] = subdomain
		}
		input.Domains[domainKey] = domain
	}

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

	// Compare the entire model tree.
	// If input and output model do not match, to explore this very deep tree of data, prune back to just the model and then iterate by layering in children one tier at a time  until the full tree exists again.
	assert.Equal(suite.T(), input, output, `Input and output model do not match, to explore this very deep tree of data, prune back to just the model and then iterate by layering in children one tier at a time until the full tree exists again`)
}
