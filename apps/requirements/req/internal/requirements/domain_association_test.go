package requirements

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestDomainAssociationSuite(t *testing.T) {
	suite.Run(t, new(DomainAssociationSuite))
}

type DomainAssociationSuite struct {
	suite.Suite
}

func (suite *DomainAssociationSuite) TestNew() {
	tests := []struct {
		key               string
		problemDomainKey  string
		solutionDomainKey string
		umlComment        string
		obj               DomainAssociation
		errstr            string
	}{
		// OK.
		{
			key:               "Key",
			problemDomainKey:  "ProblemDomainKey",
			solutionDomainKey: "SolutionDomainKey",
			umlComment:        "UmlComment",
			obj: DomainAssociation{
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
			obj: DomainAssociation{
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
		obj, err := NewDomainAssociation(test.key, test.problemDomainKey, test.solutionDomainKey, test.umlComment)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.obj, obj, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), obj, testName)
		}
	}
}
