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

func TestSubdomainSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(SubdomainSuite))
}

type SubdomainSuite struct {
	suite.Suite
	db            *sql.DB
	model         req_model.Model
	domain        model_domain.Domain
	subdomainKey  identity.Key
	subdomainKeyB identity.Key
}

func (suite *SubdomainSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))

	// Create the subdomain keys for reuse.
	suite.subdomainKey = helper.Must(identity.NewSubdomainKey(suite.domain.Key, "key"))
	suite.subdomainKeyB = helper.Must(identity.NewSubdomainKey(suite.domain.Key, "key_b"))
}

func (suite *SubdomainSuite) TestLoad() {

	// Nothing in database yet.
	domainKey, subdomain, err := LoadSubdomain(suite.db, suite.model.Key, suite.subdomainKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), domainKey)
	assert.Empty(suite.T(), subdomain)

	_, err = dbExec(suite.db, `
		INSERT INTO subdomain
			(
				model_key,
				domain_key,
				subdomain_key,
				name,
				details,
				uml_comment
			)
		VALUES
			(
				'model_key',
				'domain/domain_key',
				'domain/domain_key/subdomain/key',
				'Name',
				'Details',
				'UmlComment'
			)
	`)
	assert.Nil(suite.T(), err)

	domainKey, subdomain, err = LoadSubdomain(suite.db, suite.model.Key, suite.subdomainKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.domain.Key, domainKey)
	assert.Equal(suite.T(), model_domain.Subdomain{
		Key:        suite.subdomainKey,
		Name:       "Name",
		Details:    "Details",
		UmlComment: "UmlComment",
	}, subdomain)
}

func (suite *SubdomainSuite) TestAdd() {

	err := AddSubdomain(suite.db, suite.model.Key, suite.domain.Key, model_domain.Subdomain{
		Key:        suite.subdomainKey,
		Name:       "Name",
		Details:    "Details",
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	domainKey, subdomain, err := LoadSubdomain(suite.db, suite.model.Key, suite.subdomainKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.domain.Key, domainKey)
	assert.Equal(suite.T(), model_domain.Subdomain{
		Key:        suite.subdomainKey,
		Name:       "Name",
		Details:    "Details",
		UmlComment: "UmlComment",
	}, subdomain)
}

func (suite *SubdomainSuite) TestUpdate() {

	err := AddSubdomain(suite.db, suite.model.Key, suite.domain.Key, model_domain.Subdomain{
		Key:        suite.subdomainKey,
		Name:       "Name",
		Details:    "Details",
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateSubdomain(suite.db, suite.model.Key, model_domain.Subdomain{
		Key:        suite.subdomainKey,
		Name:       "NameX",
		Details:    "DetailsX",
		UmlComment: "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	domainKey, subdomain, err := LoadSubdomain(suite.db, suite.model.Key, suite.subdomainKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.domain.Key, domainKey)
	assert.Equal(suite.T(), model_domain.Subdomain{
		Key:        suite.subdomainKey,
		Name:       "NameX",
		Details:    "DetailsX",
		UmlComment: "UmlCommentX",
	}, subdomain)
}

func (suite *SubdomainSuite) TestRemove() {

	err := AddSubdomain(suite.db, suite.model.Key, suite.domain.Key, model_domain.Subdomain{
		Key:        suite.subdomainKey,
		Name:       "Name",
		Details:    "Details",
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = RemoveSubdomain(suite.db, suite.model.Key, suite.subdomainKey)
	assert.Nil(suite.T(), err)

	domainKey, subdomain, err := LoadSubdomain(suite.db, suite.model.Key, suite.subdomainKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), domainKey)
	assert.Empty(suite.T(), subdomain)
}

func (suite *SubdomainSuite) TestQuery() {

	err := AddSubdomain(suite.db, suite.model.Key, suite.domain.Key, model_domain.Subdomain{
		Key:        suite.subdomainKeyB,
		Name:       "NameX",
		Details:    "DetailsX",
		UmlComment: "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	err = AddSubdomain(suite.db, suite.model.Key, suite.domain.Key, model_domain.Subdomain{
		Key:        suite.subdomainKey,
		Name:       "Name",
		Details:    "Details",
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	subdomains, err := QuerySubdomains(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[identity.Key][]model_domain.Subdomain{
		suite.domain.Key: {
			{
				Key:        suite.subdomainKey,
				Name:       "Name",
				Details:    "Details",
				UmlComment: "UmlComment",
			},
			{
				Key:        suite.subdomainKeyB,
				Name:       "NameX",
				Details:    "DetailsX",
				UmlComment: "UmlCommentX",
			},
		},
	}, subdomains)
}

//==================================================
// Test objects for other tests.
//==================================================

func t_AddSubdomain(t *testing.T, dbOrTx DbOrTx, modelKey string, domainKey identity.Key, subdomainKey identity.Key) (subdomain model_domain.Subdomain) {

	err := AddSubdomain(dbOrTx, modelKey, domainKey, model_domain.Subdomain{
		Key:        subdomainKey,
		Name:       subdomainKey.String(),
		Details:    "Details",
		UmlComment: "UmlComment",
	})
	assert.Nil(t, err)

	_, subdomain, err = LoadSubdomain(dbOrTx, modelKey, subdomainKey)
	assert.Nil(t, err)

	return subdomain
}
