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

func TestActionParameterSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(ActionParameterSuite))
}

type ActionParameterSuite struct {
	suite.Suite
	db        *sql.DB
	model     req_model.Model
	domain    model_domain.Domain
	subdomain model_domain.Subdomain
	class     model_class.Class
	action    model_state.Action
	actionKey identity.Key
}

func (suite *ActionParameterSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_key")))
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewClassKey(suite.subdomain.Key, "class_key")))

	// Create the action key for reuse.
	suite.actionKey = helper.Must(identity.NewActionKey(suite.class.Key, "action_key"))
	suite.action = t_AddAction(suite.T(), suite.db, suite.model.Key, suite.class.Key, suite.actionKey)
}

func (suite *ActionParameterSuite) TestLoad() {

	// Nothing in database yet.
	param, err := LoadActionParameter(suite.db, suite.model.Key, suite.actionKey, "amount")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), param)

	_, err = dbExec(suite.db, `
		INSERT INTO action_parameter
			(
				model_key,
				action_key,
				parameter_key,
				name,
				sort_order,
				data_type_rules
			)
		VALUES
			(
				'model_key',
				'domain/domain_key/subdomain/subdomain_key/class/class_key/action/action_key',
				'amount',
				'Amount',
				1,
				'Nat'
			)
	`)
	assert.Nil(suite.T(), err)

	param, err = LoadActionParameter(suite.db, suite.model.Key, suite.actionKey, "amount")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_state.Parameter{
		Name:          "Amount",
		SortOrder:     1,
		DataTypeRules: "Nat",
	}, param)
}

func (suite *ActionParameterSuite) TestAdd() {

	err := AddActionParameter(suite.db, suite.model.Key, suite.actionKey, model_state.Parameter{
		Name:          "Amount",
		SortOrder:     1,
		DataTypeRules: "Nat",
	})
	assert.Nil(suite.T(), err)

	param, err := LoadActionParameter(suite.db, suite.model.Key, suite.actionKey, "amount")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_state.Parameter{
		Name:          "Amount",
		SortOrder:     1,
		DataTypeRules: "Nat",
	}, param)
}

func (suite *ActionParameterSuite) TestUpdate() {

	err := AddActionParameter(suite.db, suite.model.Key, suite.actionKey, model_state.Parameter{
		Name:          "Amount",
		SortOrder:     1,
		DataTypeRules: "Nat",
	})
	assert.Nil(suite.T(), err)

	err = UpdateActionParameter(suite.db, suite.model.Key, suite.actionKey, model_state.Parameter{
		Name:          "Amount",
		SortOrder:     2,
		DataTypeRules: "Int",
	})
	assert.Nil(suite.T(), err)

	param, err := LoadActionParameter(suite.db, suite.model.Key, suite.actionKey, "amount")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_state.Parameter{
		Name:          "Amount",
		SortOrder:     2,
		DataTypeRules: "Int",
	}, param)
}

func (suite *ActionParameterSuite) TestRemove() {

	err := AddActionParameter(suite.db, suite.model.Key, suite.actionKey, model_state.Parameter{
		Name:          "Amount",
		SortOrder:     1,
		DataTypeRules: "Nat",
	})
	assert.Nil(suite.T(), err)

	err = RemoveActionParameter(suite.db, suite.model.Key, suite.actionKey, "amount")
	assert.Nil(suite.T(), err)

	param, err := LoadActionParameter(suite.db, suite.model.Key, suite.actionKey, "amount")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), param)
}

func (suite *ActionParameterSuite) TestQuery() {

	err := AddActionParameters(suite.db, suite.model.Key, map[identity.Key][]model_state.Parameter{
		suite.actionKey: {
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

	params, err := QueryActionParameters(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[identity.Key][]model_state.Parameter{
		suite.actionKey: {
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

func t_AddActionParameter(t *testing.T, dbOrTx DbOrTx, modelKey string, actionKey identity.Key, name string, sortOrder int) (param model_state.Parameter) {

	paramKey, err := preenKey(name)
	assert.Nil(t, err)

	err = AddActionParameter(dbOrTx, modelKey, actionKey, model_state.Parameter{
		Name:          name,
		SortOrder:     sortOrder,
		DataTypeRules: "Nat",
	})
	assert.Nil(t, err)

	param, err = LoadActionParameter(dbOrTx, modelKey, actionKey, paramKey)
	assert.Nil(t, err)

	return param
}

func (suite *ActionParameterSuite) TestVerifyTestObjects() {

	param := t_AddActionParameter(suite.T(), suite.db, suite.model.Key, suite.actionKey, "Amount", 0)
	assert.Equal(suite.T(), model_state.Parameter{
		Name:          "Amount",
		SortOrder:     0,
		DataTypeRules: "Nat",
	}, param)
}
