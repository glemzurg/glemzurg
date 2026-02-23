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
	// This works because identity.Key no longer contains pointer fields.
	assert.Equal(suite.T(), input, output)
}
