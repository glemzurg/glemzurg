package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/stretchr/testify/require"
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
	suite.Require().ErrorIs(err, ErrNotFound)
	suite.Empty(model)

	err = dbExec(suite.db, `
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
	suite.Require().NoError(err)

	model, err = LoadModel(suite.db, "Key") // Test case-insensitive.
	suite.Require().NoError(err)
	suite.Equal(core.Model{
		Key:     "key", // Test case-insensitive.
		Name:    "Name",
		Details: "Details",
	}, model)
}

func (suite *ModelSuite) TestAdd() {
	err := AddModel(suite.db, core.Model{
		Key:     "KeY", // Test case-insensitive.
		Name:    "Name",
		Details: "Details",
	})
	suite.Require().NoError(err)

	model, err := LoadModel(suite.db, "key")
	suite.Require().NoError(err)
	suite.Equal(core.Model{
		Key:     "key", // Test case-insensitive.
		Name:    "Name",
		Details: "Details",
	}, model)
}

func (suite *ModelSuite) TestUpdate() {
	err := AddModel(suite.db, core.Model{
		Key:     "KeY", // Test case-insensitive.
		Name:    "Name",
		Details: "Details",
	})
	suite.Require().NoError(err)

	err = UpdateModel(suite.db, core.Model{
		Key:     "kEy", // Test case-insensitive.
		Name:    "NameX",
		Details: "DetailsX",
	})
	suite.Require().NoError(err)

	model, err := LoadModel(suite.db, "key")
	suite.Require().NoError(err)
	suite.Equal(core.Model{
		Key:     "key", // Test case-insensitive.
		Name:    "NameX",
		Details: "DetailsX",
	}, model)
}

func (suite *ModelSuite) TestRemove() {
	err := AddModel(suite.db, core.Model{
		Key:     "KeY", // Test case-insensitive.
		Name:    "Name",
		Details: "Details",
	})
	suite.Require().NoError(err)

	err = RemoveModel(suite.db, "kEy") // Test case-insensitive.
	suite.Require().NoError(err)

	model, err := LoadModel(suite.db, "key")
	suite.Require().ErrorIs(err, ErrNotFound)
	suite.Empty(model)
}

func (suite *ModelSuite) TestQuery() {
	err := AddModel(suite.db, core.Model{
		Key:     "keyx",
		Name:    "NameX",
		Details: "DetailsX",
	})
	suite.Require().NoError(err)

	err = AddModel(suite.db, core.Model{
		Key:     "key",
		Name:    "Name",
		Details: "Details",
	})
	suite.Require().NoError(err)

	models, err := QueryModels(suite.db)
	suite.Require().NoError(err)
	suite.Equal([]core.Model{
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

func t_AddModel(t *testing.T, dbOrTx DbOrTx) (model core.Model) {
	err := AddModel(dbOrTx, core.Model{
		Key:     "model_key",
		Name:    "Name",
		Details: "Details",
	})
	require.NoError(t, err)

	model, err = LoadModel(dbOrTx, "model_key")
	require.NoError(t, err)

	return model
}
