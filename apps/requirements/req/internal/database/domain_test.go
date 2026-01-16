package database

import (
	"database/sql"
	"strings"
	"testing"

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
	db    *sql.DB
	model req_model.Model
}

func (suite *DomainSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
}

func (suite *DomainSuite) TestLoad() {

	// Nothing in database yet.
	domain, err := LoadDomain(suite.db, strings.ToUpper(suite.model.Key), "Key")
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
				'key',
				'Name',
				'Details',
				true,
				'UmlComment'
			)
	`)
	assert.Nil(suite.T(), err)

	domain, err = LoadDomain(suite.db, strings.ToUpper(suite.model.Key), "Key") // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_domain.Domain{
		Key:        "key", // Test case-insensitive.
		Name:       "Name",
		Details:    "Details",
		Realized:   true,
		UmlComment: "UmlComment",
	}, domain)
}

func (suite *DomainSuite) TestAdd() {

	err := AddDomain(suite.db, strings.ToUpper(suite.model.Key), model_domain.Domain{
		Key:        "KeY", // Test case-insensitive.
		Name:       "Name",
		Details:    "Details",
		Realized:   true,
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	domain, err := LoadDomain(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_domain.Domain{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		Realized:   true,
		UmlComment: "UmlComment",
	}, domain)
}

func (suite *DomainSuite) TestUpdate() {

	err := AddDomain(suite.db, suite.model.Key, model_domain.Domain{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		Realized:   true,
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateDomain(suite.db, strings.ToUpper(suite.model.Key), model_domain.Domain{
		Key:        "kEy", // Test case-insensitive.
		Name:       "NameX",
		Details:    "DetailsX",
		Realized:   false,
		UmlComment: "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	domain, err := LoadDomain(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_domain.Domain{
		Key:        "key", // Test case-insensitive.
		Name:       "NameX",
		Details:    "DetailsX",
		Realized:   false,
		UmlComment: "UmlCommentX",
	}, domain)
}

func (suite *DomainSuite) TestRemove() {

	err := AddDomain(suite.db, suite.model.Key, model_domain.Domain{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		Realized:   true,
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = RemoveDomain(suite.db, strings.ToUpper(suite.model.Key), "kEy") // Test case-insensitive.
	assert.Nil(suite.T(), err)

	domain, err := LoadDomain(suite.db, suite.model.Key, "key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), domain)
}

func (suite *DomainSuite) TestQuery() {

	err := AddDomain(suite.db, suite.model.Key, model_domain.Domain{
		Key:        "keyx",
		Name:       "NameX",
		Details:    "DetailsX",
		Realized:   false,
		UmlComment: "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	err = AddDomain(suite.db, suite.model.Key, model_domain.Domain{
		Key:        "key",
		Name:       "Name",
		Details:    "Details",
		Realized:   true,
		UmlComment: "UmlComment",
	})
	assert.Nil(suite.T(), err)

	domains, err := QueryDomains(suite.db, strings.ToUpper(suite.model.Key)) // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), []model_domain.Domain{
		{
			Key:        "key",
			Name:       "Name",
			Details:    "Details",
			Realized:   true,
			UmlComment: "UmlComment",
		},
		{

			Key:        "keyx",
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

func t_AddDomain(t *testing.T, dbOrTx DbOrTx, modelKey string, opts ...string) (domain model_domain.Domain) {

	// If there is an optional parameter it is the key.
	key := "domain_key"
	if len(opts) > 0 {
		key = opts[0]
	}

	err := AddDomain(dbOrTx, modelKey, model_domain.Domain{
		Key:        key,
		Name:       "Name",
		Details:    "Details",
		Realized:   false,
		UmlComment: "UmlComment",
	})
	assert.Nil(t, err)

	domain, err = LoadDomain(dbOrTx, modelKey, "domain_key")
	assert.Nil(t, err)

	return domain
}
