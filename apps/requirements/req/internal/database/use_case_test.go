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

func TestUseCaseSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(UseCaseSuite))
}

type UseCaseSuite struct {
	suite.Suite
	db               *sql.DB
	model            req_model.Model
	domain           model_domain.Domain
	subdomain        model_domain.Subdomain
	generalization   model_use_case.Generalization
	generalizationB  model_use_case.Generalization
	useCaseKey       identity.Key
	useCaseKeyB      identity.Key
}

func (suite *UseCaseSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_key")))
	suite.generalization = t_AddUseCaseGeneralization(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewUseCaseGeneralizationKey(suite.subdomain.Key, "generalization_key")))
	suite.generalizationB = t_AddUseCaseGeneralization(suite.T(), suite.db, suite.model.Key, suite.subdomain.Key, helper.Must(identity.NewUseCaseGeneralizationKey(suite.subdomain.Key, "generalization_key_b")))

	// Create the use case keys for reuse.
	suite.useCaseKey = helper.Must(identity.NewUseCaseKey(suite.subdomain.Key, "key"))
	suite.useCaseKeyB = helper.Must(identity.NewUseCaseKey(suite.subdomain.Key, "key_b"))
}

func (suite *UseCaseSuite) TestLoad() {

	// Nothing in database yet.
	subdomainKey, useCase, err := LoadUseCase(suite.db, suite.model.Key, suite.useCaseKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), subdomainKey)
	assert.Empty(suite.T(), useCase)

	_, err = dbExec(suite.db, `
		INSERT INTO use_case
			(
				model_key,
				subdomain_key,
				use_case_key,
				name,
				details,
				level,
				read_only,
				superclass_of_key,
				subclass_of_key,
				uml_comment
			)
		VALUES
			(
				'model_key',
				'domain/domain_key/subdomain/subdomain_key',
				'domain/domain_key/subdomain/subdomain_key/usecase/key',
				'Name',
				'Details',
				'sea',
				true,
				'domain/domain_key/subdomain/subdomain_key/ucgeneralization/generalization_key',
				'domain/domain_key/subdomain/subdomain_key/ucgeneralization/generalization_key_b',
				'UmlComment'
			)
	`)
	assert.Nil(suite.T(), err)

	subdomainKey, useCase, err = LoadUseCase(suite.db, suite.model.Key, suite.useCaseKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.subdomain.Key, subdomainKey)
	assert.Equal(suite.T(), model_use_case.UseCase{
		Key:             suite.useCaseKey,
		Name:            "Name",
		Details:         "Details",
		Level:           "sea",
		ReadOnly:        true,
		SuperclassOfKey: &suite.generalization.Key,
		SubclassOfKey:   &suite.generalizationB.Key,
		UmlComment:      "UmlComment",
	}, useCase)
}

func (suite *UseCaseSuite) TestAdd() {

	err := AddUseCase(suite.db, suite.model.Key, suite.subdomain.Key, model_use_case.UseCase{
		Key:             suite.useCaseKey,
		Name:            "Name",
		Details:         "Details",
		Level:           "mud",
		ReadOnly:        true,
		SuperclassOfKey: &suite.generalization.Key,
		SubclassOfKey:   &suite.generalizationB.Key,
		UmlComment:      "UmlComment",
	})
	assert.Nil(suite.T(), err)

	subdomainKey, useCase, err := LoadUseCase(suite.db, suite.model.Key, suite.useCaseKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.subdomain.Key, subdomainKey)
	assert.Equal(suite.T(), model_use_case.UseCase{
		Key:             suite.useCaseKey,
		Name:            "Name",
		Details:         "Details",
		Level:           "mud",
		ReadOnly:        true,
		SuperclassOfKey: &suite.generalization.Key,
		SubclassOfKey:   &suite.generalizationB.Key,
		UmlComment:      "UmlComment",
	}, useCase)
}

func (suite *UseCaseSuite) TestAddNulls() {

	err := AddUseCase(suite.db, suite.model.Key, suite.subdomain.Key, model_use_case.UseCase{
		Key:             suite.useCaseKey,
		Name:            "Name",
		Details:         "",
		Level:           "sea",
		ReadOnly:        false,
		SuperclassOfKey: nil,
		SubclassOfKey:   nil,
		UmlComment:      "",
	})
	assert.Nil(suite.T(), err)

	subdomainKey, useCase, err := LoadUseCase(suite.db, suite.model.Key, suite.useCaseKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.subdomain.Key, subdomainKey)
	assert.Equal(suite.T(), model_use_case.UseCase{
		Key:             suite.useCaseKey,
		Name:            "Name",
		Details:         "",
		Level:           "sea",
		ReadOnly:        false,
		SuperclassOfKey: nil,
		SubclassOfKey:   nil,
		UmlComment:      "",
	}, useCase)
}

func (suite *UseCaseSuite) TestUpdate() {

	err := AddUseCase(suite.db, suite.model.Key, suite.subdomain.Key, model_use_case.UseCase{
		Key:             suite.useCaseKey,
		Name:            "Name",
		Details:         "Details",
		Level:           "sea",
		ReadOnly:        true,
		SuperclassOfKey: &suite.generalization.Key,
		SubclassOfKey:   &suite.generalizationB.Key,
		UmlComment:      "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateUseCase(suite.db, suite.model.Key, model_use_case.UseCase{
		Key:             suite.useCaseKey,
		Name:            "NameX",
		Details:         "DetailsX",
		Level:           "sky",
		ReadOnly:        false,
		SuperclassOfKey: &suite.generalizationB.Key,
		SubclassOfKey:   &suite.generalization.Key,
		UmlComment:      "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	subdomainKey, useCase, err := LoadUseCase(suite.db, suite.model.Key, suite.useCaseKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.subdomain.Key, subdomainKey)
	assert.Equal(suite.T(), model_use_case.UseCase{
		Key:             suite.useCaseKey,
		Name:            "NameX",
		Details:         "DetailsX",
		Level:           "sky",
		ReadOnly:        false,
		SuperclassOfKey: &suite.generalizationB.Key,
		SubclassOfKey:   &suite.generalization.Key,
		UmlComment:      "UmlCommentX",
	}, useCase)
}

func (suite *UseCaseSuite) TestUpdateNulls() {

	err := AddUseCase(suite.db, suite.model.Key, suite.subdomain.Key, model_use_case.UseCase{
		Key:             suite.useCaseKey,
		Name:            "Name",
		Details:         "Details",
		Level:           "sea",
		ReadOnly:        true,
		SuperclassOfKey: &suite.generalization.Key,
		SubclassOfKey:   &suite.generalizationB.Key,
		UmlComment:      "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateUseCase(suite.db, suite.model.Key, model_use_case.UseCase{
		Key:             suite.useCaseKey,
		Name:            "NameX",
		Details:         "",
		Level:           "mud",
		ReadOnly:        false,
		SuperclassOfKey: nil,
		SubclassOfKey:   nil,
		UmlComment:      "",
	})
	assert.Nil(suite.T(), err)

	subdomainKey, useCase, err := LoadUseCase(suite.db, suite.model.Key, suite.useCaseKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.subdomain.Key, subdomainKey)
	assert.Equal(suite.T(), model_use_case.UseCase{
		Key:             suite.useCaseKey,
		Name:            "NameX",
		Details:         "",
		Level:           "mud",
		ReadOnly:        false,
		SuperclassOfKey: nil,
		SubclassOfKey:   nil,
		UmlComment:      "",
	}, useCase)
}

func (suite *UseCaseSuite) TestRemove() {

	err := AddUseCase(suite.db, suite.model.Key, suite.subdomain.Key, model_use_case.UseCase{
		Key:        suite.useCaseKey,
		Name:       "Name",
		Details:    "Details",
		Level:      "sea",
		ReadOnly:   true,
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = RemoveUseCase(suite.db, suite.model.Key, suite.useCaseKey)
	assert.Nil(suite.T(), err)

	subdomainKey, useCase, err := LoadUseCase(suite.db, suite.model.Key, suite.useCaseKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), subdomainKey)
	assert.Empty(suite.T(), useCase)
}

func (suite *UseCaseSuite) TestQuery() {

	err := AddUseCases(suite.db, suite.model.Key, map[identity.Key]identity.Key{
		suite.useCaseKeyB: suite.subdomain.Key,
		suite.useCaseKey:  suite.subdomain.Key,
	}, []model_use_case.UseCase{
		{
			Key:             suite.useCaseKeyB,
			Name:            "NameX",
			Details:         "DetailsX",
			Level:           "sea",
			ReadOnly:        true,
			SuperclassOfKey: &suite.generalizationB.Key,
			SubclassOfKey:   &suite.generalization.Key,
			UmlComment:      "UmlCommentX",
		},
		{
			Key:             suite.useCaseKey,
			Name:            "Name",
			Details:         "Details",
			Level:           "sea",
			ReadOnly:        true,
			SuperclassOfKey: &suite.generalization.Key,
			SubclassOfKey:   &suite.generalizationB.Key,
			UmlComment:      "UmlComment",
		},
	})
	assert.Nil(suite.T(), err)

	subdomainKeys, useCases, err := QueryUseCases(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[identity.Key]identity.Key{
		suite.useCaseKeyB: suite.subdomain.Key,
		suite.useCaseKey:  suite.subdomain.Key,
	}, subdomainKeys)
	assert.Equal(suite.T(), []model_use_case.UseCase{
		{
			Key:             suite.useCaseKey,
			Name:            "Name",
			Details:         "Details",
			Level:           "sea",
			ReadOnly:        true,
			SuperclassOfKey: &suite.generalization.Key,
			SubclassOfKey:   &suite.generalizationB.Key,
			UmlComment:      "UmlComment",
		},
		{
			Key:             suite.useCaseKeyB,
			Name:            "NameX",
			Details:         "DetailsX",
			Level:           "sea",
			ReadOnly:        true,
			SuperclassOfKey: &suite.generalizationB.Key,
			SubclassOfKey:   &suite.generalization.Key,
			UmlComment:      "UmlCommentX",
		},
	}, useCases)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddUseCase(t *testing.T, dbOrTx DbOrTx, modelKey string, subdomainKey identity.Key, useCaseKey identity.Key) (useCase model_use_case.UseCase) {

	err := AddUseCase(dbOrTx, modelKey, subdomainKey, model_use_case.UseCase{
		Key:        useCaseKey,
		Name:       useCaseKey.String(),
		Details:    "Details",
		Level:      "sea",
		ReadOnly:   true,
		UmlComment: "UmlComment",
	})
	assert.Nil(t, err)

	_, useCase, err = LoadUseCase(dbOrTx, modelKey, useCaseKey)
	assert.Nil(t, err)

	return useCase
}
