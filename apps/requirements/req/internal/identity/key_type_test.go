package identity

import (
	"fmt"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestKeyTypeSuite(t *testing.T) {
	suite.Run(t, new(KeyTypeSuite))
}

type KeyTypeSuite struct {
	suite.Suite
}

func (suite *KeyTypeSuite) TestNewActorKey() {
	tests := []struct {
		subKey   string
		expected Key
		errstr   string
	}{
		// OK.
		{
			subKey:   "actor1",
			expected: helper.Must(newRootKey(KEY_TYPE_ACTOR, "actor1")),
		},

		// Errors.
		{
			subKey: "",
			errstr: "cannot be blank",
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		key, err := NewActorKey(test.subKey)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.expected, key, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Equal(suite.T(), Key{}, key, testName)
		}
	}
}

func (suite *KeyTypeSuite) TestNewDomainKey() {
	tests := []struct {
		subKey   string
		expected Key
		errstr   string
	}{
		// OK.
		{
			subKey:   "domain1",
			expected: helper.Must(newRootKey(KEY_TYPE_DOMAIN, "domain1")),
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
			assert.Equal(suite.T(), Key{}, key, testName)
		}
	}
}

func (suite *KeyTypeSuite) TestNewDomainAssociationKey() {

	domainKey, err := NewDomainKey("domain1")
	assert.Nil(suite.T(), err)

	tests := []struct {
		domainKey Key
		subKey    string
		expected  Key
		errstr    string
	}{
		// OK.
		{
			domainKey: domainKey,
			subKey:    "1",
			expected:  helper.Must(newKey(domainKey.String(), KEY_TYPE_ASSOCIATION, "1")),
		},

		// Errors.
		{
			domainKey: Key{},
			subKey:    "1",
			errstr:    "parent key cannot be of type '' for 'association' key",
		},
		{
			domainKey: helper.Must(NewActorKey("actor1")),
			subKey:    "1",
			errstr:    "parent key cannot be of type 'actor' for 'association' key",
		},
		{
			domainKey: domainKey,
			subKey:    "",
			errstr:    "cannot be blank",
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		key, err := NewDomainAssociationKey(test.domainKey, test.subKey)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.expected, key, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Equal(suite.T(), Key{}, key, testName)
		}
	}
}

func (suite *KeyTypeSuite) TestNewSubdomainKey() {

	domainKey, err := NewDomainKey("domain1")
	assert.Nil(suite.T(), err)

	tests := []struct {
		domainKey Key
		subKey    string
		expected  Key
		errstr    string
	}{
		// OK.
		{
			domainKey: domainKey,
			subKey:    "subdomain1",
			expected:  helper.Must(newKey(domainKey.String(), KEY_TYPE_SUBDOMAIN, "subdomain1")),
		},

		// Errors.
		{
			domainKey: Key{},
			subKey:    "subdomain1",
			errstr:    "parent key cannot be of type '' for 'subdomain' key",
		},
		{
			domainKey: helper.Must(NewActorKey("actor1")),
			subKey:    "subdomain1",
			errstr:    "parent key cannot be of type 'actor' for 'subdomain' key",
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
			assert.Equal(suite.T(), Key{}, key, testName)
		}
	}
}
