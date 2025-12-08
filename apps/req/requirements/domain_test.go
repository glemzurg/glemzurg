package requirements

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestDomainSuite(t *testing.T) {
	suite.Run(t, new(DomainSuite))
}

type DomainSuite struct {
	suite.Suite
}

func (suite *DomainSuite) TestNew() {
	tests := []struct {
		key        string
		name       string
		details    string
		realized   bool
		umlComment string
		obj        Domain
		errstr     string
	}{
		// OK.
		{
			key:        "Key",
			name:       "Name",
			details:    "Details",
			realized:   true,
			umlComment: "UmlComment",
			obj: Domain{
				Key:        "Key",
				Name:       "Name",
				Details:    "Details",
				Realized:   true,
				UmlComment: "UmlComment",
			},
		},
		{
			key:        "Key",
			name:       "Name",
			details:    "",
			realized:   false,
			umlComment: "",
			obj: Domain{
				Key:        "Key",
				Name:       "Name",
				Details:    "",
				Realized:   false,
				UmlComment: "",
			},
		},

		// Error states.
		{
			key:    "",
			name:   "Name",
			errstr: `cannot be blank`,
		},
		{
			key:    "Key",
			name:   "",
			errstr: `Name: cannot be blank.`,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		obj, err := NewDomain(test.key, test.name, test.details, test.realized, test.umlComment)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.obj, obj, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), obj, testName)
		}
	}
}
