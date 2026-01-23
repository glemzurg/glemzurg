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

func TestEventSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(EventSuite))
}

type EventSuite struct {
	suite.Suite
	db        *sql.DB
	model     req_model.Model
	domain    model_domain.Domain
	subdomain model_domain.Subdomain
	class     model_class.Class
	eventKey  identity.Key
	eventKeyB identity.Key
}

func (suite *EventSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_key")))
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewClassKey(suite.subdomain.Key, "class_key")))

	// Create the event keys for reuse.
	suite.eventKey = helper.Must(identity.NewEventKey(suite.class.Key, "key"))
	suite.eventKeyB = helper.Must(identity.NewEventKey(suite.class.Key, "key_b"))
}

func (suite *EventSuite) TestLoad() {

	// Nothing in database yet.
	classKey, event, err := LoadEvent(suite.db, suite.model.Key, suite.eventKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), classKey)
	assert.Empty(suite.T(), event)

	_, err = dbExec(suite.db, `
		INSERT INTO event
			(
				model_key,
				class_key,
				event_key,
				name,
				details,
				parameters
			)
		VALUES
			(
				'model_key',
				'domain/domain_key/subdomain/subdomain_key/class/class_key',
				'domain/domain_key/subdomain/subdomain_key/class/class_key/event/key',
				'Name',
				'Details',
				'{"ParamA","SourceA","ParamB","SourceB"}'
			)
	`)
	assert.Nil(suite.T(), err)

	classKey, event, err = LoadEvent(suite.db, suite.model.Key, suite.eventKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.class.Key, classKey)
	assert.Equal(suite.T(), model_state.Event{
		Key:        suite.eventKey,
		Name:       "Name",
		Details:    "Details",
		Parameters: []model_state.EventParameter{{Name: "ParamA", Source: "SourceA"}, {Name: "ParamB", Source: "SourceB"}},
	}, event)
}

func (suite *EventSuite) TestAdd() {

	err := AddEvent(suite.db, suite.model.Key, suite.class.Key, model_state.Event{
		Key:        suite.eventKey,
		Name:       "Name",
		Details:    "Details",
		Parameters: []model_state.EventParameter{{Name: "ParamA", Source: "SourceA"}, {Name: "ParamB", Source: "SourceB"}},
	})
	assert.Nil(suite.T(), err)

	classKey, event, err := LoadEvent(suite.db, suite.model.Key, suite.eventKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.class.Key, classKey)
	assert.Equal(suite.T(), model_state.Event{
		Key:        suite.eventKey,
		Name:       "Name",
		Details:    "Details",
		Parameters: []model_state.EventParameter{{Name: "ParamA", Source: "SourceA"}, {Name: "ParamB", Source: "SourceB"}},
	}, event)
}

func (suite *EventSuite) TestAddNoParams() {

	err := AddEvent(suite.db, suite.model.Key, suite.class.Key, model_state.Event{
		Key:        suite.eventKey,
		Name:       "Name",
		Details:    "Details",
		Parameters: nil,
	})
	assert.Nil(suite.T(), err)

	classKey, event, err := LoadEvent(suite.db, suite.model.Key, suite.eventKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.class.Key, classKey)
	assert.Equal(suite.T(), model_state.Event{
		Key:        suite.eventKey,
		Name:       "Name",
		Details:    "Details",
		Parameters: nil,
	}, event)
}

func (suite *EventSuite) TestUpdate() {

	err := AddEvent(suite.db, suite.model.Key, suite.class.Key, model_state.Event{
		Key:        suite.eventKey,
		Name:       "Name",
		Details:    "Details",
		Parameters: []model_state.EventParameter{{Name: "ParamA", Source: "SourceA"}, {Name: "ParamB", Source: "SourceB"}},
	})
	assert.Nil(suite.T(), err)

	err = UpdateEvent(suite.db, suite.model.Key, suite.class.Key, model_state.Event{
		Key:        suite.eventKey,
		Name:       "NameX",
		Details:    "DetailsX",
		Parameters: []model_state.EventParameter{{Name: "ParamAX", Source: "SourceAX"}, {Name: "ParamBX", Source: "SourceBX"}},
	})
	assert.Nil(suite.T(), err)

	classKey, event, err := LoadEvent(suite.db, suite.model.Key, suite.eventKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.class.Key, classKey)
	assert.Equal(suite.T(), model_state.Event{
		Key:        suite.eventKey,
		Name:       "NameX",
		Details:    "DetailsX",
		Parameters: []model_state.EventParameter{{Name: "ParamAX", Source: "SourceAX"}, {Name: "ParamBX", Source: "SourceBX"}},
	}, event)
}

func (suite *EventSuite) TestRemove() {

	err := AddEvent(suite.db, suite.model.Key, suite.class.Key, model_state.Event{
		Key:        suite.eventKey,
		Name:       "Name",
		Details:    "Details",
		Parameters: []model_state.EventParameter{{Name: "ParamA", Source: "SourceA"}, {Name: "ParamB", Source: "SourceB"}},
	})
	assert.Nil(suite.T(), err)

	err = RemoveEvent(suite.db, suite.model.Key, suite.class.Key, suite.eventKey)
	assert.Nil(suite.T(), err)

	classKey, event, err := LoadEvent(suite.db, suite.model.Key, suite.eventKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), classKey)
	assert.Empty(suite.T(), event)
}

func (suite *EventSuite) TestQuery() {

	err := AddEvents(suite.db, suite.model.Key, map[identity.Key][]model_state.Event{
		suite.class.Key: {
			{
				Key:        suite.eventKeyB,
				Name:       "NameX",
				Details:    "DetailsX",
				Parameters: []model_state.EventParameter{{Name: "ParamAX", Source: "SourceAX"}, {Name: "ParamBX", Source: "SourceBX"}},
			},
			{
				Key:        suite.eventKey,
				Name:       "Name",
				Details:    "Details",
				Parameters: []model_state.EventParameter{{Name: "ParamA", Source: "SourceA"}, {Name: "ParamB", Source: "SourceB"}},
			},
		},
	})
	assert.Nil(suite.T(), err)

	events, err := QueryEvents(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[identity.Key][]model_state.Event{
		suite.class.Key: {
			{
				Key:        suite.eventKey,
				Name:       "Name",
				Details:    "Details",
				Parameters: []model_state.EventParameter{{Name: "ParamA", Source: "SourceA"}, {Name: "ParamB", Source: "SourceB"}},
			},
			{
				Key:        suite.eventKeyB,
				Name:       "NameX",
				Details:    "DetailsX",
				Parameters: []model_state.EventParameter{{Name: "ParamAX", Source: "SourceAX"}, {Name: "ParamBX", Source: "SourceBX"}},
			},
		},
	}, events)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddEvent(t *testing.T, dbOrTx DbOrTx, modelKey string, classKey identity.Key, eventKey identity.Key) (event model_state.Event) {

	err := AddEvent(dbOrTx, modelKey, classKey, model_state.Event{
		Key:        eventKey,
		Name:       eventKey.String(),
		Details:    "Details",
		Parameters: []model_state.EventParameter{{Name: "ParamA", Source: "SourceA"}, {Name: "ParamB", Source: "SourceB"}},
	})
	assert.Nil(t, err)

	_, event, err = LoadEvent(dbOrTx, modelKey, eventKey)
	assert.Nil(t, err)

	return event
}
