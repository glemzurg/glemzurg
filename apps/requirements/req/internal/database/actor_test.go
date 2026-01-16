package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"

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
	db        *sql.DB
	model     req_model.Model
	actorKey  identity.Key
	actorKeyB identity.Key
}

func (suite *ActorSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)

	// Create the actor keys for reuse.
	suite.actorKey = helper.Must(identity.NewActorKey("key"))
	suite.actorKeyB = helper.Must(identity.NewActorKey("key_b"))
}

func (suite *ActorSuite) TestLoad() {

	// Nothing in database yet.
	actor, err := LoadActor(suite.db, suite.model.Key, suite.actorKey)
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
				'actor/key',
				'Name',
				'Details',
				'person',
				'UmlComment'
			)
	`)
	assert.Nil(suite.T(), err)

	actor, err = LoadActor(suite.db, suite.model.Key, suite.actorKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_actor.Actor{
		Key:        suite.actorKey,
		Name:       "Name",
		Details:    "Details",
		Type:       "person",
		UmlComment: "UmlComment",
	}, actor)
}

func (suite *ActorSuite) TestAdd() {

	err := AddActor(suite.db, suite.model.Key, model_actor.Actor{
		Key:        suite.actorKey,
		Name:       "Name",
		Details:    "Details",
		Type:       "person",
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	actor, err := LoadActor(suite.db, suite.model.Key, suite.actorKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_actor.Actor{
		Key:        suite.actorKey,
		Name:       "Name",
		Details:    "Details",
		Type:       "person",
		UmlComment: "UmlComment",
	}, actor)
}

func (suite *ActorSuite) TestUpdate() {

	err := AddActor(suite.db, suite.model.Key, model_actor.Actor{
		Key:        suite.actorKey,
		Name:       "Name",
		Details:    "Details",
		Type:       "person",
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateActor(suite.db, suite.model.Key, model_actor.Actor{
		Key:        suite.actorKey,
		Name:       "NameX",
		Details:    "DetailsX",
		Type:       "system",
		UmlComment: "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	actor, err := LoadActor(suite.db, suite.model.Key, suite.actorKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_actor.Actor{
		Key:        suite.actorKey,
		Name:       "NameX",
		Details:    "DetailsX",
		Type:       "system",
		UmlComment: "UmlCommentX",
	}, actor)
}

func (suite *ActorSuite) TestRemove() {

	err := AddActor(suite.db, suite.model.Key, model_actor.Actor{
		Key:        suite.actorKey,
		Name:       "Name",
		Details:    "Details",
		Type:       "person",
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = RemoveActor(suite.db, suite.model.Key, suite.actorKey)
	assert.Nil(suite.T(), err)

	actor, err := LoadActor(suite.db, suite.model.Key, suite.actorKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), actor)
}

func (suite *ActorSuite) TestQuery() {

	err := AddActor(suite.db, suite.model.Key, model_actor.Actor{
		Key:        suite.actorKeyB,
		Name:       "NameX",
		Details:    "DetailsX",
		Type:       "system",
		UmlComment: "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	err = AddActor(suite.db, suite.model.Key, model_actor.Actor{
		Key:        suite.actorKey,
		Name:       "Name",
		Details:    "Details",
		Type:       "person",
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	actors, err := QueryActors(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), []model_actor.Actor{
		{
			Key:        suite.actorKey,
			Name:       "Name",
			Details:    "Details",
			Type:       "person",
			UmlComment: "UmlComment",
		},
		{
			Key:        suite.actorKeyB,
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

func t_AddActor(t *testing.T, dbOrTx DbOrTx, modelKey string, actorKey identity.Key) (actor model_actor.Actor) {

	err := AddActor(dbOrTx, modelKey, model_actor.Actor{
		Key:        actorKey,
		Name:       actorKey.String(),
		Details:    "Details",
		Type:       "person",
		UmlComment: "UmlComment",
	})
	assert.Nil(t, err)

	actor, err = LoadActor(dbOrTx, modelKey, actorKey)
	assert.Nil(t, err)

	return actor
}
