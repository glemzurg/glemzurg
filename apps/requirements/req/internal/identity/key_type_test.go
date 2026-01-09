package identity

import (
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
		testName string
		subKey   string
		expected Key
		errstr   string
	}{
		// OK.
		{
			testName: "ok",
			subKey:   "actor1",
			expected: helper.Must(newRootKey(KEY_TYPE_ACTOR, "actor1")),
		},

		// Errors.
		{
			testName: "error blank",
			subKey:   "",
			errstr:   "cannot be blank",
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			key, err := NewActorKey(tt.subKey)
			if tt.errstr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, key)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
				assert.Equal(t, Key{}, key)
			}
		})
		if !pass {
			break
		}
	}
}

func (suite *KeyTypeSuite) TestNewDomainKey() {
	tests := []struct {
		testName string
		subKey   string
		expected Key
		errstr   string
	}{
		// OK.
		{
			testName: "ok",
			subKey:   "domain1",
			expected: helper.Must(newRootKey(KEY_TYPE_DOMAIN, "domain1")),
		},

		// Errors.
		{
			testName: "error blank",
			subKey:   "",
			errstr:   "cannot be blank",
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			key, err := NewDomainKey(tt.subKey)
			if tt.errstr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, key)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
				assert.Equal(t, Key{}, key)
			}
		})
		if !pass {
			break
		}
	}
}

func (suite *KeyTypeSuite) TestNewDomainAssociationKey() {

	domainKey, err := NewDomainKey("domain1")
	assert.NoError(suite.T(), err)

	tests := []struct {
		testName  string
		domainKey Key
		subKey    string
		expected  Key
		errstr    string
	}{
		// OK.
		{
			testName:  "ok",
			domainKey: domainKey,
			subKey:    "1",
			expected:  helper.Must(newKey(domainKey.String(), KEY_TYPE_DOMAIN_ASSOCIATION, "1")),
		},

		// Errors.
		{
			testName:  "error empty parent",
			domainKey: Key{},
			subKey:    "1",
			errstr:    "parent key cannot be of type '' for 'association' key",
		},
		{
			testName:  "error wrong parent type",
			domainKey: helper.Must(NewActorKey("actor1")),
			subKey:    "1",
			errstr:    "parent key cannot be of type 'actor' for 'association' key",
		},
		{
			testName:  "error blank subKey",
			domainKey: domainKey,
			subKey:    "",
			errstr:    "cannot be blank",
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			key, err := NewDomainAssociationKey(tt.domainKey, tt.subKey)
			if tt.errstr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, key)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
				assert.Equal(t, Key{}, key)
			}
		})
		if !pass {
			break
		}
	}
}

func (suite *KeyTypeSuite) TestNewSubdomainKey() {

	domainKey, err := NewDomainKey("domain1")
	assert.NoError(suite.T(), err)

	tests := []struct {
		testName  string
		domainKey Key
		subKey    string
		expected  Key
		errstr    string
	}{
		// OK.
		{
			testName:  "ok",
			domainKey: domainKey,
			subKey:    "subdomain1",
			expected:  helper.Must(newKey(domainKey.String(), KEY_TYPE_SUBDOMAIN, "subdomain1")),
		},

		// Errors.
		{
			testName:  "error empty parent",
			domainKey: Key{},
			subKey:    "subdomain1",
			errstr:    "parent key cannot be of type '' for 'subdomain' key",
		},
		{
			testName:  "error wrong parent type",
			domainKey: helper.Must(NewActorKey("actor1")),
			subKey:    "subdomain1",
			errstr:    "parent key cannot be of type 'actor' for 'subdomain' key",
		},
		{
			testName:  "error blank subKey",
			domainKey: domainKey,
			subKey:    "",
			errstr:    "cannot be blank",
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			key, err := NewSubdomainKey(tt.domainKey, tt.subKey)
			if tt.errstr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, key)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
				assert.Equal(t, Key{}, key)
			}
		})
		if !pass {
			break
		}
	}
}
