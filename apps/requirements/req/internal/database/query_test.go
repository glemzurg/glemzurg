package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestQuerySuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(QuerySuite))
}

type QuerySuite struct {
	suite.Suite
	db        *sql.DB
	model     req_model.Model
	domain    model_domain.Domain
	subdomain model_domain.Subdomain
	class     model_class.Class
	queryKey  identity.Key
	queryKeyB identity.Key
}

func (suite *QuerySuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_key")))
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewClassKey(suite.subdomain.Key, "class_key")))

	// Create the query keys for reuse.
	suite.queryKey = helper.Must(identity.NewQueryKey(suite.class.Key, "key"))
	suite.queryKeyB = helper.Must(identity.NewQueryKey(suite.class.Key, "key_b"))
}

func (suite *QuerySuite) TestLoad() {

	// Nothing in database yet.
	classKey, query, err := LoadQuery(suite.db, suite.model.Key, suite.queryKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), classKey)
	assert.Empty(suite.T(), query)

	_, err = dbExec(suite.db, `
		INSERT INTO query
			(
				model_key,
				class_key,
				query_key,
				name,
				details
			)
		VALUES
			(
				'model_key',
				'domain/domain_key/subdomain/subdomain_key/class/class_key',
				'domain/domain_key/subdomain/subdomain_key/class/class_key/query/key',
				'Name',
				'Details'
			)
	`)
	assert.Nil(suite.T(), err)

	classKey, query, err = LoadQuery(suite.db, suite.model.Key, suite.queryKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.class.Key, classKey)
	assert.Equal(suite.T(), model_state.Query{
		Key:     suite.queryKey,
		Name:    "Name",
		Details: "Details",
	}, query)
}

func (suite *QuerySuite) TestAdd() {

	err := AddQuery(suite.db, suite.model.Key, suite.class.Key, model_state.Query{
		Key:     suite.queryKey,
		Name:    "Name",
		Details: "Details",
	})
	assert.Nil(suite.T(), err)

	classKey, query, err := LoadQuery(suite.db, suite.model.Key, suite.queryKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.class.Key, classKey)
	assert.Equal(suite.T(), model_state.Query{
		Key:     suite.queryKey,
		Name:    "Name",
		Details: "Details",
	}, query)
}

func (suite *QuerySuite) TestUpdate() {

	err := AddQuery(suite.db, suite.model.Key, suite.class.Key, model_state.Query{
		Key:     suite.queryKey,
		Name:    "Name",
		Details: "Details",
	})
	assert.Nil(suite.T(), err)

	err = UpdateQuery(suite.db, suite.model.Key, suite.class.Key, model_state.Query{
		Key:     suite.queryKey,
		Name:    "NameX",
		Details: "DetailsX",
	})
	assert.Nil(suite.T(), err)

	classKey, query, err := LoadQuery(suite.db, suite.model.Key, suite.queryKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.class.Key, classKey)
	assert.Equal(suite.T(), model_state.Query{
		Key:     suite.queryKey,
		Name:    "NameX",
		Details: "DetailsX",
	}, query)
}

func (suite *QuerySuite) TestRemove() {

	err := AddQuery(suite.db, suite.model.Key, suite.class.Key, model_state.Query{
		Key:     suite.queryKey,
		Name:    "Name",
		Details: "Details",
	})
	assert.Nil(suite.T(), err)

	err = RemoveQuery(suite.db, suite.model.Key, suite.class.Key, suite.queryKey)
	assert.Nil(suite.T(), err)

	classKey, query, err := LoadQuery(suite.db, suite.model.Key, suite.queryKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), classKey)
	assert.Empty(suite.T(), query)
}

func (suite *QuerySuite) TestQuery() {

	err := AddQueries(suite.db, suite.model.Key, map[identity.Key][]model_state.Query{
		suite.class.Key: {
			{
				Key:     suite.queryKeyB,
				Name:    "NameX",
				Details: "DetailsX",
			},
			{
				Key:     suite.queryKey,
				Name:    "Name",
				Details: "Details",
			},
		},
	})
	assert.Nil(suite.T(), err)

	queries, err := QueryQueries(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[identity.Key][]model_state.Query{
		suite.class.Key: {
			{
				Key:     suite.queryKey,
				Name:    "Name",
				Details: "Details",
			},
			{
				Key:     suite.queryKeyB,
				Name:    "NameX",
				Details: "DetailsX",
			},
		},
	}, queries)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddQuery(t *testing.T, dbOrTx DbOrTx, modelKey string, classKey identity.Key, queryKey identity.Key) (query model_state.Query) {

	err := AddQuery(dbOrTx, modelKey, classKey, model_state.Query{
		Key:     queryKey,
		Name:    queryKey.String(),
		Details: "Details",
	})
	assert.Nil(t, err)

	_, query, err = LoadQuery(dbOrTx, modelKey, queryKey)
	assert.Nil(t, err)

	return query
}

func (suite *QuerySuite) TestVerifyTestObjects() {

	query := t_AddQuery(suite.T(), suite.db, suite.model.Key, suite.class.Key, suite.queryKey)
	assert.Equal(suite.T(), model_state.Query{
		Key:     suite.queryKey,
		Name:    suite.queryKey.String(),
		Details: "Details",
	}, query)
}
