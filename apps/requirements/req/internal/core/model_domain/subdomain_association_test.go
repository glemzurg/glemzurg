package model_domain

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/suite"
)

func TestSubdomainAssociationSuite(t *testing.T) {
	suite.Run(t, new(SubdomainAssociationSuite))
}

type SubdomainAssociationSuite struct {
	suite.Suite
	domainKey            identity.Key
	problemSubdomainKey  identity.Key
	solutionSubdomainKey identity.Key
}

func (suite *SubdomainAssociationSuite) SetupTest() {
	suite.domainKey = helper.Must(identity.NewDomainKey("domain1"))
	suite.problemSubdomainKey = helper.Must(identity.NewSubdomainKey(suite.domainKey, "billing"))
	suite.solutionSubdomainKey = helper.Must(identity.NewSubdomainKey(suite.domainKey, "fulfillment"))
}

func (suite *SubdomainAssociationSuite) TestValidate() {
	validKey := helper.Must(identity.NewSubdomainAssociationKey(suite.domainKey, suite.problemSubdomainKey, suite.solutionSubdomainKey))

	tests := []struct {
		testName    string
		association SubdomainAssociation
		errstr      string
	}{
		{
			testName: "valid association",
			association: SubdomainAssociation{
				Key:                  validKey,
				ProblemSubdomainKey:  suite.problemSubdomainKey,
				SolutionSubdomainKey: suite.solutionSubdomainKey,
			},
		},
		{
			testName: "error empty key",
			association: SubdomainAssociation{
				Key:                  identity.Key{},
				ProblemSubdomainKey:  suite.problemSubdomainKey,
				SolutionSubdomainKey: suite.solutionSubdomainKey,
			},
			errstr: "key type is required",
		},
		{
			testName: "error wrong key type",
			association: SubdomainAssociation{
				Key:                  helper.Must(identity.NewActorKey("actor1")),
				ProblemSubdomainKey:  suite.problemSubdomainKey,
				SolutionSubdomainKey: suite.solutionSubdomainKey,
			},
			errstr: "Key: invalid key type 'actor' for subdomain association",
		},
		{
			testName: "error same problem and solution subdomain",
			association: SubdomainAssociation{
				Key:                  validKey,
				ProblemSubdomainKey:  suite.problemSubdomainKey,
				SolutionSubdomainKey: suite.problemSubdomainKey,
			},
			errstr: "ProblemSubdomainKey and SolutionSubdomainKey cannot be the same",
		},
	}
	for _, tt := range tests {
		suite.Run(tt.testName, func() {
			ctx := coreerr.NewContext("test", "")
			err := tt.association.Validate(ctx)
			if tt.errstr == "" {
				suite.Require().NoError(err)
			} else {
				suite.Require().ErrorContains(err, tt.errstr)
			}
		})
	}
}

func (suite *SubdomainAssociationSuite) TestValidateWithParent() {
	validKey := helper.Must(identity.NewSubdomainAssociationKey(suite.domainKey, suite.problemSubdomainKey, suite.solutionSubdomainKey))
	ctx := coreerr.NewContext("test", "")

	assoc := SubdomainAssociation{
		Key:                  validKey,
		ProblemSubdomainKey:  suite.problemSubdomainKey,
		SolutionSubdomainKey: suite.solutionSubdomainKey,
	}
	err := assoc.ValidateWithParent(ctx, &suite.domainKey)
	suite.Require().NoError(err)

	otherDomainKey := helper.Must(identity.NewDomainKey("other_domain"))
	err = assoc.ValidateWithParent(ctx, &otherDomainKey)
	suite.Require().Error(err)
}

func (suite *SubdomainAssociationSuite) TestValidateReferences() {
	validKey := helper.Must(identity.NewSubdomainAssociationKey(suite.domainKey, suite.problemSubdomainKey, suite.solutionSubdomainKey))
	nonExistentSubdomainKey := helper.Must(identity.NewSubdomainKey(suite.domainKey, "missing"))

	subdomains := map[identity.Key]bool{
		suite.problemSubdomainKey:  true,
		suite.solutionSubdomainKey: true,
	}

	tests := []struct {
		testName    string
		association SubdomainAssociation
		errstr      string
	}{
		{
			testName: "valid references",
			association: SubdomainAssociation{
				Key:                  validKey,
				ProblemSubdomainKey:  suite.problemSubdomainKey,
				SolutionSubdomainKey: suite.solutionSubdomainKey,
			},
		},
		{
			testName: "error missing problem subdomain",
			association: SubdomainAssociation{
				Key:                  validKey,
				ProblemSubdomainKey:  nonExistentSubdomainKey,
				SolutionSubdomainKey: suite.solutionSubdomainKey,
			},
			errstr: "references non-existent problem subdomain",
		},
		{
			testName: "error missing solution subdomain",
			association: SubdomainAssociation{
				Key:                  validKey,
				ProblemSubdomainKey:  suite.problemSubdomainKey,
				SolutionSubdomainKey: nonExistentSubdomainKey,
			},
			errstr: "references non-existent solution subdomain",
		},
	}
	for _, tt := range tests {
		suite.Run(tt.testName, func() {
			ctx := coreerr.NewContext("test", "")
			err := tt.association.ValidateReferences(ctx, suite.domainKey, subdomains)
			if tt.errstr == "" {
				suite.Require().NoError(err)
			} else {
				suite.Require().ErrorContains(err, tt.errstr)
			}
		})
	}
}
