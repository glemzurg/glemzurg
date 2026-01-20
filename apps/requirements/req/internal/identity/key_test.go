package identity

import (
	"encoding/json"
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
			expected: Key{parentKey: "domain/problem1", keyType: "dassociation", subKey: "problem1", subKey2: "solution1"},
		},
		// Class association with subdomain parent.
		// Format: domain/d/subdomain/s/cassociation/class/class_a/class/class_b
		{
			testName: "ok class association with subdomain parent",
			input:    "domain/domain_a/subdomain/subdomain_a/cassociation/class/class_a/class/class_b",
			expected: Key{parentKey: "domain/domain_a/subdomain/subdomain_a", keyType: "cassociation", subKey: "class/class_a", subKey2: "class/class_b"},
		},
		// Class association with domain parent.
		// Format: domain/d/cassociation/subdomain/s_a/class/c_a/subdomain/s_b/class/c_b
		{
			testName: "ok class association with domain parent",
			input:    "domain/domain_a/cassociation/subdomain/subdomain_a/class/class_a/subdomain/subdomain_b/class/class_b",
			expected: Key{parentKey: "domain/domain_a", keyType: "cassociation", subKey: "subdomain/subdomain_a/class/class_a", subKey2: "subdomain/subdomain_b/class/class_b"},
		},
		// Class association with model parent (no parent).
		// Format: cassociation/domain/d_a/subdomain/s_a/class/c_a/domain/d_b/subdomain/s_b/class/c_b
		{
			testName: "ok class association with model parent",
			input:    "cassociation/domain/domain_a/subdomain/subdomain_a/class/class_a/domain/domain_b/subdomain/subdomain_b/class/class_b",
			expected: Key{parentKey: "", keyType: "cassociation", subKey: "domain/domain_a/subdomain/subdomain_a/class/class_a", subKey2: "domain/domain_b/subdomain/subdomain_b/class/class_b"},
		},

		// State action key with composite subKey (when/subKey).
		{
			testName: "ok state action key",
			input:    "domain/domain_key/subdomain/subdomain_key/class/class_key/state/state_key/saction/entry/key",
			expected: Key{parentKey: "domain/domain_key/subdomain/subdomain_key/class/class_key/state/state_key", keyType: "saction", subKey: "entry/key"},
		},
		{
			testName: "ok state action key with exit",
			input:    "domain/domain_key/subdomain/subdomain_key/class/class_key/state/state_key/saction/exit/key_b",
			expected: Key{parentKey: "domain/domain_key/subdomain/subdomain_key/class/class_key/state/state_key", keyType: "saction", subKey: "exit/key_b"},
		},
		{
			testName: "ok state action key with do",
			input:    "domain/domain_key/subdomain/subdomain_key/class/class_key/state/state_key/saction/do/action_name",
			expected: Key{parentKey: "domain/domain_key/subdomain/subdomain_key/class/class_key/state/state_key", keyType: "saction", subKey: "do/action_name"},
		},
		// Transition key with composite subKey (from/event/guard/action/to).
		{
			testName: "ok transition key",
			input:    "domain/domain_key/subdomain/subdomain_key/class/class_key/transition/state_a/event_key/guard_key/action_key/state_b",
			expected: Key{parentKey: "domain/domain_key/subdomain/subdomain_key/class/class_key", keyType: "transition", subKey: "state_a/event_key/guard_key/action_key/state_b"},
		},
		{
			testName: "ok transition key with different states",
			input:    "domain/d1/subdomain/s1/class/c1/transition/from_state/my_event/my_guard/my_action/to_state",
			expected: Key{parentKey: "domain/d1/subdomain/s1/class/c1", keyType: "transition", subKey: "from_state/my_event/my_guard/my_action/to_state"},
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
			key:      Key{parentKey: "domain/problem1", keyType: "dassociation", subKey: "problem1", subKey2: "solution1"},
			expected: "domain/problem1/dassociation/problem1/solution1",
		},
		// Class association with subdomain parent.
		{
			testName: "class association with subdomain parent",
			key:      Key{parentKey: "domain/domain_a/subdomain/subdomain_a", keyType: "cassociation", subKey: "class/class_a", subKey2: "class/class_b"},
			expected: "domain/domain_a/subdomain/subdomain_a/cassociation/class/class_a/class/class_b",
		},
		// Class association with domain parent.
		{
			testName: "class association with domain parent",
			key:      Key{parentKey: "domain/domain_a", keyType: "cassociation", subKey: "subdomain/subdomain_a/class/class_a", subKey2: "subdomain/subdomain_b/class/class_b"},
			expected: "domain/domain_a/cassociation/subdomain/subdomain_a/class/class_a/subdomain/subdomain_b/class/class_b",
		},
		// Class association with model parent (no parent).
		{
			testName: "class association with model parent",
			key:      Key{parentKey: "", keyType: "cassociation", subKey: "domain/domain_a/subdomain/subdomain_a/class/class_a", subKey2: "domain/domain_b/subdomain/subdomain_b/class/class_b"},
			expected: "cassociation/domain/domain_a/subdomain/subdomain_a/class/class_a/domain/domain_b/subdomain/subdomain_b/class/class_b",
		},
		// State action key with composite subKey.
		{
			testName: "state action key",
			key:      Key{parentKey: "domain/domain_key/subdomain/subdomain_key/class/class_key/state/state_key", keyType: "saction", subKey: "entry/key"},
			expected: "domain/domain_key/subdomain/subdomain_key/class/class_key/state/state_key/saction/entry/key",
		},
		{
			testName: "state action key with exit",
			key:      Key{parentKey: "domain/domain_key/subdomain/subdomain_key/class/class_key/state/state_key", keyType: "saction", subKey: "exit/key_b"},
			expected: "domain/domain_key/subdomain/subdomain_key/class/class_key/state/state_key/saction/exit/key_b",
		},
		// Transition key with composite subKey.
		{
			testName: "transition key",
			key:      Key{parentKey: "domain/domain_key/subdomain/subdomain_key/class/class_key", keyType: "transition", subKey: "state_a/event_key/guard_key/action_key/state_b"},
			expected: "domain/domain_key/subdomain/subdomain_key/class/class_key/transition/state_a/event_key/guard_key/action_key/state_b",
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

func (suite *KeySuite) TestHasNoParent() {
	// Create hierarchy of keys.
	domainKey, _ := NewDomainKey("testdomain")
	subdomainKey, _ := NewSubdomainKey(domainKey, "testsubdomain")
	classKey, _ := NewClassKey(subdomainKey, "testclass")
	stateKey, _ := NewStateKey(classKey, "teststate")
	actorKey, _ := NewActorKey("testactor")

	tests := []struct {
		testName string
		key      Key
		expected bool
	}{
		// Root keys (no parent).
		{
			testName: "domain has no parent",
			key:      domainKey,
			expected: true,
		},
		{
			testName: "actor has no parent",
			key:      actorKey,
			expected: true,
		},

		// Keys with parents.
		{
			testName: "subdomain has parent",
			key:      subdomainKey,
			expected: false,
		},
		{
			testName: "class has parent",
			key:      classKey,
			expected: false,
		},
		{
			testName: "state has parent",
			key:      stateKey,
			expected: false,
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			result := tt.key.HasNoParent()
			assert.Equal(t, tt.expected, result)
		})
		if !pass {
			break
		}
	}
}

func (suite *KeySuite) TestIsParent() {
	// Create hierarchy of keys.
	domainKey, _ := NewDomainKey("testdomain")
	domainKey2, _ := NewDomainKey("otherdomain")
	subdomainKey, _ := NewSubdomainKey(domainKey, "testsubdomain")
	classKey, _ := NewClassKey(subdomainKey, "testclass")
	stateKey, _ := NewStateKey(classKey, "teststate")
	actorKey, _ := NewActorKey("testactor")

	tests := []struct {
		testName  string
		key       Key
		parentKey Key
		expected  bool
	}{
		// Direct parent relationships.
		{
			testName:  "subdomain has domain as parent",
			key:       subdomainKey,
			parentKey: domainKey,
			expected:  true,
		},
		{
			testName:  "class has subdomain as parent",
			key:       classKey,
			parentKey: subdomainKey,
			expected:  true,
		},
		{
			testName:  "state has class as parent",
			key:       stateKey,
			parentKey: classKey,
			expected:  true,
		},

		// Ancestor relationships (grandparent, etc.).
		{
			testName:  "class has domain as ancestor",
			key:       classKey,
			parentKey: domainKey,
			expected:  true,
		},
		{
			testName:  "state has subdomain as ancestor",
			key:       stateKey,
			parentKey: subdomainKey,
			expected:  true,
		},
		{
			testName:  "state has domain as ancestor",
			key:       stateKey,
			parentKey: domainKey,
			expected:  true,
		},

		// Not parent relationships.
		{
			testName:  "domain is not parent of itself",
			key:       domainKey,
			parentKey: domainKey,
			expected:  false,
		},
		{
			testName:  "different domain is not parent",
			key:       subdomainKey,
			parentKey: domainKey2,
			expected:  false,
		},
		{
			testName:  "child is not parent of parent",
			key:       domainKey,
			parentKey: subdomainKey,
			expected:  false,
		},
		{
			testName:  "unrelated keys are not parent-child",
			key:       actorKey,
			parentKey: domainKey,
			expected:  false,
		},
		{
			testName:  "sibling classes not parent-child",
			key:       classKey,
			parentKey: classKey,
			expected:  false,
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			result := tt.key.IsParent(tt.parentKey)
			assert.Equal(t, tt.expected, result)
		})
		if !pass {
			break
		}
	}
}

// TestParseKeyRoundTrip tests that keys created with New* functions
// can be converted to string and parsed back successfully.
func (suite *KeySuite) TestParseKeyRoundTrip() {
	// Create hierarchy of keys.
	domainKey, _ := NewDomainKey("domain_key")
	subdomainKey, _ := NewSubdomainKey(domainKey, "subdomain_key")
	classKey, _ := NewClassKey(subdomainKey, "class_key")
	stateKey, _ := NewStateKey(classKey, "state_key")

	tests := []struct {
		testName    string
		createKey   func() (Key, error)
		description string
	}{
		{
			testName: "state action key round trip",
			createKey: func() (Key, error) {
				return NewStateActionKey(stateKey, "entry", "key")
			},
			description: "StateAction key with entry/key subKey",
		},
		{
			testName: "state action key with exit round trip",
			createKey: func() (Key, error) {
				return NewStateActionKey(stateKey, "exit", "key_b")
			},
			description: "StateAction key with exit/key_b subKey",
		},
		{
			testName: "state action key with do round trip",
			createKey: func() (Key, error) {
				return NewStateActionKey(stateKey, "do", "action_name")
			},
			description: "StateAction key with do/action_name subKey",
		},
		{
			testName: "transition key round trip",
			createKey: func() (Key, error) {
				return NewTransitionKey(classKey, "state_a", "event_key", "guard_key", "action_key", "state_b")
			},
			description: "Transition key with from/event/guard/action/to subKey",
		},
		{
			testName: "transition key with different parts round trip",
			createKey: func() (Key, error) {
				return NewTransitionKey(classKey, "from_state", "my_event", "my_guard", "my_action", "to_state")
			},
			description: "Transition key with various state/event/guard/action names",
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			// Create the key.
			originalKey, err := tt.createKey()
			assert.NoError(t, err, "Failed to create key for: %s", tt.description)

			// Convert to string.
			keyStr := originalKey.String()
			assert.NotEmpty(t, keyStr, "Key string should not be empty for: %s", tt.description)

			// Parse the string back.
			parsedKey, err := ParseKey(keyStr)
			assert.NoError(t, err, "Failed to parse key string '%s' for: %s", keyStr, tt.description)

			// Verify the parsed key matches the original.
			assert.Equal(t, originalKey.parentKey, parsedKey.parentKey, "ParentKey mismatch for: %s", tt.description)
			assert.Equal(t, originalKey.keyType, parsedKey.keyType, "KeyType mismatch for: %s", tt.description)
			assert.Equal(t, originalKey.subKey, parsedKey.subKey, "SubKey mismatch for: %s", tt.description)
			assert.Equal(t, originalKey.String(), parsedKey.String(), "String() mismatch for: %s", tt.description)
		})
		if !pass {
			break
		}
	}
}

// TestJSONRoundTrip tests that keys can be marshalled to JSON and unmarshalled back.
func (suite *KeySuite) TestJSONRoundTrip() {
	// Create hierarchy of keys.
	domainKey, _ := NewDomainKey("domain_key")
	subdomainKey, _ := NewSubdomainKey(domainKey, "subdomain_key")
	classKey, _ := NewClassKey(subdomainKey, "class_key")
	classKey2, _ := NewClassKey(subdomainKey, "class_key2")
	stateKey, _ := NewStateKey(classKey, "state_key")
	useCaseKey, _ := NewUseCaseKey(subdomainKey, "use_case_key")
	scenarioKey, _ := NewScenarioKey(useCaseKey, "scenario_key")
	actorKey, _ := NewActorKey("actor_key")

	// State action and transition keys.
	stateActionKey, _ := NewStateActionKey(stateKey, "entry", "action_key")
	transitionKey, _ := NewTransitionKey(classKey, "state_a", "event_key", "guard_key", "action_key", "state_b")

	// Domain association key.
	domainKey2, _ := NewDomainKey("domain_key2")
	domainAssocKey, _ := NewDomainAssociationKey(domainKey, domainKey2)

	// Class association key.
	classAssocKey, _ := NewClassAssociationKey(subdomainKey, classKey, classKey2)

	tests := []struct {
		testName string
		key      Key
	}{
		{testName: "domain key", key: domainKey},
		{testName: "subdomain key", key: subdomainKey},
		{testName: "class key", key: classKey},
		{testName: "state key", key: stateKey},
		{testName: "actor key", key: actorKey},
		{testName: "use case key", key: useCaseKey},
		{testName: "scenario key", key: scenarioKey},
		{testName: "state action key", key: stateActionKey},
		{testName: "transition key", key: transitionKey},
		{testName: "domain association key", key: domainAssocKey},
		{testName: "class association key", key: classAssocKey},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			// Marshal to JSON.
			jsonBytes, err := json.Marshal(tt.key)
			assert.NoError(t, err, "Failed to marshal key to JSON")

			// Verify the JSON is a string (quoted).
			jsonStr := string(jsonBytes)
			assert.True(t, len(jsonStr) >= 2 && jsonStr[0] == '"' && jsonStr[len(jsonStr)-1] == '"',
				"JSON should be a quoted string, got: %s", jsonStr)

			// Unmarshal back.
			var parsedKey Key
			err = json.Unmarshal(jsonBytes, &parsedKey)
			assert.NoError(t, err, "Failed to unmarshal key from JSON")

			// Verify the parsed key matches the original.
			assert.Equal(t, tt.key, parsedKey, "Round-trip key mismatch")
			assert.Equal(t, tt.key.String(), parsedKey.String(), "String() mismatch after round-trip")
		})
		if !pass {
			break
		}
	}
}

// TestJSONUnmarshalEmpty tests that unmarshalling an empty string results in a zero-value Key.
func (suite *KeySuite) TestJSONUnmarshalEmpty() {
	var key Key
	err := json.Unmarshal([]byte(`""`), &key)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), Key{}, key)
}

// TestJSONUnmarshalInvalid tests that unmarshalling invalid JSON or key strings returns errors.
func (suite *KeySuite) TestJSONUnmarshalInvalid() {
	tests := []struct {
		testName string
		jsonStr  string
		errstr   string
	}{
		{
			testName: "invalid json",
			jsonStr:  `not json`,
			errstr:   "invalid character",
		},
		{
			testName: "invalid key format",
			jsonStr:  `"invalid"`,
			errstr:   "invalid key format",
		},
		{
			testName: "unknown key type",
			jsonStr:  `"unknown/something"`,
			errstr:   "must be a valid value",
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			var key Key
			err := json.Unmarshal([]byte(tt.jsonStr), &key)
			assert.ErrorContains(t, err, tt.errstr)
		})
		if !pass {
			break
		}
	}
}

// TestTextMarshalRoundTrip tests that keys can be marshalled to text and unmarshalled back.
// This is required for Key to be used as a map key in JSON marshalling/unmarshalling.
func (suite *KeySuite) TestTextMarshalRoundTrip() {
	// Create hierarchy of keys.
	domainKey, _ := NewDomainKey("domain_key")
	subdomainKey, _ := NewSubdomainKey(domainKey, "subdomain_key")
	classKey, _ := NewClassKey(subdomainKey, "class_key")
	classKey2, _ := NewClassKey(subdomainKey, "class_key2")
	stateKey, _ := NewStateKey(classKey, "state_key")
	useCaseKey, _ := NewUseCaseKey(subdomainKey, "use_case_key")
	scenarioKey, _ := NewScenarioKey(useCaseKey, "scenario_key")
	actorKey, _ := NewActorKey("actor_key")

	// State action and transition keys.
	stateActionKey, _ := NewStateActionKey(stateKey, "entry", "action_key")
	transitionKey, _ := NewTransitionKey(classKey, "state_a", "event_key", "guard_key", "action_key", "state_b")

	// Domain association key.
	domainKey2, _ := NewDomainKey("domain_key2")
	domainAssocKey, _ := NewDomainAssociationKey(domainKey, domainKey2)

	// Class association key.
	classAssocKey, _ := NewClassAssociationKey(subdomainKey, classKey, classKey2)

	tests := []struct {
		testName string
		key      Key
	}{
		{testName: "domain key", key: domainKey},
		{testName: "subdomain key", key: subdomainKey},
		{testName: "class key", key: classKey},
		{testName: "state key", key: stateKey},
		{testName: "actor key", key: actorKey},
		{testName: "use case key", key: useCaseKey},
		{testName: "scenario key", key: scenarioKey},
		{testName: "state action key", key: stateActionKey},
		{testName: "transition key", key: transitionKey},
		{testName: "domain association key", key: domainAssocKey},
		{testName: "class association key", key: classAssocKey},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			// Marshal to text.
			textBytes, err := tt.key.MarshalText()
			assert.NoError(t, err, "Failed to marshal key to text")

			// Verify the text matches the String() output.
			assert.Equal(t, tt.key.String(), string(textBytes), "MarshalText should return String()")

			// Unmarshal back.
			var parsedKey Key
			err = parsedKey.UnmarshalText(textBytes)
			assert.NoError(t, err, "Failed to unmarshal key from text")

			// Verify the parsed key matches the original.
			assert.Equal(t, tt.key, parsedKey, "Round-trip key mismatch")
			assert.Equal(t, tt.key.String(), parsedKey.String(), "String() mismatch after round-trip")
		})
		if !pass {
			break
		}
	}
}

// TestTextUnmarshalEmpty tests that unmarshalling an empty string results in a zero-value Key.
func (suite *KeySuite) TestTextUnmarshalEmpty() {
	var key Key
	err := key.UnmarshalText([]byte(""))
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), Key{}, key)
}

// TestTextUnmarshalInvalid tests that unmarshalling invalid key strings returns errors.
func (suite *KeySuite) TestTextUnmarshalInvalid() {
	tests := []struct {
		testName string
		text     string
		errstr   string
	}{
		{
			testName: "invalid key format",
			text:     "invalid",
			errstr:   "invalid key format",
		},
		{
			testName: "unknown key type",
			text:     "unknown/something",
			errstr:   "must be a valid value",
		},
	}
	for _, tt := range tests {
		pass := suite.T().Run(tt.testName, func(t *testing.T) {
			var key Key
			err := key.UnmarshalText([]byte(tt.text))
			assert.ErrorContains(t, err, tt.errstr)
		})
		if !pass {
			break
		}
	}
}

// TestJSONMapKeyRoundTrip tests that Key can be used as a map key in JSON marshalling/unmarshalling.
// This verifies that MarshalText and UnmarshalText work correctly for map keys.
func (suite *KeySuite) TestJSONMapKeyRoundTrip() {
	// Create test keys.
	domainKey, _ := NewDomainKey("domain_key")
	subdomainKey, _ := NewSubdomainKey(domainKey, "subdomain_key")
	classKey, _ := NewClassKey(subdomainKey, "class_key")
	stateKey, _ := NewStateKey(classKey, "state_key")

	// Create a map with Key as the key type.
	originalMap := map[Key]string{
		domainKey:    "domain value",
		subdomainKey: "subdomain value",
		classKey:     "class value",
		stateKey:     "state value",
	}

	// Marshal the map to JSON.
	jsonBytes, err := json.Marshal(originalMap)
	assert.NoError(suite.T(), err, "Failed to marshal map to JSON")

	// Unmarshal back.
	var parsedMap map[Key]string
	err = json.Unmarshal(jsonBytes, &parsedMap)
	assert.NoError(suite.T(), err, "Failed to unmarshal map from JSON")

	// Verify the parsed map matches the original.
	assert.Equal(suite.T(), len(originalMap), len(parsedMap), "Map length mismatch")
	for key, value := range originalMap {
		parsedValue, ok := parsedMap[key]
		assert.True(suite.T(), ok, "Key not found in parsed map: %s", key.String())
		assert.Equal(suite.T(), value, parsedValue, "Value mismatch for key: %s", key.String())
	}
}
