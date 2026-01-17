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

func TestStateActionSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(StateActionSuite))
}

type StateActionSuite struct {
	suite.Suite
	db              *sql.DB
	model           req_model.Model
	domain          model_domain.Domain
	subdomain       model_domain.Subdomain
	class           model_class.Class
	state           model_state.State
	action          model_state.Action
	actionB         model_state.Action
	stateActionKey  identity.Key
	stateActionKeyB identity.Key
}

func (suite *StateActionSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_key")))
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewClassKey(suite.subdomain.Key, "class_key")))
	suite.state = t_AddState(suite.T(), suite.db, suite.model.Key, suite.class.Key, helper.Must(identity.NewStateKey(suite.class.Key, "state_key")))
	suite.action = t_AddAction(suite.T(), suite.db, suite.model.Key, suite.class.Key, helper.Must(identity.NewActionKey(suite.class.Key, "action_key")))
	suite.actionB = t_AddAction(suite.T(), suite.db, suite.model.Key, suite.class.Key, helper.Must(identity.NewActionKey(suite.class.Key, "action_key_b")))

	// Create the state action keys for reuse.
	suite.stateActionKey = helper.Must(identity.NewStateActionKey(suite.state.Key, "entry", "key"))
	suite.stateActionKeyB = helper.Must(identity.NewStateActionKey(suite.state.Key, "exit", "key_b"))
}

func (suite *StateActionSuite) TestLoad() {

	// Nothing in database yet.
	stateKey, stateAction, err := LoadStateAction(suite.db, suite.model.Key, suite.stateActionKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), stateKey)
	assert.Empty(suite.T(), stateAction)

	_, err = dbExec(suite.db, `
		INSERT INTO state_action
			(
				model_key,
				state_key,
				state_action_key,
				action_key,
				action_when
			)
		VALUES
			(
				'model_key',
				'domain/domain_key/subdomain/subdomain_key/class/class_key/state/state_key',
				'domain/domain_key/subdomain/subdomain_key/class/class_key/state/state_key/saction/entry/key',
				'domain/domain_key/subdomain/subdomain_key/class/class_key/action/action_key',
				'entry'
			)
	`)
	assert.Nil(suite.T(), err)

	stateKey, stateAction, err = LoadStateAction(suite.db, suite.model.Key, suite.stateActionKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.state.Key, stateKey)
	assert.Equal(suite.T(), model_state.StateAction{
		Key:       suite.stateActionKey,
		ActionKey: suite.action.Key,
		When:      "entry",
	}, stateAction)
}

func (suite *StateActionSuite) TestAdd() {

	err := AddStateAction(suite.db, suite.model.Key, suite.state.Key, model_state.StateAction{
		Key:       suite.stateActionKey,
		ActionKey: suite.action.Key,
		When:      "entry",
	})
	assert.Nil(suite.T(), err)

	stateKey, stateAction, err := LoadStateAction(suite.db, suite.model.Key, suite.stateActionKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.state.Key, stateKey)
	assert.Equal(suite.T(), model_state.StateAction{
		Key:       suite.stateActionKey,
		ActionKey: suite.action.Key,
		When:      "entry",
	}, stateAction)
}

func (suite *StateActionSuite) TestUpdate() {

	err := AddStateAction(suite.db, suite.model.Key, suite.state.Key, model_state.StateAction{
		Key:       suite.stateActionKey,
		ActionKey: suite.action.Key,
		When:      "do",
	})
	assert.Nil(suite.T(), err)

	err = UpdateStateAction(suite.db, suite.model.Key, suite.state.Key, model_state.StateAction{
		Key:       suite.stateActionKey,
		ActionKey: suite.actionB.Key,
		When:      "exit",
	})
	assert.Nil(suite.T(), err)

	stateKey, stateAction, err := LoadStateAction(suite.db, suite.model.Key, suite.stateActionKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.state.Key, stateKey)
	assert.Equal(suite.T(), model_state.StateAction{
		Key:       suite.stateActionKey,
		ActionKey: suite.actionB.Key,
		When:      "exit",
	}, stateAction)
}

func (suite *StateActionSuite) TestRemove() {

	err := AddStateAction(suite.db, suite.model.Key, suite.state.Key, model_state.StateAction{
		Key:       suite.stateActionKey,
		ActionKey: suite.action.Key,
		When:      "entry",
	})
	assert.Nil(suite.T(), err)

	err = RemoveStateAction(suite.db, suite.model.Key, suite.state.Key, suite.stateActionKey)
	assert.Nil(suite.T(), err)

	stateKey, stateAction, err := LoadStateAction(suite.db, suite.model.Key, suite.stateActionKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), stateKey)
	assert.Empty(suite.T(), stateAction)
}

func (suite *StateActionSuite) TestQuery() {

	err := AddStateAction(suite.db, suite.model.Key, suite.state.Key, model_state.StateAction{
		Key:       suite.stateActionKeyB,
		ActionKey: suite.action.Key,
		When:      "exit",
	})
	assert.Nil(suite.T(), err)

	err = AddStateAction(suite.db, suite.model.Key, suite.state.Key, model_state.StateAction{
		Key:       suite.stateActionKey,
		ActionKey: suite.action.Key,
		When:      "entry",
	})
	assert.Nil(suite.T(), err)

	stateActions, err := QueryStateActions(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[identity.Key][]model_state.StateAction{
		suite.state.Key: {
			{
				Key:       suite.stateActionKey,
				ActionKey: suite.action.Key,
				When:      "entry",
			},
			{
				Key:       suite.stateActionKeyB,
				ActionKey: suite.action.Key,
				When:      "exit",
			},
		},
	}, stateActions)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddStateAction(t *testing.T, dbOrTx DbOrTx, modelKey string, stateKey identity.Key, stateActionKey identity.Key, actionKey identity.Key, when string) (stateAction model_state.StateAction) {

	err := AddStateAction(dbOrTx, modelKey, stateKey, model_state.StateAction{
		Key:       stateActionKey,
		ActionKey: actionKey,
		When:      when,
	})
	assert.Nil(t, err)

	_, stateAction, err = LoadStateAction(dbOrTx, modelKey, stateActionKey)
	assert.Nil(t, err)

	return stateAction
}
