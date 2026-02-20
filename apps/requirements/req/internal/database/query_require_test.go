package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestQueryRequireSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(QueryRequireSuite))
}

type QueryRequireSuite struct {
	suite.Suite
	db        *sql.DB
	model     req_model.Model
	domain    model_domain.Domain
	subdomain model_domain.Subdomain
	class     model_class.Class
	query     model_state.Query
	logic     model_logic.Logic
	logicB    model_logic.Logic
	queryKey  identity.Key
	logicKey  identity.Key
	logicKeyB identity.Key
}

func (suite *QueryRequireSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_key")))
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewClassKey(suite.subdomain.Key, "class_key")))
	suite.queryKey = helper.Must(identity.NewQueryKey(suite.class.Key, "query_key"))
	suite.query = t_AddQuery(suite.T(), suite.db, suite.model.Key, suite.class.Key, suite.queryKey)

	// Create logic rows (query require keys are children of query key).
	suite.logicKey = helper.Must(identity.NewQueryRequireKey(suite.queryKey, "req_a"))
	suite.logicKeyB = helper.Must(identity.NewQueryRequireKey(suite.queryKey, "req_b"))
	suite.logic = t_AddLogic(suite.T(), suite.db, suite.model.Key, suite.logicKey)
	suite.logicB = t_AddLogic(suite.T(), suite.db, suite.model.Key, suite.logicKeyB)
}

func (suite *QueryRequireSuite) TestLoad() {

	// Logic row exists from SetupTest, but no query_require join row yet.
	_, err := LoadQueryRequire(suite.db, suite.model.Key, suite.queryKey, suite.logicKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)

	// Insert the query_require join row.
	_, err = dbExec(suite.db, `
		INSERT INTO query_require
			(model_key, query_key, logic_key)
		VALUES
			($1, $2, $3)
	`, suite.model.Key, suite.queryKey.String(), suite.logicKey.String())
	assert.Nil(suite.T(), err)

	key, err := LoadQueryRequire(suite.db, suite.model.Key, suite.queryKey, suite.logicKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.logicKey, key)
}

func (suite *QueryRequireSuite) TestAdd() {

	err := AddQueryRequire(suite.db, suite.model.Key, suite.queryKey, suite.logicKey)
	assert.Nil(suite.T(), err)

	key, err := LoadQueryRequire(suite.db, suite.model.Key, suite.queryKey, suite.logicKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.logicKey, key)
}

func (suite *QueryRequireSuite) TestRemove() {

	err := AddQueryRequire(suite.db, suite.model.Key, suite.queryKey, suite.logicKey)
	assert.Nil(suite.T(), err)

	err = RemoveQueryRequire(suite.db, suite.model.Key, suite.queryKey, suite.logicKey)
	assert.Nil(suite.T(), err)

	// Query require should be gone.
	_, err = LoadQueryRequire(suite.db, suite.model.Key, suite.queryKey, suite.logicKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
}

func (suite *QueryRequireSuite) TestQuery() {

	err := AddQueryRequires(suite.db, suite.model.Key, map[identity.Key][]identity.Key{
		suite.queryKey: {suite.logicKeyB, suite.logicKey},
	})
	assert.Nil(suite.T(), err)

	requires, err := QueryQueryRequires(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[identity.Key][]identity.Key{
		suite.queryKey: {suite.logicKey, suite.logicKeyB},
	}, requires)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddQueryRequire(t *testing.T, dbOrTx DbOrTx, modelKey string, queryKey identity.Key, logicKey identity.Key) identity.Key {

	err := AddQueryRequire(dbOrTx, modelKey, queryKey, logicKey)
	assert.Nil(t, err)

	key, err := LoadQueryRequire(dbOrTx, modelKey, queryKey, logicKey)
	assert.Nil(t, err)

	return key
}
