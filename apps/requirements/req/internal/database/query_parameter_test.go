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

func TestQueryParameterSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(QueryParameterSuite))
}

type QueryParameterSuite struct {
	suite.Suite
	db        *sql.DB
	model     req_model.Model
	domain    model_domain.Domain
	subdomain model_domain.Subdomain
	class     model_class.Class
	query     model_state.Query
	queryKey  identity.Key
}

func (suite *QueryParameterSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_key")))
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewClassKey(suite.subdomain.Key, "class_key")))

	// Create the query key for reuse.
	suite.queryKey = helper.Must(identity.NewQueryKey(suite.class.Key, "query_key"))
	suite.query = t_AddQuery(suite.T(), suite.db, suite.model.Key, suite.class.Key, suite.queryKey)
}

func (suite *QueryParameterSuite) TestLoad() {

	// Nothing in database yet.
	param, err := LoadQueryParameter(suite.db, suite.model.Key, suite.queryKey, "amount")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), param)

	_, err = dbExec(suite.db, `
		INSERT INTO query_parameter
			(
				model_key,
				query_key,
				parameter_key,
				name,
				sort_order,
				data_type_rules
			)
		VALUES
			(
				'model_key',
				'domain/domain_key/subdomain/subdomain_key/class/class_key/query/query_key',
				'amount',
				'Amount',
				1,
				'Nat'
			)
	`)
	assert.Nil(suite.T(), err)

	param, err = LoadQueryParameter(suite.db, suite.model.Key, suite.queryKey, "amount")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_state.Parameter{
		Name:          "Amount",
		SortOrder:     1,
		DataTypeRules: "Nat",
	}, param)
}

func (suite *QueryParameterSuite) TestAdd() {

	err := AddQueryParameter(suite.db, suite.model.Key, suite.queryKey, model_state.Parameter{
		Name:          "Amount",
		SortOrder:     1,
		DataTypeRules: "Nat",
	})
	assert.Nil(suite.T(), err)

	param, err := LoadQueryParameter(suite.db, suite.model.Key, suite.queryKey, "amount")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_state.Parameter{
		Name:          "Amount",
		SortOrder:     1,
		DataTypeRules: "Nat",
	}, param)
}

func (suite *QueryParameterSuite) TestUpdate() {

	err := AddQueryParameter(suite.db, suite.model.Key, suite.queryKey, model_state.Parameter{
		Name:          "Amount",
		SortOrder:     1,
		DataTypeRules: "Nat",
	})
	assert.Nil(suite.T(), err)

	err = UpdateQueryParameter(suite.db, suite.model.Key, suite.queryKey, model_state.Parameter{
		Name:          "Amount",
		SortOrder:     2,
		DataTypeRules: "Int",
	})
	assert.Nil(suite.T(), err)

	param, err := LoadQueryParameter(suite.db, suite.model.Key, suite.queryKey, "amount")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_state.Parameter{
		Name:          "Amount",
		SortOrder:     2,
		DataTypeRules: "Int",
	}, param)
}

func (suite *QueryParameterSuite) TestRemove() {

	err := AddQueryParameter(suite.db, suite.model.Key, suite.queryKey, model_state.Parameter{
		Name:          "Amount",
		SortOrder:     1,
		DataTypeRules: "Nat",
	})
	assert.Nil(suite.T(), err)

	err = RemoveQueryParameter(suite.db, suite.model.Key, suite.queryKey, "amount")
	assert.Nil(suite.T(), err)

	param, err := LoadQueryParameter(suite.db, suite.model.Key, suite.queryKey, "amount")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), param)
}

func (suite *QueryParameterSuite) TestQuery() {

	err := AddQueryParameters(suite.db, suite.model.Key, map[identity.Key][]model_state.Parameter{
		suite.queryKey: {
			{
				Name:          "Bravo",
				SortOrder:     1,
				DataTypeRules: "Int",
			},
			{
				Name:          "Alpha",
				SortOrder:     0,
				DataTypeRules: "Nat",
			},
		},
	})
	assert.Nil(suite.T(), err)

	params, err := QueryQueryParameters(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[identity.Key][]model_state.Parameter{
		suite.queryKey: {
			{
				Name:          "Alpha",
				SortOrder:     0,
				DataTypeRules: "Nat",
			},
			{
				Name:          "Bravo",
				SortOrder:     1,
				DataTypeRules: "Int",
			},
		},
	}, params)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddQueryParameter(t *testing.T, dbOrTx DbOrTx, modelKey string, queryKey identity.Key, name string, sortOrder int) (param model_state.Parameter) {

	paramKey, err := preenKey(name)
	assert.Nil(t, err)

	err = AddQueryParameter(dbOrTx, modelKey, queryKey, model_state.Parameter{
		Name:          name,
		SortOrder:     sortOrder,
		DataTypeRules: "Nat",
	})
	assert.Nil(t, err)

	param, err = LoadQueryParameter(dbOrTx, modelKey, queryKey, paramKey)
	assert.Nil(t, err)

	return param
}

func (suite *QueryParameterSuite) TestVerifyTestObjects() {

	param := t_AddQueryParameter(suite.T(), suite.db, suite.model.Key, suite.queryKey, "Amount", 0)
	assert.Equal(suite.T(), model_state.Parameter{
		Name:          "Amount",
		SortOrder:     0,
		DataTypeRules: "Nat",
	}, param)
}
