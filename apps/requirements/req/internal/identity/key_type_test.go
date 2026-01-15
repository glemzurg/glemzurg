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
				subKey:    solution1SubKey,
				subKey2:   nil,
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

func (suite *KeyTypeSuite) TestNewClassAssociationKey() {

	// Create keys for testing.
	domain1Key := helper.Must(NewDomainKey("domain1"))
	domain2Key := helper.Must(NewDomainKey("domain2"))

	subdomain1Key := helper.Must(NewSubdomainKey(domain1Key, "subdomain1"))
	subdomain2Key := helper.Must(NewSubdomainKey(domain1Key, "subdomain2"))
	subdomain3Key := helper.Must(NewSubdomainKey(domain2Key, "subdomain3"))

	// Classes in subdomain1.
	class1Key := helper.Must(NewClassKey(subdomain1Key, "class1"))
	class2Key := helper.Must(NewClassKey(subdomain1Key, "class2"))

	// Classes in subdomain2 (same domain as subdomain1).
	class3Key := helper.Must(NewClassKey(subdomain2Key, "class3"))

	// Classes in subdomain3 (different domain).
	class4Key := helper.Must(NewClassKey(subdomain3Key, "class4"))

	// Helper for creating string pointers.
	strPtr := func(s string) *string { return &s }

	tests := []struct {
		testName     string
		parentKey    Key
		fromClassKey Key
		toClassKey   Key
		expected     Key
		errstr       string
	}{
		// OK: Parent is subdomain - both classes in same subdomain.
		{
			testName:     "ok subdomain parent",
			parentKey:    subdomain1Key,
			fromClassKey: class1Key,
			toClassKey:   class2Key,
			expected: Key{
				parentKey: subdomain1Key.String(),
				keyType:   KEY_TYPE_CLASS_ASSOCIATION,
				subKey:    "class/class1",
				subKey2:   strPtr("class/class2"),
			},
		},

		// OK: Parent is domain - classes in different subdomains of same domain.
		{
			testName:     "ok domain parent",
			parentKey:    domain1Key,
			fromClassKey: class1Key,
			toClassKey:   class3Key,
			expected: Key{
				parentKey: domain1Key.String(),
				keyType:   KEY_TYPE_CLASS_ASSOCIATION,
				subKey:    "subdomain/subdomain1/class/class1",
				subKey2:   strPtr("subdomain/subdomain2/class/class3"),
			},
		},

		// OK: Parent is model (empty) - classes in different domains.
		{
			testName:     "ok model parent",
			parentKey:    Key{},
			fromClassKey: class1Key,
			toClassKey:   class4Key,
			expected: Key{
				parentKey: "",
				keyType:   KEY_TYPE_CLASS_ASSOCIATION,
				subKey:    "domain/domain1/subdomain/subdomain1/class/class1",
				subKey2:   strPtr("domain/domain2/subdomain/subdomain3/class/class4"),
			},
		},

		// Errors: Wrong key types for classes.
		{
			testName:     "error from class wrong type",
			parentKey:    subdomain1Key,
			fromClassKey: helper.Must(NewActorKey("actor1")),
			toClassKey:   class2Key,
			errstr:       "from class key cannot be of type 'actor' for 'cassociation' key",
		},
		{
			testName:     "error to class wrong type",
			parentKey:    subdomain1Key,
			fromClassKey: class1Key,
			toClassKey:   helper.Must(NewActorKey("actor1")),
			errstr:       "to class key cannot be of type 'actor' for 'cassociation' key",
		},

		// Errors: Wrong parent type.
		{
			testName:     "error wrong parent type",
			parentKey:    helper.Must(NewActorKey("actor1")),
			fromClassKey: class1Key,
			toClassKey:   class2Key,
			errstr:       "parent key cannot be of type 'actor' for 'cassociation' key",
		},

		// Errors: Subdomain parent - class not in subdomain.
		{
			testName:     "error subdomain parent from class not in subdomain",
			parentKey:    subdomain1Key,
			fromClassKey: class3Key, // class3 is in subdomain2, not subdomain1.
			toClassKey:   class2Key,
			errstr:       "from class key 'domain/domain1/subdomain/subdomain2/class/class3' is not in subdomain",
		},
		{
			testName:     "error subdomain parent to class not in subdomain",
			parentKey:    subdomain1Key,
			fromClassKey: class1Key,
			toClassKey:   class3Key, // class3 is in subdomain2, not subdomain1.
			errstr:       "to class key 'domain/domain1/subdomain/subdomain2/class/class3' is not in subdomain",
		},

		// Errors: Domain parent - class not in domain.
		{
			testName:     "error domain parent from class not in domain",
			parentKey:    domain1Key,
			fromClassKey: class4Key, // class4 is in domain2, not domain1.
			toClassKey:   class3Key,
			errstr:       "from class key 'domain/domain2/subdomain/subdomain3/class/class4' is not in domain",
		},
		{
			testName:     "error domain parent to class not in domain",
			parentKey:    domain1Key,
			fromClassKey: class1Key,
			toClassKey:   class4Key, // class4 is in domain2, not domain1.
			errstr:       "to class key 'domain/domain2/subdomain/subdomain3/class/class4' is not in domain",
		},

		// Errors: Domain parent - classes in same subdomain (should use subdomain parent).
		{
			testName:     "error domain parent classes in same subdomain",
			parentKey:    domain1Key,
			fromClassKey: class1Key,
			toClassKey:   class2Key, // Both in subdomain1.
			errstr:       "classes are in the same subdomain 'subdomain1', use subdomain as parent instead",
		},

		// Errors: Model parent - classes in same domain (should use domain parent).
		{
			testName:     "error model parent classes in same domain",
			parentKey:    Key{},
			fromClassKey: class1Key,
			toClassKey:   class3Key, // Both in domain1.
			errstr:       "classes are in the same domain 'domain1', use domain as parent instead",
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			key, err := NewClassAssociationKey(tt.parentKey, tt.fromClassKey, tt.toClassKey)
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
