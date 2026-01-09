package model_domain

import (
	"fmt"
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
		key               identity.Key
		problemDomainKey  identity.Key
		solutionDomainKey identity.Key
		umlComment        string
		obj               Association
		errstr            string
	}{
		// OK.
		{
			key:               helper.Must(identity.NewDomainAssociationKey(suite.problemDomainKey, "1")),
			problemDomainKey:  suite.problemDomainKey,
			solutionDomainKey: suite.solutionDomainKey,
			umlComment:        "UmlComment",
			obj: Association{
				Key:               helper.Must(identity.NewDomainAssociationKey(suite.problemDomainKey, "1")),
				ProblemDomainKey:  suite.problemDomainKey,
				SolutionDomainKey: suite.solutionDomainKey,
				UmlComment:        "UmlComment",
			},
		},
		{
			key:               helper.Must(identity.NewDomainAssociationKey(suite.problemDomainKey, "2")),
			problemDomainKey:  suite.problemDomainKey,
			solutionDomainKey: suite.solutionDomainKey,
			umlComment:        "",
			obj: Association{
				Key:               helper.Must(identity.NewDomainAssociationKey(suite.problemDomainKey, "2")),
				ProblemDomainKey:  suite.problemDomainKey,
				SolutionDomainKey: suite.solutionDomainKey,
				UmlComment:        "",
			},
		},

		// Error states.
		{
			key:               identity.Key{},
			problemDomainKey:  suite.problemDomainKey,
			solutionDomainKey: suite.solutionDomainKey,
			umlComment:        "UmlComment",
			errstr:            `Key: (keyType: cannot be blank;`,
		},
		{
			key:               helper.Must(identity.NewActorKey("actor1")),
			problemDomainKey:  suite.problemDomainKey,
			solutionDomainKey: suite.solutionDomainKey,
			umlComment:        "UmlComment",
			errstr:            `Key: (keyType: cannot be blankxx;`,
		},
		{
			key:               helper.Must(identity.NewDomainAssociationKey(suite.problemDomainKey, "1")),
			problemDomainKey:  identity.Key{},
			solutionDomainKey: suite.solutionDomainKey,
			umlComment:        "UmlComment",
			errstr:            `ProblemDomainKey: (keyType: cannot be blank;`,
		},
		{
			key:               helper.Must(identity.NewDomainAssociationKey(suite.problemDomainKey, "1")),
			problemDomainKey:  helper.Must(identity.NewActorKey("actor1")),
			solutionDomainKey: suite.solutionDomainKey,
			umlComment:        "UmlComment",
			errstr:            `ProblemDomainKey: (keyType: cannot be blankxxx;`,
		},
		{
			key:               helper.Must(identity.NewDomainAssociationKey(suite.problemDomainKey, "1")),
			problemDomainKey:  suite.problemDomainKey,
			solutionDomainKey: identity.Key{},
			umlComment:        "UmlComment",
			errstr:            `SolutionDomainKey: (keyType: cannot be blank;`,
		},
		{
			key:               helper.Must(identity.NewDomainAssociationKey(suite.problemDomainKey, "1")),
			problemDomainKey:  suite.problemDomainKey,
			solutionDomainKey: helper.Must(identity.NewActorKey("actor1")),
			umlComment:        "UmlComment",
			errstr:            `SolutionDomainKey: (keyType: cannot be blank;`,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		obj, err := NewAssociation(test.key, test.problemDomainKey, test.solutionDomainKey, test.umlComment)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.obj, obj, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), obj, testName)
		}
	}
}
