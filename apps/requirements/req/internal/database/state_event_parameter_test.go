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

func TestEventParameterSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(EventParameterSuite))
}

type EventParameterSuite struct {
	suite.Suite
	db        *sql.DB
	model     req_model.Model
	domain    model_domain.Domain
	subdomain model_domain.Subdomain
	class     model_class.Class
	event     model_state.Event
	eventKey  identity.Key
}

func (suite *EventParameterSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_key")))
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewClassKey(suite.subdomain.Key, "class_key")))

	// Create the event key for reuse.
	suite.eventKey = helper.Must(identity.NewEventKey(suite.class.Key, "event_key"))
	suite.event = t_AddEvent(suite.T(), suite.db, suite.model.Key, suite.class.Key, suite.eventKey)
}

func (suite *EventParameterSuite) TestLoad() {

	// Nothing in database yet.
	param, err := LoadEventParameter(suite.db, suite.model.Key, suite.eventKey, "amount")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), param)

	_, err = dbExec(suite.db, `
		INSERT INTO event_parameter
			(
				model_key,
				event_key,
				parameter_key,
				name,
				sort_order,
				data_type_rules
			)
		VALUES
			(
				'model_key',
				$1,
				'amount',
				'Amount',
				1,
				'Nat'
			)
	`, suite.eventKey.String())
	assert.Nil(suite.T(), err)

	param, err = LoadEventParameter(suite.db, suite.model.Key, suite.eventKey, "amount")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_state.Parameter{
		Name:          "Amount",
		SortOrder:     1,
		DataTypeRules: "Nat",
	}, param)
}

func (suite *EventParameterSuite) TestAdd() {

	err := AddEventParameter(suite.db, suite.model.Key, suite.eventKey, model_state.Parameter{
		Name:          "Amount",
		SortOrder:     1,
		DataTypeRules: "Nat",
	})
	assert.Nil(suite.T(), err)

	param, err := LoadEventParameter(suite.db, suite.model.Key, suite.eventKey, "amount")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_state.Parameter{
		Name:          "Amount",
		SortOrder:     1,
		DataTypeRules: "Nat",
	}, param)
}

func (suite *EventParameterSuite) TestUpdate() {

	err := AddEventParameter(suite.db, suite.model.Key, suite.eventKey, model_state.Parameter{
		Name:          "Amount",
		SortOrder:     1,
		DataTypeRules: "Nat",
	})
	assert.Nil(suite.T(), err)

	err = UpdateEventParameter(suite.db, suite.model.Key, suite.eventKey, model_state.Parameter{
		Name:          "Amount",
		SortOrder:     2,
		DataTypeRules: "Int",
	})
	assert.Nil(suite.T(), err)

	param, err := LoadEventParameter(suite.db, suite.model.Key, suite.eventKey, "amount")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_state.Parameter{
		Name:          "Amount",
		SortOrder:     2,
		DataTypeRules: "Int",
	}, param)
}

func (suite *EventParameterSuite) TestRemove() {

	err := AddEventParameter(suite.db, suite.model.Key, suite.eventKey, model_state.Parameter{
		Name:          "Amount",
		SortOrder:     1,
		DataTypeRules: "Nat",
	})
	assert.Nil(suite.T(), err)

	err = RemoveEventParameter(suite.db, suite.model.Key, suite.eventKey, "amount")
	assert.Nil(suite.T(), err)

	param, err := LoadEventParameter(suite.db, suite.model.Key, suite.eventKey, "amount")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), param)
}

func (suite *EventParameterSuite) TestQuery() {

	err := AddEventParameters(suite.db, suite.model.Key, map[identity.Key][]model_state.Parameter{
		suite.eventKey: {
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

	params, err := QueryEventParameters(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[identity.Key][]model_state.Parameter{
		suite.eventKey: {
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

func t_AddEventParameter(t *testing.T, dbOrTx DbOrTx, modelKey string, eventKey identity.Key, name string, sortOrder int) (param model_state.Parameter) {

	paramKey, err := preenKey(name)
	assert.Nil(t, err)

	err = AddEventParameter(dbOrTx, modelKey, eventKey, model_state.Parameter{
		Name:          name,
		SortOrder:     sortOrder,
		DataTypeRules: "Nat",
	})
	assert.Nil(t, err)

	param, err = LoadEventParameter(dbOrTx, modelKey, eventKey, paramKey)
	assert.Nil(t, err)

	return param
}

func (suite *EventParameterSuite) TestVerifyTestObjects() {

	param := t_AddEventParameter(suite.T(), suite.db, suite.model.Key, suite.eventKey, "Amount", 0)
	assert.Equal(suite.T(), model_state.Parameter{
		Name:          "Amount",
		SortOrder:     0,
		DataTypeRules: "Nat",
	}, param)
}
