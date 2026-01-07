package database

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_domain"

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
	db     *sql.DB
	model  requirements.Model
	domain model_domain.Domain
}

func (suite *SubdomainSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key)
}

func (suite *SubdomainSuite) TestLoad() {

	// Nothing in database yet.
	domainKey, subdomain, err := LoadSubdomain(suite.db, strings.ToUpper(suite.model.Key), "Key")
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
				'domain_key',
				'key',
				'Name',
				'Details',
				'UmlComment'
			)
	`)
	assert.Nil(suite.T(), err)

	domainKey, subdomain, err = LoadSubdomain(suite.db, strings.ToUpper(suite.model.Key), "Key") // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), `domain_key`, domainKey)
	assert.Equal(suite.T(), model_domain.Subdomain{
		Key:        "key", // Test case-insensitive.
		Name:       "Name",
		Details:    "Details",
		UmlComment: "UmlComment",
	}, subdomain)
}

func (suite *SubdomainSuite) TestAdd() {

	err := AddSubdomain(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper(suite.domain.Key), model_domain.Subdomain{
		Key:        "KeY", // Test case-insensitive.
		Name:       "Name",
		Details:    "Details",
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	domainKey, subdomain, err := LoadSubdomain(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), `domain_key`, domainKey)
	assert.Equal(suite.T(), model_domain.Subdomain{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		UmlComment: "UmlComment",
	}, subdomain)
}

func (suite *SubdomainSuite) TestUpdate() {

	err := AddSubdomain(suite.db, suite.model.Key, suite.domain.Key, model_domain.Subdomain{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateSubdomain(suite.db, strings.ToUpper(suite.model.Key), model_domain.Subdomain{
		Key:        "kEy", // Test case-insensitive.
		Name:       "NameX",
		Details:    "DetailsX",
		UmlComment: "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	domainKey, subdomain, err := LoadSubdomain(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), `domain_key`, domainKey)
	assert.Equal(suite.T(), model_domain.Subdomain{
		Key:        "key", // Test case-insensitive.
		Name:       "NameX",
		Details:    "DetailsX",
		UmlComment: "UmlCommentX",
	}, subdomain)
}

func (suite *SubdomainSuite) TestRemove() {

	err := AddSubdomain(suite.db, suite.model.Key, suite.domain.Key, model_domain.Subdomain{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = RemoveSubdomain(suite.db, strings.ToUpper(suite.model.Key), "kEy") // Test case-insensitive.
	assert.Nil(suite.T(), err)

	domainKey, subdomain, err := LoadSubdomain(suite.db, suite.model.Key, "key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), domainKey)
	assert.Empty(suite.T(), subdomain)
}

func (suite *SubdomainSuite) TestQuery() {

	err := AddSubdomain(suite.db, suite.model.Key, suite.domain.Key, model_domain.Subdomain{
		Key:        "keyx",
		Name:       "NameX",
		Details:    "DetailsX",
		UmlComment: "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	err = AddSubdomain(suite.db, suite.model.Key, suite.domain.Key, model_domain.Subdomain{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	subdomains, err := QuerySubdomains(suite.db, strings.ToUpper(suite.model.Key)) // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), map[string][]model_domain.Subdomain{
		suite.domain.Key: {
			{
				Key:        "key",
				Name:       "Name",
				Details:    "Details",
				UmlComment: "UmlComment",
			},
			{

				Key:        "keyx",
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

func t_AddSubdomain(t *testing.T, dbOrTx DbOrTx, modelKey, domainKey string) (subdomain model_domain.Subdomain) {

	err := AddSubdomain(dbOrTx, modelKey, domainKey, model_domain.Subdomain{
		Key:        "subdomain_key",
		Name:       "Name",
		Details:    "Details",
		UmlComment: "UmlComment",
	})
	assert.Nil(t, err)

	_, subdomain, err = LoadSubdomain(dbOrTx, "model_key", "subdomain_key")
	assert.Nil(t, err)

	return subdomain
}
