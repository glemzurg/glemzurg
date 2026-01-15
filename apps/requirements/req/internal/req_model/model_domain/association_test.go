package model_domain

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestAssociationSuite(t *testing.T) {
	suite.Run(t, new(AssociationSuite))
}

type AssociationSuite struct {
	suite.Suite
	problemDomainKey  identity.Key
	solutionDomainKey identity.Key
}

func (suite *AssociationSuite) SetupTest() {
	suite.problemDomainKey = helper.Must(identity.NewDomainKey("domain1"))
	suite.solutionDomainKey = helper.Must(identity.NewDomainKey("domain2"))
}

func (suite *AssociationSuite) TestNew() {

	tests := []struct {
		testName          string
		key               identity.Key
		problemDomainKey  identity.Key
		solutionDomainKey identity.Key
		umlComment        string
		obj               Association
		errstr            string
	}{
		// OK.
		{
			testName:          "ok with comment",
			key:               helper.Must(identity.NewDomainAssociationKey(suite.problemDomainKey, suite.solutionDomainKey)),
			problemDomainKey:  suite.problemDomainKey,
			solutionDomainKey: suite.solutionDomainKey,
			umlComment:        "UmlComment",
			obj: Association{
				Key:               helper.Must(identity.NewDomainAssociationKey(suite.problemDomainKey, suite.solutionDomainKey)),
				ProblemDomainKey:  suite.problemDomainKey,
				SolutionDomainKey: suite.solutionDomainKey,
				UmlComment:        "UmlComment",
			},
		},
		{
			testName:          "ok minimal",
			key:               helper.Must(identity.NewDomainAssociationKey(suite.problemDomainKey, suite.solutionDomainKey)),
			problemDomainKey:  suite.problemDomainKey,
			solutionDomainKey: suite.solutionDomainKey,
			umlComment:        "",
			obj: Association{
				Key:               helper.Must(identity.NewDomainAssociationKey(suite.problemDomainKey, suite.solutionDomainKey)),
				ProblemDomainKey:  suite.problemDomainKey,
				SolutionDomainKey: suite.solutionDomainKey,
				UmlComment:        "",
			},
		},

		// Error states.
		{
			testName:          "error empty key",
			key:               identity.Key{},
			problemDomainKey:  suite.problemDomainKey,
			solutionDomainKey: suite.solutionDomainKey,
			umlComment:        "UmlComment",
			errstr:            `Key: (keyType: cannot be blank;`,
		},
		{
			testName:          "error wrong key type",
			key:               helper.Must(identity.NewActorKey("actor1")),
			problemDomainKey:  suite.problemDomainKey,
			solutionDomainKey: suite.solutionDomainKey,
			umlComment:        "UmlComment",
			errstr:            `Key: invalid key type 'actor' for domain association.`,
		},
		{
			testName:          "error empty problem key",
			key:               helper.Must(identity.NewDomainAssociationKey(suite.problemDomainKey, suite.solutionDomainKey)),
			problemDomainKey:  identity.Key{},
			solutionDomainKey: suite.solutionDomainKey,
			umlComment:        "UmlComment",
			errstr:            `ProblemDomainKey: (keyType: cannot be blank;`,
		},
		{
			testName:          "error wrong problem key type",
			key:               helper.Must(identity.NewDomainAssociationKey(suite.problemDomainKey, suite.solutionDomainKey)),
			problemDomainKey:  helper.Must(identity.NewActorKey("actor1")),
			solutionDomainKey: suite.solutionDomainKey,
			umlComment:        "UmlComment",
			errstr:            `ProblemDomainKey: invalid key type 'actor' for domain.`,
		},
		{
			testName:          "error empty solution key",
			key:               helper.Must(identity.NewDomainAssociationKey(suite.problemDomainKey, suite.solutionDomainKey)),
			problemDomainKey:  suite.problemDomainKey,
			solutionDomainKey: identity.Key{},
			umlComment:        "UmlComment",
			errstr:            `SolutionDomainKey: (keyType: cannot be blank;`,
		},
		{
			testName:          "error wrong solution key type",
			key:               helper.Must(identity.NewDomainAssociationKey(suite.problemDomainKey, suite.solutionDomainKey)),
			problemDomainKey:  suite.problemDomainKey,
			solutionDomainKey: helper.Must(identity.NewActorKey("actor1")),
			umlComment:        "UmlComment",
			errstr:            `SolutionDomainKey: invalid key type 'actor' for domain.`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			obj, err := NewAssociation(tt.key, tt.problemDomainKey, tt.solutionDomainKey, tt.umlComment)
			if tt.errstr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.obj, obj)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
				assert.Empty(t, obj)
			}
		})
	}
}
