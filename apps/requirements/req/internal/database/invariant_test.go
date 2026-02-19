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

func TestInvariantSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(InvariantSuite))
}

type InvariantSuite struct {
	suite.Suite
	db            *sql.DB
	model         req_model.Model
	logic         model_logic.Logic
	invariantKey  identity.Key
	invariantKeyB identity.Key
}

func (suite *InvariantSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.logic = t_AddLogic(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewInvariantKey("key")))
	suite.logicB = t_AddLogic(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewInvariantKey("key_b"))

	// Create the invariant keys for reuse.
	suite.invariantKey = suite.logic.Key
	suite.invariantKeyB = suite.logicB.Key
}

func (suite *InvariantSuite) TestLoad() {

	// Logic row exists from SetupTest, but no invariant join row yet.
	logic, err := LoadInvariant(suite.db, suite.model.Key, suite.invariantKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), logic)

	// Insert the invariant join row.
	_, err = dbExec(suite.db, `
		INSERT INTO invariant
			(model_key, logic_key)
		VALUES
			($1, $2)
	`, suite.model.Key, suite.invariantKey.String())
	assert.Nil(suite.T(), err)

	logic, err = LoadInvariant(suite.db, suite.model.Key, suite.invariantKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.logic, logic)
}

func (suite *InvariantSuite) TestAdd() {

	err := AddInvariant(suite.db, suite.model.Key, model_logic.Logic{
		Key:           suite.invariantKey,
		Description:   "Description",
		Notation:      "tla_plus",
		Specification: "Specification",
	})
	assert.Nil(suite.T(), err)

	logic, err := LoadInvariant(suite.db, suite.model.Key, suite.invariantKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_logic.Logic{
		Key:           suite.invariantKey,
		Description:   "Description",
		Notation:      "tla_plus",
		Specification: "Specification",
	}, logic)
}

func (suite *InvariantSuite) TestAddNulls() {

	err := AddInvariant(suite.db, suite.model.Key, model_logic.Logic{
		Key:           suite.invariantKeyB,
		Description:   "Description",
		Notation:      "tla_plus",
		Specification: "",
	})
	assert.Nil(suite.T(), err)

	logic, err := LoadInvariant(suite.db, suite.model.Key, suite.invariantKeyB)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_logic.Logic{
		Key:           suite.invariantKeyB,
		Description:   "Description",
		Notation:      "tla_plus",
		Specification: "",
	}, logic)
}

func (suite *InvariantSuite) TestRemove() {

	err := AddInvariant(suite.db, suite.model.Key, model_logic.Logic{
		Key:           suite.invariantKeyB,
		Description:   "Description",
		Notation:      "tla_plus",
		Specification: "Specification",
	})
	assert.Nil(suite.T(), err)

	err = RemoveInvariant(suite.db, suite.model.Key, suite.invariantKeyB)
	assert.Nil(suite.T(), err)

	// Invariant should be gone.
	logic, err := LoadInvariant(suite.db, suite.model.Key, suite.invariantKeyB)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), logic)

	// Logic row should also be gone.
	logic, err = LoadLogic(suite.db, suite.model.Key, suite.invariantKeyB)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), logic)
}

func (suite *InvariantSuite) TestQuery() {

	invariantKeyC := helper.Must(identity.NewInvariantKey("key_c"))

	err := AddInvariants(suite.db, suite.model.Key, []model_logic.Logic{
		{
			Key:           invariantKeyC,
			Description:   "DescriptionX",
			Notation:      "tla_plus",
			Specification: "SpecificationX",
		},
		{
			Key:           suite.invariantKeyB,
			Description:   "Description",
			Notation:      "tla_plus",
			Specification: "Specification",
		},
	})
	assert.Nil(suite.T(), err)

	logics, err := QueryInvariants(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), []model_logic.Logic{
		{
			Key:           suite.invariantKeyB,
			Description:   "Description",
			Notation:      "tla_plus",
			Specification: "Specification",
		},
		{
			Key:           invariantKeyC,
			Description:   "DescriptionX",
			Notation:      "tla_plus",
			Specification: "SpecificationX",
		},
	}, logics)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddInvariant(t *testing.T, dbOrTx DbOrTx, modelKey string, logicKey identity.Key) (logic model_logic.Logic) {

	err := AddInvariant(dbOrTx, modelKey, model_logic.Logic{
		Key:           logicKey,
		Description:   logicKey.String(),
		Notation:      "tla_plus",
		Specification: "Specification",
	})
	assert.Nil(t, err)

	logic, err = LoadInvariant(dbOrTx, modelKey, logicKey)
	assert.Nil(t, err)

	return logic
}
