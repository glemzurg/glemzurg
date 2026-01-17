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

func TestDomainAssociationSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(DomainAssociationSuite))
}

type DomainAssociationSuite struct {
	suite.Suite
	db              *sql.DB
	model           req_model.Model
	domain          model_domain.Domain
	domainB         model_domain.Domain
	associationKey  identity.Key
	associationKeyB identity.Key
}

func (suite *DomainAssociationSuite) SetupTest() {

	// Clear the database.
	suite.db = t_ResetDatabase(suite.T())

	// Add any objects needed for tests.
	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	suite.domainB = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key_b")))

	// Create the association key for reuse.
	suite.associationKey = helper.Must(identity.NewDomainAssociationKey(suite.domain.Key, suite.domainB.Key))
	suite.associationKeyB = helper.Must(identity.NewDomainAssociationKey(suite.domainB.Key, suite.domain.Key))
}

func (suite *DomainAssociationSuite) TestLoad() {

	// Nothing in database yet.
	association, err := LoadDomainAssociation(suite.db, suite.model.Key, suite.associationKey)
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
				'domain/domain_key/dassociation/domain_key_b',
				'domain/domain_key',
				'domain/domain_key_b',
				'UmlComment'
			)
	`)
	assert.Nil(suite.T(), err)

	association, err = LoadDomainAssociation(suite.db, suite.model.Key, suite.associationKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_domain.Association{
		Key:               suite.associationKey,
		ProblemDomainKey:  suite.domain.Key,
		SolutionDomainKey: suite.domainB.Key,
		UmlComment:        "UmlComment",
	}, association)
}

func (suite *DomainAssociationSuite) TestAdd() {

	err := AddDomainAssociation(suite.db, suite.model.Key, model_domain.Association{
		Key:               suite.associationKey,
		ProblemDomainKey:  suite.domain.Key,
		SolutionDomainKey: suite.domainB.Key,
		UmlComment:        "UmlComment",
	})
	assert.Nil(suite.T(), err)

	association, err := LoadDomainAssociation(suite.db, suite.model.Key, suite.associationKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_domain.Association{
		Key:               suite.associationKey,
		ProblemDomainKey:  suite.domain.Key,
		SolutionDomainKey: suite.domainB.Key,
		UmlComment:        "UmlComment",
	}, association)
}

func (suite *DomainAssociationSuite) TestUpdate() {

	err := AddDomainAssociation(suite.db, suite.model.Key, model_domain.Association{
		Key:               suite.associationKey,
		ProblemDomainKey:  suite.domain.Key,
		SolutionDomainKey: suite.domainB.Key,
		UmlComment:        "UmlComment",
	})
	assert.Nil(suite.T(), err)

	// Update swaps problem and solution domains.
	err = UpdateDomainAssociation(suite.db, suite.model.Key, model_domain.Association{
		Key:               suite.associationKey,
		ProblemDomainKey:  suite.domainB.Key,
		SolutionDomainKey: suite.domain.Key,
		UmlComment:        "UmlCommentX",
	})
	assert.Nil(suite.T(), err)

	association, err := LoadDomainAssociation(suite.db, suite.model.Key, suite.associationKey)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), model_domain.Association{
		Key:               suite.associationKey,
		ProblemDomainKey:  suite.domainB.Key,
		SolutionDomainKey: suite.domain.Key,
		UmlComment:        "UmlCommentX",
	}, association)
}

func (suite *DomainAssociationSuite) TestRemove() {

	err := AddDomainAssociation(suite.db, suite.model.Key, model_domain.Association{
		Key:               suite.associationKey,
		ProblemDomainKey:  suite.domain.Key,
		SolutionDomainKey: suite.domainB.Key,
		UmlComment:        "UmlComment",
	})
	assert.Nil(suite.T(), err)

	err = RemoveDomainAssociation(suite.db, suite.model.Key, suite.associationKey)
	assert.Nil(suite.T(), err)

	association, err := LoadDomainAssociation(suite.db, suite.model.Key, suite.associationKey)
	assert.ErrorIs(suite.T(), err, ErrNotFound)
	assert.Empty(suite.T(), association)
}

func (suite *DomainAssociationSuite) TestQuery() {

	err := AddDomainAssociations(suite.db, suite.model.Key, []model_domain.Association{
		{
			Key:               suite.associationKeyB,
			ProblemDomainKey:  suite.domainB.Key,
			SolutionDomainKey: suite.domain.Key,
			UmlComment:        "UmlCommentX",
		},
		{
			Key:               suite.associationKey,
			ProblemDomainKey:  suite.domain.Key,
			SolutionDomainKey: suite.domainB.Key,
			UmlComment:        "UmlComment",
		},
	})
	assert.Nil(suite.T(), err)

	associations, err := QueryDomainAssociations(suite.db, suite.model.Key)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), []model_domain.Association{
		{
			Key:               suite.associationKey,
			ProblemDomainKey:  suite.domain.Key,
			SolutionDomainKey: suite.domainB.Key,
			UmlComment:        "UmlComment",
		},
		{
			Key:               suite.associationKeyB,
			ProblemDomainKey:  suite.domainB.Key,
			SolutionDomainKey: suite.domain.Key,
			UmlComment:        "UmlCommentX",
		},
	}, associations)
}
