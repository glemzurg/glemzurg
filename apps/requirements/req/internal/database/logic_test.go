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

func TestLogicSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(LogicSuite))
}

type LogicSuite struct {
	suite.Suite
	db        *sql.DB
	model     req_model.Model
	logicKey  identity.Key
	logicKeyB identity.Key
}

func (suite *LogicSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)

	// Create the keys for reuse.
	suite.logicKey = helper.Must(identity.NewInvariantKey("0"))
	suite.logicKeyB = helper.Must(identity.NewInvariantKey("1"))
}

func (suite *LogicSuite) TestLoad() {

	// Nothing in database yet.
	logic, err := LoadLogic(suite.db, suite.model.Key, suite.logicKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), logic)

	_, err = dbExec(suite.db, `
		INSERT INTO logic
			(
				model_key,
				logic_key,
				description,
				notation,
				specification,
				sort_order
			)
		VALUES
			(
				'model_key',
				'invariant/0',
				'Description',
				'tla_plus',
				'Specification',
				0
			)
	`)
	assert.Nil(suite.T(), err)

	logic, err = LoadLogic(suite.db, suite.model.Key, suite.logicKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_logic.Logic{
		Key:           suite.logicKey,
		Description:   "Description",
		Notation:      "tla_plus",
		Specification: "Specification",
	}, logic)
}

func (suite *LogicSuite) TestAdd() {

	err := AddLogic(suite.db, suite.model.Key, model_logic.Logic{
		Key:           suite.logicKey,
		Description:   "Description",
		Notation:      "tla_plus",
		Specification: "Specification",
	})
	assert.Nil(suite.T(), err)

	logic, err := LoadLogic(suite.db, suite.model.Key, suite.logicKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_logic.Logic{
		Key:           suite.logicKey,
		Description:   "Description",
		Notation:      "tla_plus",
		Specification: "Specification",
	}, logic)
}

func (suite *LogicSuite) TestAddNulls() {

	err := AddLogic(suite.db, suite.model.Key, model_logic.Logic{
		Key:           suite.logicKey,
		Description:   "Description",
		Notation:      "tla_plus",
		Specification: "",
	})
	assert.Nil(suite.T(), err)

	logic, err := LoadLogic(suite.db, suite.model.Key, suite.logicKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_logic.Logic{
		Key:           suite.logicKey,
		Description:   "Description",
		Notation:      "tla_plus",
		Specification: "",
	}, logic)
}

func (suite *LogicSuite) TestUpdate() {

	err := AddLogic(suite.db, suite.model.Key, model_logic.Logic{
		Key:           suite.logicKey,
		Description:   "Description",
		Notation:      "tla_plus",
		Specification: "Specification",
	})
	assert.Nil(suite.T(), err)

	err = UpdateLogic(suite.db, suite.model.Key, model_logic.Logic{
		Key:           suite.logicKey,
		Description:   "DescriptionX",
		Notation:      "tla_plus",
		Specification: "SpecificationX",
	}, 0)
	assert.Nil(suite.T(), err)

	logic, err := LoadLogic(suite.db, suite.model.Key, suite.logicKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_logic.Logic{
		Key:           suite.logicKey,
		Description:   "DescriptionX",
		Notation:      "tla_plus",
		Specification: "SpecificationX",
	}, logic)
}

func (suite *LogicSuite) TestUpdateNulls() {

	err := AddLogic(suite.db, suite.model.Key, model_logic.Logic{
		Key:           suite.logicKey,
		Description:   "Description",
		Notation:      "tla_plus",
		Specification: "Specification",
	})
	assert.Nil(suite.T(), err)

	err = UpdateLogic(suite.db, suite.model.Key, model_logic.Logic{
		Key:           suite.logicKey,
		Description:   "DescriptionX",
		Notation:      "tla_plus",
		Specification: "",
	}, 0)
	assert.Nil(suite.T(), err)

	logic, err := LoadLogic(suite.db, suite.model.Key, suite.logicKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_logic.Logic{
		Key:           suite.logicKey,
		Description:   "DescriptionX",
		Notation:      "tla_plus",
		Specification: "",
	}, logic)
}

func (suite *LogicSuite) TestRemove() {

	err := AddLogic(suite.db, suite.model.Key, model_logic.Logic{
		Key:           suite.logicKey,
		Description:   "Description",
		Notation:      "tla_plus",
		Specification: "Specification",
	})
	assert.Nil(suite.T(), err)

	err = RemoveLogic(suite.db, suite.model.Key, suite.logicKey)
	assert.Nil(suite.T(), err)

	logic, err := LoadLogic(suite.db, suite.model.Key, suite.logicKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), logic)
}

func (suite *LogicSuite) TestQuery() {

	err := AddLogics(suite.db, suite.model.Key, []model_logic.Logic{
		{
			Key:           suite.logicKeyB,
			Description:   "DescriptionX",
			Notation:      "tla_plus",
			Specification: "SpecificationX",
		},
		{
			Key:           suite.logicKey,
			Description:   "Description",
			Notation:      "tla_plus",
			Specification: "Specification",
		},
	}, map[string]int{
		suite.logicKeyB.String(): 0,
		suite.logicKey.String():  1,
	})
	assert.Nil(suite.T(), err)

	logics, err := QueryLogics(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), []model_logic.Logic{
		{
			Key:           suite.logicKey,
			Description:   "Description",
			Notation:      "tla_plus",
			Specification: "Specification",
		},
		{
			Key:           suite.logicKeyB,
			Description:   "DescriptionX",
			Notation:      "tla_plus",
			Specification: "SpecificationX",
		},
	}, logics)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddLogic(t *testing.T, dbOrTx DbOrTx, modelKey string, logicKey identity.Key) (logic model_logic.Logic) {

	err := AddLogic(dbOrTx, modelKey, model_logic.Logic{
		Key:           logicKey,
		Description:   logicKey.String(),
		Notation:      "tla_plus",
		Specification: "Specification",
	})
	assert.Nil(t, err)

	logic, err = LoadLogic(dbOrTx, modelKey, logicKey)
	assert.Nil(t, err)

	return logic
}
