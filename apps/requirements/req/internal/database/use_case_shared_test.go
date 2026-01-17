package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestUseCaseSharedSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(UseCaseSharedSuite))
}

type UseCaseSharedSuite struct {
	suite.Suite
	db          *sql.DB
	model       req_model.Model
	domain      model_domain.Domain
	subdomain   model_domain.Subdomain
	seaUseCase  model_use_case.UseCase
	mudUseCase  model_use_case.UseCase
	mudUseCaseB model_use_case.UseCase
}

func (suite *UseCaseSharedSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_key")))
	suite.seaUseCase = t_AddUseCase(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewUseCaseKey(suite.subdomain.Key, "sea_use_case_key")))
	suite.mudUseCase = t_AddUseCase(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewUseCaseKey(suite.subdomain.Key, "mud_use_case_key")))
	suite.mudUseCaseB = t_AddUseCase(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewUseCaseKey(suite.subdomain.Key, "mud_use_case_key_b")))
}

func (suite *UseCaseSharedSuite) TestLoad() {

	// Nothing in database yet.
	useCaseShared, err := LoadUseCaseShared(suite.db, suite.model.Key, suite.seaUseCase.Key, suite.mudUseCase.Key)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), useCaseShared)

	_, err = dbExec(suite.db, `
		INSERT INTO use_case_shared
			(
				model_key,
				sea_use_case_key,
				mud_use_case_key,
				share_type,
				uml_comment
			)
		VALUES
			(
				'model_key',
				'domain/domain_key/subdomain/subdomain_key/usecase/sea_use_case_key',
				'domain/domain_key/subdomain/subdomain_key/usecase/mud_use_case_key',
				'include',
				'UmlComment'
			)
	`)
	assert.Nil(suite.T(), err)

	useCaseShared, err = LoadUseCaseShared(suite.db, suite.model.Key, suite.seaUseCase.Key, suite.mudUseCase.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_use_case.UseCaseShared{
		ShareType:  "include",
		UmlComment: "UmlComment",
	}, useCaseShared)
}

func (suite *UseCaseSharedSuite) TestAdd() {

	err := AddUseCaseShared(suite.db, suite.model.Key, suite.seaUseCase.Key, suite.mudUseCase.Key, model_use_case.UseCaseShared{
		ShareType:  "include",
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	useCaseShared, err := LoadUseCaseShared(suite.db, suite.model.Key, suite.seaUseCase.Key, suite.mudUseCase.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_use_case.UseCaseShared{
		ShareType:  "include",
		UmlComment: "UmlComment",
	}, useCaseShared)
}

func (suite *UseCaseSharedSuite) TestUpdate() {

	err := AddUseCaseShared(suite.db, suite.model.Key, suite.seaUseCase.Key, suite.mudUseCase.Key, model_use_case.UseCaseShared{
		ShareType:  "include",
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateUseCaseShared(suite.db, suite.model.Key, suite.seaUseCase.Key, suite.mudUseCase.Key, model_use_case.UseCaseShared{
		ShareType:  "extend",
		UmlComment: "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	useCaseShared, err := LoadUseCaseShared(suite.db, suite.model.Key, suite.seaUseCase.Key, suite.mudUseCase.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_use_case.UseCaseShared{
		ShareType:  "extend",
		UmlComment: "UmlCommentX",
	}, useCaseShared)
}

func (suite *UseCaseSharedSuite) TestRemove() {

	err := AddUseCaseShared(suite.db, suite.model.Key, suite.seaUseCase.Key, suite.mudUseCase.Key, model_use_case.UseCaseShared{
		ShareType:  "include",
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = RemoveUseCaseShared(suite.db, suite.model.Key, suite.seaUseCase.Key, suite.mudUseCase.Key)
	assert.Nil(suite.T(), err)

	useCaseShared, err := LoadUseCaseShared(suite.db, suite.model.Key, suite.seaUseCase.Key, suite.mudUseCase.Key)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), useCaseShared)
}

func (suite *UseCaseSharedSuite) TestQuery() {

	err := AddUseCaseShared(suite.db, suite.model.Key, suite.seaUseCase.Key, suite.mudUseCase.Key, model_use_case.UseCaseShared{
		ShareType:  "include",
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = AddUseCaseShared(suite.db, suite.model.Key, suite.seaUseCase.Key, suite.mudUseCaseB.Key, model_use_case.UseCaseShared{
		ShareType:  "extend",
		UmlComment: "UmlCommentB",
	})
	assert.Nil(suite.T(), err)

	useCaseShareds, err := QueryUseCaseShareds(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[identity.Key]map[identity.Key]model_use_case.UseCaseShared{
		suite.seaUseCase.Key: {
			suite.mudUseCase.Key: {
				ShareType:  "include",
				UmlComment: "UmlComment",
			},
			suite.mudUseCaseB.Key: {
				ShareType:  "extend",
				UmlComment: "UmlCommentB",
			},
		},
	}, useCaseShareds)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddUseCaseShared(t *testing.T, dbOrTx DbOrTx, modelKey string, seaUseCaseKey identity.Key, mudUseCaseKey identity.Key, shareType string) (useCaseShared model_use_case.UseCaseShared) {

	err := AddUseCaseShared(dbOrTx, modelKey, seaUseCaseKey, mudUseCaseKey, model_use_case.UseCaseShared{
		ShareType:  shareType,
		UmlComment: "UmlComment",
	})
	assert.Nil(t, err)

	useCaseShared, err = LoadUseCaseShared(dbOrTx, modelKey, seaUseCaseKey, mudUseCaseKey)
	assert.Nil(t, err)

	return useCaseShared
}
