package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestQueryGuaranteeSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(QueryGuaranteeSuite))
}

type QueryGuaranteeSuite struct {
	suite.Suite
	db        *sql.DB
	model     core.Model
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

func (suite *QueryGuaranteeSuite) SetupTest() {
	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_key")))
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewClassKey(suite.subdomain.Key, "class_key")))
	suite.queryKey = helper.Must(identity.NewQueryKey(suite.class.Key, "query_key"))
	suite.query = t_AddQuery(suite.T(), suite.db, suite.model.Key, suite.class.Key, suite.queryKey)

	// Create logic rows (query guarantee keys are children of query key).
	suite.logicKey = helper.Must(identity.NewQueryGuaranteeKey(suite.queryKey, "guar_a"))
	suite.logicKeyB = helper.Must(identity.NewQueryGuaranteeKey(suite.queryKey, "guar_b"))
	suite.logic = t_AddLogic(suite.T(), suite.db, suite.model.Key, suite.logicKey)
	suite.logicB = t_AddLogic(suite.T(), suite.db, suite.model.Key, suite.logicKeyB)
}

func (suite *QueryGuaranteeSuite) TestLoad() {
	// Logic row exists from SetupTest, but no query_guarantee join row yet.
	_, err := LoadQueryGuarantee(suite.db, suite.model.Key, suite.queryKey, suite.logicKey)
	suite.ErrorIs(err, ErrNotFound)

	// Insert the query_guarantee join row.
	err = dbExec(suite.db, `
		INSERT INTO query_guarantee
			(model_key, query_key, logic_key)
		VALUES
			(
				'model_key',
				'domain/domain_key/subdomain/subdomain_key/class/class_key/query/query_key',
				'domain/domain_key/subdomain/subdomain_key/class/class_key/query/query_key/qguarantee/guar_a'
			)
	`)
	assert.Nil(suite.T(), err)

	key, err := LoadQueryGuarantee(suite.db, suite.model.Key, suite.queryKey, suite.logicKey)
	assert.Nil(suite.T(), err)
	suite.Equal(suite.logicKey, key)
}

func (suite *QueryGuaranteeSuite) TestAdd() {
	err := AddQueryGuarantee(suite.db, suite.model.Key, suite.queryKey, suite.logicKey)
	assert.Nil(suite.T(), err)

	key, err := LoadQueryGuarantee(suite.db, suite.model.Key, suite.queryKey, suite.logicKey)
	assert.Nil(suite.T(), err)
	suite.Equal(suite.logicKey, key)
}

func (suite *QueryGuaranteeSuite) TestRemove() {
	err := AddQueryGuarantee(suite.db, suite.model.Key, suite.queryKey, suite.logicKey)
	assert.Nil(suite.T(), err)

	err = RemoveQueryGuarantee(suite.db, suite.model.Key, suite.queryKey, suite.logicKey)
	assert.Nil(suite.T(), err)

	// Query guarantee should be gone.
	_, err = LoadQueryGuarantee(suite.db, suite.model.Key, suite.queryKey, suite.logicKey)
	suite.ErrorIs(err, ErrNotFound)
}

func (suite *QueryGuaranteeSuite) TestQuery() {
	err := AddQueryGuarantees(suite.db, suite.model.Key, map[identity.Key][]identity.Key{
		suite.queryKey: {suite.logicKeyB, suite.logicKey},
	})
	assert.Nil(suite.T(), err)

	guarantees, err := QueryQueryGuarantees(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	suite.Equal(map[identity.Key][]identity.Key{
		suite.queryKey: {suite.logicKey, suite.logicKeyB},
	}, guarantees)
}
