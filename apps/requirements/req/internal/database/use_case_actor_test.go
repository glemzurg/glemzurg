package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
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
	suite.Run(t, new(UseCaseActorSuite))
}

type UseCaseActorSuite struct {
	suite.Suite
	db        *sql.DB
	model     req_model.Model
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
	suite.actor = t_AddActor(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewActorKey("actor_key")))
	suite.actorB = t_AddActor(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewActorKey("actor_key_b")))
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_key")))
	suite.useCase = t_AddUseCase(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewUseCaseKey(suite.subdomain.Key, "use_case_key")))
}

func (suite *UseCaseActorSuite) TestLoad() {

	// Nothing in database yet.
	actor, err := LoadUseCaseActor(suite.db, suite.model.Key, suite.useCase.Key, suite.actor.Key)
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
				'domain/domain_key/subdomain/subdomain_key/usecase/use_case_key',
				'actor/actor_key',
				'UmlComment'
			)
	`)
	assert.Nil(suite.T(), err)

	actor, err = LoadUseCaseActor(suite.db, suite.model.Key, suite.useCase.Key, suite.actor.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_use_case.Actor{
		UmlComment: "UmlComment",
	}, actor)
}

func (suite *UseCaseActorSuite) TestAdd() {

	err := AddUseCaseActor(suite.db, suite.model.Key, suite.useCase.Key, suite.actor.Key, model_use_case.Actor{
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

	err = UpdateUseCaseActor(suite.db, suite.model.Key, suite.useCase.Key, suite.actor.Key, model_use_case.Actor{
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

	err = RemoveUseCaseActor(suite.db, suite.model.Key, suite.useCase.Key, suite.actor.Key)
	assert.Nil(suite.T(), err)

	actor, err := LoadUseCaseActor(suite.db, suite.model.Key, suite.useCase.Key, suite.actor.Key)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), actor)
}

func (suite *UseCaseActorSuite) TestQuery() {

	err := AddUseCaseActors(suite.db, suite.model.Key, map[identity.Key]map[identity.Key]model_use_case.Actor{
		suite.useCase.Key: {
			suite.actor.Key: {
				UmlComment: "UmlComment",
			},
			suite.actorB.Key: {
				UmlComment: "UmlCommentB",
			},
		},
	})
	assert.Nil(suite.T(), err)

	actors, err := QueryUseCaseActors(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[identity.Key]map[identity.Key]model_use_case.Actor{
		suite.useCase.Key: {
			suite.actor.Key: {
				UmlComment: "UmlComment",
			},
			suite.actorB.Key: {
				UmlComment: "UmlCommentB",
			},
		},
	}, actors)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddUseCaseActor(t *testing.T, dbOrTx DbOrTx, modelKey string, useCaseKey identity.Key, actorKey identity.Key) (actor model_use_case.Actor) {

	err := AddUseCaseActor(dbOrTx, modelKey, useCaseKey, actorKey, model_use_case.Actor{
		UmlComment: "UmlComment",
	})
	assert.Nil(t, err)

	actor, err = LoadUseCaseActor(dbOrTx, modelKey, useCaseKey, actorKey)
	assert.Nil(t, err)

	return actor
}
