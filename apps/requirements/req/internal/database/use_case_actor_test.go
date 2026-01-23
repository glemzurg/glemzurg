package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
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
	class     model_class.Class
	classB    model_class.Class
	domain    model_domain.Domain
	subdomain model_domain.Subdomain
	useCase   model_use_case.UseCase
}

func (suite *UseCaseActorSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_key")))
	suite.class = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewClassKey(suite.subdomain.Key, "class_key")))
	suite.classB = t_AddClass(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewClassKey(suite.subdomain.Key, "class_key_b")))
	suite.useCase = t_AddUseCase(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewUseCaseKey(suite.subdomain.Key, "use_case_key")))
}

func (suite *UseCaseActorSuite) TestLoad() {

	// Nothing in database yet.
	actor, err := LoadUseCaseActor(suite.db, suite.model.Key, suite.useCase.Key, suite.class.Key)
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
				'domain/domain_key/subdomain/subdomain_key/class/class_key',
				'UmlComment'
			)
	`)
	assert.Nil(suite.T(), err)

	actor, err = LoadUseCaseActor(suite.db, suite.model.Key, suite.useCase.Key, suite.class.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_use_case.Actor{
		UmlComment: "UmlComment",
	}, actor)
}

func (suite *UseCaseActorSuite) TestAdd() {

	err := AddUseCaseActor(suite.db, suite.model.Key, suite.useCase.Key, suite.class.Key, model_use_case.Actor{
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	actor, err := LoadUseCaseActor(suite.db, suite.model.Key, suite.useCase.Key, suite.class.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_use_case.Actor{
		UmlComment: "UmlComment",
	}, actor)
}

func (suite *UseCaseActorSuite) TestUpdate() {

	err := AddUseCaseActor(suite.db, suite.model.Key, suite.useCase.Key, suite.class.Key, model_use_case.Actor{
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateUseCaseActor(suite.db, suite.model.Key, suite.useCase.Key, suite.class.Key, model_use_case.Actor{
		UmlComment: "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	actor, err := LoadUseCaseActor(suite.db, suite.model.Key, suite.useCase.Key, suite.class.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_use_case.Actor{
		UmlComment: "UmlCommentX",
	}, actor)
}

func (suite *UseCaseActorSuite) TestRemove() {

	err := AddUseCaseActor(suite.db, suite.model.Key, suite.useCase.Key, suite.class.Key, model_use_case.Actor{
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = RemoveUseCaseActor(suite.db, suite.model.Key, suite.useCase.Key, suite.class.Key)
	assert.Nil(suite.T(), err)

	actor, err := LoadUseCaseActor(suite.db, suite.model.Key, suite.useCase.Key, suite.class.Key)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), actor)
}

func (suite *UseCaseActorSuite) TestQuery() {

	err := AddUseCaseActors(suite.db, suite.model.Key, map[identity.Key]map[identity.Key]model_use_case.Actor{
		suite.useCase.Key: {
			suite.class.Key: {
				UmlComment: "UmlComment",
			},
			suite.classB.Key: {
				UmlComment: "UmlCommentB",
			},
		},
	})
	assert.Nil(suite.T(), err)

	actors, err := QueryUseCaseActors(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[identity.Key]map[identity.Key]model_use_case.Actor{
		suite.useCase.Key: {
			suite.class.Key: {
				UmlComment: "UmlComment",
			},
			suite.classB.Key: {
				UmlComment: "UmlCommentB",
			},
		},
	}, actors)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddUseCaseActor(t *testing.T, dbOrTx DbOrTx, modelKey string, useCaseKey identity.Key, classKey identity.Key) (actor model_use_case.Actor) {

	err := AddUseCaseActor(dbOrTx, modelKey, useCaseKey, classKey, model_use_case.Actor{
		UmlComment: "UmlComment",
	})
	assert.Nil(t, err)

	actor, err = LoadUseCaseActor(dbOrTx, modelKey, useCaseKey, classKey)
	assert.Nil(t, err)

	return actor
}
