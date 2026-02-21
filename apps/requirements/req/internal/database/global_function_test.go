package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestGlobalFunctionSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(GlobalFunctionSuite))
}

type GlobalFunctionSuite struct {
	suite.Suite
	db     *sql.DB
	model  req_model.Model
	logic  model_logic.Logic
	logicB model_logic.Logic
	gfKey  identity.Key
	gfKeyB identity.Key
}

func (suite *GlobalFunctionSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.logic = t_AddLogic(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewGlobalFunctionKey("key")))
	suite.logicB = t_AddLogic(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewGlobalFunctionKey("key_b")))

	// Create the keys for reuse.
	suite.gfKey = suite.logic.Key
	suite.gfKeyB = suite.logicB.Key
}

func (suite *GlobalFunctionSuite) TestLoad() {

	// Logic row exists from SetupTest, but no global function row yet.
	_, err := LoadGlobalFunction(suite.db, suite.model.Key, suite.gfKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)

	// Insert the global function row with raw SQL.
	_, err = dbExec(suite.db, `
		INSERT INTO global_function
			(model_key, logic_key, name, comment, parameters)
		VALUES
			('model_key', 'gfunc/key', '_Max', 'Returns the maximum', '{x,y}')
	`)
	assert.Nil(suite.T(), err)

	gf, err := LoadGlobalFunction(suite.db, suite.model.Key, suite.gfKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_logic.GlobalFunction{
		Key:        suite.gfKey,
		Name:       "_Max",
		Comment:    "Returns the maximum",
		Parameters: []string{"x", "y"},
	}, gf)
}

func (suite *GlobalFunctionSuite) TestAdd() {

	err := AddGlobalFunction(suite.db, suite.model.Key, model_logic.GlobalFunction{
		Key:        suite.gfKey,
		Name:       "_Max",
		Comment:    "Returns the maximum",
		Parameters: []string{"x", "y"},
	})
	assert.Nil(suite.T(), err)

	gf, err := LoadGlobalFunction(suite.db, suite.model.Key, suite.gfKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_logic.GlobalFunction{
		Key:        suite.gfKey,
		Name:       "_Max",
		Comment:    "Returns the maximum",
		Parameters: []string{"x", "y"},
	}, gf)
}

func (suite *GlobalFunctionSuite) TestAddNulls() {

	err := AddGlobalFunction(suite.db, suite.model.Key, model_logic.GlobalFunction{
		Key:        suite.gfKey,
		Name:       "_Max",
		Comment:    "",
		Parameters: nil,
	})
	assert.Nil(suite.T(), err)

	gf, err := LoadGlobalFunction(suite.db, suite.model.Key, suite.gfKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_logic.GlobalFunction{
		Key:        suite.gfKey,
		Name:       "_Max",
		Comment:    "",
		Parameters: nil,
	}, gf)
}

func (suite *GlobalFunctionSuite) TestUpdate() {

	err := AddGlobalFunction(suite.db, suite.model.Key, model_logic.GlobalFunction{
		Key:        suite.gfKey,
		Name:       "_Max",
		Comment:    "Returns the maximum",
		Parameters: []string{"x", "y"},
	})
	assert.Nil(suite.T(), err)

	err = UpdateGlobalFunction(suite.db, suite.model.Key, model_logic.GlobalFunction{
		Key:        suite.gfKey,
		Name:       "_Min",
		Comment:    "Returns the minimum",
		Parameters: []string{"a", "b"},
	})
	assert.Nil(suite.T(), err)

	gf, err := LoadGlobalFunction(suite.db, suite.model.Key, suite.gfKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_logic.GlobalFunction{
		Key:        suite.gfKey,
		Name:       "_Min",
		Comment:    "Returns the minimum",
		Parameters: []string{"a", "b"},
	}, gf)
}

func (suite *GlobalFunctionSuite) TestUpdateNulls() {

	err := AddGlobalFunction(suite.db, suite.model.Key, model_logic.GlobalFunction{
		Key:        suite.gfKey,
		Name:       "_Max",
		Comment:    "Returns the maximum",
		Parameters: []string{"x", "y"},
	})
	assert.Nil(suite.T(), err)

	err = UpdateGlobalFunction(suite.db, suite.model.Key, model_logic.GlobalFunction{
		Key:        suite.gfKey,
		Name:       "_Min",
		Comment:    "",
		Parameters: nil,
	})
	assert.Nil(suite.T(), err)

	gf, err := LoadGlobalFunction(suite.db, suite.model.Key, suite.gfKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_logic.GlobalFunction{
		Key:        suite.gfKey,
		Name:       "_Min",
		Comment:    "",
		Parameters: nil,
	}, gf)
}

func (suite *GlobalFunctionSuite) TestRemove() {

	err := AddGlobalFunction(suite.db, suite.model.Key, model_logic.GlobalFunction{
		Key:        suite.gfKey,
		Name:       "_Max",
		Comment:    "Returns the maximum",
		Parameters: []string{"x", "y"},
	})
	assert.Nil(suite.T(), err)

	err = RemoveGlobalFunction(suite.db, suite.model.Key, suite.gfKey)
	assert.Nil(suite.T(), err)

	// Global function should be gone.
	_, err = LoadGlobalFunction(suite.db, suite.model.Key, suite.gfKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
}

func (suite *GlobalFunctionSuite) TestQuery() {

	err := AddGlobalFunctions(suite.db, suite.model.Key, []model_logic.GlobalFunction{
		{
			Key:        suite.gfKeyB,
			Name:       "_Min",
			Comment:    "Returns the minimum",
			Parameters: []string{"a", "b"},
		},
		{
			Key:        suite.gfKey,
			Name:       "_Max",
			Comment:    "Returns the maximum",
			Parameters: []string{"x", "y"},
		},
	})
	assert.Nil(suite.T(), err)

	gfs, err := QueryGlobalFunctions(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), []model_logic.GlobalFunction{
		{
			Key:        suite.gfKey,
			Name:       "_Max",
			Comment:    "Returns the maximum",
			Parameters: []string{"x", "y"},
		},
		{
			Key:        suite.gfKeyB,
			Name:       "_Min",
			Comment:    "Returns the minimum",
			Parameters: []string{"a", "b"},
		},
	}, gfs)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddGlobalFunction(t *testing.T, dbOrTx DbOrTx, modelKey string, logicKey identity.Key, name string, comment string, parameters []string) model_logic.GlobalFunction {

	err := AddGlobalFunction(dbOrTx, modelKey, model_logic.GlobalFunction{
		Key:        logicKey,
		Name:       name,
		Comment:    comment,
		Parameters: parameters,
	})
	assert.Nil(t, err)

	gf, err := LoadGlobalFunction(dbOrTx, modelKey, logicKey)
	assert.Nil(t, err)

	return gf
}
