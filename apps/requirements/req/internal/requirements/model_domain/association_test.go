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
}

func (suite *AssociationSuite) TestNew() {

	problemDomainKey := helper.Must(identity.NewRootKey(identity.KEY_TYPE_DOMAIN, "domain1"))
	solutionDomainKey := helper.Must(identity.NewRootKey(identity.KEY_TYPE_DOMAIN, "domain2"))

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
			key:               helper.Must(NewAssociationKey(problemDomainKey, "1")),
			problemDomainKey:  problemDomainKey,
			solutionDomainKey: solutionDomainKey,
			umlComment:        "UmlComment",
			obj: Association{
				Key:               helper.Must(NewAssociationKey(problemDomainKey, "1")),
				ProblemDomainKey:  problemDomainKey,
				SolutionDomainKey: solutionDomainKey,
				UmlComment:        "UmlComment",
			},
		},
		{
			key:               helper.Must(NewAssociationKey(problemDomainKey, "2")),
			problemDomainKey:  problemDomainKey,
			solutionDomainKey: solutionDomainKey,
			umlComment:        "",
			obj: Association{
				Key:               helper.Must(NewAssociationKey(problemDomainKey, "2")),
				ProblemDomainKey:  problemDomainKey,
				SolutionDomainKey: solutionDomainKey,
				UmlComment:        "",
			},
		},

		// Error states.
		{
			key:               identity.Key{},
			problemDomainKey:  problemDomainKey,
			solutionDomainKey: solutionDomainKey,
			umlComment:        "UmlComment",
			errstr:            `Key: (childType: cannot be blank; subKey: cannot be blank.).`,
		},
		{
			key:               helper.Must(identity.NewKey(problemDomainKey.String(), "class", "1")),
			problemDomainKey:  problemDomainKey,
			solutionDomainKey: solutionDomainKey,
			umlComment:        "UmlComment",
			errstr:            "Key: invalid child type for association.",
		},
		{
			key:               helper.Must(NewAssociationKey(problemDomainKey, "1")),
			problemDomainKey:  identity.Key{},
			solutionDomainKey: solutionDomainKey,
			umlComment:        "UmlComment",
			errstr:            `ProblemDomainKey: (childType: cannot be blank; subKey: cannot be blank.).`,
		},
		{
			key:               helper.Must(NewAssociationKey(problemDomainKey, "1")),
			problemDomainKey:  problemDomainKey,
			solutionDomainKey: identity.Key{},
			umlComment:        "UmlComment",
			errstr:            `SolutionDomainKey: (childType: cannot be blank; subKey: cannot be blank.).`,
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
func (suite *AssociationSuite) TestNewAssociationKey() {
	domainKey := helper.Must(identity.NewRootKey(identity.KEY_TYPE_DOMAIN, "domain1"))

	tests := []struct {
		domainKey identity.Key
		subKey    string
		expected  identity.Key
		errstr    string
	}{
		// OK.
		{
			domainKey: domainKey,
			subKey:    "1",
			expected:  helper.Must(identity.NewKey(domainKey.String(), identity.KEY_TYPE_ASSOCIATION, "1")),
		},
		{
			domainKey: domainKey,
			subKey:    "2",
			expected:  helper.Must(identity.NewKey(domainKey.String(), identity.KEY_TYPE_ASSOCIATION, "2")),
		},

		// OK case: blank parentKey.
		{
			domainKey: identity.Key{},
			subKey:    "1",
			expected:  helper.Must(identity.NewKey("", identity.KEY_TYPE_ASSOCIATION, "1")),
		},
		{
			domainKey: domainKey,
			subKey:    "",
			errstr:    "cannot be blank",
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		key, err := NewAssociationKey(test.domainKey, test.subKey)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.expected, key, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Equal(suite.T(), identity.Key{}, key, testName)
		}
	}
}
