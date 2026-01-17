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

func TestTransitionSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(TransitionSuite))
}

type TransitionSuite struct {
	suite.Suite
	db             *sql.DB
	model          req_model.Model
	domain         model_domain.Domain
	subdomain      model_domain.Subdomain
	class          model_class.Class
	stateA         model_state.State
	stateB         model_state.State
	event          model_state.Event
	eventB         model_state.Event
	guard          model_state.Guard
	guardB         model_state.Guard
	action         model_state.Action
	actionB        model_state.Action
	transitionKey  identity.Key
	transitionKeyB identity.Key
}

func (suite *TransitionSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_key")))
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewClassKey(suite.subdomain.Key, "class_key")))
	suite.stateA = t_AddState(suite.T(), suite.db, suite.model.Key, suite.class.Key, helper.Must(identity.NewStateKey(suite.class.Key, "state_key_a")))
	suite.stateB = t_AddState(suite.T(), suite.db, suite.model.Key, suite.class.Key, helper.Must(identity.NewStateKey(suite.class.Key, "state_key_b")))
	suite.event = t_AddEvent(suite.T(), suite.db, suite.model.Key, suite.class.Key, helper.Must(identity.NewEventKey(suite.class.Key, "event_key")))
	suite.eventB = t_AddEvent(suite.T(), suite.db, suite.model.Key, suite.class.Key, helper.Must(identity.NewEventKey(suite.class.Key, "event_key_b")))
	suite.guard = t_AddGuard(suite.T(), suite.db, suite.model.Key, suite.class.Key, helper.Must(identity.NewGuardKey(suite.class.Key, "guard_key")))
	suite.guardB = t_AddGuard(suite.T(), suite.db, suite.model.Key, suite.class.Key, helper.Must(identity.NewGuardKey(suite.class.Key, "guard_key_b")))
	suite.action = t_AddAction(suite.T(), suite.db, suite.model.Key, suite.class.Key, helper.Must(identity.NewActionKey(suite.class.Key, "action_key")))
	suite.actionB = t_AddAction(suite.T(), suite.db, suite.model.Key, suite.class.Key, helper.Must(identity.NewActionKey(suite.class.Key, "action_key_b")))

	// Create the transition keys for reuse.
	// NewTransitionKey(classKey, from, event, guard, action, to)
	suite.transitionKey = helper.Must(identity.NewTransitionKey(suite.class.Key, "state_key_a", "event_key", "guard_key", "action_key", "state_key_b"))
	suite.transitionKeyB = helper.Must(identity.NewTransitionKey(suite.class.Key, "state_key_b", "event_key", "guard_key", "action_key", "state_key_a"))
}

func (suite *TransitionSuite) TestLoad() {

	// Nothing in database yet.
	classKey, transition, err := LoadTransition(suite.db, suite.model.Key, suite.transitionKey)
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
				'domain/domain_key/subdomain/subdomain_key/class/class_key',
				'domain/domain_key/subdomain/subdomain_key/class/class_key/transition/state_key_a/event_key/guard_key/action_key/state_key_b',
				'domain/domain_key/subdomain/subdomain_key/class/class_key/state/state_key_a',
				'domain/domain_key/subdomain/subdomain_key/class/class_key/event/event_key',
				'domain/domain_key/subdomain/subdomain_key/class/class_key/guard/guard_key',
				'domain/domain_key/subdomain/subdomain_key/class/class_key/action/action_key',
				'domain/domain_key/subdomain/subdomain_key/class/class_key/state/state_key_b',
				'UmlComment'
			)
	`)
	assert.Nil(suite.T(), err)

	classKey, transition, err = LoadTransition(suite.db, suite.model.Key, suite.transitionKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.class.Key, classKey)
	assert.Equal(suite.T(), model_state.Transition{
		Key:          suite.transitionKey,
		FromStateKey: &suite.stateA.Key,
		EventKey:     suite.event.Key,
		GuardKey:     &suite.guard.Key,
		ActionKey:    &suite.action.Key,
		ToStateKey:   &suite.stateB.Key,
		UmlComment:   "UmlComment",
	}, transition)
}

func (suite *TransitionSuite) TestAdd() {

	err := AddTransition(suite.db, suite.model.Key, suite.class.Key, model_state.Transition{
		Key:          suite.transitionKey,
		FromStateKey: &suite.stateA.Key,
		EventKey:     suite.event.Key,
		GuardKey:     &suite.guard.Key,
		ActionKey:    &suite.action.Key,
		ToStateKey:   &suite.stateB.Key,
		UmlComment:   "UmlComment",
	})
	assert.Nil(suite.T(), err)

	classKey, transition, err := LoadTransition(suite.db, suite.model.Key, suite.transitionKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.class.Key, classKey)
	assert.Equal(suite.T(), model_state.Transition{
		Key:          suite.transitionKey,
		FromStateKey: &suite.stateA.Key,
		EventKey:     suite.event.Key,
		GuardKey:     &suite.guard.Key,
		ActionKey:    &suite.action.Key,
		ToStateKey:   &suite.stateB.Key,
		UmlComment:   "UmlComment",
	}, transition)
}

func (suite *TransitionSuite) TestAddNulls() {

	err := AddTransition(suite.db, suite.model.Key, suite.class.Key, model_state.Transition{
		Key:          suite.transitionKey,
		FromStateKey: nil,
		EventKey:     suite.event.Key,
		GuardKey:     nil,
		ActionKey:    nil,
		ToStateKey:   nil,
		UmlComment:   "UmlComment",
	})
	assert.Nil(suite.T(), err)

	classKey, transition, err := LoadTransition(suite.db, suite.model.Key, suite.transitionKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.class.Key, classKey)
	assert.Equal(suite.T(), model_state.Transition{
		Key:          suite.transitionKey,
		FromStateKey: nil,
		EventKey:     suite.event.Key,
		GuardKey:     nil,
		ActionKey:    nil,
		ToStateKey:   nil,
		UmlComment:   "UmlComment",
	}, transition)
}

func (suite *TransitionSuite) TestUpdate() {

	err := AddTransition(suite.db, suite.model.Key, suite.class.Key, model_state.Transition{
		Key:          suite.transitionKey,
		FromStateKey: &suite.stateA.Key,
		EventKey:     suite.event.Key,
		GuardKey:     &suite.guard.Key,
		ActionKey:    &suite.action.Key,
		ToStateKey:   &suite.stateB.Key,
		UmlComment:   "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateTransition(suite.db, suite.model.Key, suite.class.Key, model_state.Transition{
		Key:          suite.transitionKey,
		FromStateKey: &suite.stateB.Key,
		EventKey:     suite.eventB.Key,
		GuardKey:     &suite.guardB.Key,
		ActionKey:    &suite.actionB.Key,
		ToStateKey:   &suite.stateA.Key,
		UmlComment:   "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	classKey, transition, err := LoadTransition(suite.db, suite.model.Key, suite.transitionKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.class.Key, classKey)
	assert.Equal(suite.T(), model_state.Transition{
		Key:          suite.transitionKey,
		FromStateKey: &suite.stateB.Key,
		EventKey:     suite.eventB.Key,
		GuardKey:     &suite.guardB.Key,
		ActionKey:    &suite.actionB.Key,
		ToStateKey:   &suite.stateA.Key,
		UmlComment:   "UmlCommentX",
	}, transition)
}

func (suite *TransitionSuite) TestUpdateNulls() {

	err := AddTransition(suite.db, suite.model.Key, suite.class.Key, model_state.Transition{
		Key:          suite.transitionKey,
		FromStateKey: &suite.stateA.Key,
		EventKey:     suite.event.Key,
		GuardKey:     &suite.guard.Key,
		ActionKey:    &suite.action.Key,
		ToStateKey:   &suite.stateB.Key,
		UmlComment:   "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateTransition(suite.db, suite.model.Key, suite.class.Key, model_state.Transition{
		Key:          suite.transitionKey,
		FromStateKey: nil,
		EventKey:     suite.event.Key,
		GuardKey:     nil,
		ActionKey:    nil,
		ToStateKey:   nil,
		UmlComment:   "UmlComment",
	})
	assert.Nil(suite.T(), err)

	classKey, transition, err := LoadTransition(suite.db, suite.model.Key, suite.transitionKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.class.Key, classKey)
	assert.Equal(suite.T(), model_state.Transition{
		Key:          suite.transitionKey,
		FromStateKey: nil,
		EventKey:     suite.event.Key,
		GuardKey:     nil,
		ActionKey:    nil,
		ToStateKey:   nil,
		UmlComment:   "UmlComment",
	}, transition)
}

func (suite *TransitionSuite) TestRemove() {

	err := AddTransition(suite.db, suite.model.Key, suite.class.Key, model_state.Transition{
		Key:          suite.transitionKey,
		FromStateKey: &suite.stateA.Key,
		EventKey:     suite.event.Key,
		GuardKey:     &suite.guard.Key,
		ActionKey:    &suite.action.Key,
		ToStateKey:   &suite.stateB.Key,
		UmlComment:   "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = RemoveTransition(suite.db, suite.model.Key, suite.class.Key, suite.transitionKey)
	assert.Nil(suite.T(), err)

	classKey, transition, err := LoadTransition(suite.db, suite.model.Key, suite.transitionKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), classKey)
	assert.Empty(suite.T(), transition)
}

func (suite *TransitionSuite) TestQuery() {

	err := AddTransition(suite.db, suite.model.Key, suite.class.Key, model_state.Transition{
		Key:          suite.transitionKeyB,
		FromStateKey: &suite.stateB.Key,
		EventKey:     suite.event.Key,
		GuardKey:     &suite.guard.Key,
		ActionKey:    &suite.action.Key,
		ToStateKey:   &suite.stateA.Key,
		UmlComment:   "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	err = AddTransition(suite.db, suite.model.Key, suite.class.Key, model_state.Transition{
		Key:          suite.transitionKey,
		FromStateKey: &suite.stateA.Key,
		EventKey:     suite.event.Key,
		GuardKey:     &suite.guard.Key,
		ActionKey:    &suite.action.Key,
		ToStateKey:   &suite.stateB.Key,
		UmlComment:   "UmlComment",
	})
	assert.Nil(suite.T(), err)

	transitions, err := QueryTransitions(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[identity.Key][]model_state.Transition{
		suite.class.Key: {
			{
				Key:          suite.transitionKey,
				FromStateKey: &suite.stateA.Key,
				EventKey:     suite.event.Key,
				GuardKey:     &suite.guard.Key,
				ActionKey:    &suite.action.Key,
				ToStateKey:   &suite.stateB.Key,
				UmlComment:   "UmlComment",
			},
			{
				Key:          suite.transitionKeyB,
				FromStateKey: &suite.stateB.Key,
				EventKey:     suite.event.Key,
				GuardKey:     &suite.guard.Key,
				ActionKey:    &suite.action.Key,
				ToStateKey:   &suite.stateA.Key,
				UmlComment:   "UmlCommentX",
			},
		},
	}, transitions)
}
