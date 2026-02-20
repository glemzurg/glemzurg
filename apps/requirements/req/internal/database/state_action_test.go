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

func TestActionSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(ActionSuite))
}

type ActionSuite struct {
	suite.Suite
	db         *sql.DB
	model      req_model.Model
	domain     model_domain.Domain
	subdomain  model_domain.Subdomain
	class      model_class.Class
	actionKey  identity.Key
	actionKeyB identity.Key
}

func (suite *ActionSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_key")))
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewClassKey(suite.subdomain.Key, "class_key")))

	// Create the action keys for reuse.
	suite.actionKey = helper.Must(identity.NewActionKey(suite.class.Key, "key"))
	suite.actionKeyB = helper.Must(identity.NewActionKey(suite.class.Key, "key_b"))
}

func (suite *ActionSuite) TestLoad() {

	// Nothing in database yet.
	classKey, action, err := LoadAction(suite.db, suite.model.Key, suite.actionKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), classKey)
	assert.Empty(suite.T(), action)

	_, err = dbExec(suite.db, `
		INSERT INTO action
			(
				model_key,
				class_key,
				action_key,
				name,
				details
			)
		VALUES
			(
				'model_key',
				$1,
				$2,
				'Name',
				'Details'
			)
	`, suite.class.Key.String(), suite.actionKey.String())
	assert.Nil(suite.T(), err)

	classKey, action, err = LoadAction(suite.db, suite.model.Key, suite.actionKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.class.Key, classKey)
	assert.Equal(suite.T(), model_state.Action{
		Key:     suite.actionKey,
		Name:    "Name",
		Details: "Details",
	}, action)
}

func (suite *ActionSuite) TestAdd() {

	err := AddAction(suite.db, suite.model.Key, suite.class.Key, model_state.Action{
		Key:     suite.actionKey,
		Name:    "Name",
		Details: "Details",
	})
	assert.Nil(suite.T(), err)

	classKey, action, err := LoadAction(suite.db, suite.model.Key, suite.actionKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.class.Key, classKey)
	assert.Equal(suite.T(), model_state.Action{
		Key:     suite.actionKey,
		Name:    "Name",
		Details: "Details",
	}, action)
}

func (suite *ActionSuite) TestUpdate() {

	err := AddAction(suite.db, suite.model.Key, suite.class.Key, model_state.Action{
		Key:     suite.actionKey,
		Name:    "Name",
		Details: "Details",
	})
	assert.Nil(suite.T(), err)

	err = UpdateAction(suite.db, suite.model.Key, suite.class.Key, model_state.Action{
		Key:     suite.actionKey,
		Name:    "NameX",
		Details: "DetailsX",
	})
	assert.Nil(suite.T(), err)

	classKey, action, err := LoadAction(suite.db, suite.model.Key, suite.actionKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.class.Key, classKey)
	assert.Equal(suite.T(), model_state.Action{
		Key:     suite.actionKey,
		Name:    "NameX",
		Details: "DetailsX",
	}, action)
}

func (suite *ActionSuite) TestRemove() {

	err := AddAction(suite.db, suite.model.Key, suite.class.Key, model_state.Action{
		Key:     suite.actionKey,
		Name:    "Name",
		Details: "Details",
	})
	assert.Nil(suite.T(), err)

	err = RemoveAction(suite.db, suite.model.Key, suite.class.Key, suite.actionKey)
	assert.Nil(suite.T(), err)

	classKey, action, err := LoadAction(suite.db, suite.model.Key, suite.actionKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), classKey)
	assert.Empty(suite.T(), action)
}

func (suite *ActionSuite) TestQuery() {

	err := AddActions(suite.db, suite.model.Key, map[identity.Key][]model_state.Action{
		suite.class.Key: {
			{
				Key:     suite.actionKeyB,
				Name:    "NameX",
				Details: "DetailsX",
			},
			{
				Key:     suite.actionKey,
				Name:    "Name",
				Details: "Details",
			},
		},
	})
	assert.Nil(suite.T(), err)

	actions, err := QueryActions(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[identity.Key][]model_state.Action{
		suite.class.Key: {
			{
				Key:     suite.actionKey,
				Name:    "Name",
				Details: "Details",
			},
			{
				Key:     suite.actionKeyB,
				Name:    "NameX",
				Details: "DetailsX",
			},
		},
	}, actions)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddAction(t *testing.T, dbOrTx DbOrTx, modelKey string, classKey identity.Key, actionKey identity.Key) (action model_state.Action) {

	err := AddAction(dbOrTx, modelKey, classKey, model_state.Action{
		Key:     actionKey,
		Name:    actionKey.String(),
		Details: "Details",
	})
	assert.Nil(t, err)

	_, action, err = LoadAction(dbOrTx, modelKey, actionKey)
	assert.Nil(t, err)

	return action
}
