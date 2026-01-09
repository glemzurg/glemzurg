package model_domain

import (
	"fmt"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
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

	domainKey := helper.Must(identity.NewRootKey(identity.KEY_TYPE_DOMAIN, "domain1"))

	tests := []struct {
		key        identity.Key
		name       string
		details    string
		umlComment string
		obj        Subdomain
		errstr     string
	}{
		// OK.
		{
			key:        helper.Must(NewSubdomainKey(domainKey, "subdomain1")),
			name:       "Name",
			details:    "Details",
			umlComment: "UmlComment",
			obj: Subdomain{
				Key:        helper.Must(NewSubdomainKey(domainKey, "subdomain1")),
				Name:       "Name",
				Details:    "Details",
				UmlComment: "UmlComment",
			},
		},
		{
			key:        helper.Must(NewSubdomainKey(domainKey, "subdomain1")),
			name:       "Name",
			details:    "",
			umlComment: "",
			obj: Subdomain{
				Key:        helper.Must(NewSubdomainKey(domainKey, "subdomain1")),
				Name:       "Name",
				Details:    "",
				UmlComment: "",
			},
		},

		// Error states.
		{
			key:     identity.Key{},
			name:    "Name",
			details: "Details",
			errstr:  "Key: (childType: cannot be blank; subKey: cannot be blank.).",
		},
		{
			key:     helper.Must(identity.NewKey(domainKey.String(), "class", "subdomain1")),
			name:    "Name",
			details: "Details",
			errstr:  "Key: invalid child type for subdomain.",
		},
		{
			key:     helper.Must(NewSubdomainKey(domainKey, "subdomain1")),
			name:    "",
			details: "Details",
			errstr:  "Name: cannot be blank.",
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

func (suite *SubdomainSuite) TestNewSubdomainKey() {
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
			subKey:    "subdomain1",
			expected:  helper.Must(identity.NewKey(domainKey.String(), identity.KEY_TYPE_SUBDOMAIN, "subdomain1")),
		},
		{
			domainKey: domainKey,
			subKey:    "subdomain2",
			expected:  helper.Must(identity.NewKey(domainKey.String(), identity.KEY_TYPE_SUBDOMAIN, "subdomain2")),
		},

		// OK case: blank parentKey.
		{
			domainKey: identity.Key{},
			subKey:    "subdomain1",
			expected:  helper.Must(identity.NewKey("", identity.KEY_TYPE_SUBDOMAIN, "subdomain1")),
		},
		{
			domainKey: domainKey,
			subKey:    "",
			errstr:    "cannot be blank",
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		key, err := NewSubdomainKey(test.domainKey, test.subKey)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.expected, key, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Equal(suite.T(), identity.Key{}, key, testName)
		}
	}
}
