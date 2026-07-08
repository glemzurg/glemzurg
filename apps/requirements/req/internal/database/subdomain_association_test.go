package database

import (
	"database/sql"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/stretchr/testify/suite"
)

func TestSubdomainAssociationSuite(t *testing.T) {
	if !*_runDatabaseTests {
		t.Skip("Skipping database test; run `go test ./internal/database/... -dbtests`")
	}
	suite.Run(t, new(SubdomainAssociationSuite))
}

type SubdomainAssociationSuite struct {
	suite.Suite
	db              *sql.DB
	model           core.Model
	domain          model_domain.Domain
	subdomainA      model_domain.Subdomain
	subdomainB      model_domain.Subdomain
	associationKey  identity.Key
	associationKeyB identity.Key
}

func (suite *SubdomainAssociationSuite) SetupTest() {
	suite.db = t_ResetDatabase(suite.T())

	suite.model = t_AddModel(suite.T(), suite.db)
	suite.domain = t_AddDomain(suite.T(), suite.db, suite.model.Key, helper.Must(identity.NewDomainKey("domain_key")))
	subdomainAKey := helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_a"))
	subdomainBKey := helper.Must(identity.NewSubdomainKey(suite.domain.Key, "subdomain_b"))
	suite.subdomainA = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, subdomainAKey)
	suite.subdomainB = t_AddSubdomain(suite.T(), suite.db, suite.model.Key, suite.domain.Key, subdomainBKey)

	suite.associationKey = helper.Must(identity.NewSubdomainAssociationKey(suite.domain.Key, suite.subdomainA.Key, suite.subdomainB.Key))
	suite.associationKeyB = helper.Must(identity.NewSubdomainAssociationKey(suite.domain.Key, suite.subdomainB.Key, suite.subdomainA.Key))
}

func (suite *SubdomainAssociationSuite) TestAddAndQuery() {
	err := AddSubdomainAssociations(suite.db, suite.model.Key, []model_domain.SubdomainAssociation{
		{
			Key:                  suite.associationKey,
			ProblemSubdomainKey:  suite.subdomainA.Key,
			SolutionSubdomainKey: suite.subdomainB.Key,
			UmlComment:           "UmlComment",
		},
		{
			Key:                  suite.associationKeyB,
			ProblemSubdomainKey:  suite.subdomainB.Key,
			SolutionSubdomainKey: suite.subdomainA.Key,
			UmlComment:           "UmlCommentX",
		},
	})
	suite.Require().NoError(err)

	associations, err := QuerySubdomainAssociations(suite.db, suite.model.Key)
	suite.Require().NoError(err)
	suite.Equal([]model_domain.SubdomainAssociation{
		{
			Key:                  suite.associationKey,
			ProblemSubdomainKey:  suite.subdomainA.Key,
			SolutionSubdomainKey: suite.subdomainB.Key,
			UmlComment:           "UmlComment",
		},
		{
			Key:                  suite.associationKeyB,
			ProblemSubdomainKey:  suite.subdomainB.Key,
			SolutionSubdomainKey: suite.subdomainA.Key,
			UmlComment:           "UmlCommentX",
		},
	}, associations)
}

func (suite *SubdomainAssociationSuite) TestLoadUpdateRemove() {
	err := AddSubdomainAssociation(suite.db, suite.model.Key, model_domain.SubdomainAssociation{
		Key:                  suite.associationKey,
		ProblemSubdomainKey:  suite.subdomainA.Key,
		SolutionSubdomainKey: suite.subdomainB.Key,
		UmlComment:           "UmlComment",
	})
	suite.Require().NoError(err)

	association, err := LoadSubdomainAssociation(suite.db, suite.model.Key, suite.associationKey)
	suite.Require().NoError(err)
	suite.Equal("UmlComment", association.UmlComment)

	err = UpdateSubdomainAssociation(suite.db, suite.model.Key, model_domain.SubdomainAssociation{
		Key:                  suite.associationKey,
		ProblemSubdomainKey:  suite.subdomainB.Key,
		SolutionSubdomainKey: suite.subdomainA.Key,
		UmlComment:           "Updated",
	})
	suite.Require().NoError(err)

	association, err = LoadSubdomainAssociation(suite.db, suite.model.Key, suite.associationKey)
	suite.Require().NoError(err)
	suite.Equal("Updated", association.UmlComment)
	suite.Equal(suite.subdomainB.Key, association.ProblemSubdomainKey)

	err = RemoveSubdomainAssociation(suite.db, suite.model.Key, suite.associationKey)
	suite.Require().NoError(err)

	_, err = LoadSubdomainAssociation(suite.db, suite.model.Key, suite.associationKey)
	suite.Require().ErrorIs(err, ErrNotFound)
}
