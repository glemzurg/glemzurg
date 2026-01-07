package database

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_actor"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestActorSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(ActorSuite))
}

type ActorSuite struct {
	suite.Suite
	db    *sql.DB
	model requirements.Model
}

func (suite *ActorSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
}

func (suite *ActorSuite) TestLoad() {

	// Nothing in database yet.
	actor, err := LoadActor(suite.db, strings.ToUpper(suite.model.Key), "Key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), actor)

	_, err = dbExec(suite.db, `
		INSERT INTO actor
			(
				model_key,
				actor_key,
				name,
				details,
				actor_type,
				uml_comment
			)
		VALUES
			(
				'model_key',
				'key',
				'Name',
				'Details',
				'person',
				'UmlComment'
			)
	`)
	assert.Nil(suite.T(), err)

	actor, err = LoadActor(suite.db, strings.ToUpper(suite.model.Key), "Key") // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_actor.Actor{
		Key:        "key", // Test case-insensitive.
		Name:       "Name",
		Details:    "Details",
		Type:       "person",
		UmlComment: "UmlComment",
	}, actor)
}

func (suite *ActorSuite) TestAdd() {

	err := AddActor(suite.db, strings.ToUpper(suite.model.Key), model_actor.Actor{
		Key:        "KeY", // Test case-insensitive.
		Name:       "Name",
		Details:    "Details",
		Type:       "person",
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	actor, err := LoadActor(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_actor.Actor{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		Type:       "person",
		UmlComment: "UmlComment",
	}, actor)
}

func (suite *ActorSuite) TestUpdate() {

	err := AddActor(suite.db, suite.model.Key, model_actor.Actor{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		Type:       "person",
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateActor(suite.db, strings.ToUpper(suite.model.Key), model_actor.Actor{
		Key:        "kEy", // Test case-insensitive.
		Name:       "NameX",
		Details:    "DetailsX",
		Type:       "system",
		UmlComment: "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	actor, err := LoadActor(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_actor.Actor{
		Key:        "key", // Test case-insensitive.
		Name:       "NameX",
		Details:    "DetailsX",
		Type:       "system",
		UmlComment: "UmlCommentX",
	}, actor)
}

func (suite *ActorSuite) TestRemove() {

	err := AddActor(suite.db, suite.model.Key, model_actor.Actor{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		Type:       "person",
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = RemoveActor(suite.db, strings.ToUpper(suite.model.Key), "kEy") // Test case-insensitive.
	assert.Nil(suite.T(), err)

	actor, err := LoadActor(suite.db, suite.model.Key, "key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), actor)
}

func (suite *ActorSuite) TestQuery() {

	err := AddActor(suite.db, suite.model.Key, model_actor.Actor{
		Key:        "keyx",
		Name:       "NameX",
		Details:    "DetailsX",
		Type:       "system",
		UmlComment: "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	err = AddActor(suite.db, suite.model.Key, model_actor.Actor{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		Type:       "person",
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	actors, err := QueryActors(suite.db, strings.ToUpper(suite.model.Key)) // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), []model_actor.Actor{
		{
			Key:        "key",
			Name:       "Name",
			Details:    "Details",
			Type:       "person",
			UmlComment: "UmlComment",
		},
		{

			Key:        "keyx",
			Name:       "NameX",
			Details:    "DetailsX",
			Type:       "system",
			UmlComment: "UmlCommentX",
		},
	}, actors)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddActor(t *testing.T, dbOrTx DbOrTx, modelKey, actorKey string) (actor model_actor.Actor) {

	err := AddActor(dbOrTx, modelKey, model_actor.Actor{
		Key:        actorKey,
		Name:       actorKey,
		Details:    "Details",
		Type:       "person",
		UmlComment: "UmlComment",
	})
	assert.Nil(t, err)

	actor, err = LoadActor(dbOrTx, modelKey, actorKey)
	assert.Nil(t, err)

	return actor
}
