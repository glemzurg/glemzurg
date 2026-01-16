package database

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
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
	db        *sql.DB
	model     requirements.Model
	domain    model_domain.Domain
	subdomain model_domain.Subdomain
	class     model_class.Class
	state     model_state.State
	action    model_state.Action
	actionB   model_state.Action
}

func (suite *StateActionSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key)
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key)
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, "state_key")
	suite.state = t_AddState(suite.T(), suite.db, suite.model.Key, suite.class.Key, "state_key")
	suite.action = t_AddAction(suite.T(), suite.db, suite.model.Key, suite.class.Key, "action_key")
	suite.actionB = t_AddAction(suite.T(), suite.db, suite.model.Key, suite.class.Key, "action_key_b")
}

func (suite *StateActionSuite) TestLoad() {

	// Nothing in database yet.
	stateKey, stateAction, err := LoadStateAction(suite.db, strings.ToUpper(suite.model.Key), "Key")
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
				'state_key',
				'key',
				'action_key',
				'entry'
			)
	`)
	assert.Nil(suite.T(), err)

	stateKey, stateAction, err = LoadStateAction(suite.db, strings.ToUpper(suite.model.Key), "Key") // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "state_key", stateKey)
	assert.Equal(suite.T(), model_state.StateAction{
		Key:       "key", // Test case-insensitive.
		ActionKey: "action_key",
		When:      "entry",
	}, stateAction)
}

func (suite *StateActionSuite) TestAdd() {

	err := AddStateAction(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.state.Key), model_state.StateAction{
		Key:       "KeY",        // Test case-insensitive.
		ActionKey: "action_KEY", // Test case-insensitive.
		When:      "entry",
	})
	assert.Nil(suite.T(), err)

	stateKey, stateAction, err := LoadStateAction(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "state_key", stateKey)
	assert.Equal(suite.T(), model_state.StateAction{
		Key:       "key",
		ActionKey: "action_key",
		When:      "entry",
	}, stateAction)
}

func (suite *StateActionSuite) TestUpdate() {

	err := AddStateAction(suite.db, suite.model.Key, suite.state.Key, model_state.StateAction{
		Key:       "key",
		ActionKey: "action_key",
		When:      "do",
	})
	assert.Nil(suite.T(), err)

	err = UpdateStateAction(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.state.Key), model_state.StateAction{
		Key:       "KeY",          // Test case-insensitive.
		ActionKey: "action_KEY_b", // Test case-insensitive.
		When:      "exit",
	})
	assert.Nil(suite.T(), err)

	stateKey, stateAction, err := LoadStateAction(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "state_key", stateKey)
	assert.Equal(suite.T(), model_state.StateAction{
		Key:       "key",
		ActionKey: "action_key_b",
		When:      "exit",
	}, stateAction)
}

func (suite *StateActionSuite) TestRemove() {

	err := AddStateAction(suite.db, suite.model.Key, suite.state.Key, model_state.StateAction{
		Key:       "key",
		ActionKey: "action_key",
		When:      "entry",
	})
	assert.Nil(suite.T(), err)

	err = RemoveStateAction(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.state.Key), strings.ToUpper("key")) // Test case-insensitive.
	assert.Nil(suite.T(), err)

	stateKey, stateAction, err := LoadStateAction(suite.db, suite.model.Key, "key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), stateKey)
	assert.Empty(suite.T(), stateAction)
}

func (suite *StateActionSuite) TestQuery() {

	err := AddStateAction(suite.db, suite.model.Key, suite.state.Key, model_state.StateAction{
		Key:       "keyx",
		ActionKey: "action_key",
		When:      "exit",
	})
	assert.Nil(suite.T(), err)

	err = AddStateAction(suite.db, suite.model.Key, suite.state.Key, model_state.StateAction{
		Key:       "key",
		ActionKey: "action_key",
		When:      "entry",
	})
	assert.Nil(suite.T(), err)

	stateActions, err := QueryStateActions(suite.db, strings.ToUpper(suite.model.Key)) // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[string][]model_state.StateAction{
		"state_key": []model_state.StateAction{
			{
				Key:       "key",
				ActionKey: "action_key",
				When:      "entry",
			},
			{
				Key:       "keyx",
				ActionKey: "action_key",
				When:      "exit",
			},
		},
	}, stateActions)
}
