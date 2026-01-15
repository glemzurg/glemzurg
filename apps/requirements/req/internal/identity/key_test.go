package identity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestKeySuite(t *testing.T) {
	suite.Run(t, new(KeySuite))
}

type KeySuite struct {
	suite.Suite
}

func (suite *KeySuite) TestNewKey() {
	tests := []struct {
		testName  string
		parentKey string
		keyType   string
		subKey    string
		expected  Key
		errstr    string
	}{
		// OK cases.
		{
			testName:  "ok root",
			parentKey: "",
			keyType:   "domain",
			subKey:    "rootkey",
			expected:  Key{parentKey: "", keyType: "domain", subKey: "rootkey"},
		},
		{
			testName:  "ok nested",
			parentKey: "domain/domain1",
			keyType:   "subdomain",
			subKey:    "subdomain1",
			expected:  Key{parentKey: "domain/domain1", keyType: "subdomain", subKey: "subdomain1"},
		},
		{
			testName:  "ok with spaces and case insensitivity",
			parentKey: " PARENT ",
			keyType:   "class",
			subKey:    " KEY ",
			expected:  Key{parentKey: "parent", keyType: "class", subKey: "key"},
		},

		// Error cases: verify that validate is being called.
		{
			testName:  "validate being called",
			parentKey: "something",
			keyType:   "", // Trigger validation error.
			subKey:    "somethingelse",
			errstr:    "keyType: cannot be blank.",
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			key, err := newKey(tt.parentKey, tt.keyType, tt.subKey)
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

func (suite *KeySuite) TestParseKey() {
	solution1SubKey := "solution1"
	tests := []struct {
		testName string
		input    string
		expected Key
		errstr   string
	}{
		// OK cases.
		{
			testName: "ok simple",
			input:    "domain/domain1",
			expected: Key{parentKey: "", keyType: "domain", subKey: "domain1"},
		},
		{
			testName: "ok nested",
			input:    "domain/domain1/subdomain/subdomain1",
			expected: Key{parentKey: "domain/domain1", keyType: "subdomain", subKey: "subdomain1"},
		},
		{
			testName: "ok deep",
			input:    "domain/domain1/subdomain/subdomain1/class/thing1",
			expected: Key{parentKey: "domain/domain1/subdomain/subdomain1", keyType: "class", subKey: "thing1"},
		},
		{
			testName: "ok with spaces",
			input:    " DOMAIN / DOMAIN1  /  SUBDOMAIN  /  SUBDOMAIN1  ", // with spaces
			expected: Key{parentKey: "domain/domain1", keyType: "subdomain", subKey: "subdomain1"},
		},
		{
			testName: "ok domain association with subKey2",
			input:    "domain/problem1/dassociation/problem1/solution1",
			expected: Key{parentKey: "domain/problem1", keyType: "dassociation", subKey: "problem1", subKey2: &solution1SubKey},
		},

		// Error cases: invalid format.
		{
			testName: "error empty",
			input:    "", // empty string
			errstr:   "invalid key format",
		},
		{
			testName: "error empty keyType",
			input:    "domain/domain1/subdomain/subdomain1//thing1", // empty keyType
			errstr:   "keyType: cannot be blank.",
		},
		{
			testName: "error unknown keyType",
			input:    "domain/domain1/subdomain/subdomain1/unknown/thing1", // unknown keyType
			errstr:   "keyType: must be a valid value.",
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			key, err := ParseKey(tt.input)
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

func (suite *KeySuite) TestString() {
	subKey2 := "solution1"
	tests := []struct {
		testName string
		key      Key
		expected string
	}{
		{
			testName: "with parent",
			key:      Key{parentKey: "domain/domain1", keyType: "class", subKey: "thing1"},
			expected: "domain/domain1/class/thing1",
		},
		{
			testName: "root",
			key:      Key{parentKey: "", keyType: "domain", subKey: "domain1"},
			expected: "domain/domain1",
		},
		{
			testName: "with subKey2",
			key:      Key{parentKey: "domain/problem1", keyType: "dassociation", subKey: "problem1", subKey2: &subKey2},
			expected: "domain/problem1/dassociation/problem1/solution1",
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.key.String())
		})
		if !pass {
			break
		}
	}
}

func (suite *KeySuite) TestValidate() {
	tests := []struct {
		testName string
		key      Key
		errstr   string
	}{
		// OK cases, test for each key type.
		{
			testName: "ok actor",
			key:      Key{parentKey: "", keyType: "actor", subKey: "actor1"},
		},
		{
			testName: "ok domain",
			key:      Key{parentKey: "", keyType: "domain", subKey: "domain1"},
		},
		{
			testName: "ok domain association",
			key:      Key{parentKey: "domain/domain1", keyType: "dassociation", subKey: "1"},
		},
		{
			testName: "ok subdomain",
			key:      Key{parentKey: "domain/domain1", keyType: "subdomain", subKey: "subdomain1"},
		},
		{
			testName: "ok use case",
			key:      Key{parentKey: ".../subdomain/subdomain1", keyType: "usecase", subKey: "usecase1"},
		},

		{
			testName: "ok class",
			key:      Key{parentKey: ".../subdomain/subdomain1", keyType: "class", subKey: "thing1"},
		},

		// Error cases for all keys.
		{
			testName: "error no subKey",
			key:      Key{parentKey: "domain/domain1", keyType: "subdomain", subKey: ""},
			errstr:   "cannot be blank",
		},
		{
			testName: "error no keyType",
			key:      Key{parentKey: "domain/domain1", keyType: "", subKey: "subdomain1"},
			errstr:   "cannot be blank",
		},
		{
			testName: "error invalid keyType",
			key:      Key{parentKey: "domain/domain1", keyType: "unknown", subKey: "something1"},
			errstr:   "keyType: must be a valid value.",
		},

		// Error cases: specific key types.
		{
			testName: "error parentKey for actor",
			key:      Key{parentKey: "notallowed", keyType: "actor", subKey: "actor1"},
			errstr:   "parentKey: parentKey must be blank for 'actor' keys, cannot be 'notallowed'.",
		},
		{
			testName: "error parentKey for domain",
			key:      Key{parentKey: "notallowed", keyType: "domain", subKey: "domain1"},
			errstr:   "parentKey: parentKey must be blank for 'domain' keys, cannot be 'notallowed'.",
		},
		{
			testName: "error missing parentKey for domain association",
			key:      Key{parentKey: "", keyType: "dassociation", subKey: "1"},
			errstr:   "parentKey: parentKey must be non-blank for 'dassociation' keys.",
		},
		{
			testName: "error missing parentKey for subdomain",
			key:      Key{parentKey: "", keyType: "subdomain", subKey: "subdomain1"},
			errstr:   "parentKey: parentKey must be non-blank for 'subdomain' keys.",
		},
		{
			testName: "error missing parentKey for usecase",
			key:      Key{parentKey: "", keyType: "usecase", subKey: "usecase1"},
			errstr:   "parentKey: parentKey must be non-blank for 'usecase' keys.",
		},
		{
			testName: "error missing parentKey for class",
			key:      Key{parentKey: "", keyType: "class", subKey: "thing1"},
			errstr:   "parentKey: parentKey must be non-blank for 'class' keys.",
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.key.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

func (suite *KeySuite) TestValidateParent() {
	// Create some test keys.
	domainKey, _ := NewDomainKey("testdomain")
	domainKey2, _ := NewDomainKey("testdomain2")
	subdomainKey, _ := NewSubdomainKey(domainKey, "testsubdomain")
	subdomainKey2, _ := NewSubdomainKey(domainKey, "testsubdomain2")
	classKey, _ := NewClassKey(subdomainKey, "testclass")
	classKey2, _ := NewClassKey(subdomainKey2, "testclass2")
	useCaseKey, _ := NewUseCaseKey(subdomainKey, "testusecase")
	scenarioKey, _ := NewScenarioKey(useCaseKey, "testscenario")
	stateKey, _ := NewStateKey(classKey, "teststate")
	actorKey, _ := NewActorKey("testactor")

	// Class associations at different levels.
	subdomainCassocKey, _ := NewClassAssociationKey(subdomainKey, classKey, classKey)
	domainCassocKey, _ := NewClassAssociationKey(domainKey, classKey, classKey2)

	// For model-level class association, we need classes from different domains.
	domainKey3, _ := NewDomainKey("testdomain3")
	subdomainKey3, _ := NewSubdomainKey(domainKey3, "testsubdomain3")
	classKey3, _ := NewClassKey(subdomainKey3, "testclass3")
	modelCassocKey, _ := NewClassAssociationKey(Key{}, classKey, classKey3)

	tests := []struct {
		testName string
		key      Key
		parent   *Key
		errstr   string
	}{
		// Root keys (no parent).
		{
			testName: "ok actor nil parent",
			key:      actorKey,
			parent:   nil,
		},
		{
			testName: "ok domain nil parent",
			key:      domainKey,
			parent:   nil,
		},
		{
			testName: "error actor with parent",
			key:      actorKey,
			parent:   &domainKey,
			errstr:   "should not have a parent",
		},
		{
			testName: "error domain with parent",
			key:      domainKey,
			parent:   &domainKey2,
			errstr:   "should not have a parent",
		},

		// Subdomain requires domain parent.
		{
			testName: "ok subdomain with domain parent",
			key:      subdomainKey,
			parent:   &domainKey,
		},
		{
			testName: "error subdomain nil parent",
			key:      subdomainKey,
			parent:   nil,
			errstr:   "requires a parent of type 'domain'",
		},
		{
			testName: "error subdomain wrong parent type",
			key:      subdomainKey,
			parent:   &subdomainKey2,
			errstr:   "requires parent of type 'domain', but got 'subdomain'",
		},
		{
			testName: "error subdomain wrong parent key",
			key:      subdomainKey,
			parent:   &domainKey2,
			errstr:   "does not match expected parent",
		},

		// Class requires subdomain parent.
		{
			testName: "ok class with subdomain parent",
			key:      classKey,
			parent:   &subdomainKey,
		},
		{
			testName: "error class nil parent",
			key:      classKey,
			parent:   nil,
			errstr:   "requires a parent of type 'subdomain'",
		},
		{
			testName: "error class wrong parent type",
			key:      classKey,
			parent:   &domainKey,
			errstr:   "requires parent of type 'subdomain', but got 'domain'",
		},

		// State requires class parent.
		{
			testName: "ok state with class parent",
			key:      stateKey,
			parent:   &classKey,
		},
		{
			testName: "error state nil parent",
			key:      stateKey,
			parent:   nil,
			errstr:   "requires a parent of type 'class'",
		},

		// Scenario requires use case parent.
		{
			testName: "ok scenario with usecase parent",
			key:      scenarioKey,
			parent:   &useCaseKey,
		},
		{
			testName: "error scenario nil parent",
			key:      scenarioKey,
			parent:   nil,
			errstr:   "requires a parent of type 'usecase'",
		},

		// Class association at subdomain level.
		{
			testName: "ok subdomain cassoc with subdomain parent",
			key:      subdomainCassocKey,
			parent:   &subdomainKey,
		},
		{
			testName: "error subdomain cassoc with nil parent",
			key:      subdomainCassocKey,
			parent:   nil,
			errstr:   "requires a parent of type 'subdomain'",
		},
		{
			testName: "error subdomain cassoc with domain parent",
			key:      subdomainCassocKey,
			parent:   &domainKey,
			errstr:   "requires parent of type 'subdomain', but got 'domain'",
		},

		// Class association at domain level.
		{
			testName: "ok domain cassoc with domain parent",
			key:      domainCassocKey,
			parent:   &domainKey,
		},
		{
			testName: "error domain cassoc with nil parent",
			key:      domainCassocKey,
			parent:   nil,
			errstr:   "requires a parent of type 'domain'",
		},
		{
			testName: "error domain cassoc with subdomain parent",
			key:      domainCassocKey,
			parent:   &subdomainKey,
			errstr:   "requires parent of type 'domain', but got 'subdomain'",
		},

		// Class association at model level.
		{
			testName: "ok model cassoc with nil parent",
			key:      modelCassocKey,
			parent:   nil,
		},
		{
			testName: "error model cassoc with domain parent",
			key:      modelCassocKey,
			parent:   &domainKey,
			errstr:   "should not have a parent",
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.key.ValidateParent(tt.parent)
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
		if !pass {
			break
		}
	}
}
