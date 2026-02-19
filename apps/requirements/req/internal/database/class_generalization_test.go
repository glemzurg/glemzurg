package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestGeneralizationSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(GeneralizationSuite))
}

type GeneralizationSuite struct {
	suite.Suite
	db                 *sql.DB
	model              req_model.Model
	domain             model_domain.Domain
	subdomain          model_domain.Subdomain
	generalizationKey  identity.Key
	generalizationKeyB identity.Key
}

func (suite *GeneralizationSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.subdomain = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_key")))

	// Create the generalization keys for reuse.
	suite.generalizationKey = helper.Must(identity.NewGeneralizationKey(suite.subdomain.Key, "key"))
	suite.generalizationKeyB = helper.Must(identity.NewGeneralizationKey(suite.subdomain.Key, "key_b"))
}

func (suite *GeneralizationSuite) TestLoad() {

	// Nothing in database yet.
	generalization, err := LoadGeneralization(suite.db, suite.model.Key, suite.generalizationKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), generalization)

	_, err = dbExec(suite.db, `
		INSERT INTO class_generalization
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
				'domain/domain_key/subdomain/subdomain_key/generalization/key',
				'Name',
				'Details',
				true,
				false,
				'UmlComment'
			)
	`)
	assert.Nil(suite.T(), err)

	generalization, err = LoadGeneralization(suite.db, suite.model.Key, suite.generalizationKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_class.Generalization{
		Key:        suite.generalizationKey,
		Name:       "Name",
		Details:    "Details",
		IsComplete: true,
		IsStatic:   false,
		UmlComment: "UmlComment",
	}, generalization)
}

func (suite *GeneralizationSuite) TestAdd() {

	err := AddGeneralization(suite.db, suite.model.Key, model_class.Generalization{
		Key:        suite.generalizationKey,
		Name:       "Name",
		Details:    "Details",
		IsComplete: true,
		IsStatic:   false,
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	generalization, err := LoadGeneralization(suite.db, suite.model.Key, suite.generalizationKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_class.Generalization{
		Key:        suite.generalizationKey,
		Name:       "Name",
		Details:    "Details",
		IsComplete: true,
		IsStatic:   false,
		UmlComment: "UmlComment",
	}, generalization)
}

func (suite *GeneralizationSuite) TestUpdate() {

	err := AddGeneralization(suite.db, suite.model.Key, model_class.Generalization{
		Key:        suite.generalizationKey,
		Name:       "Name",
		Details:    "Details",
		IsComplete: true,
		IsStatic:   false,
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateGeneralization(suite.db, suite.model.Key, model_class.Generalization{
		Key:        suite.generalizationKey,
		Name:       "NameX",
		Details:    "DetailsX",
		IsComplete: false,
		IsStatic:   true,
		UmlComment: "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	generalization, err := LoadGeneralization(suite.db, suite.model.Key, suite.generalizationKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_class.Generalization{
		Key:        suite.generalizationKey,
		Name:       "NameX",
		Details:    "DetailsX",
		IsComplete: false,
		IsStatic:   true,
		UmlComment: "UmlCommentX",
	}, generalization)
}

func (suite *GeneralizationSuite) TestRemove() {

	err := AddGeneralization(suite.db, suite.model.Key, model_class.Generalization{
		Key:        suite.generalizationKey,
		Name:       "Name",
		Details:    "Details",
		IsComplete: true,
		IsStatic:   false,
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = RemoveGeneralization(suite.db, suite.model.Key, suite.generalizationKey)
	assert.Nil(suite.T(), err)

	generalization, err := LoadGeneralization(suite.db, suite.model.Key, suite.generalizationKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), generalization)
}

func (suite *GeneralizationSuite) TestQuery() {

	err := AddGeneralizations(suite.db, suite.model.Key, []model_class.Generalization{
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

	generalizations, err := QueryGeneralizations(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), []model_class.Generalization{
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

func t_AddGeneralization(t *testing.T, dbOrTx DbOrTx, modelKey string, generalizationKey identity.Key) (generalization model_class.Generalization) {

	err := AddGeneralization(dbOrTx, modelKey, model_class.Generalization{
		Key:        generalizationKey,
		Name:       generalizationKey.String(),
		Details:    "Details",
		IsComplete: true,
		IsStatic:   true,
		UmlComment: "UmlComment",
	})
	assert.Nil(t, err)

	generalization, err = LoadGeneralization(dbOrTx, modelKey, generalizationKey)
	assert.Nil(t, err)

	return generalization
}
