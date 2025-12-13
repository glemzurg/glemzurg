package database

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

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
	model     requirements.Model
	domain    requirements.Domain
	subdomain requirements.Subdomain
	class     requirements.Class
}

func (suite *EventSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key)
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key)
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, "class_key")
}

func (suite *EventSuite) TestLoad() {

	// Nothing in database yet.
	classKey, event, err := LoadEvent(suite.db, strings.ToUpper(suite.model.Key), "Key")
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
				'class_key',
				'key',
				'Name',
				'Details',
				'{"ParamA","SourceA","ParamB","SourceB"}'
			)
	`)
	assert.Nil(suite.T(), err)

	classKey, event, err = LoadEvent(suite.db, strings.ToUpper(suite.model.Key), "Key") // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "class_key", classKey)
	assert.Equal(suite.T(), requirements.Event{
		Key:        "key", // Test case-insensitive.
		Name:       "Name",
		Details:    "Details",
		Parameters: []requirements.EventParameter{{Name: "ParamA", Source: "SourceA"}, {Name: "ParamB", Source: "SourceB"}},
	}, event)
}

func (suite *EventSuite) TestAdd() {

	err := AddEvent(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.class.Key), requirements.Event{
		Key:        "KeY", // Test case-insensitive.
		Name:       "Name",
		Details:    "Details",
		Parameters: []requirements.EventParameter{{Name: "ParamA", Source: "SourceA"}, {Name: "ParamB", Source: "SourceB"}},
	})
	assert.Nil(suite.T(), err)

	classKey, event, err := LoadEvent(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "class_key", classKey)
	assert.Equal(suite.T(), requirements.Event{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		Parameters: []requirements.EventParameter{{Name: "ParamA", Source: "SourceA"}, {Name: "ParamB", Source: "SourceB"}},
	}, event)
}

func (suite *EventSuite) TestAddNoParams() {

	err := AddEvent(suite.db, suite.model.Key, suite.class.Key, requirements.Event{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		Parameters: nil,
	})
	assert.Nil(suite.T(), err)

	classKey, event, err := LoadEvent(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "class_key", classKey)
	assert.Equal(suite.T(), requirements.Event{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		Parameters: nil,
	}, event)
}

func (suite *EventSuite) TestUpdate() {

	err := AddEvent(suite.db, suite.model.Key, suite.class.Key, requirements.Event{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		Parameters: []requirements.EventParameter{{Name: "ParamA", Source: "SourceA"}, {Name: "ParamB", Source: "SourceB"}},
	})
	assert.Nil(suite.T(), err)

	err = UpdateEvent(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.class.Key), requirements.Event{
		Key:        "KeY", // Test case-insensitive.
		Name:       "NameX",
		Details:    "DetailsX",
		Parameters: []requirements.EventParameter{{Name: "ParamAX", Source: "SourceAX"}, {Name: "ParamBX", Source: "SourceBX"}},
	})
	assert.Nil(suite.T(), err)

	classKey, event, err := LoadEvent(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "class_key", classKey)
	assert.Equal(suite.T(), requirements.Event{
		Key:        "key",
		Name:       "NameX",
		Details:    "DetailsX",
		Parameters: []requirements.EventParameter{{Name: "ParamAX", Source: "SourceAX"}, {Name: "ParamBX", Source: "SourceBX"}},
	}, event)
}

func (suite *EventSuite) TestRemove() {

	err := AddEvent(suite.db, suite.model.Key, suite.class.Key, requirements.Event{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		Parameters: []requirements.EventParameter{{Name: "ParamA", Source: "SourceA"}, {Name: "ParamB", Source: "SourceB"}},
	})
	assert.Nil(suite.T(), err)

	err = RemoveEvent(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.class.Key), strings.ToUpper("key")) // Test case-insensitive.
	assert.Nil(suite.T(), err)

	classKey, event, err := LoadEvent(suite.db, suite.model.Key, "key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), classKey)
	assert.Empty(suite.T(), event)
}

func (suite *EventSuite) TestQuery() {

	err := AddEvent(suite.db, suite.model.Key, suite.class.Key, requirements.Event{
		Key:        "keyx",
		Name:       "NameX",
		Details:    "DetailsX",
		Parameters: []requirements.EventParameter{{Name: "ParamAX", Source: "SourceAX"}, {Name: "ParamBX", Source: "SourceBX"}},
	})
	assert.Nil(suite.T(), err)

	err = AddEvent(suite.db, suite.model.Key, suite.class.Key, requirements.Event{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		Parameters: []requirements.EventParameter{{Name: "ParamA", Source: "SourceA"}, {Name: "ParamB", Source: "SourceB"}},
	})
	assert.Nil(suite.T(), err)

	events, err := QueryEvents(suite.db, strings.ToUpper(suite.model.Key)) // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[string][]requirements.Event{
		"class_key": []requirements.Event{
			{
				Key:        "key",
				Name:       "Name",
				Details:    "Details",
				Parameters: []requirements.EventParameter{{Name: "ParamA", Source: "SourceA"}, {Name: "ParamB", Source: "SourceB"}},
			},
			{
				Key:        "keyx",
				Name:       "NameX",
				Details:    "DetailsX",
				Parameters: []requirements.EventParameter{{Name: "ParamAX", Source: "SourceAX"}, {Name: "ParamBX", Source: "SourceBX"}},
			},
		},
	}, events)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddEvent(t *testing.T, dbOrTx DbOrTx, modelKey, classKey, eventKey string) (event requirements.Event) {

	err := AddEvent(dbOrTx, modelKey, classKey, requirements.Event{
		Key:        eventKey,
		Name:       "Name",
		Details:    "Details",
		Parameters: []requirements.EventParameter{{Name: "ParamA", Source: "SourceA"}, {Name: "ParamB", Source: "SourceB"}},
	})
	assert.Nil(t, err)

	_, event, err = LoadEvent(dbOrTx, modelKey, eventKey)
	assert.Nil(t, err)

	return event
}
