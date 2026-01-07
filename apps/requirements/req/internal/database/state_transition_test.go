package database

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_state"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestTransitionSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(TransitionSuite))
}

type TransitionSuite struct {
	suite.Suite
	db        *sql.DB
	model     requirements.Model
	domain    model_domain.Domain
	subdomain model_domain.Subdomain
	class     model_class.Class
	stateA    model_state.State
	stateB    model_state.State
	event     model_state.Event
	eventB    model_state.Event
	guard     model_state.Guard
	guardB    model_state.Guard
	action    model_state.Action
	actionB   model_state.Action
}

func (suite *TransitionSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key)
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key)
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, "class_key")
	suite.stateA = t_AddState(suite.T(), suite.db, suite.model.Key, suite.class.Key, "state_key_a")
	suite.stateB = t_AddState(suite.T(), suite.db, suite.model.Key, suite.class.Key, "state_key_b")
	suite.event = t_AddEvent(suite.T(), suite.db, suite.model.Key, suite.class.Key, "event_key")
	suite.eventB = t_AddEvent(suite.T(), suite.db, suite.model.Key, suite.class.Key, "event_key_b")
	suite.guard = t_AddGuard(suite.T(), suite.db, suite.model.Key, suite.class.Key, "guard_key")
	suite.guardB = t_AddGuard(suite.T(), suite.db, suite.model.Key, suite.class.Key, "guard_key_b")
	suite.action = t_AddAction(suite.T(), suite.db, suite.model.Key, suite.class.Key, "action_key")
	suite.actionB = t_AddAction(suite.T(), suite.db, suite.model.Key, suite.class.Key, "action_key_b")
}

func (suite *TransitionSuite) TestLoad() {

	// Nothing in database yet.
	classKey, transition, err := LoadTransition(suite.db, strings.ToUpper(suite.model.Key), "Key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), classKey)
	assert.Empty(suite.T(), transition)

	_, err = dbExec(suite.db, `
		INSERT INTO transition
			(
				model_key,
				class_key,
				transition_key,
				from_state_key,
				event_key,
				guard_key,
				action_key,
				to_state_key,
				uml_comment
			)
		VALUES
			(
				'model_key',
				'class_key',
				'key',
				'state_key_a',
				'event_key',
				'guard_key',
				'action_key',
				'state_key_b',
				'UmlComment'
			)
	`)
	assert.Nil(suite.T(), err)

	classKey, transition, err = LoadTransition(suite.db, strings.ToUpper(suite.model.Key), "Key") // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "class_key", classKey)
	assert.Equal(suite.T(), model_state.Transition{
		Key:          "key", // Test case-insensitive.
		FromStateKey: "state_key_a",
		EventKey:     "event_key",
		GuardKey:     "guard_key",
		ActionKey:    "action_key",
		ToStateKey:   "state_key_b",
		UmlComment:   "UmlComment",
	}, transition)
}

func (suite *TransitionSuite) TestAdd() {

	err := AddTransition(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.class.Key), model_state.Transition{
		Key:          "KeY",         // Test case-insensitive.
		FromStateKey: "state_KEY_a", // Test case-insensitive.
		EventKey:     "event_KEY",   // Test case-insensitive.
		GuardKey:     "guard_KEY",   // Test case-insensitive.
		ActionKey:    "action_KEY",  // Test case-insensitive.
		ToStateKey:   "state_KEY_b", // Test case-insensitive.
		UmlComment:   "UmlComment",
	})
	assert.Nil(suite.T(), err)

	classKey, transition, err := LoadTransition(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "class_key", classKey)
	assert.Equal(suite.T(), model_state.Transition{
		Key:          "key",
		FromStateKey: "state_key_a",
		EventKey:     "event_key",
		GuardKey:     "guard_key",
		ActionKey:    "action_key",
		ToStateKey:   "state_key_b",
		UmlComment:   "UmlComment",
	}, transition)
}

func (suite *TransitionSuite) TestAddNulls() {

	err := AddTransition(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.class.Key), model_state.Transition{
		Key:          "key",
		FromStateKey: "",
		EventKey:     "event_key",
		GuardKey:     "",
		ActionKey:    "",
		ToStateKey:   "",
		UmlComment:   "UmlComment"})
	assert.Nil(suite.T(), err)

	classKey, transition, err := LoadTransition(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "class_key", classKey)
	assert.Equal(suite.T(), model_state.Transition{
		Key:          "key",
		FromStateKey: "",
		EventKey:     "event_key",
		GuardKey:     "",
		ActionKey:    "",
		ToStateKey:   "",
		UmlComment:   "UmlComment",
	}, transition)
}

func (suite *TransitionSuite) TestUpdate() {

	err := AddTransition(suite.db, suite.model.Key, suite.class.Key, model_state.Transition{
		Key:          "key",
		FromStateKey: "state_key_a",
		EventKey:     "event_key",
		GuardKey:     "guard_key",
		ActionKey:    "action_key",
		ToStateKey:   "state_key_b",
		UmlComment:   "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateTransition(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.class.Key), model_state.Transition{
		Key:          "KeY",          // Test case-insensitive.
		FromStateKey: "state_KEY_b",  // Test case-insensitive.
		EventKey:     "event_KEY_b",  // Test case-insensitive.
		GuardKey:     "guard_KEY_b",  // Test case-insensitive.
		ActionKey:    "action_KEY_b", // Test case-insensitive.
		ToStateKey:   "state_KEY_a",  // Test case-insensitive.
		UmlComment:   "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	classKey, transition, err := LoadTransition(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "class_key", classKey)
	assert.Equal(suite.T(), model_state.Transition{
		Key:          "key",
		FromStateKey: "state_key_b",
		EventKey:     "event_key_b",
		GuardKey:     "guard_key_b",
		ActionKey:    "action_key_b",
		ToStateKey:   "state_key_a",
		UmlComment:   "UmlCommentX",
	}, transition)
}

func (suite *TransitionSuite) TestUpdateNulls() {

	err := AddTransition(suite.db, suite.model.Key, suite.class.Key, model_state.Transition{
		Key:          "key",
		FromStateKey: "state_key_a",
		EventKey:     "event_key",
		GuardKey:     "guard_key",
		ActionKey:    "action_key",
		ToStateKey:   "state_key_b",
		UmlComment:   "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateTransition(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.class.Key), model_state.Transition{
		Key:          "key",
		FromStateKey: "",
		EventKey:     "event_key",
		GuardKey:     "",
		ActionKey:    "",
		ToStateKey:   "",
		UmlComment:   "UmlComment",
	})
	assert.Nil(suite.T(), err)

	classKey, transition, err := LoadTransition(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "class_key", classKey)
	assert.Equal(suite.T(), model_state.Transition{
		Key:          "key",
		FromStateKey: "",
		EventKey:     "event_key",
		GuardKey:     "",
		ActionKey:    "",
		ToStateKey:   "",
		UmlComment:   "UmlComment",
	}, transition)
}

func (suite *TransitionSuite) TestRemove() {

	err := AddTransition(suite.db, suite.model.Key, suite.class.Key, model_state.Transition{
		Key:          "key",
		FromStateKey: "state_key_a",
		EventKey:     "event_key",
		GuardKey:     "guard_key",
		ActionKey:    "action_key",
		ToStateKey:   "state_key_b",
		UmlComment:   "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = RemoveTransition(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.class.Key), strings.ToUpper("key")) // Test case-insensitive.
	assert.Nil(suite.T(), err)

	classKey, transition, err := LoadTransition(suite.db, suite.model.Key, "key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), classKey)
	assert.Empty(suite.T(), transition)
}

func (suite *TransitionSuite) TestQuery() {

	err := AddTransition(suite.db, suite.model.Key, suite.class.Key, model_state.Transition{
		Key:          "keyx",
		FromStateKey: "state_key_a",
		EventKey:     "event_key",
		GuardKey:     "guard_key",
		ActionKey:    "action_key",
		ToStateKey:   "state_key_b",
		UmlComment:   "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	err = AddTransition(suite.db, suite.model.Key, suite.class.Key, model_state.Transition{
		Key:          "key",
		FromStateKey: "state_key_a",
		EventKey:     "event_key",
		GuardKey:     "guard_key",
		ActionKey:    "action_key",
		ToStateKey:   "state_key_b",
		UmlComment:   "UmlComment",
	})
	assert.Nil(suite.T(), err)

	transitions, err := QueryTransitions(suite.db, strings.ToUpper(suite.model.Key)) // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[string][]model_state.Transition{
		"class_key": []model_state.Transition{
			{
				Key:          "key",
				FromStateKey: "state_key_a",
				EventKey:     "event_key",
				GuardKey:     "guard_key",
				ActionKey:    "action_key",
				ToStateKey:   "state_key_b",
				UmlComment:   "UmlComment",
			},
			{
				Key:          "keyx",
				FromStateKey: "state_key_a",
				EventKey:     "event_key",
				GuardKey:     "guard_key",
				ActionKey:    "action_key",
				ToStateKey:   "state_key_b",
				UmlComment:   "UmlCommentX",
			},
		},
	}, transitions)
}
