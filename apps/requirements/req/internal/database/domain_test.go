package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestDomainSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(DomainSuite))
}

type DomainSuite struct {
	suite.Suite
	db         *sql.DB
	model      req_model.Model
	domainKey  identity.Key
	domainKeyB identity.Key
}

func (suite *DomainSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)

	// Create the domain keys for reuse.
	suite.domainKey = helper.Must(identity.NewDomainKey("key"))
	suite.domainKeyB = helper.Must(identity.NewDomainKey("key_b"))
}

func (suite *DomainSuite) TestLoad() {

	// Nothing in database yet.
	domain, err := LoadDomain(suite.db, suite.model.Key, suite.domainKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), domain)

	_, err = dbExec(suite.db, `
		INSERT INTO domain
			(
				model_key,
				domain_key,
				name,
				details,
				realized,
				uml_comment
			)
		VALUES
			(
				'model_key',
				'domain/key',
				'Name',
				'Details',
				true,
				'UmlComment'
			)
	`)
	assert.Nil(suite.T(), err)

	domain, err = LoadDomain(suite.db, suite.model.Key, suite.domainKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_domain.Domain{
		Key:        suite.domainKey,
		Name:       "Name",
		Details:    "Details",
		Realized:   true,
		UmlComment: "UmlComment",
	}, domain)
}

func (suite *DomainSuite) TestAdd() {

	err := AddDomain(suite.db, suite.model.Key, model_domain.Domain{
		Key:        suite.domainKey,
		Name:       "Name",
		Details:    "Details",
		Realized:   true,
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	domain, err := LoadDomain(suite.db, suite.model.Key, suite.domainKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_domain.Domain{
		Key:        suite.domainKey,
		Name:       "Name",
		Details:    "Details",
		Realized:   true,
		UmlComment: "UmlComment",
	}, domain)
}

func (suite *DomainSuite) TestUpdate() {

	err := AddDomain(suite.db, suite.model.Key, model_domain.Domain{
		Key:        suite.domainKey,
		Name:       "Name",
		Details:    "Details",
		Realized:   true,
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateDomain(suite.db, suite.model.Key, model_domain.Domain{
		Key:        suite.domainKey,
		Name:       "NameX",
		Details:    "DetailsX",
		Realized:   false,
		UmlComment: "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	domain, err := LoadDomain(suite.db, suite.model.Key, suite.domainKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_domain.Domain{
		Key:        suite.domainKey,
		Name:       "NameX",
		Details:    "DetailsX",
		Realized:   false,
		UmlComment: "UmlCommentX",
	}, domain)
}

func (suite *DomainSuite) TestRemove() {

	err := AddDomain(suite.db, suite.model.Key, model_domain.Domain{
		Key:        suite.domainKey,
		Name:       "Name",
		Details:    "Details",
		Realized:   true,
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = RemoveDomain(suite.db, suite.model.Key, suite.domainKey)
	assert.Nil(suite.T(), err)

	domain, err := LoadDomain(suite.db, suite.model.Key, suite.domainKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), domain)
}

func (suite *DomainSuite) TestQuery() {

	err := AddDomains(suite.db, suite.model.Key, []model_domain.Domain{
		{
			Key:        suite.domainKeyB,
			Name:       "NameX",
			Details:    "DetailsX",
			Realized:   false,
			UmlComment: "UmlCommentX",
		},
		{
			Key:        suite.domainKey,
			Name:       "Name",
			Details:    "Details",
			Realized:   true,
			UmlComment: "UmlComment",
		},
	})
	assert.Nil(suite.T(), err)

	domains, err := QueryDomains(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), []model_domain.Domain{
		{
			Key:        suite.domainKey,
			Name:       "Name",
			Details:    "Details",
			Realized:   true,
			UmlComment: "UmlComment",
		},
		{
			Key:        suite.domainKeyB,
			Name:       "NameX",
			Details:    "DetailsX",
			Realized:   false,
			UmlComment: "UmlCommentX",
		},
	}, domains)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddDomain(t *testing.T, dbOrTx DbOrTx, modelKey string, domainKey identity.Key) (domain model_domain.Domain) {

	err := AddDomain(dbOrTx, modelKey, model_domain.Domain{
		Key:        domainKey,
		Name:       domainKey.String(),
		Details:    "Details",
		Realized:   false,
		UmlComment: "UmlComment",
	})
	assert.Nil(t, err)

	domain, err = LoadDomain(dbOrTx, modelKey, domainKey)
	assert.Nil(t, err)

	return domain
}
