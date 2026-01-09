package model_domain

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

func TestDomainSuite(t *testing.T) {
	suite.Run(t, new(DomainSuite))
}

type DomainSuite struct {
	suite.Suite
}

func (suite *DomainSuite) TestNewDomainKey() {
	tests := []struct {
		subKey   string
		expected identity.Key
		errstr   string
	}{
		// OK.
		{
			subKey:   "domain1",
			expected: helper.Must(identity.NewRootKey(identity.KEY_TYPE_DOMAIN, "domain1")),
		},

		// Errors.
		{
			subKey: "",
			errstr: "cannot be blank",
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		key, err := NewDomainKey(test.subKey)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.expected, key, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Equal(suite.T(), identity.Key{}, key, testName)
		}
	}
}

func (suite *DomainSuite) TestNew() {
	tests := []struct {
		key        identity.Key
		name       string
		details    string
		realized   bool
		umlComment string
		obj        Domain
		errstr     string
	}{
		// OK.
		{
			key:        helper.Must(identity.NewRootKey(identity.KEY_TYPE_DOMAIN, "domain1")),
			name:       "Name",
			details:    "Details",
			realized:   true,
			umlComment: "UmlComment",
			obj: Domain{
				Key:        helper.Must(identity.NewRootKey(identity.KEY_TYPE_DOMAIN, "domain1")),
				Name:       "Name",
				Details:    "Details",
				Realized:   true,
				UmlComment: "UmlComment",
			},
		},
		{
			key:        helper.Must(identity.NewRootKey(identity.KEY_TYPE_DOMAIN, "domain1")),
			name:       "Name",
			details:    "",
			realized:   false,
			umlComment: "",
			obj: Domain{
				Key:        helper.Must(identity.NewRootKey(identity.KEY_TYPE_DOMAIN, "domain1")),
				Name:       "Name",
				Details:    "",
				Realized:   false,
				UmlComment: "",
			},
		},

		// Error states.
		{
			key:    identity.Key{},
			name:   "Name",
			errstr: "keyType: cannot be blank",
		},
		{
			key:    helper.Must(identity.NewRootKey(identity.KEY_TYPE_DOMAIN, "domain1")),
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
