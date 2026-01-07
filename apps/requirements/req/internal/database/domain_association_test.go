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

func TestDomainAssociationSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(DomainAssociationSuite))
}

type DomainAssociationSuite struct {
	suite.Suite
	db      *sql.DB
	model   requirements.Model
	domain  model_domain.Domain
	domainB model_domain.Domain
}

func (suite *DomainAssociationSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key)
	suite.domainB = t_AddDomain(suite.T(), suite.db, suite.model.Key, "domain_key_b")
}

func (suite *DomainAssociationSuite) TestLoad() {

	// Nothing in database yet.
	association, err := LoadDomainAssociation(suite.db, strings.ToUpper(suite.model.Key), "Key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), association)

	_, err = dbExec(suite.db, `
		INSERT INTO domain_association
			(
				model_key,
				association_key,
				problem_domain_key,
				solution_domain_key,
				uml_comment
			)
		VALUES
			(
				'model_key',
				'key',
				'domain_key',
				'domain_key_b',
				'UmlComment'
			)
	`)
	assert.Nil(suite.T(), err)

	association, err = LoadDomainAssociation(suite.db, strings.ToUpper(suite.model.Key), "Key") // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_domain.DomainAssociation{
		Key:               "key", // Test case-insensitive.
		ProblemDomainKey:  "domain_key",
		SolutionDomainKey: "domain_key_b",
		UmlComment:        "UmlComment",
	}, association)
}

func (suite *DomainAssociationSuite) TestAdd() {

	err := AddDomainAssociation(suite.db, strings.ToUpper(suite.model.Key), model_domain.DomainAssociation{
		Key:               "KeY", // Test case-insensitive.
		ProblemDomainKey:  "domain_KEY",
		SolutionDomainKey: "doMAIN_key_b",
		UmlComment:        "UmlComment",
	})
	assert.Nil(suite.T(), err)

	association, err := LoadDomainAssociation(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_domain.DomainAssociation{
		Key:               "key",
		ProblemDomainKey:  "domain_key",
		SolutionDomainKey: "domain_key_b",
		UmlComment:        "UmlComment",
	}, association)
}

func (suite *DomainAssociationSuite) TestUpdate() {

	err := AddDomainAssociation(suite.db, suite.model.Key, model_domain.DomainAssociation{
		Key:               "key",
		ProblemDomainKey:  "domain_key",
		SolutionDomainKey: "domain_key_b",
		UmlComment:        "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = UpdateDomainAssociation(suite.db, strings.ToUpper(suite.model.Key), model_domain.DomainAssociation{
		Key:               "KeY",          // Test case-insensitive.
		ProblemDomainKey:  "doMAIN_key_b", // Test case-insensitive.
		SolutionDomainKey: "domain_KEY",   // Test case-insensitive.
		UmlComment:        "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	association, err := LoadDomainAssociation(suite.db, suite.model.Key, "key")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_domain.DomainAssociation{
		Key:               "key",
		ProblemDomainKey:  "domain_key_b",
		SolutionDomainKey: "domain_key",
		UmlComment:        "UmlCommentX",
	}, association)
}

func (suite *DomainAssociationSuite) TestRemove() {

	err := AddDomainAssociation(suite.db, suite.model.Key, model_domain.DomainAssociation{
		Key:               "key",
		ProblemDomainKey:  "domain_key",
		SolutionDomainKey: "domain_key_b",
		UmlComment:        "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = RemoveDomainAssociation(suite.db, strings.ToUpper(suite.model.Key), strings.ToUpper("key")) // Test case-insensitive.
	assert.Nil(suite.T(), err)

	association, err := LoadDomainAssociation(suite.db, suite.model.Key, "key")
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), association)
}

func (suite *DomainAssociationSuite) TestQuery() {

	err := AddDomainAssociation(suite.db, suite.model.Key, model_domain.DomainAssociation{
		Key:               "keyx",
		ProblemDomainKey:  "domain_key",
		SolutionDomainKey: "domain_key_b",
		UmlComment:        "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	err = AddDomainAssociation(suite.db, suite.model.Key, model_domain.DomainAssociation{
		Key:               "key",
		ProblemDomainKey:  "domain_key",
		SolutionDomainKey: "domain_key_b",
		UmlComment:        "UmlComment",
	})
	assert.Nil(suite.T(), err)

	associations, err := QueryDomainAssociations(suite.db, strings.ToUpper(suite.model.Key)) // Test case-insensitive.
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), []model_domain.DomainAssociation{
		{
			Key:               "key",
			ProblemDomainKey:  "domain_key",
			SolutionDomainKey: "domain_key_b",
			UmlComment:        "UmlComment",
		},
		{
			Key:               "keyx",
			ProblemDomainKey:  "domain_key",
			SolutionDomainKey: "domain_key_b",
			UmlComment:        "UmlCommentX",
		},
	}, associations)
}
