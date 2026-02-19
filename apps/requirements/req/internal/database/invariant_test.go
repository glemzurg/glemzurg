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
	logicB        model_logic.Logic
	invariantKey  identity.Key
	invariantKeyB identity.Key
}

func (suite *InvariantSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.logic = t_AddLogic(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewInvariantKey("key")))
	suite.logicB = t_AddLogic(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewInvariantKey("key_b")))

	// Create the invariant keys for reuse.
	suite.invariantKey = suite.logic.Key
	suite.invariantKeyB = suite.logicB.Key
}

func (suite *InvariantSuite) TestLoad() {

	// Logic row exists from SetupTest, but no invariant join row yet.
	_, err := LoadInvariant(suite.db, suite.model.Key, suite.invariantKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)

	// Insert the invariant join row.
	err = AddInvariant(suite.db, suite.model.Key, suite.invariantKey)
	assert.Nil(suite.T(), err)

	key, err := LoadInvariant(suite.db, suite.model.Key, suite.invariantKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.invariantKey, key)
}

func (suite *InvariantSuite) TestAdd() {

	err := AddInvariant(suite.db, suite.model.Key, suite.invariantKey)
	assert.Nil(suite.T(), err)

	key, err := LoadInvariant(suite.db, suite.model.Key, suite.invariantKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.invariantKey, key)
}

func (suite *InvariantSuite) TestRemove() {

	err := AddInvariant(suite.db, suite.model.Key, suite.invariantKey)
	assert.Nil(suite.T(), err)

	err = RemoveInvariant(suite.db, suite.model.Key, suite.invariantKey)
	assert.Nil(suite.T(), err)

	// Invariant should be gone.
	_, err = LoadInvariant(suite.db, suite.model.Key, suite.invariantKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
}

func (suite *InvariantSuite) TestQuery() {

	err := AddInvariants(suite.db, suite.model.Key, []identity.Key{
		suite.invariantKeyB,
		suite.invariantKey,
	})
	assert.Nil(suite.T(), err)

	keys, err := QueryInvariants(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), []identity.Key{
		suite.invariantKey,
		suite.invariantKeyB,
	}, keys)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddInvariant(t *testing.T, dbOrTx DbOrTx, modelKey string, logicKey identity.Key) identity.Key {

	err := AddInvariant(dbOrTx, modelKey, logicKey)
	assert.Nil(t, err)

	key, err := LoadInvariant(dbOrTx, modelKey, logicKey)
	assert.Nil(t, err)

	return key
}
