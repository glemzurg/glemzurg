package identity

import (
	"encoding/json"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
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
			expected:  Key{ParentKey: "", KeyType: "domain", SubKey: "rootkey"},
		},
		{
			testName:  "ok nested",
			parentKey: "domain/domain1",
			keyType:   "subdomain",
			subKey:    "subdomain1",
			expected:  Key{ParentKey: "domain/domain1", KeyType: "subdomain", SubKey: "subdomain1"},
		},
		{
			testName:  "ok with spaces and case insensitivity",
			parentKey: " PARENT ",
			keyType:   "class",
			subKey:    " KEY ",
			expected:  Key{ParentKey: "parent", KeyType: "class", SubKey: "key"},
		},

		// Error cases: verify that validate is being called.
		{
			testName:  "validate being called",
			parentKey: "something",
			keyType:   "", // Trigger validation error.
			subKey:    "somethingelse",
			errstr:    "'KeyType' failed on the 'required' tag",
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
			expected: Key{ParentKey: "", KeyType: "domain", SubKey: "domain1"},
		},
		{
			testName: "ok nested",
			input:    "domain/domain1/subdomain/subdomain1",
			expected: Key{ParentKey: "domain/domain1", KeyType: "subdomain", SubKey: "subdomain1"},
		},
		{
			testName: "ok deep",
			input:    "domain/domain1/subdomain/subdomain1/class/thing1",
			expected: Key{ParentKey: "domain/domain1/subdomain/subdomain1", KeyType: "class", SubKey: "thing1"},
		},
		{
			testName: "ok with spaces",
			input:    " DOMAIN / DOMAIN1  /  SUBDOMAIN  /  SUBDOMAIN1  ", // with spaces
			expected: Key{ParentKey: "domain/domain1", KeyType: "subdomain", SubKey: "subdomain1"},
		},
		{
			testName: "ok domain association with subKey2",
			input:    "dassociation/problem1/solution1",
			expected: Key{ParentKey: "", KeyType: "dassociation", SubKey: "problem1", SubKey2: "solution1"},
		},
		// Class association with subdomain parent.
		// Format: domain/d/subdomain/s/cassociation/class/class_a/class/class_b/name
		{
			testName: "ok class association with subdomain parent",
			input:    "domain/domain_a/subdomain/subdomain_a/cassociation/class/class_a/class/class_b/assoc_name",
			expected: Key{ParentKey: "domain/domain_a/subdomain/subdomain_a", KeyType: "cassociation", SubKey: "class/class_a", SubKey2: "class/class_b", SubKey3: "assoc_name"},
		},
		// Class association with domain parent.
		// Format: domain/d/cassociation/subdomain/s_a/class/c_a/subdomain/s_b/class/c_b/name
		{
			testName: "ok class association with domain parent",
			input:    "domain/domain_a/cassociation/subdomain/subdomain_a/class/class_a/subdomain/subdomain_b/class/class_b/assoc_name",
			expected: Key{ParentKey: "domain/domain_a", KeyType: "cassociation", SubKey: "subdomain/subdomain_a/class/class_a", SubKey2: "subdomain/subdomain_b/class/class_b", SubKey3: "assoc_name"},
		},
		// Class association with model parent (no parent).
		// Format: cassociation/domain/d_a/subdomain/s_a/class/c_a/domain/d_b/subdomain/s_b/class/c_b/name
		{
			testName: "ok class association with model parent",
			input:    "cassociation/domain/domain_a/subdomain/subdomain_a/class/class_a/domain/domain_b/subdomain/subdomain_b/class/class_b/assoc_name",
			expected: Key{ParentKey: "", KeyType: "cassociation", SubKey: "domain/domain_a/subdomain/subdomain_a/class/class_a", SubKey2: "domain/domain_b/subdomain/subdomain_b/class/class_b", SubKey3: "assoc_name"},
		},

		// State action key with composite subKey (when/subKey).
		{
			testName: "ok state action key",
			input:    "domain/domain_key/subdomain/subdomain_key/class/class_key/state/state_key/saction/entry/key",
			expected: Key{ParentKey: "domain/domain_key/subdomain/subdomain_key/class/class_key/state/state_key", KeyType: "saction", SubKey: "entry/key"},
		},
		{
			testName: "ok state action key with exit",
			input:    "domain/domain_key/subdomain/subdomain_key/class/class_key/state/state_key/saction/exit/key_b",
			expected: Key{ParentKey: "domain/domain_key/subdomain/subdomain_key/class/class_key/state/state_key", KeyType: "saction", SubKey: "exit/key_b"},
		},
		{
			testName: "ok state action key with do",
			input:    "domain/domain_key/subdomain/subdomain_key/class/class_key/state/state_key/saction/do/action_name",
			expected: Key{ParentKey: "domain/domain_key/subdomain/subdomain_key/class/class_key/state/state_key", KeyType: "saction", SubKey: "do/action_name"},
		},
		// Transition key with composite subKey (from/event/guard/action/to).
		{
			testName: "ok transition key",
			input:    "domain/domain_key/subdomain/subdomain_key/class/class_key/transition/state_a/event_key/guard_key/action_key/state_b",
			expected: Key{ParentKey: "domain/domain_key/subdomain/subdomain_key/class/class_key", KeyType: "transition", SubKey: "state_a/event_key/guard_key/action_key/state_b"},
		},
		{
			testName: "ok transition key with different states",
			input:    "domain/d1/subdomain/s1/class/c1/transition/from_state/my_event/my_guard/my_action/to_state",
			expected: Key{ParentKey: "domain/d1/subdomain/s1/class/c1", KeyType: "transition", SubKey: "from_state/my_event/my_guard/my_action/to_state"},
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
			errstr:   "'KeyType' failed on the 'required' tag",
		},
		{
			testName: "error unknown keyType",
			input:    "domain/domain1/subdomain/subdomain1/unknown/thing1", // unknown keyType
			errstr:   "'KeyType' failed on the 'oneof' tag",
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
			key:      Key{ParentKey: "domain/domain1", KeyType: "class", SubKey: "thing1"},
			expected: "domain/domain1/class/thing1",
		},
		{
			testName: "root",
			key:      Key{ParentKey: "", KeyType: "domain", SubKey: "domain1"},
			expected: "domain/domain1",
		},
		{
			testName: "with subKey2",
			key:      Key{ParentKey: "", KeyType: "dassociation", SubKey: "problem1", SubKey2: "solution1"},
			expected: "dassociation/problem1/solution1",
		},
		// Class association with subdomain parent.
		{
			testName: "class association with subdomain parent",
			key:      Key{ParentKey: "domain/domain_a/subdomain/subdomain_a", KeyType: "cassociation", SubKey: "class/class_a", SubKey2: "class/class_b", SubKey3: "assoc_name"},
			expected: "domain/domain_a/subdomain/subdomain_a/cassociation/class/class_a/class/class_b/assoc_name",
		},
		// Class association with domain parent.
		{
			testName: "class association with domain parent",
			key:      Key{ParentKey: "domain/domain_a", KeyType: "cassociation", SubKey: "subdomain/subdomain_a/class/class_a", SubKey2: "subdomain/subdomain_b/class/class_b", SubKey3: "assoc_name"},
			expected: "domain/domain_a/cassociation/subdomain/subdomain_a/class/class_a/subdomain/subdomain_b/class/class_b/assoc_name",
		},
		// Class association with model parent (no parent).
		{
			testName: "class association with model parent",
			key:      Key{ParentKey: "", KeyType: "cassociation", SubKey: "domain/domain_a/subdomain/subdomain_a/class/class_a", SubKey2: "domain/domain_b/subdomain/subdomain_b/class/class_b", SubKey3: "assoc_name"},
			expected: "cassociation/domain/domain_a/subdomain/subdomain_a/class/class_a/domain/domain_b/subdomain/subdomain_b/class/class_b/assoc_name",
		},
		// State action key with composite subKey.
		{
			testName: "state action key",
			key:      Key{ParentKey: "domain/domain_key/subdomain/subdomain_key/class/class_key/state/state_key", KeyType: "saction", SubKey: "entry/key"},
			expected: "domain/domain_key/subdomain/subdomain_key/class/class_key/state/state_key/saction/entry/key",
		},
		{
			testName: "state action key with exit",
			key:      Key{ParentKey: "domain/domain_key/subdomain/subdomain_key/class/class_key/state/state_key", KeyType: "saction", SubKey: "exit/key_b"},
			expected: "domain/domain_key/subdomain/subdomain_key/class/class_key/state/state_key/saction/exit/key_b",
		},
		// Transition key with composite subKey.
		{
			testName: "transition key",
			key:      Key{ParentKey: "domain/domain_key/subdomain/subdomain_key/class/class_key", KeyType: "transition", SubKey: "state_a/event_key/guard_key/action_key/state_b"},
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
			key:      Key{ParentKey: "", KeyType: "actor", SubKey: "actor1"},
		},
		{
			testName: "ok domain",
			key:      Key{ParentKey: "", KeyType: "domain", SubKey: "domain1"},
		},
		{
			testName: "ok domain association",
			key:      Key{ParentKey: "", KeyType: "dassociation", SubKey: "1", SubKey2: "2"},
		},
		{
			testName: "ok subdomain",
			key:      Key{ParentKey: "domain/domain1", KeyType: "subdomain", SubKey: "subdomain1"},
		},
		{
			testName: "ok use case",
			key:      Key{ParentKey: ".../subdomain/subdomain1", KeyType: "usecase", SubKey: "usecase1"},
		},

		{
			testName: "ok class",
			key:      Key{ParentKey: ".../subdomain/subdomain1", KeyType: "class", SubKey: "thing1"},
		},

		// Error cases for all keys.
		{
			testName: "error no subKey",
			key:      Key{ParentKey: "domain/domain1", KeyType: "subdomain", SubKey: ""},
			errstr:   "'SubKey' failed on the 'required' tag",
		},
		{
			testName: "error no keyType",
			key:      Key{ParentKey: "domain/domain1", KeyType: "", SubKey: "subdomain1"},
			errstr:   "'KeyType' failed on the 'required' tag",
		},
		{
			testName: "error invalid keyType",
			key:      Key{ParentKey: "domain/domain1", KeyType: "unknown", SubKey: "something1"},
			errstr:   "'KeyType' failed on the 'oneof' tag",
		},

		// Error cases: specific key types.
		{
			testName: "error parentKey for actor",
			key:      Key{ParentKey: "notallowed", KeyType: "actor", SubKey: "actor1"},
			errstr:   "parentKey must be blank for 'actor' keys, cannot be 'notallowed'",
		},
		{
			testName: "error parentKey for domain",
			key:      Key{ParentKey: "notallowed", KeyType: "domain", SubKey: "domain1"},
			errstr:   "parentKey must be blank for 'domain' keys, cannot be 'notallowed'",
		},
		{
			testName: "error domain association with parentKey",
			key:      Key{ParentKey: "notallowed", KeyType: "dassociation", SubKey: "1", SubKey2: "2"},
			errstr:   "parentKey must be blank for 'dassociation' keys, cannot be 'notallowed'",
		},
		{
			testName: "error missing parentKey for subdomain",
			key:      Key{ParentKey: "", KeyType: "subdomain", SubKey: "subdomain1"},
			errstr:   "parentKey must be non-blank for 'subdomain' keys",
		},
		{
			testName: "error missing parentKey for usecase",
			key:      Key{ParentKey: "", KeyType: "usecase", SubKey: "usecase1"},
			errstr:   "parentKey must be non-blank for 'usecase' keys",
		},
		{
			testName: "error missing parentKey for class",
			key:      Key{ParentKey: "", KeyType: "class", SubKey: "thing1"},
			errstr:   "parentKey must be non-blank for 'class' keys",
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
	domainKey := helper.Must(NewDomainKey("testdomain"))
	domainKey2 := helper.Must(NewDomainKey("testdomain2"))
	subdomainKey := helper.Must(NewSubdomainKey(domainKey, "testsubdomain"))
	subdomainKey2 := helper.Must(NewSubdomainKey(domainKey, "testsubdomain2"))
	classKey := helper.Must(NewClassKey(subdomainKey, "testclass"))
	classKey2 := helper.Must(NewClassKey(subdomainKey2, "testclass2"))
	useCaseKey := helper.Must(NewUseCaseKey(subdomainKey, "testusecase"))
	generalizationKey := helper.Must(NewGeneralizationKey(subdomainKey, "testgen"))
	scenarioKey := helper.Must(NewScenarioKey(useCaseKey, "testscenario"))
	scenarioObjectKey := helper.Must(NewScenarioObjectKey(scenarioKey, "testsobject"))
	stateKey := helper.Must(NewStateKey(classKey, "teststate"))
	eventKey := helper.Must(NewEventKey(classKey, "testevent"))
	guardKey := helper.Must(NewGuardKey(classKey, "testguard"))
	actionKey := helper.Must(NewActionKey(classKey, "testaction"))
	queryKey := helper.Must(NewQueryKey(classKey, "testquery"))
	transitionKey := helper.Must(NewTransitionKey(classKey, "state_a", "testevent", "", "", "state_b"))
	attributeKey := helper.Must(NewAttributeKey(classKey, "testattr"))
	stateActionKey := helper.Must(NewStateActionKey(stateKey, "entry", "testaction"))
	invariantKey := helper.Must(NewInvariantKey("testinvariant"))
	actionRequireKey := helper.Must(NewActionRequireKey(actionKey, "testreq"))
	actionGuaranteeKey := helper.Must(NewActionGuaranteeKey(actionKey, "testguar"))
	actionSafetyKey := helper.Must(NewActionSafetyKey(actionKey, "testsafety"))
	queryRequireKey := helper.Must(NewQueryRequireKey(queryKey, "testreq"))
	queryGuaranteeKey := helper.Must(NewQueryGuaranteeKey(queryKey, "testguar"))
	attributeDerivationKey := helper.Must(NewAttributeDerivationKey(attributeKey, "testderiv"))
	actorKey := helper.Must(NewActorKey("testactor"))
	actorGeneralizationKey := helper.Must(NewActorGeneralizationKey("testactorgen"))
	globalFuncKey := helper.Must(NewGlobalFunctionKey("_max"))
	domainAssocKey := helper.Must(NewDomainAssociationKey(domainKey, domainKey2))

	// Class associations at different levels.
	subdomainCassocKey := helper.Must(NewClassAssociationKey(subdomainKey, classKey, classKey, "subdomain assoc"))
	domainCassocKey := helper.Must(NewClassAssociationKey(domainKey, classKey, classKey2, "domain assoc"))

	// For model-level class association, we need classes from different domains.
	domainKey3 := helper.Must(NewDomainKey("testdomain3"))
	subdomainKey3 := helper.Must(NewSubdomainKey(domainKey3, "testsubdomain3"))
	classKey3 := helper.Must(NewClassKey(subdomainKey3, "testclass3"))
	modelCassocKey := helper.Must(NewClassAssociationKey(Key{}, classKey, classKey3, "model assoc"))

	// Keys for wrong-parent tests.
	otherActionKey := helper.Must(NewActionKey(classKey, "otheraction"))
	otherQueryKey := helper.Must(NewQueryKey(classKey, "otherquery"))
	otherAttributeKey := helper.Must(NewAttributeKey(classKey, "otherattr"))

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
		{
			testName: "ok domain association nil parent",
			key:      domainAssocKey,
			parent:   nil,
		},
		{
			testName: "error domain association with parent",
			key:      domainAssocKey,
			parent:   &domainKey,
			errstr:   "should not have a parent",
		},
		{
			testName: "ok actor generalization nil parent",
			key:      actorGeneralizationKey,
			parent:   nil,
		},
		{
			testName: "error actor generalization with parent",
			key:      actorGeneralizationKey,
			parent:   &domainKey,
			errstr:   "should not have a parent",
		},

		// Global function is a root key (no parent).
		{
			testName: "ok global function nil parent",
			key:      globalFuncKey,
			parent:   nil,
		},
		{
			testName: "error global function with parent",
			key:      globalFuncKey,
			parent:   &domainKey,
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

		// Invariant is a root key (no parent).
		{
			testName: "ok invariant nil parent",
			key:      invariantKey,
			parent:   nil,
		},
		{
			testName: "error invariant with parent",
			key:      invariantKey,
			parent:   &domainKey,
			errstr:   "should not have a parent",
		},

		// UseCase requires subdomain parent.
		{
			testName: "ok usecase with subdomain parent",
			key:      useCaseKey,
			parent:   &subdomainKey,
		},
		{
			testName: "error usecase nil parent",
			key:      useCaseKey,
			parent:   nil,
			errstr:   "requires a parent of type 'subdomain'",
		},
		{
			testName: "error usecase wrong parent type",
			key:      useCaseKey,
			parent:   &domainKey,
			errstr:   "requires parent of type 'subdomain', but got 'domain'",
		},

		// Generalization requires subdomain parent.
		{
			testName: "ok generalization with subdomain parent",
			key:      generalizationKey,
			parent:   &subdomainKey,
		},
		{
			testName: "error generalization nil parent",
			key:      generalizationKey,
			parent:   nil,
			errstr:   "requires a parent of type 'subdomain'",
		},
		{
			testName: "error generalization wrong parent type",
			key:      generalizationKey,
			parent:   &classKey,
			errstr:   "requires parent of type 'subdomain', but got 'class'",
		},

		// ScenarioObject requires scenario parent.
		{
			testName: "ok scenario object with scenario parent",
			key:      scenarioObjectKey,
			parent:   &scenarioKey,
		},
		{
			testName: "error scenario object nil parent",
			key:      scenarioObjectKey,
			parent:   nil,
			errstr:   "requires a parent of type 'scenario'",
		},
		{
			testName: "error scenario object wrong parent type",
			key:      scenarioObjectKey,
			parent:   &useCaseKey,
			errstr:   "requires parent of type 'scenario', but got 'usecase'",
		},

		// Event requires class parent.
		{
			testName: "ok event with class parent",
			key:      eventKey,
			parent:   &classKey,
		},
		{
			testName: "error event nil parent",
			key:      eventKey,
			parent:   nil,
			errstr:   "requires a parent of type 'class'",
		},
		{
			testName: "error event wrong parent type",
			key:      eventKey,
			parent:   &subdomainKey,
			errstr:   "requires parent of type 'class', but got 'subdomain'",
		},

		// Guard requires class parent.
		{
			testName: "ok guard with class parent",
			key:      guardKey,
			parent:   &classKey,
		},
		{
			testName: "error guard nil parent",
			key:      guardKey,
			parent:   nil,
			errstr:   "requires a parent of type 'class'",
		},

		// Action requires class parent.
		{
			testName: "ok action with class parent",
			key:      actionKey,
			parent:   &classKey,
		},
		{
			testName: "error action nil parent",
			key:      actionKey,
			parent:   nil,
			errstr:   "requires a parent of type 'class'",
		},

		// Query requires class parent.
		{
			testName: "ok query with class parent",
			key:      queryKey,
			parent:   &classKey,
		},
		{
			testName: "error query nil parent",
			key:      queryKey,
			parent:   nil,
			errstr:   "requires a parent of type 'class'",
		},

		// Transition requires class parent.
		{
			testName: "ok transition with class parent",
			key:      transitionKey,
			parent:   &classKey,
		},
		{
			testName: "error transition nil parent",
			key:      transitionKey,
			parent:   nil,
			errstr:   "requires a parent of type 'class'",
		},

		// Attribute requires class parent.
		{
			testName: "ok attribute with class parent",
			key:      attributeKey,
			parent:   &classKey,
		},
		{
			testName: "error attribute nil parent",
			key:      attributeKey,
			parent:   nil,
			errstr:   "requires a parent of type 'class'",
		},

		// StateAction requires state parent.
		{
			testName: "ok state action with state parent",
			key:      stateActionKey,
			parent:   &stateKey,
		},
		{
			testName: "error state action nil parent",
			key:      stateActionKey,
			parent:   nil,
			errstr:   "requires a parent of type 'state'",
		},
		{
			testName: "error state action wrong parent type",
			key:      stateActionKey,
			parent:   &classKey,
			errstr:   "requires parent of type 'state', but got 'class'",
		},

		// ActionRequire requires action parent.
		{
			testName: "ok action require with action parent",
			key:      actionRequireKey,
			parent:   &actionKey,
		},
		{
			testName: "error action require nil parent",
			key:      actionRequireKey,
			parent:   nil,
			errstr:   "requires a parent of type 'action'",
		},
		{
			testName: "error action require wrong parent type",
			key:      actionRequireKey,
			parent:   &classKey,
			errstr:   "requires parent of type 'action', but got 'class'",
		},
		{
			testName: "error action require wrong parent key",
			key:      actionRequireKey,
			parent:   &otherActionKey,
			errstr:   "does not match expected parent",
		},

		// ActionGuarantee requires action parent.
		{
			testName: "ok action guarantee with action parent",
			key:      actionGuaranteeKey,
			parent:   &actionKey,
		},
		{
			testName: "error action guarantee nil parent",
			key:      actionGuaranteeKey,
			parent:   nil,
			errstr:   "requires a parent of type 'action'",
		},
		{
			testName: "error action guarantee wrong parent type",
			key:      actionGuaranteeKey,
			parent:   &queryKey,
			errstr:   "requires parent of type 'action', but got 'query'",
		},

		// ActionSafety requires action parent.
		{
			testName: "ok action safety with action parent",
			key:      actionSafetyKey,
			parent:   &actionKey,
		},
		{
			testName: "error action safety nil parent",
			key:      actionSafetyKey,
			parent:   nil,
			errstr:   "requires a parent of type 'action'",
		},
		{
			testName: "error action safety wrong parent type",
			key:      actionSafetyKey,
			parent:   &classKey,
			errstr:   "requires parent of type 'action', but got 'class'",
		},

		// QueryRequire requires query parent.
		{
			testName: "ok query require with query parent",
			key:      queryRequireKey,
			parent:   &queryKey,
		},
		{
			testName: "error query require nil parent",
			key:      queryRequireKey,
			parent:   nil,
			errstr:   "requires a parent of type 'query'",
		},
		{
			testName: "error query require wrong parent type",
			key:      queryRequireKey,
			parent:   &actionKey,
			errstr:   "requires parent of type 'query', but got 'action'",
		},
		{
			testName: "error query require wrong parent key",
			key:      queryRequireKey,
			parent:   &otherQueryKey,
			errstr:   "does not match expected parent",
		},

		// QueryGuarantee requires query parent.
		{
			testName: "ok query guarantee with query parent",
			key:      queryGuaranteeKey,
			parent:   &queryKey,
		},
		{
			testName: "error query guarantee nil parent",
			key:      queryGuaranteeKey,
			parent:   nil,
			errstr:   "requires a parent of type 'query'",
		},
		{
			testName: "error query guarantee wrong parent type",
			key:      queryGuaranteeKey,
			parent:   &classKey,
			errstr:   "requires parent of type 'query', but got 'class'",
		},

		// AttributeDerivation requires attribute parent.
		{
			testName: "ok attribute derivation with attribute parent",
			key:      attributeDerivationKey,
			parent:   &attributeKey,
		},
		{
			testName: "error attribute derivation nil parent",
			key:      attributeDerivationKey,
			parent:   nil,
			errstr:   "requires a parent of type 'attribute'",
		},
		{
			testName: "error attribute derivation wrong parent type",
			key:      attributeDerivationKey,
			parent:   &classKey,
			errstr:   "requires parent of type 'attribute', but got 'class'",
		},
		{
			testName: "error attribute derivation wrong parent key",
			key:      attributeDerivationKey,
			parent:   &otherAttributeKey,
			errstr:   "does not match expected parent",
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
	domainKey := helper.Must(NewDomainKey("testdomain"))
	subdomainKey := helper.Must(NewSubdomainKey(domainKey, "testsubdomain"))
	classKey := helper.Must(NewClassKey(subdomainKey, "testclass"))
	stateKey := helper.Must(NewStateKey(classKey, "teststate"))
	actorKey := helper.Must(NewActorKey("testactor"))

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
	domainKey := helper.Must(NewDomainKey("testdomain"))
	domainKey2 := helper.Must(NewDomainKey("otherdomain"))
	subdomainKey := helper.Must(NewSubdomainKey(domainKey, "testsubdomain"))
	classKey := helper.Must(NewClassKey(subdomainKey, "testclass"))
	stateKey := helper.Must(NewStateKey(classKey, "teststate"))
	actorKey := helper.Must(NewActorKey("testactor"))

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
	domainKey := helper.Must(NewDomainKey("domain_key"))
	subdomainKey := helper.Must(NewSubdomainKey(domainKey, "subdomain_key"))
	classKey := helper.Must(NewClassKey(subdomainKey, "class_key"))
	stateKey := helper.Must(NewStateKey(classKey, "state_key"))

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
			assert.Equal(t, originalKey.ParentKey, parsedKey.ParentKey, "ParentKey mismatch for: %s", tt.description)
			assert.Equal(t, originalKey.KeyType, parsedKey.KeyType, "KeyType mismatch for: %s", tt.description)
			assert.Equal(t, originalKey.SubKey, parsedKey.SubKey, "SubKey mismatch for: %s", tt.description)
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
	domainKey := helper.Must(NewDomainKey("domain_key"))
	subdomainKey := helper.Must(NewSubdomainKey(domainKey, "subdomain_key"))
	classKey := helper.Must(NewClassKey(subdomainKey, "class_key"))
	classKey2 := helper.Must(NewClassKey(subdomainKey, "class_key2"))
	stateKey := helper.Must(NewStateKey(classKey, "state_key"))
	useCaseKey := helper.Must(NewUseCaseKey(subdomainKey, "use_case_key"))
	scenarioKey := helper.Must(NewScenarioKey(useCaseKey, "scenario_key"))
	actorKey := helper.Must(NewActorKey("actor_key"))

	// State action and transition keys.
	stateActionKey := helper.Must(NewStateActionKey(stateKey, "entry", "action_key"))
	transitionKey := helper.Must(NewTransitionKey(classKey, "state_a", "event_key", "guard_key", "action_key", "state_b"))

	// Domain association key.
	domainKey2 := helper.Must(NewDomainKey("domain_key2"))
	domainAssocKey := helper.Must(NewDomainAssociationKey(domainKey, domainKey2))

	// Class association key.
	classAssocKey := helper.Must(NewClassAssociationKey(subdomainKey, classKey, classKey2, "json test assoc"))

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
			errstr:   "'KeyType' failed on the 'oneof' tag",
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
	domainKey := helper.Must(NewDomainKey("domain_key"))
	subdomainKey := helper.Must(NewSubdomainKey(domainKey, "subdomain_key"))
	classKey := helper.Must(NewClassKey(subdomainKey, "class_key"))
	classKey2 := helper.Must(NewClassKey(subdomainKey, "class_key2"))
	stateKey := helper.Must(NewStateKey(classKey, "state_key"))
	useCaseKey := helper.Must(NewUseCaseKey(subdomainKey, "use_case_key"))
	scenarioKey := helper.Must(NewScenarioKey(useCaseKey, "scenario_key"))
	actorKey := helper.Must(NewActorKey("actor_key"))

	// State action and transition keys.
	stateActionKey := helper.Must(NewStateActionKey(stateKey, "entry", "action_key"))
	transitionKey := helper.Must(NewTransitionKey(classKey, "state_a", "event_key", "guard_key", "action_key", "state_b"))

	// Domain association key.
	domainKey2 := helper.Must(NewDomainKey("domain_key2"))
	domainAssocKey := helper.Must(NewDomainAssociationKey(domainKey, domainKey2))

	// Class association key.
	classAssocKey := helper.Must(NewClassAssociationKey(subdomainKey, classKey, classKey2, "text test assoc"))

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
			errstr:   "'KeyType' failed on the 'oneof' tag",
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
	domainKey := helper.Must(NewDomainKey("domain_key"))
	subdomainKey := helper.Must(NewSubdomainKey(domainKey, "subdomain_key"))
	classKey := helper.Must(NewClassKey(subdomainKey, "class_key"))
	stateKey := helper.Must(NewStateKey(classKey, "state_key"))

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
