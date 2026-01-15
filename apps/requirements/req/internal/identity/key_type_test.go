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

	problemDomainKey, err := NewDomainKey("problem1")
	assert.NoError(suite.T(), err)

	solutionDomainKey, err := NewDomainKey("solution1")
	assert.NoError(suite.T(), err)

	solution1SubKey := "solution1"
	tests := []struct {
		testName          string
		problemDomainKey  Key
		solutionDomainKey Key
		expected          Key
		errstr            string
	}{
		// OK.
		{
			testName:          "ok",
			problemDomainKey:  problemDomainKey,
			solutionDomainKey: solutionDomainKey,
			expected: Key{
				parentKey: problemDomainKey.String(),
				keyType:   KEY_TYPE_DOMAIN_ASSOCIATION,
				subKey:    "problem1",
				subKey2:   &solution1SubKey,
			},
		},

		// Errors.
		{
			testName:          "error empty problem domain",
			problemDomainKey:  Key{},
			solutionDomainKey: solutionDomainKey,
			errstr:            "problem domain key cannot be of type '' for 'dassociation' key",
		},
		{
			testName:          "error wrong problem domain type",
			problemDomainKey:  helper.Must(NewActorKey("actor1")),
			solutionDomainKey: solutionDomainKey,
			errstr:            "problem domain key cannot be of type 'actor' for 'dassociation' key",
		},
		{
			testName:          "error empty solution domain",
			problemDomainKey:  problemDomainKey,
			solutionDomainKey: Key{},
			errstr:            "solution domain key cannot be of type '' for 'dassociation' key",
		},
		{
			testName:          "error wrong solution domain type",
			problemDomainKey:  problemDomainKey,
			solutionDomainKey: helper.Must(NewActorKey("actor1")),
			errstr:            "solution domain key cannot be of type 'actor' for 'dassociation' key",
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			key, err := NewDomainAssociationKey(tt.problemDomainKey, tt.solutionDomainKey)
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

func (suite *KeyTypeSuite) TestNewUseCaseKey() {

	domainKey, err := NewDomainKey("domain1")
	assert.NoError(suite.T(), err)

	subdomainKey, err := NewSubdomainKey(domainKey, "subdomain1")
	assert.NoError(suite.T(), err)

	tests := []struct {
		testName     string
		subdomainKey Key
		subKey       string
		expected     Key
		errstr       string
	}{
		// OK.
		{
			testName:     "ok",
			subdomainKey: subdomainKey,
			subKey:       "usecase1",
			expected:     helper.Must(newKey(subdomainKey.String(), KEY_TYPE_USE_CASE, "usecase1")),
		},

		// Errors.
		{
			testName:     "error empty parent",
			subdomainKey: Key{},
			subKey:       "usecase1",
			errstr:       "parent key cannot be of type '' for 'usecase' key",
		},
		{
			testName:     "error wrong parent type",
			subdomainKey: helper.Must(NewActorKey("actor1")),
			subKey:       "usecase1",
			errstr:       "parent key cannot be of type 'actor' for 'usecase' key",
		},
		{
			testName:     "error blank subKey",
			subdomainKey: subdomainKey,
			subKey:       "",
			errstr:       "cannot be blank",
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			key, err := NewUseCaseKey(tt.subdomainKey, tt.subKey)
			if tt.errstr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, key)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
				assert.Equal(t, Key{}, key)
			}
		})
	}
}
