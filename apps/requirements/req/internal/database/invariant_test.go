package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

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
	model         core.Model
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
	suite.logic = t_AddLogic(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewInvariantKey("0")))
	suite.logicB = t_AddLogic(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewInvariantKey("1")))

	// Create the keys for reuse.
	suite.invariantKey = suite.logic.Key
	suite.invariantKeyB = suite.logicB.Key
}

func (suite *InvariantSuite) TestLoad() {
	// Logic row exists from SetupTest, but no invariant join row yet.
	_, err := LoadInvariant(suite.db, suite.model.Key, suite.invariantKey)
	suite.ErrorIs(err, ErrNotFound)

	// Insert the invariant join row.
	err = dbExec(suite.db, `
		INSERT INTO invariant
			(model_key, logic_key)
		VALUES
			('model_key', 'invariant/0')
	`)
	suite.Require().NoError(err)

	key, err := LoadInvariant(suite.db, suite.model.Key, suite.invariantKey)
	suite.Require().NoError(err)
	suite.Equal(suite.invariantKey, key)
}

func (suite *InvariantSuite) TestAdd() {
	err := AddInvariant(suite.db, suite.model.Key, suite.invariantKey)
	suite.Require().NoError(err)

	key, err := LoadInvariant(suite.db, suite.model.Key, suite.invariantKey)
	suite.Require().NoError(err)
	suite.Equal(suite.invariantKey, key)
}

func (suite *InvariantSuite) TestRemove() {
	err := AddInvariant(suite.db, suite.model.Key, suite.invariantKey)
	suite.Require().NoError(err)

	err = RemoveInvariant(suite.db, suite.model.Key, suite.invariantKey)
	suite.Require().NoError(err)

	// Invariant should be gone.
	_, err = LoadInvariant(suite.db, suite.model.Key, suite.invariantKey)
	suite.ErrorIs(err, ErrNotFound)
}

func (suite *InvariantSuite) TestQuery() {
	err := AddInvariants(suite.db, suite.model.Key, []identity.Key{
		suite.invariantKeyB,
		suite.invariantKey,
	})
	suite.Require().NoError(err)

	keys, err := QueryInvariants(suite.db, suite.model.Key)
	suite.Require().NoError(err)
	suite.Equal([]identity.Key{
		suite.invariantKey,
		suite.invariantKeyB,
	}, keys)
}
