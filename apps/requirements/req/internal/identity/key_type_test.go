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
			errstr:   "'SubKey' failed on the 'required' tag",
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
			errstr:   "'SubKey' failed on the 'required' tag",
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

func (suite *KeyTypeSuite) TestNewInvariantKey() {
	tests := []struct {
		testName string
		subKey   string
		expected Key
		errstr   string
	}{
		// OK.
		{
			testName: "ok",
			subKey:   "_InvariantA",
			expected: helper.Must(newRootKey(KEY_TYPE_INVARIANT, "_InvariantA")),
		},

		// Errors.
		{
			testName: "error blank",
			subKey:   "",
			errstr:   "'SubKey' failed on the 'required' tag",
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			key, err := NewInvariantKey(tt.subKey)
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
				ParentKey: "",
				KeyType:   KEY_TYPE_DOMAIN_ASSOCIATION,
				SubKey:    "problem1",
				SubKey2:   "solution1",
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
			errstr:    "'SubKey' failed on the 'required' tag",
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

func (suite *KeyTypeSuite) TestNewAttributeDerivationKey() {

	domainKey, err := NewDomainKey("domain1")
	assert.NoError(suite.T(), err)

	subdomainKey, err := NewSubdomainKey(domainKey, "subdomain1")
	assert.NoError(suite.T(), err)

	classKey, err := NewClassKey(subdomainKey, "class1")
	assert.NoError(suite.T(), err)

	attributeKey, err := NewAttributeKey(classKey, "attribute1")
	assert.NoError(suite.T(), err)

	tests := []struct {
		testName     string
		attributeKey Key
		subKey       string
		expected     Key
		errstr       string
	}{
		// OK.
		{
			testName:     "ok",
			attributeKey: attributeKey,
			subKey:       "attribute1",
			expected:     helper.Must(newKey(attributeKey.String(), KEY_TYPE_ATTRIBUTE_DERIVATION, "attribute1")),
		},

		// Errors.
		{
			testName:     "error empty parent",
			attributeKey: Key{},
			subKey:       "attribute1",
			errstr:       "parent key cannot be of type '' for 'aderive' key",
		},
		{
			testName:     "error wrong parent type",
			attributeKey: helper.Must(NewActorKey("actor1")),
			subKey:       "subdomain1",
			errstr:       "parent key cannot be of type 'actor' for 'aderive' key",
		},
		{
			testName:     "error blank subKey",
			attributeKey: attributeKey,
			subKey:       "",
			errstr:       "'SubKey' failed on the 'required' tag",
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			key, err := NewAttributeDerivationKey(tt.attributeKey, tt.subKey)
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

func (suite *KeyTypeSuite) TestNewActionRequireKey() {

	domainKey, err := NewDomainKey("domain1")
	assert.NoError(suite.T(), err)

	subdomainKey, err := NewSubdomainKey(domainKey, "subdomain1")
	assert.NoError(suite.T(), err)

	classKey, err := NewClassKey(subdomainKey, "class1")
	assert.NoError(suite.T(), err)

	actionKey, err := NewActionKey(classKey, "action1")
	assert.NoError(suite.T(), err)

	tests := []struct {
		testName  string
		actionKey Key
		subKey    string
		expected  Key
		errstr    string
	}{
		// OK.
		{
			testName:  "ok",
			actionKey: actionKey,
			subKey:    "1",
			expected:  helper.Must(newKey(actionKey.String(), KEY_TYPE_ACTION_REQUIRE, "1")),
		},

		// Errors.
		{
			testName:  "error empty parent",
			actionKey: Key{},
			subKey:    "1",
			errstr:    "parent key cannot be of type '' for 'arequire' key",
		},
		{
			testName:  "error wrong parent type",
			actionKey: helper.Must(NewActorKey("actor1")),
			subKey:    "1",
			errstr:    "parent key cannot be of type 'actor' for 'arequire' key",
		},
		{
			testName:  "error blank subKey",
			actionKey: actionKey,
			subKey:    "",
			errstr:    "'SubKey' failed on the 'required' tag",
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			key, err := NewActionRequireKey(tt.actionKey, tt.subKey)
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

func (suite *KeyTypeSuite) TestNewActionGuaranteeKey() {

	domainKey, err := NewDomainKey("domain1")
	assert.NoError(suite.T(), err)

	subdomainKey, err := NewSubdomainKey(domainKey, "subdomain1")
	assert.NoError(suite.T(), err)

	classKey, err := NewClassKey(subdomainKey, "class1")
	assert.NoError(suite.T(), err)

	actionKey, err := NewActionKey(classKey, "action1")
	assert.NoError(suite.T(), err)

	tests := []struct {
		testName  string
		actionKey Key
		subKey    string
		expected  Key
		errstr    string
	}{
		// OK.
		{
			testName:  "ok",
			actionKey: actionKey,
			subKey:    "1",
			expected:  helper.Must(newKey(actionKey.String(), KEY_TYPE_ACTION_GUARANTEE, "1")),
		},

		// Errors.
		{
			testName:  "error empty parent",
			actionKey: Key{},
			subKey:    "1",
			errstr:    "parent key cannot be of type '' for 'aguarantee' key",
		},
		{
			testName:  "error wrong parent type",
			actionKey: helper.Must(NewActorKey("actor1")),
			subKey:    "1",
			errstr:    "parent key cannot be of type 'actor' for 'aguarantee' key",
		},
		{
			testName:  "error blank subKey",
			actionKey: actionKey,
			subKey:    "",
			errstr:    "'SubKey' failed on the 'required' tag",
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			key, err := NewActionGuaranteeKey(tt.actionKey, tt.subKey)
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

func (suite *KeyTypeSuite) TestNewActionSafetyKey() {

	domainKey, err := NewDomainKey("domain1")
	assert.NoError(suite.T(), err)

	subdomainKey, err := NewSubdomainKey(domainKey, "subdomain1")
	assert.NoError(suite.T(), err)

	classKey, err := NewClassKey(subdomainKey, "class1")
	assert.NoError(suite.T(), err)

	actionKey, err := NewActionKey(classKey, "action1")
	assert.NoError(suite.T(), err)

	tests := []struct {
		testName  string
		actionKey Key
		subKey    string
		expected  Key
		errstr    string
	}{
		// OK.
		{
			testName:  "ok",
			actionKey: actionKey,
			subKey:    "1",
			expected:  helper.Must(newKey(actionKey.String(), KEY_TYPE_ACTION_SAFETY, "1")),
		},

		// Errors.
		{
			testName:  "error empty parent",
			actionKey: Key{},
			subKey:    "1",
			errstr:    "parent key cannot be of type '' for 'asafety' key",
		},
		{
			testName:  "error wrong parent type",
			actionKey: helper.Must(NewActorKey("actor1")),
			subKey:    "1",
			errstr:    "parent key cannot be of type 'actor' for 'asafety' key",
		},
		{
			testName:  "error blank subKey",
			actionKey: actionKey,
			subKey:    "",
			errstr:    "'SubKey' failed on the 'required' tag",
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			key, err := NewActionSafetyKey(tt.actionKey, tt.subKey)
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

func (suite *KeyTypeSuite) TestNewQueryRequireKey() {

	domainKey, err := NewDomainKey("domain1")
	assert.NoError(suite.T(), err)

	subdomainKey, err := NewSubdomainKey(domainKey, "subdomain1")
	assert.NoError(suite.T(), err)

	classKey, err := NewClassKey(subdomainKey, "class1")
	assert.NoError(suite.T(), err)

	queryKey, err := NewQueryKey(classKey, "query1")
	assert.NoError(suite.T(), err)

	tests := []struct {
		testName string
		queryKey Key
		subKey   string
		expected Key
		errstr   string
	}{
		// OK.
		{
			testName: "ok",
			queryKey: queryKey,
			subKey:   "1",
			expected: helper.Must(newKey(queryKey.String(), KEY_TYPE_QUERY_REQUIRE, "1")),
		},

		// Errors.
		{
			testName: "error empty parent",
			queryKey: Key{},
			subKey:   "1",
			errstr:   "parent key cannot be of type '' for 'qrequire' key",
		},
		{
			testName: "error wrong parent type",
			queryKey: helper.Must(NewActorKey("actor1")),
			subKey:   "1",
			errstr:   "parent key cannot be of type 'actor' for 'qrequire' key",
		},
		{
			testName: "error blank subKey",
			queryKey: queryKey,
			subKey:   "",
			errstr:   "'SubKey' failed on the 'required' tag",
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			key, err := NewQueryRequireKey(tt.queryKey, tt.subKey)
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

func (suite *KeyTypeSuite) TestNewQueryGuaranteeKey() {

	domainKey, err := NewDomainKey("domain1")
	assert.NoError(suite.T(), err)

	subdomainKey, err := NewSubdomainKey(domainKey, "subdomain1")
	assert.NoError(suite.T(), err)

	classKey, err := NewClassKey(subdomainKey, "class1")
	assert.NoError(suite.T(), err)

	queryKey, err := NewQueryKey(classKey, "query1")
	assert.NoError(suite.T(), err)

	tests := []struct {
		testName string
		queryKey Key
		subKey   string
		expected Key
		errstr   string
	}{
		// OK.
		{
			testName: "ok",
			queryKey: queryKey,
			subKey:   "1",
			expected: helper.Must(newKey(queryKey.String(), KEY_TYPE_QUERY_GUARANTEE, "1")),
		},

		// Errors.
		{
			testName: "error empty parent",
			queryKey: Key{},
			subKey:   "1",
			errstr:   "parent key cannot be of type '' for 'qguarantee' key",
		},
		{
			testName: "error wrong parent type",
			queryKey: helper.Must(NewActorKey("actor1")),
			subKey:   "1",
			errstr:   "parent key cannot be of type 'actor' for 'qguarantee' key",
		},
		{
			testName: "error blank subKey",
			queryKey: queryKey,
			subKey:   "",
			errstr:   "'SubKey' failed on the 'required' tag",
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			key, err := NewQueryGuaranteeKey(tt.queryKey, tt.subKey)
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
			errstr:       "'SubKey' failed on the 'required' tag",
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

	tests := []struct {
		testName     string
		parentKey    Key
		fromClassKey Key
		toClassKey   Key
		name         string
		expected     Key
		errstr       string
	}{
		// OK: Parent is subdomain - both classes in same subdomain.
		{
			testName:     "ok subdomain parent",
			parentKey:    subdomain1Key,
			fromClassKey: class1Key,
			toClassKey:   class2Key,
			name:         "My Association",
			expected: Key{
				ParentKey: subdomain1Key.String(),
				KeyType:   KEY_TYPE_CLASS_ASSOCIATION,
				SubKey:    "class/class1",
				SubKey2:   "class/class2",
				SubKey3:   "my_association",
			},
		},

		// OK: Parent is domain - classes in different subdomains of same domain.
		{
			testName:     "ok domain parent",
			parentKey:    domain1Key,
			fromClassKey: class1Key,
			toClassKey:   class3Key,
			name:         "Cross Subdomain Link",
			expected: Key{
				ParentKey: domain1Key.String(),
				KeyType:   KEY_TYPE_CLASS_ASSOCIATION,
				SubKey:    "subdomain/subdomain1/class/class1",
				SubKey2:   "subdomain/subdomain2/class/class3",
				SubKey3:   "cross_subdomain_link",
			},
		},

		// OK: Parent is model (empty) - classes in different domains.
		{
			testName:     "ok model parent",
			parentKey:    Key{},
			fromClassKey: class1Key,
			toClassKey:   class4Key,
			name:         "Cross Domain Link",
			expected: Key{
				ParentKey: "",
				KeyType:   KEY_TYPE_CLASS_ASSOCIATION,
				SubKey:    "domain/domain1/subdomain/subdomain1/class/class1",
				SubKey2:   "domain/domain2/subdomain/subdomain3/class/class4",
				SubKey3:   "cross_domain_link",
			},
		},

		// OK: Name distillation - leading/trailing spaces.
		{
			testName:     "ok name with leading trailing spaces",
			parentKey:    subdomain1Key,
			fromClassKey: class1Key,
			toClassKey:   class2Key,
			name:         "  Spaced Name  ",
			expected: Key{
				ParentKey: subdomain1Key.String(),
				KeyType:   KEY_TYPE_CLASS_ASSOCIATION,
				SubKey:    "class/class1",
				SubKey2:   "class/class2",
				SubKey3:   "spaced_name",
			},
		},

		// OK: Name distillation - multiple internal spaces.
		{
			testName:     "ok name with multiple internal spaces",
			parentKey:    subdomain1Key,
			fromClassKey: class1Key,
			toClassKey:   class2Key,
			name:         "Multiple   Spaces   Here",
			expected: Key{
				ParentKey: subdomain1Key.String(),
				KeyType:   KEY_TYPE_CLASS_ASSOCIATION,
				SubKey:    "class/class1",
				SubKey2:   "class/class2",
				SubKey3:   "multiple___spaces___here",
			},
		},

		// OK: Name distillation - forward slashes converted to tildes.
		{
			testName:     "ok name with forward slashes",
			parentKey:    subdomain1Key,
			fromClassKey: class1Key,
			toClassKey:   class2Key,
			name:         "Parent/Child Relationship",
			expected: Key{
				ParentKey: subdomain1Key.String(),
				KeyType:   KEY_TYPE_CLASS_ASSOCIATION,
				SubKey:    "class/class1",
				SubKey2:   "class/class2",
				SubKey3:   "parent~child_relationship",
			},
		},

		// Errors: Empty name.
		{
			testName:     "error empty name",
			parentKey:    subdomain1Key,
			fromClassKey: class1Key,
			toClassKey:   class2Key,
			name:         "",
			errstr:       "name cannot be empty for class association key",
		},

		// Errors: Whitespace-only name.
		{
			testName:     "error whitespace only name",
			parentKey:    subdomain1Key,
			fromClassKey: class1Key,
			toClassKey:   class2Key,
			name:         "   ",
			errstr:       "name cannot be empty for class association key",
		},

		// Errors: Wrong key types for classes.
		{
			testName:     "error from class wrong type",
			parentKey:    subdomain1Key,
			fromClassKey: helper.Must(NewActorKey("actor1")),
			toClassKey:   class2Key,
			name:         "Test",
			errstr:       "from class key cannot be of type 'actor' for 'cassociation' key",
		},
		{
			testName:     "error to class wrong type",
			parentKey:    subdomain1Key,
			fromClassKey: class1Key,
			toClassKey:   helper.Must(NewActorKey("actor1")),
			name:         "Test",
			errstr:       "to class key cannot be of type 'actor' for 'cassociation' key",
		},

		// Errors: Wrong parent type.
		{
			testName:     "error wrong parent type",
			parentKey:    helper.Must(NewActorKey("actor1")),
			fromClassKey: class1Key,
			toClassKey:   class2Key,
			name:         "Test",
			errstr:       "parent key cannot be of type 'actor' for 'cassociation' key",
		},

		// Errors: Subdomain parent - class not in subdomain.
		{
			testName:     "error subdomain parent from class not in subdomain",
			parentKey:    subdomain1Key,
			fromClassKey: class3Key, // class3 is in subdomain2, not subdomain1.
			toClassKey:   class2Key,
			name:         "Test",
			errstr:       "from class key 'domain/domain1/subdomain/subdomain2/class/class3' is not in subdomain",
		},
		{
			testName:     "error subdomain parent to class not in subdomain",
			parentKey:    subdomain1Key,
			fromClassKey: class1Key,
			toClassKey:   class3Key, // class3 is in subdomain2, not subdomain1.
			name:         "Test",
			errstr:       "to class key 'domain/domain1/subdomain/subdomain2/class/class3' is not in subdomain",
		},

		// Errors: Domain parent - class not in domain.
		{
			testName:     "error domain parent from class not in domain",
			parentKey:    domain1Key,
			fromClassKey: class4Key, // class4 is in domain2, not domain1.
			toClassKey:   class3Key,
			name:         "Test",
			errstr:       "from class key 'domain/domain2/subdomain/subdomain3/class/class4' is not in domain",
		},
		{
			testName:     "error domain parent to class not in domain",
			parentKey:    domain1Key,
			fromClassKey: class1Key,
			toClassKey:   class4Key, // class4 is in domain2, not domain1.
			name:         "Test",
			errstr:       "to class key 'domain/domain2/subdomain/subdomain3/class/class4' is not in domain",
		},

		// Errors: Domain parent - classes in same subdomain (should use subdomain parent).
		{
			testName:     "error domain parent classes in same subdomain",
			parentKey:    domain1Key,
			fromClassKey: class1Key,
			toClassKey:   class2Key, // Both in subdomain1.
			name:         "Test",
			errstr:       "classes are in the same subdomain 'subdomain1', use subdomain as parent instead",
		},

		// Errors: Model parent - classes in same domain (should use domain parent).
		{
			testName:     "error model parent classes in same domain",
			parentKey:    Key{},
			fromClassKey: class1Key,
			toClassKey:   class3Key, // Both in domain1.
			name:         "Test",
			errstr:       "classes are in the same domain 'domain1', use domain as parent instead",
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			key, err := NewClassAssociationKey(tt.parentKey, tt.fromClassKey, tt.toClassKey, tt.name)
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

func (suite *KeyTypeSuite) TestNewStateActionKey() {

	domainKey := helper.Must(NewDomainKey("domain1"))
	subdomainKey := helper.Must(NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(NewClassKey(subdomainKey, "class1"))
	stateKey := helper.Must(NewStateKey(classKey, "state1"))

	tests := []struct {
		testName string
		stateKey Key
		when     string
		subKey   string
		expected Key
		errstr   string
	}{
		// OK.
		{
			testName: "ok entry",
			stateKey: stateKey,
			when:     "entry",
			subKey:   "action1",
			expected: helper.Must(newKey(stateKey.String(), KEY_TYPE_STATE_ACTION, "entry/action1")),
		},
		{
			testName: "ok exit",
			stateKey: stateKey,
			when:     "exit",
			subKey:   "action2",
			expected: helper.Must(newKey(stateKey.String(), KEY_TYPE_STATE_ACTION, "exit/action2")),
		},
		{
			testName: "ok do",
			stateKey: stateKey,
			when:     "do",
			subKey:   "action3",
			expected: helper.Must(newKey(stateKey.String(), KEY_TYPE_STATE_ACTION, "do/action3")),
		},

		// Errors.
		{
			testName: "error empty parent",
			stateKey: Key{},
			when:     "entry",
			subKey:   "action1",
			errstr:   "parent key cannot be of type '' for 'saction' key",
		},
		{
			testName: "error wrong parent type",
			stateKey: helper.Must(NewActorKey("actor1")),
			when:     "entry",
			subKey:   "action1",
			errstr:   "parent key cannot be of type 'actor' for 'saction' key",
		},
		{
			testName: "error empty when",
			stateKey: stateKey,
			when:     "",
			subKey:   "action1",
			errstr:   "when cannot be empty for state action key",
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			key, err := NewStateActionKey(tt.stateKey, tt.when, tt.subKey)
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

func (suite *KeyTypeSuite) TestNewTransitionKey() {

	domainKey := helper.Must(NewDomainKey("domain1"))
	subdomainKey := helper.Must(NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(NewClassKey(subdomainKey, "class1"))

	tests := []struct {
		testName string
		classKey Key
		from     string
		event    string
		guard    string
		action   string
		to       string
		expected Key
		errstr   string
	}{
		// OK.
		{
			testName: "ok all fields",
			classKey: classKey,
			from:     "state1",
			event:    "event1",
			guard:    "guard1",
			action:   "action1",
			to:       "state2",
			expected: helper.Must(newKey(classKey.String(), KEY_TYPE_TRANSITION, "state1/event1/guard1/action1/state2")),
		},
		{
			testName: "ok from blank defaults to initial",
			classKey: classKey,
			from:     "",
			event:    "event1",
			guard:    "guard1",
			action:   "action1",
			to:       "state2",
			expected: helper.Must(newKey(classKey.String(), KEY_TYPE_TRANSITION, "initial/event1/guard1/action1/state2")),
		},
		{
			testName: "ok to blank defaults to final",
			classKey: classKey,
			from:     "state1",
			event:    "event1",
			guard:    "guard1",
			action:   "action1",
			to:       "",
			expected: helper.Must(newKey(classKey.String(), KEY_TYPE_TRANSITION, "state1/event1/guard1/action1/final")),
		},
		{
			testName: "ok empty guard and action",
			classKey: classKey,
			from:     "state1",
			event:    "event1",
			guard:    "",
			action:   "",
			to:       "state2",
			expected: helper.Must(newKey(classKey.String(), KEY_TYPE_TRANSITION, "state1/event1///state2")),
		},

		// Errors.
		{
			testName: "error empty parent",
			classKey: Key{},
			from:     "state1",
			event:    "event1",
			guard:    "",
			action:   "",
			to:       "state2",
			errstr:   "parent key cannot be of type '' for 'transition' key",
		},
		{
			testName: "error wrong parent type",
			classKey: helper.Must(NewActorKey("actor1")),
			from:     "state1",
			event:    "event1",
			guard:    "",
			action:   "",
			to:       "state2",
			errstr:   "parent key cannot be of type 'actor' for 'transition' key",
		},
		{
			testName: "error empty event",
			classKey: classKey,
			from:     "state1",
			event:    "",
			guard:    "",
			action:   "",
			to:       "state2",
			errstr:   "event cannot be empty for transition key",
		},
		{
			testName: "error both from and to blank (initial to final)",
			classKey: classKey,
			from:     "",
			event:    "event1",
			guard:    "",
			action:   "",
			to:       "",
			errstr:   "cannot transition directly from initial to final",
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			key, err := NewTransitionKey(tt.classKey, tt.from, tt.event, tt.guard, tt.action, tt.to)
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
