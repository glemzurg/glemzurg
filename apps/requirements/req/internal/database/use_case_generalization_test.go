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

func TestUseCaseGeneralizationSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(UseCaseGeneralizationSuite))
}

type UseCaseGeneralizationSuite struct {
	suite.Suite
	db                 *sql.DB
	model              req_model.Model
	domain             model_domain.Domain
	subdomain          model_domain.Subdomain
	generalizationKey  identity.Key
	generalizationKeyB identity.Key
}

func (suite *UseCaseGeneralizationSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_key")))

	// Create the generalization keys for reuse.
	suite.generalizationKey = helper.Must(identity.NewUseCaseGeneralizationKey(suite.subdomain.Key, "key"))
	suite.generalizationKeyB = helper.Must(identity.NewUseCaseGeneralizationKey(suite.subdomain.Key, "key_b"))
}

func (suite *UseCaseGeneralizationSuite) TestLoad() {

	// Nothing in database yet.
	subdomainKey, generalization, err := LoadUseCaseGeneralization(suite.db, suite.model.Key, suite.generalizationKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), subdomainKey)
	assert.Empty(suite.T(), generalization)

	_, err = dbExec(suite.db, `
		INSERT INTO use_case_generalization
			(
				model_key,
				subdomain_key,
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
				'domain/domain_key/subdomain/subdomain_key',
				'domain/domain_key/subdomain/subdomain_key/ucgeneralization/key',
				'Name',
				'Details',
				true,
				false,
				'UmlComment'
			)
	`)
	assert.Nil(suite.T(), err)

	subdomainKey, generalization, err = LoadUseCaseGeneralization(suite.db, suite.model.Key, suite.generalizationKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.subdomain.Key, subdomainKey)
	assert.Equal(suite.T(), model_use_case.Generalization{
		Key:        suite.generalizationKey,
		Name:       "Name",
		Details:    "Details",
		IsComplete: true,
		IsStatic:   false,
		UmlComment: "UmlComment",
	}, generalization)
}

func (suite *UseCaseGeneralizationSuite) TestAdd() {

	err := AddUseCaseGeneralization(suite.db, suite.model.Key, suite.subdomain.Key, model_use_case.Generalization{
		Key:        suite.generalizationKey,
		Name:       "Name",
		Details:    "Details",
		IsComplete: true,
		IsStatic:   false,
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	subdomainKey, generalization, err := LoadUseCaseGeneralization(suite.db, suite.model.Key, suite.generalizationKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.subdomain.Key, subdomainKey)
	assert.Equal(suite.T(), model_use_case.Generalization{
		Key:        suite.generalizationKey,
		Name:       "Name",
		Details:    "Details",
		IsComplete: true,
		IsStatic:   false,
		UmlComment: "UmlComment",
	}, generalization)
}

func (suite *UseCaseGeneralizationSuite) TestAddNulls() {

	err := AddUseCaseGeneralization(suite.db, suite.model.Key, suite.subdomain.Key, model_use_case.Generalization{
		Key:        suite.generalizationKey,
		Name:       "Name",
		Details:    "",
		IsComplete: false,
		IsStatic:   false,
		UmlComment: "",
	})
	assert.Nil(suite.T(), err)

	subdomainKey, generalization, err := LoadUseCaseGeneralization(suite.db, suite.model.Key, suite.generalizationKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.subdomain.Key, subdomainKey)
	assert.Equal(suite.T(), model_use_case.Generalization{
		Key:        suite.generalizationKey,
		Name:       "Name",
		Details:    "",
		IsComplete: false,
		IsStatic:   false,
		UmlComment: "",
	}, generalization)
}

func (suite *UseCaseGeneralizationSuite) TestUpdate() {

	err := AddUseCaseGeneralization(suite.db, suite.model.Key, suite.subdomain.Key, model_use_case.Generalization{
		Key:        suite.generalizationKey,
		Name:       "Name",
		Details:    "Details",
		IsComplete: true,
		IsStatic:   false,
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateUseCaseGeneralization(suite.db, suite.model.Key, model_use_case.Generalization{
		Key:        suite.generalizationKey,
		Name:       "NameX",
		Details:    "DetailsX",
		IsComplete: false,
		IsStatic:   true,
		UmlComment: "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	subdomainKey, generalization, err := LoadUseCaseGeneralization(suite.db, suite.model.Key, suite.generalizationKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.subdomain.Key, subdomainKey)
	assert.Equal(suite.T(), model_use_case.Generalization{
		Key:        suite.generalizationKey,
		Name:       "NameX",
		Details:    "DetailsX",
		IsComplete: false,
		IsStatic:   true,
		UmlComment: "UmlCommentX",
	}, generalization)
}

func (suite *UseCaseGeneralizationSuite) TestUpdateNulls() {

	err := AddUseCaseGeneralization(suite.db, suite.model.Key, suite.subdomain.Key, model_use_case.Generalization{
		Key:        suite.generalizationKey,
		Name:       "Name",
		Details:    "Details",
		IsComplete: true,
		IsStatic:   true,
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateUseCaseGeneralization(suite.db, suite.model.Key, model_use_case.Generalization{
		Key:        suite.generalizationKey,
		Name:       "NameX",
		Details:    "",
		IsComplete: false,
		IsStatic:   false,
		UmlComment: "",
	})
	assert.Nil(suite.T(), err)

	subdomainKey, generalization, err := LoadUseCaseGeneralization(suite.db, suite.model.Key, suite.generalizationKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.subdomain.Key, subdomainKey)
	assert.Equal(suite.T(), model_use_case.Generalization{
		Key:        suite.generalizationKey,
		Name:       "NameX",
		Details:    "",
		IsComplete: false,
		IsStatic:   false,
		UmlComment: "",
	}, generalization)
}

func (suite *UseCaseGeneralizationSuite) TestRemove() {

	err := AddUseCaseGeneralization(suite.db, suite.model.Key, suite.subdomain.Key, model_use_case.Generalization{
		Key:        suite.generalizationKey,
		Name:       "Name",
		Details:    "Details",
		IsComplete: true,
		IsStatic:   false,
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = RemoveUseCaseGeneralization(suite.db, suite.model.Key, suite.generalizationKey)
	assert.Nil(suite.T(), err)

	subdomainKey, generalization, err := LoadUseCaseGeneralization(suite.db, suite.model.Key, suite.generalizationKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), subdomainKey)
	assert.Empty(suite.T(), generalization)
}

func (suite *UseCaseGeneralizationSuite) TestQuery() {

	err := AddUseCaseGeneralizations(suite.db, suite.model.Key, map[identity.Key][]model_use_case.Generalization{
		suite.subdomain.Key: {
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
		},
	})
	assert.Nil(suite.T(), err)

	generalizations, err := QueryUseCaseGeneralizations(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[identity.Key][]model_use_case.Generalization{
		suite.subdomain.Key: {
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
		},
	}, generalizations)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddUseCaseGeneralization(t *testing.T, dbOrTx DbOrTx, modelKey string, subdomainKey identity.Key, generalizationKey identity.Key) (generalization model_use_case.Generalization) {

	err := AddUseCaseGeneralization(dbOrTx, modelKey, subdomainKey, model_use_case.Generalization{
		Key:        generalizationKey,
		Name:       generalizationKey.String(),
		Details:    "Details",
		IsComplete: true,
		IsStatic:   true,
		UmlComment: "UmlComment",
	})
	assert.Nil(t, err)

	_, generalization, err = LoadUseCaseGeneralization(dbOrTx, modelKey, generalizationKey)
	assert.Nil(t, err)

	return generalization
}
