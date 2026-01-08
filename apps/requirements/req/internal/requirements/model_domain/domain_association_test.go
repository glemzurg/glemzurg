package model_domain

import (
	"fmt"
	"testing"

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
		tests := []struct {
			key               string
			problemDomainKey  string
			solutionDomainKey string
			umlComment        string
			obj               Association
			errstr            string
		}{
			// OK.
			{
				key:               "Key",
				problemDomainKey:  "ProblemDomainKey",
				solutionDomainKey: "SolutionDomainKey",
				umlComment:        "UmlComment",
				obj: Association{
					Key:               "Key",
					ProblemDomainKey:  "ProblemDomainKey",
					SolutionDomainKey: "SolutionDomainKey",
					UmlComment:        "UmlComment",
				},
			},
			{
				key:               "Key",
				problemDomainKey:  "ProblemDomainKey",
				solutionDomainKey: "SolutionDomainKey",
				umlComment:        "",
				obj: Association{
					Key:               "Key",
					ProblemDomainKey:  "ProblemDomainKey",
					SolutionDomainKey: "SolutionDomainKey",
					UmlComment:        "",
				},
			},

			// Error states.
			{
				key:               "",
				problemDomainKey:  "ProblemDomainKey",
				solutionDomainKey: "SolutionDomainKey",
				umlComment:        "UmlComment",
				errstr:            `Key: cannot be blank`,
			},
			{
				key:               "Key",
				problemDomainKey:  "",
				solutionDomainKey: "SolutionDomainKey",
				umlComment:        "UmlComment",
				errstr:            `ProblemDomainKey: cannot be blank`,
			},
			{
				key:               "Key",
				problemDomainKey:  "ProblemDomainKey",
				solutionDomainKey: "",
				umlComment:        "UmlComment",
				errstr:            `SolutionDomainKey: cannot be blank`,
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
