package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestModelSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(ModelSuite))
}

type ModelSuite struct {
	suite.Suite
	db *sql.DB
}

func (suite *ModelSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())
}

func (suite *ModelSuite) TestLoad() {

	// Nothing in database yet.
	model, err := LoadModel(suite.db, "Key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), model)

	_, err = dbExec(suite.db, `
		INSERT INTO model
			(
				model_key,
				name,
				details
			)
		VALUES
			(
				'key',
				'Name',
				'Details'
			)
	`)
	assert.Nil(suite.T(), err)

	model, err = LoadModel(suite.db, "Key") // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), requirements.Model{
		Key:     "key", // Test case-insensitive.
		Name:    "Name",
		Details: "Details",
	}, model)
}

func (suite *ModelSuite) TestAdd() {

	err := AddModel(suite.db, requirements.Model{
		Key:     "KeY", // Test case-insensitive.
		Name:    "Name",
		Details: "Details",
	})
	assert.Nil(suite.T(), err)

	model, err := LoadModel(suite.db, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), requirements.Model{
		Key:     "key", // Test case-insensitive.
		Name:    "Name",
		Details: "Details",
	}, model)
}

func (suite *ModelSuite) TestUpdate() {

	err := AddModel(suite.db, requirements.Model{
		Key:     "KeY", // Test case-insensitive.
		Name:    "Name",
		Details: "Details",
	})
	assert.Nil(suite.T(), err)

	err = UpdateModel(suite.db, requirements.Model{
		Key:     "kEy", // Test case-insensitive.
		Name:    "NameX",
		Details: "DetailsX",
	})
	assert.Nil(suite.T(), err)

	model, err := LoadModel(suite.db, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), requirements.Model{
		Key:     "key", // Test case-insensitive.
		Name:    "NameX",
		Details: "DetailsX",
	}, model)
}

func (suite *ModelSuite) TestRemove() {

	err := AddModel(suite.db, requirements.Model{
		Key:     "KeY", // Test case-insensitive.
		Name:    "Name",
		Details: "Details",
	})
	assert.Nil(suite.T(), err)

	err = RemoveModel(suite.db, "kEy") // Test case-insensitive.
	assert.Nil(suite.T(), err)

	model, err := LoadModel(suite.db, "key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), model)
}

func (suite *ModelSuite) TestQuery() {

	err := AddModel(suite.db, requirements.Model{
		Key:     "keyx",
		Name:    "NameX",
		Details: "DetailsX",
	})
	assert.Nil(suite.T(), err)

	err = AddModel(suite.db, requirements.Model{
		Key:     "key",
		Name:    "Name",
		Details: "Details",
	})
	assert.Nil(suite.T(), err)

	models, err := QueryModels(suite.db)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), []requirements.Model{
		{
			Key:     "key",
			Name:    "Name",
			Details: "Details",
		},
		{
			Key:     "keyx",
			Name:    "NameX",
			Details: "DetailsX",
		},
	}, models)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddModel(t *testing.T, dbOrTx DbOrTx) (model requirements.Model) {
	err := AddModel(dbOrTx, requirements.Model{
		Key:     "model_key",
		Name:    "Name",
		Details: "Details",
	})
	assert.Nil(t, err)

	model, err = LoadModel(dbOrTx, "model_key")
	assert.Nil(t, err)

	return model
}
