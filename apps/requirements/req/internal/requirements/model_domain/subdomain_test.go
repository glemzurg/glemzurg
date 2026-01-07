package model_domain

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestSubdomainSuite(t *testing.T) {
	suite.Run(t, new(SubdomainSuite))
}

type SubdomainSuite struct {
	suite.Suite
}

func (suite *SubdomainSuite) TestNew() {
	tests := []struct {
		key        string
		name       string
		details    string
		umlComment string
		obj        Subdomain
		errstr     string
	}{
		// OK.
		{
			key:        "Key",
			name:       "Name",
			details:    "Details",
			umlComment: "UmlComment",
			obj: Subdomain{
				Key:        "Key",
				Name:       "Name",
				Details:    "Details",
				UmlComment: "UmlComment",
			},
		},
		{
			key:        "Key",
			name:       "Name",
			details:    "",
			umlComment: "",
			obj: Subdomain{
				Key:        "Key",
				Name:       "Name",
				Details:    "",
				UmlComment: "",
			},
		},

		// Error states.
		{
			key:     "",
			name:    "Name",
			details: "Details",
			errstr:  `cannot be blank`,
		},
		{
			key:     "Key",
			name:    "",
			details: "Details",
			errstr:  `Name: cannot be blank.`,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		obj, err := NewSubdomain(test.key, test.name, test.details, test.umlComment)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.obj, obj, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), obj, testName)
		}
	}
}
