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

func TestActorGeneralizationSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(ActorGeneralizationSuite))
}

type ActorGeneralizationSuite struct {
	suite.Suite
	db                 *sql.DB
	model              req_model.Model
	generalizationKey  identity.Key
	generalizationKeyB identity.Key
}

func (suite *ActorGeneralizationSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)

	// Create the generalization keys for reuse.
	suite.generalizationKey = helper.Must(identity.NewActorGeneralizationKey("key"))
	suite.generalizationKeyB = helper.Must(identity.NewActorGeneralizationKey("key_b"))
}

func (suite *ActorGeneralizationSuite) TestLoad() {

	// Nothing in database yet.
	generalization, err := LoadActorGeneralization(suite.db, suite.model.Key, suite.generalizationKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), generalization)

	_, err = dbExec(suite.db, `
		INSERT INTO actor_generalization
			(
				model_key,
				generalization_key,
				name,
				details,
				is_complete,
				is_static,
				uml_comment
			)
		VALUES
			(
				'model_key',
				'ageneralization/key',
				'Name',
				'Details',
				true,
				false,
				'UmlComment'
			)
	`)
	assert.Nil(suite.T(), err)

	generalization, err = LoadActorGeneralization(suite.db, suite.model.Key, suite.generalizationKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_actor.Generalization{
		Key:        suite.generalizationKey,
		Name:       "Name",
		Details:    "Details",
		IsComplete: true,
		IsStatic:   false,
		UmlComment: "UmlComment",
	}, generalization)
}

func (suite *ActorGeneralizationSuite) TestAdd() {

	err := AddActorGeneralization(suite.db, suite.model.Key, model_actor.Generalization{
		Key:        suite.generalizationKey,
		Name:       "Name",
		Details:    "Details",
		IsComplete: true,
		IsStatic:   false,
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	generalization, err := LoadActorGeneralization(suite.db, suite.model.Key, suite.generalizationKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_actor.Generalization{
		Key:        suite.generalizationKey,
		Name:       "Name",
		Details:    "Details",
		IsComplete: true,
		IsStatic:   false,
		UmlComment: "UmlComment",
	}, generalization)
}

func (suite *ActorGeneralizationSuite) TestAddNulls() {

	err := AddActorGeneralization(suite.db, suite.model.Key, model_actor.Generalization{
		Key:        suite.generalizationKey,
		Name:       "Name",
		Details:    "",
		IsComplete: false,
		IsStatic:   false,
		UmlComment: "",
	})
	assert.Nil(suite.T(), err)

	generalization, err := LoadActorGeneralization(suite.db, suite.model.Key, suite.generalizationKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_actor.Generalization{
		Key:        suite.generalizationKey,
		Name:       "Name",
		Details:    "",
		IsComplete: false,
		IsStatic:   false,
		UmlComment: "",
	}, generalization)
}

func (suite *ActorGeneralizationSuite) TestUpdate() {

	err := AddActorGeneralization(suite.db, suite.model.Key, model_actor.Generalization{
		Key:        suite.generalizationKey,
		Name:       "Name",
		Details:    "Details",
		IsComplete: true,
		IsStatic:   false,
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateActorGeneralization(suite.db, suite.model.Key, model_actor.Generalization{
		Key:        suite.generalizationKey,
		Name:       "NameX",
		Details:    "DetailsX",
		IsComplete: false,
		IsStatic:   true,
		UmlComment: "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	generalization, err := LoadActorGeneralization(suite.db, suite.model.Key, suite.generalizationKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_actor.Generalization{
		Key:        suite.generalizationKey,
		Name:       "NameX",
		Details:    "DetailsX",
		IsComplete: false,
		IsStatic:   true,
		UmlComment: "UmlCommentX",
	}, generalization)
}

func (suite *ActorGeneralizationSuite) TestUpdateNulls() {

	err := AddActorGeneralization(suite.db, suite.model.Key, model_actor.Generalization{
		Key:        suite.generalizationKey,
		Name:       "Name",
		Details:    "Details",
		IsComplete: true,
		IsStatic:   true,
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateActorGeneralization(suite.db, suite.model.Key, model_actor.Generalization{
		Key:        suite.generalizationKey,
		Name:       "NameX",
		Details:    "",
		IsComplete: false,
		IsStatic:   false,
		UmlComment: "",
	})
	assert.Nil(suite.T(), err)

	generalization, err := LoadActorGeneralization(suite.db, suite.model.Key, suite.generalizationKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_actor.Generalization{
		Key:        suite.generalizationKey,
		Name:       "NameX",
		Details:    "",
		IsComplete: false,
		IsStatic:   false,
		UmlComment: "",
	}, generalization)
}

func (suite *ActorGeneralizationSuite) TestRemove() {

	err := AddActorGeneralization(suite.db, suite.model.Key, model_actor.Generalization{
		Key:        suite.generalizationKey,
		Name:       "Name",
		Details:    "Details",
		IsComplete: true,
		IsStatic:   false,
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = RemoveActorGeneralization(suite.db, suite.model.Key, suite.generalizationKey)
	assert.Nil(suite.T(), err)

	generalization, err := LoadActorGeneralization(suite.db, suite.model.Key, suite.generalizationKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), generalization)
}

func (suite *ActorGeneralizationSuite) TestQuery() {

	err := AddActorGeneralizations(suite.db, suite.model.Key, []model_actor.Generalization{
		{
			Key:        suite.generalizationKeyB,
			Name:       "NameX",
			Details:    "DetailsX",
			IsComplete: false,
			IsStatic:   true,
			UmlComment: "UmlCommentX",
		},
		{
			Key:        suite.generalizationKey,
			Name:       "Name",
			Details:    "Details",
			IsComplete: true,
			IsStatic:   false,
			UmlComment: "UmlComment",
		},
	})
	assert.Nil(suite.T(), err)

	generalizations, err := QueryActorGeneralizations(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), []model_actor.Generalization{
		{
			Key:        suite.generalizationKey,
			Name:       "Name",
			Details:    "Details",
			IsComplete: true,
			IsStatic:   false,
			UmlComment: "UmlComment",
		},
		{
			Key:        suite.generalizationKeyB,
			Name:       "NameX",
			Details:    "DetailsX",
			IsComplete: false,
			IsStatic:   true,
			UmlComment: "UmlCommentX",
		},
	}, generalizations)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddActorGeneralization(t *testing.T, dbOrTx DbOrTx, modelKey string, generalizationKey identity.Key) (generalization model_actor.Generalization) {

	err := AddActorGeneralization(dbOrTx, modelKey, model_actor.Generalization{
		Key:        generalizationKey,
		Name:       generalizationKey.String(),
		Details:    "Details",
		IsComplete: true,
		IsStatic:   true,
		UmlComment: "UmlComment",
	})
	assert.Nil(t, err)

	generalization, err = LoadActorGeneralization(dbOrTx, modelKey, generalizationKey)
	assert.Nil(t, err)

	return generalization
}
