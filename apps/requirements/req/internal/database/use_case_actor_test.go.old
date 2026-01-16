package database

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestUseCaseActorSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(ActorSuite))
}

type UseCaseActorSuite struct {
	suite.Suite
	db        *sql.DB
	model     requirements.Model
	actor     model_actor.Actor
	actorB    model_actor.Actor
	domain    model_domain.Domain
	subdomain model_domain.Subdomain
	useCase   model_use_case.UseCase
}

func (suite *UseCaseActorSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.actor = t_AddActor(suite.T(), suite.db, suite.model.Key, "actor_key")
	suite.actorB = t_AddActor(suite.T(), suite.db, suite.model.Key, "actor_key_b")
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key)
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key)
	suite.useCase = t_AddUseCase(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, "use_case_key")
}

func (suite *UseCaseActorSuite) TestLoad() {

	// Nothing in database yet.
	actor, err := LoadUseCaseActor(suite.db, strings.ToUpper(suite.model.Key), "Use_Case_Key", "Actor_Key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), actor)

	_, err = dbExec(suite.db, `
		INSERT INTO use_case_actor
			(
				model_key,
				use_case_key,
				actor_key,
				uml_comment
			)
		VALUES
			(
				'model_key',
				'use_case_key',
				'actor_key',
				'UmlComment'
			)
	`)
	assert.Nil(suite.T(), err)

	actor, err = LoadUseCaseActor(suite.db, strings.ToUpper(suite.model.Key), "Use_Case_Key", "Actor_Key") // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_use_case.Actor{
		UmlComment: "UmlComment",
	}, actor)
}

func (suite *UseCaseActorSuite) TestAdd() {

	err := AddUseCaseActor(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.useCase.Key), strings.ToUpper(suite.actor.Key), model_use_case.Actor{
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	actor, err := LoadUseCaseActor(suite.db, suite.model.Key, suite.useCase.Key, suite.actor.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_use_case.Actor{
		UmlComment: "UmlComment",
	}, actor)
}

func (suite *UseCaseActorSuite) TestUpdate() {

	err := AddUseCaseActor(suite.db, suite.model.Key, suite.useCase.Key, suite.actor.Key, model_use_case.Actor{
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateUseCaseActor(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.useCase.Key), strings.ToUpper(suite.actor.Key), model_use_case.Actor{
		UmlComment: "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	actor, err := LoadUseCaseActor(suite.db, suite.model.Key, suite.useCase.Key, suite.actor.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_use_case.Actor{
		UmlComment: "UmlCommentX",
	}, actor)
}

func (suite *UseCaseActorSuite) TestRemove() {

	err := AddUseCaseActor(suite.db, suite.model.Key, suite.useCase.Key, suite.actor.Key, model_use_case.Actor{
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = RemoveUseCaseActor(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.useCase.Key), strings.ToUpper(suite.actor.Key)) // Test case-insensitive.
	assert.Nil(suite.T(), err)

	actor, err := LoadUseCaseActor(suite.db, suite.model.Key, suite.useCase.Key, suite.actor.Key)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), actor)
}

func (suite *UseCaseActorSuite) TestQuery() {

	err := AddUseCaseActor(suite.db, suite.model.Key, suite.useCase.Key, suite.actor.Key, model_use_case.Actor{
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = AddUseCaseActor(suite.db, suite.model.Key, suite.useCase.Key, suite.actorB.Key, model_use_case.Actor{
		UmlComment: "UmlCommentB",
	})
	assert.Nil(suite.T(), err)

	actors, err := QueryActors(suite.db, strings.ToUpper(suite.model.Key)) // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[string]map[string]model_use_case.Actor{
		"use_case_key": map[string]model_use_case.Actor{
			"actor_key": {
				UmlComment: "UmlComment",
			},
			"actor_key_b": {
				UmlComment: "UmlCommentB",
			},
		},
	}, actors)
}
