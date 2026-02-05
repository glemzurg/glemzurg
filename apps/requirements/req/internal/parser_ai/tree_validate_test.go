package parser_ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TreeValidateSuite tests the validateModelTree function.
type TreeValidateSuite struct {
	suite.Suite
}

func TestTreeValidateSuite(t *testing.T) {
	suite.Run(t, new(TreeValidateSuite))
}

// TestValidTree verifies that a valid tree passes validation.
func (suite *TreeValidateSuite) TestValidTree() {
	t := suite.T()

	model := t_buildValidModelTree()
	err := validateModelTree(model)
	assert.NoError(t, err)
}

// TestClassActorNotFound verifies error when class references missing actor.
func (suite *TreeValidateSuite) TestClassActorNotFound() {
	t := suite.T()

	model := t_buildMinimalModelTree()
	// Add class with invalid actor reference
	model.Domains["domain1"].Subdomains["subdomain1"].Classes["class1"].ActorKey = "nonexistent_actor"

	err := validateModelTree(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeClassActorNotFound, parseErr.Code)
	assert.Equal(t, "actor_key", parseErr.Field)
}

// TestClassActorValid verifies valid actor reference passes.
func (suite *TreeValidateSuite) TestClassActorValid() {
	t := suite.T()

	model := t_buildMinimalModelTree()
	model.Actors["customer"] = &inputActor{Name: "Customer", Type: "person"}
	model.Domains["domain1"].Subdomains["subdomain1"].Classes["class1"].ActorKey = "customer"

	err := validateModelTree(model)
	assert.NoError(t, err)
}

// TestClassIndexAttrNotFound verifies error when index references missing attribute.
func (suite *TreeValidateSuite) TestClassIndexAttrNotFound() {
	t := suite.T()

	model := t_buildMinimalModelTree()
	// Add index referencing non-existent attribute
	model.Domains["domain1"].Subdomains["subdomain1"].Classes["class1"].Indexes = [][]string{{"missing_attr"}}

	err := validateModelTree(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeClassIndexAttrNotFound, parseErr.Code)
	assert.Equal(t, "indexes[0][0]", parseErr.Field)
}

// TestClassIndexDuplicateAttr verifies error when index has duplicate attributes.
func (suite *TreeValidateSuite) TestClassIndexDuplicateAttr() {
	t := suite.T()

	model := t_buildMinimalModelTree()
	// Add attribute and index with duplicate
	model.Domains["domain1"].Subdomains["subdomain1"].Classes["class1"].Attributes = map[string]*inputAttribute{
		"id": {Name: "ID", DataTypeRules: "int"},
	}
	model.Domains["domain1"].Subdomains["subdomain1"].Classes["class1"].Indexes = [][]string{{"id", "id"}}

	err := validateModelTree(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeClassIndexAttrNotFound, parseErr.Code)
	assert.Contains(t, parseErr.Message, "duplicate")
}

// TestStateMachineStateActionNotFound verifies error when state action references missing action.
func (suite *TreeValidateSuite) TestStateMachineStateActionNotFound() {
	t := suite.T()

	model := t_buildMinimalModelTree()
	class := model.Domains["domain1"].Subdomains["subdomain1"].Classes["class1"]
	class.StateMachine = &inputStateMachine{
		States: map[string]*inputState{
			"pending": {
				Name: "Pending",
				Actions: []inputStateAction{
					{ActionKey: "missing_action", When: "entry"},
				},
			},
		},
		Events:      map[string]*inputEvent{},
		Guards:      map[string]*inputGuard{},
		Transitions: []inputTransition{},
	}
	class.Actions = map[string]*inputAction{}

	err := validateModelTree(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeStateMachineActionNotFound, parseErr.Code)
	assert.Equal(t, "states.pending.actions[0].action_key", parseErr.Field)
}

// TestTransitionNoStates verifies error when transition has neither from nor to state.
func (suite *TreeValidateSuite) TestTransitionNoStates() {
	t := suite.T()

	model := t_buildMinimalModelTree()
	class := model.Domains["domain1"].Subdomains["subdomain1"].Classes["class1"]
	class.StateMachine = &inputStateMachine{
		States: map[string]*inputState{
			"pending": {Name: "Pending"},
		},
		Events: map[string]*inputEvent{
			"create": {Name: "create"},
		},
		Guards: map[string]*inputGuard{},
		Transitions: []inputTransition{
			{
				FromStateKey: nil,
				ToStateKey:   nil,
				EventKey:     "create",
			},
		},
	}
	class.Actions = map[string]*inputAction{}

	err := validateModelTree(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeTransitionNoStates, parseErr.Code)
	assert.Equal(t, "transitions[0]", parseErr.Field)
}

// TestTransitionFromStateNotFound verifies error when transition from_state_key doesn't exist.
func (suite *TreeValidateSuite) TestTransitionFromStateNotFound() {
	t := suite.T()

	model := t_buildMinimalModelTree()
	class := model.Domains["domain1"].Subdomains["subdomain1"].Classes["class1"]
	missingState := "missing_state"
	toState := "pending"
	class.StateMachine = &inputStateMachine{
		States: map[string]*inputState{
			"pending": {Name: "Pending"},
		},
		Events: map[string]*inputEvent{
			"proceed": {Name: "proceed"},
		},
		Guards: map[string]*inputGuard{},
		Transitions: []inputTransition{
			{
				FromStateKey: &missingState,
				ToStateKey:   &toState,
				EventKey:     "proceed",
			},
		},
	}
	class.Actions = map[string]*inputAction{}

	err := validateModelTree(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeStateMachineStateNotFound, parseErr.Code)
	assert.Equal(t, "transitions[0].from_state_key", parseErr.Field)
}

// TestTransitionToStateNotFound verifies error when transition to_state_key doesn't exist.
func (suite *TreeValidateSuite) TestTransitionToStateNotFound() {
	t := suite.T()

	model := t_buildMinimalModelTree()
	class := model.Domains["domain1"].Subdomains["subdomain1"].Classes["class1"]
	fromState := "pending"
	missingState := "missing_state"
	class.StateMachine = &inputStateMachine{
		States: map[string]*inputState{
			"pending": {Name: "Pending"},
		},
		Events: map[string]*inputEvent{
			"proceed": {Name: "proceed"},
		},
		Guards: map[string]*inputGuard{},
		Transitions: []inputTransition{
			{
				FromStateKey: &fromState,
				ToStateKey:   &missingState,
				EventKey:     "proceed",
			},
		},
	}
	class.Actions = map[string]*inputAction{}

	err := validateModelTree(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeStateMachineStateNotFound, parseErr.Code)
	assert.Equal(t, "transitions[0].to_state_key", parseErr.Field)
}

// TestTransitionEventNotFound verifies error when transition event_key doesn't exist.
func (suite *TreeValidateSuite) TestTransitionEventNotFound() {
	t := suite.T()

	model := t_buildMinimalModelTree()
	class := model.Domains["domain1"].Subdomains["subdomain1"].Classes["class1"]
	toState := "pending"
	class.StateMachine = &inputStateMachine{
		States: map[string]*inputState{
			"pending": {Name: "Pending"},
		},
		Events:      map[string]*inputEvent{},
		Guards:      map[string]*inputGuard{},
		Transitions: []inputTransition{
			{
				FromStateKey: nil,
				ToStateKey:   &toState,
				EventKey:     "missing_event",
			},
		},
	}
	class.Actions = map[string]*inputAction{}

	err := validateModelTree(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeStateMachineEventNotFound, parseErr.Code)
	assert.Equal(t, "transitions[0].event_key", parseErr.Field)
}

// TestTransitionGuardNotFound verifies error when transition guard_key doesn't exist.
func (suite *TreeValidateSuite) TestTransitionGuardNotFound() {
	t := suite.T()

	model := t_buildMinimalModelTree()
	class := model.Domains["domain1"].Subdomains["subdomain1"].Classes["class1"]
	toState := "pending"
	missingGuard := "missing_guard"
	class.StateMachine = &inputStateMachine{
		States: map[string]*inputState{
			"pending": {Name: "Pending"},
		},
		Events: map[string]*inputEvent{
			"create": {Name: "create"},
		},
		Guards: map[string]*inputGuard{},
		Transitions: []inputTransition{
			{
				FromStateKey: nil,
				ToStateKey:   &toState,
				EventKey:     "create",
				GuardKey:     &missingGuard,
			},
		},
	}
	class.Actions = map[string]*inputAction{}

	err := validateModelTree(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeStateMachineGuardNotFound, parseErr.Code)
	assert.Equal(t, "transitions[0].guard_key", parseErr.Field)
}

// TestTransitionActionNotFound verifies error when transition action_key doesn't exist.
func (suite *TreeValidateSuite) TestTransitionActionNotFound() {
	t := suite.T()

	model := t_buildMinimalModelTree()
	class := model.Domains["domain1"].Subdomains["subdomain1"].Classes["class1"]
	toState := "pending"
	missingAction := "missing_action"
	class.StateMachine = &inputStateMachine{
		States: map[string]*inputState{
			"pending": {Name: "Pending"},
		},
		Events: map[string]*inputEvent{
			"create": {Name: "create"},
		},
		Guards:      map[string]*inputGuard{},
		Transitions: []inputTransition{
			{
				FromStateKey: nil,
				ToStateKey:   &toState,
				EventKey:     "create",
				ActionKey:    &missingAction,
			},
		},
	}
	class.Actions = map[string]*inputAction{}

	err := validateModelTree(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeStateMachineActionNotFound, parseErr.Code)
	assert.Equal(t, "transitions[0].action_key", parseErr.Field)
}

// TestActionUnreferenced verifies error when an action exists but is not referenced.
func (suite *TreeValidateSuite) TestActionUnreferenced() {
	t := suite.T()

	model := t_buildMinimalModelTree()
	class := model.Domains["domain1"].Subdomains["subdomain1"].Classes["class1"]
	toState := "pending"
	class.StateMachine = &inputStateMachine{
		States: map[string]*inputState{
			"pending": {Name: "Pending"},
		},
		Events: map[string]*inputEvent{
			"create": {Name: "create"},
		},
		Guards:      map[string]*inputGuard{},
		Transitions: []inputTransition{
			{
				FromStateKey: nil,
				ToStateKey:   &toState,
				EventKey:     "create",
				// No action_key - the action is not referenced
			},
		},
	}
	// Add an action that is not referenced anywhere
	class.Actions = map[string]*inputAction{
		"unreferenced_action": {Name: "Unreferenced Action"},
	}

	err := validateModelTree(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeActionUnreferenced, parseErr.Code)
	assert.Equal(t, "action_key", parseErr.Field)
	assert.Contains(t, parseErr.Message, "unreferenced_action")
	assert.Contains(t, parseErr.Message, "not referenced")
}

// TestActionReferencedByStateAction verifies that action referenced by state action passes.
func (suite *TreeValidateSuite) TestActionReferencedByStateAction() {
	t := suite.T()

	model := t_buildMinimalModelTree()
	class := model.Domains["domain1"].Subdomains["subdomain1"].Classes["class1"]
	toState := "pending"
	class.StateMachine = &inputStateMachine{
		States: map[string]*inputState{
			"pending": {
				Name: "Pending",
				Actions: []inputStateAction{
					{ActionKey: "my_action", When: "entry"},
				},
			},
		},
		Events: map[string]*inputEvent{
			"create": {Name: "create"},
		},
		Guards:      map[string]*inputGuard{},
		Transitions: []inputTransition{
			{
				FromStateKey: nil,
				ToStateKey:   &toState,
				EventKey:     "create",
			},
		},
	}
	class.Actions = map[string]*inputAction{
		"my_action": {Name: "My Action"},
	}

	err := validateModelTree(model)
	assert.NoError(t, err)
}

// TestActionReferencedByTransition verifies that action referenced by transition passes.
func (suite *TreeValidateSuite) TestActionReferencedByTransition() {
	t := suite.T()

	model := t_buildMinimalModelTree()
	class := model.Domains["domain1"].Subdomains["subdomain1"].Classes["class1"]
	toState := "pending"
	actionKey := "my_action"
	class.StateMachine = &inputStateMachine{
		States: map[string]*inputState{
			"pending": {Name: "Pending"},
		},
		Events: map[string]*inputEvent{
			"create": {Name: "create"},
		},
		Guards:      map[string]*inputGuard{},
		Transitions: []inputTransition{
			{
				FromStateKey: nil,
				ToStateKey:   &toState,
				EventKey:     "create",
				ActionKey:    &actionKey,
			},
		},
	}
	class.Actions = map[string]*inputAction{
		"my_action": {Name: "My Action"},
	}

	err := validateModelTree(model)
	assert.NoError(t, err)
}

// TestGenSuperclassNotFound verifies error when generalization superclass doesn't exist.
func (suite *TreeValidateSuite) TestGenSuperclassNotFound() {
	t := suite.T()

	model := t_buildMinimalModelTree()
	subdomain := model.Domains["domain1"].Subdomains["subdomain1"]
	subdomain.Classes["book"] = &inputClass{Name: "Book"}
	subdomain.Generalizations = map[string]*inputGeneralization{
		"medium": {
			Name:          "Medium",
			SuperclassKey: "missing_class",
			SubclassKeys:  []string{"book"},
		},
	}

	err := validateModelTree(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeGenSuperclassNotFound, parseErr.Code)
	assert.Equal(t, "superclass_key", parseErr.Field)
}

// TestGenSubclassNotFound verifies error when generalization subclass doesn't exist.
func (suite *TreeValidateSuite) TestGenSubclassNotFound() {
	t := suite.T()

	model := t_buildMinimalModelTree()
	subdomain := model.Domains["domain1"].Subdomains["subdomain1"]
	subdomain.Classes["product"] = &inputClass{Name: "Product"}
	subdomain.Generalizations = map[string]*inputGeneralization{
		"medium": {
			Name:          "Medium",
			SuperclassKey: "product",
			SubclassKeys:  []string{"missing_class"},
		},
	}

	err := validateModelTree(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeGenSubclassNotFound, parseErr.Code)
	assert.Equal(t, "subclass_keys[0]", parseErr.Field)
}

// TestGenSubclassDuplicate verifies error when generalization has duplicate subclass.
func (suite *TreeValidateSuite) TestGenSubclassDuplicate() {
	t := suite.T()

	model := t_buildMinimalModelTree()
	subdomain := model.Domains["domain1"].Subdomains["subdomain1"]
	subdomain.Classes["product"] = &inputClass{Name: "Product"}
	subdomain.Classes["book"] = &inputClass{Name: "Book"}
	subdomain.Generalizations = map[string]*inputGeneralization{
		"medium": {
			Name:          "Medium",
			SuperclassKey: "product",
			SubclassKeys:  []string{"book", "book"},
		},
	}

	err := validateModelTree(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeGenSubclassDuplicate, parseErr.Code)
	assert.Equal(t, "subclass_keys[1]", parseErr.Field)
}

// TestGenSuperclassIsSubclass verifies error when superclass is also listed as subclass.
func (suite *TreeValidateSuite) TestGenSuperclassIsSubclass() {
	t := suite.T()

	model := t_buildMinimalModelTree()
	subdomain := model.Domains["domain1"].Subdomains["subdomain1"]
	subdomain.Classes["product"] = &inputClass{Name: "Product"}
	subdomain.Classes["book"] = &inputClass{Name: "Book"}
	subdomain.Generalizations = map[string]*inputGeneralization{
		"medium": {
			Name:          "Medium",
			SuperclassKey: "product",
			SubclassKeys:  []string{"book", "product"},
		},
	}

	err := validateModelTree(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeGenSuperclassIsSubclass, parseErr.Code)
}

// TestSubdomainAssocFromClassNotFound verifies error when subdomain association from_class_key doesn't exist.
func (suite *TreeValidateSuite) TestSubdomainAssocFromClassNotFound() {
	t := suite.T()

	model := t_buildMinimalModelTree()
	subdomain := model.Domains["domain1"].Subdomains["subdomain1"]
	subdomain.Associations = map[string]*inputAssociation{
		"test_assoc": {
			Name:             "Test Association",
			FromClassKey:     "missing_class",
			FromMultiplicity: "1",
			ToClassKey:       "class1",
			ToMultiplicity:   "*",
		},
	}

	err := validateModelTree(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeAssocFromClassNotFound, parseErr.Code)
	assert.Equal(t, "from_class_key", parseErr.Field)
}

// TestSubdomainAssocToClassNotFound verifies error when subdomain association to_class_key doesn't exist.
func (suite *TreeValidateSuite) TestSubdomainAssocToClassNotFound() {
	t := suite.T()

	model := t_buildMinimalModelTree()
	subdomain := model.Domains["domain1"].Subdomains["subdomain1"]
	subdomain.Associations = map[string]*inputAssociation{
		"test_assoc": {
			Name:             "Test Association",
			FromClassKey:     "class1",
			FromMultiplicity: "1",
			ToClassKey:       "missing_class",
			ToMultiplicity:   "*",
		},
	}

	err := validateModelTree(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeAssocToClassNotFound, parseErr.Code)
	assert.Equal(t, "to_class_key", parseErr.Field)
}

// TestSubdomainAssocClassNotFound verifies error when subdomain association association_class_key doesn't exist.
func (suite *TreeValidateSuite) TestSubdomainAssocClassNotFound() {
	t := suite.T()

	model := t_buildMinimalModelTree()
	subdomain := model.Domains["domain1"].Subdomains["subdomain1"]
	subdomain.Classes["class2"] = &inputClass{Name: "Class 2"}
	missingClass := "missing_class"
	subdomain.Associations = map[string]*inputAssociation{
		"test_assoc": {
			Name:                "Test Association",
			FromClassKey:        "class1",
			FromMultiplicity:    "1",
			ToClassKey:          "class2",
			ToMultiplicity:      "*",
			AssociationClassKey: &missingClass,
		},
	}

	err := validateModelTree(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeAssocClassNotFound, parseErr.Code)
	assert.Equal(t, "association_class_key", parseErr.Field)
}

// TestSubdomainAssocClassSameAsFromClass verifies error when association_class_key equals from_class_key.
func (suite *TreeValidateSuite) TestSubdomainAssocClassSameAsFromClass() {
	t := suite.T()

	model := t_buildMinimalModelTree()
	subdomain := model.Domains["domain1"].Subdomains["subdomain1"]
	subdomain.Classes["class2"] = &inputClass{Name: "Class 2"}
	sameAsFrom := "class1"
	subdomain.Associations = map[string]*inputAssociation{
		"test_assoc": {
			Name:                "Test Association",
			FromClassKey:        "class1",
			FromMultiplicity:    "1",
			ToClassKey:          "class2",
			ToMultiplicity:      "*",
			AssociationClassKey: &sameAsFrom,
		},
	}

	err := validateModelTree(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeAssocClassSameAsEndpoint, parseErr.Code)
	assert.Equal(t, "association_class_key", parseErr.Field)
	assert.Contains(t, parseErr.Message, "from_class_key")
}

// TestSubdomainAssocClassSameAsToClass verifies error when association_class_key equals to_class_key.
func (suite *TreeValidateSuite) TestSubdomainAssocClassSameAsToClass() {
	t := suite.T()

	model := t_buildMinimalModelTree()
	subdomain := model.Domains["domain1"].Subdomains["subdomain1"]
	subdomain.Classes["class2"] = &inputClass{Name: "Class 2"}
	sameAsTo := "class2"
	subdomain.Associations = map[string]*inputAssociation{
		"test_assoc": {
			Name:                "Test Association",
			FromClassKey:        "class1",
			FromMultiplicity:    "1",
			ToClassKey:          "class2",
			ToMultiplicity:      "*",
			AssociationClassKey: &sameAsTo,
		},
	}

	err := validateModelTree(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeAssocClassSameAsEndpoint, parseErr.Code)
	assert.Equal(t, "association_class_key", parseErr.Field)
	assert.Contains(t, parseErr.Message, "to_class_key")
}

// TestSubdomainAssocMultiplicityInvalid verifies error when multiplicity format is invalid.
func (suite *TreeValidateSuite) TestSubdomainAssocMultiplicityInvalid() {
	t := suite.T()

	model := t_buildMinimalModelTree()
	subdomain := model.Domains["domain1"].Subdomains["subdomain1"]
	subdomain.Classes["class2"] = &inputClass{Name: "Class 2"}
	subdomain.Associations = map[string]*inputAssociation{
		"test_assoc": {
			Name:             "Test Association",
			FromClassKey:     "class1",
			FromMultiplicity: "invalid",
			ToClassKey:       "class2",
			ToMultiplicity:   "*",
		},
	}

	err := validateModelTree(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeAssocMultiplicityInvalid, parseErr.Code)
	assert.Equal(t, "from_multiplicity", parseErr.Field)
}

// TestDomainAssocFromClassNotFound verifies error when domain association from_class_key doesn't exist.
func (suite *TreeValidateSuite) TestDomainAssocFromClassNotFound() {
	t := suite.T()

	model := t_buildMinimalModelTree()
	// Add another subdomain with a class
	model.Domains["domain1"].Subdomains["subdomain2"] = &inputSubdomain{
		Name:            "Subdomain 2",
		Classes:         map[string]*inputClass{"class2": {Name: "Class 2"}},
		Generalizations: map[string]*inputGeneralization{},
		Associations:    map[string]*inputAssociation{},
	}
	model.Domains["domain1"].Associations = map[string]*inputAssociation{
		"test_assoc": {
			Name:             "Test Association",
			FromClassKey:     "subdomain1/missing_class",
			FromMultiplicity: "1",
			ToClassKey:       "subdomain2/class2",
			ToMultiplicity:   "*",
		},
	}

	err := validateModelTree(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeAssocFromClassNotFound, parseErr.Code)
	assert.Equal(t, "from_class_key", parseErr.Field)
}

// TestDomainAssocInvalidKeyFormat verifies error when domain association has invalid key format.
func (suite *TreeValidateSuite) TestDomainAssocInvalidKeyFormat() {
	t := suite.T()

	model := t_buildMinimalModelTree()
	model.Domains["domain1"].Associations = map[string]*inputAssociation{
		"test_assoc": {
			Name:             "Test Association",
			FromClassKey:     "class1", // Wrong format - should be subdomain/class
			FromMultiplicity: "1",
			ToClassKey:       "subdomain1/class1",
			ToMultiplicity:   "*",
		},
	}

	err := validateModelTree(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeAssocFromClassNotFound, parseErr.Code)
}

// TestModelAssocFromClassNotFound verifies error when model association from_class_key doesn't exist.
func (suite *TreeValidateSuite) TestModelAssocFromClassNotFound() {
	t := suite.T()

	model := t_buildMinimalModelTree()
	// Add another domain with subdomain and class
	model.Domains["domain2"] = &inputDomain{
		Name: "Domain 2",
		Subdomains: map[string]*inputSubdomain{
			"subdomain1": {
				Name:            "Subdomain 1",
				Classes:         map[string]*inputClass{"class1": {Name: "Class 1"}},
				Generalizations: map[string]*inputGeneralization{},
				Associations:    map[string]*inputAssociation{},
			},
		},
		Associations: map[string]*inputAssociation{},
	}
	model.Associations = map[string]*inputAssociation{
		"test_assoc": {
			Name:             "Test Association",
			FromClassKey:     "domain1/subdomain1/missing_class",
			FromMultiplicity: "1",
			ToClassKey:       "domain2/subdomain1/class1",
			ToMultiplicity:   "*",
		},
	}

	err := validateModelTree(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeAssocFromClassNotFound, parseErr.Code)
	assert.Equal(t, "from_class_key", parseErr.Field)
}

// TestModelAssocToClassNotFound verifies error when model association to_class_key doesn't exist.
func (suite *TreeValidateSuite) TestModelAssocToClassNotFound() {
	t := suite.T()

	model := t_buildMinimalModelTree()
	model.Domains["domain2"] = &inputDomain{
		Name: "Domain 2",
		Subdomains: map[string]*inputSubdomain{
			"subdomain1": {
				Name:            "Subdomain 1",
				Classes:         map[string]*inputClass{},
				Generalizations: map[string]*inputGeneralization{},
				Associations:    map[string]*inputAssociation{},
			},
		},
		Associations: map[string]*inputAssociation{},
	}
	model.Associations = map[string]*inputAssociation{
		"test_assoc": {
			Name:             "Test Association",
			FromClassKey:     "domain1/subdomain1/class1",
			FromMultiplicity: "1",
			ToClassKey:       "domain2/subdomain1/missing_class",
			ToMultiplicity:   "*",
		},
	}

	err := validateModelTree(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeAssocToClassNotFound, parseErr.Code)
	assert.Equal(t, "to_class_key", parseErr.Field)
}

// TestModelAssocInvalidKeyFormat verifies error when model association has invalid key format.
func (suite *TreeValidateSuite) TestModelAssocInvalidKeyFormat() {
	t := suite.T()

	model := t_buildMinimalModelTree()
	model.Associations = map[string]*inputAssociation{
		"test_assoc": {
			Name:             "Test Association",
			FromClassKey:     "subdomain1/class1", // Wrong format - should be domain/subdomain/class
			FromMultiplicity: "1",
			ToClassKey:       "domain1/subdomain1/class1",
			ToMultiplicity:   "*",
		},
	}

	err := validateModelTree(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeAssocFromClassNotFound, parseErr.Code)
}

// TestMultiplicityValidation tests various multiplicity formats.
func (suite *TreeValidateSuite) TestMultiplicityValidation() {
	t := suite.T()

	tests := []struct {
		name     string
		mult     string
		expected bool
	}{
		{"exactly_one", "1", true},
		{"zero_or_one", "0..1", true},
		{"zero_or_more", "*", true},
		{"one_or_more", "1..*", true},
		{"exactly_three", "3", true},
		{"range", "2..5", true},
		{"three_or_more", "3..*", true},
		{"zero_to_three", "0..3", true},
		{"invalid_text", "one", false},
		{"invalid_range", "5..3", false},
		{"empty", "", false},
		{"invalid_separator", "1-*", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMultiplicity(tt.mult)
			if tt.expected {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// ======================================
// Completeness Validation Tests
// ======================================

// TestCompletenessValidModel verifies that a complete valid model passes completeness validation.
func (suite *TreeValidateSuite) TestCompletenessValidModel() {
	t := suite.T()

	model := t_buildCompleteModelTree()
	err := validateModelCompleteness(model)
	assert.NoError(t, err)
}

// TestCompletenessModelNoActors verifies error when model has no actors.
func (suite *TreeValidateSuite) TestCompletenessModelNoActors() {
	t := suite.T()

	model := t_buildCompleteModelTree()
	model.Actors = map[string]*inputActor{} // Remove all actors

	err := validateModelCompleteness(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeModelNoActors, parseErr.Code)
	assert.Equal(t, "actors", parseErr.Field)
	assert.Contains(t, parseErr.Message, "at least one actor")
	assert.Contains(t, parseErr.Message, "actors/") // Check for guidance about file location
}

// TestCompletenessModelNoDomains verifies error when model has no domains.
func (suite *TreeValidateSuite) TestCompletenessModelNoDomains() {
	t := suite.T()

	model := t_buildCompleteModelTree()
	model.Domains = map[string]*inputDomain{} // Remove all domains

	err := validateModelCompleteness(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeModelNoDomains, parseErr.Code)
	assert.Equal(t, "domains", parseErr.Field)
	assert.Contains(t, parseErr.Message, "at least one domain")
	assert.Contains(t, parseErr.Message, "domains/") // Check for guidance about file location
}

// TestCompletenessDomainNoSubdomains verifies error when domain has no subdomains.
func (suite *TreeValidateSuite) TestCompletenessDomainNoSubdomains() {
	t := suite.T()

	model := t_buildCompleteModelTree()
	model.Domains["orders"].Subdomains = map[string]*inputSubdomain{} // Remove all subdomains

	err := validateModelCompleteness(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeDomainNoSubdomains, parseErr.Code)
	assert.Equal(t, "subdomains", parseErr.Field)
	assert.Contains(t, parseErr.Message, "at least one subdomain")
	assert.Contains(t, parseErr.Message, "orders") // Check for specific domain name
}

// TestCompletenessSubdomainTooFewClasses verifies error when subdomain has less than 2 classes.
func (suite *TreeValidateSuite) TestCompletenessSubdomainTooFewClasses() {
	t := suite.T()

	model := t_buildCompleteModelTree()
	// Keep only one class
	model.Domains["orders"].Subdomains["default"].Classes = map[string]*inputClass{
		"order": t_buildCompleteClass(),
	}
	// Update association to use remaining classes
	model.Domains["orders"].Subdomains["default"].Associations = map[string]*inputAssociation{
		"self_ref": {
			Name:             "Self Ref",
			FromClassKey:     "order",
			FromMultiplicity: "1",
			ToClassKey:       "order",
			ToMultiplicity:   "*",
		},
	}

	err := validateModelCompleteness(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeSubdomainTooFewClasses, parseErr.Code)
	assert.Equal(t, "classes", parseErr.Field)
	assert.Contains(t, parseErr.Message, "at least 2 classes")
	assert.Contains(t, parseErr.Message, "has 1")
}

// TestCompletenessSubdomainNoAssociations verifies error when subdomain has no associations.
func (suite *TreeValidateSuite) TestCompletenessSubdomainNoAssociations() {
	t := suite.T()

	model := t_buildCompleteModelTree()
	model.Domains["orders"].Subdomains["default"].Associations = map[string]*inputAssociation{} // Remove all associations

	err := validateModelCompleteness(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeSubdomainNoAssociations, parseErr.Code)
	assert.Equal(t, "associations", parseErr.Field)
	assert.Contains(t, parseErr.Message, "at least one association")
	assert.Contains(t, parseErr.Message, "associations/") // Check for guidance about file location
}

// TestCompletenessClassNoAttributes verifies error when class has no attributes.
func (suite *TreeValidateSuite) TestCompletenessClassNoAttributes() {
	t := suite.T()

	model := t_buildCompleteModelTree()
	model.Domains["orders"].Subdomains["default"].Classes["order"].Attributes = map[string]*inputAttribute{} // Remove all attributes

	err := validateModelCompleteness(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeClassNoAttributes, parseErr.Code)
	assert.Equal(t, "attributes", parseErr.Field)
	assert.Contains(t, parseErr.Message, "at least one attribute")
	assert.Contains(t, parseErr.Message, "order") // Check for specific class name
}

// TestCompletenessClassNoStateMachine verifies error when class has no state machine.
func (suite *TreeValidateSuite) TestCompletenessClassNoStateMachine() {
	t := suite.T()

	model := t_buildCompleteModelTree()
	model.Domains["orders"].Subdomains["default"].Classes["order"].StateMachine = nil // Remove state machine

	err := validateModelCompleteness(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeClassNoStateMachine, parseErr.Code)
	assert.Equal(t, "state_machine", parseErr.Field)
	assert.Contains(t, parseErr.Message, "must have a state machine")
	assert.Contains(t, parseErr.Message, "state_machine.json") // Check for guidance about file
}

// TestCompletenessStateMachineNoTransitions verifies error when state machine has no transitions.
func (suite *TreeValidateSuite) TestCompletenessStateMachineNoTransitions() {
	t := suite.T()

	model := t_buildCompleteModelTree()
	model.Domains["orders"].Subdomains["default"].Classes["order"].StateMachine.Transitions = []inputTransition{} // Remove all transitions

	err := validateModelCompleteness(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeStateMachineNoTransitions, parseErr.Code)
	assert.Equal(t, "transitions", parseErr.Field)
	assert.Contains(t, parseErr.Message, "at least one transition")
}

// TestCompletenessAllErrorsProvideGuidance verifies all completeness errors provide helpful guidance.
func (suite *TreeValidateSuite) TestCompletenessAllErrorsProvideGuidance() {
	t := suite.T()

	// Test each error type and verify it contains actionable guidance
	tests := []struct {
		name            string
		buildModel      func() *inputModel
		expectedCode    int
		shouldContain   []string
	}{
		{
			name: "no_actors",
			buildModel: func() *inputModel {
				m := t_buildCompleteModelTree()
				m.Actors = map[string]*inputActor{}
				return m
			},
			expectedCode: ErrTreeModelNoActors,
			shouldContain: []string{
				"actors represent the users, systems, or external entities",
				"actors/",
				".actor.json",
			},
		},
		{
			name: "no_domains",
			buildModel: func() *inputModel {
				m := t_buildCompleteModelTree()
				m.Domains = map[string]*inputDomain{}
				return m
			},
			expectedCode: ErrTreeModelNoDomains,
			shouldContain: []string{
				"domains are high-level subject areas",
				"domains/",
				"domain.json",
			},
		},
		{
			name: "no_subdomains",
			buildModel: func() *inputModel {
				m := t_buildCompleteModelTree()
				m.Domains["orders"].Subdomains = map[string]*inputSubdomain{}
				return m
			},
			expectedCode: ErrTreeDomainNoSubdomains,
			shouldContain: []string{
				"subdomains organize classes",
				"subdomain.json",
			},
		},
		{
			name: "too_few_classes",
			buildModel: func() *inputModel {
				m := t_buildCompleteModelTree()
				m.Domains["orders"].Subdomains["default"].Classes = map[string]*inputClass{
					"order": t_buildCompleteClass(),
				}
				m.Domains["orders"].Subdomains["default"].Associations = map[string]*inputAssociation{
					"self_ref": {
						Name:             "Self Ref",
						FromClassKey:     "order",
						FromMultiplicity: "1",
						ToClassKey:       "order",
						ToMultiplicity:   "*",
					},
				}
				return m
			},
			expectedCode: ErrTreeSubdomainTooFewClasses,
			shouldContain: []string{
				"needs multiple classes to represent meaningful relationships",
				"classes/",
				"class.json",
			},
		},
		{
			name: "no_associations",
			buildModel: func() *inputModel {
				m := t_buildCompleteModelTree()
				m.Domains["orders"].Subdomains["default"].Associations = map[string]*inputAssociation{}
				return m
			},
			expectedCode: ErrTreeSubdomainNoAssociations,
			shouldContain: []string{
				"associations describe how classes relate",
				"associations/",
				".assoc.json",
			},
		},
		{
			name: "no_attributes",
			buildModel: func() *inputModel {
				m := t_buildCompleteModelTree()
				m.Domains["orders"].Subdomains["default"].Classes["order"].Attributes = map[string]*inputAttribute{}
				return m
			},
			expectedCode: ErrTreeClassNoAttributes,
			shouldContain: []string{
				"attributes describe the data properties",
				"attributes",
			},
		},
		{
			name: "no_state_machine",
			buildModel: func() *inputModel {
				m := t_buildCompleteModelTree()
				m.Domains["orders"].Subdomains["default"].Classes["order"].StateMachine = nil
				return m
			},
			expectedCode: ErrTreeClassNoStateMachine,
			shouldContain: []string{
				"state machines describe the lifecycle and behavior",
				"state_machine.json",
			},
		},
		{
			name: "no_transitions",
			buildModel: func() *inputModel {
				m := t_buildCompleteModelTree()
				m.Domains["orders"].Subdomains["default"].Classes["order"].StateMachine.Transitions = []inputTransition{}
				return m
			},
			expectedCode: ErrTreeStateMachineNoTransitions,
			shouldContain: []string{
				"transitions describe how the class moves between states",
				"transitions",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := tt.buildModel()
			err := validateModelCompleteness(model)
			require.Error(t, err)

			parseErr, ok := err.(*ParseError)
			require.True(t, ok, "error should be a ParseError")
			assert.Equal(t, tt.expectedCode, parseErr.Code, "error code should match")

			// Verify all expected guidance strings are present
			for _, s := range tt.shouldContain {
				assert.Contains(t, parseErr.Message, s,
					"error message should contain guidance: %s", s)
			}
		})
	}
}

// t_buildMinimalModelTree creates a minimal valid model tree for testing.
func t_buildMinimalModelTree() *inputModel {
	return &inputModel{
		Name:   "Test Model",
		Actors: map[string]*inputActor{},
		Domains: map[string]*inputDomain{
			"domain1": {
				Name: "Domain 1",
				Subdomains: map[string]*inputSubdomain{
					"subdomain1": {
						Name: "Subdomain 1",
						Classes: map[string]*inputClass{
							"class1": {
								Name:       "Class 1",
								Attributes: map[string]*inputAttribute{},
							},
						},
						Generalizations: map[string]*inputGeneralization{},
						Associations:    map[string]*inputAssociation{},
					},
				},
				Associations: map[string]*inputAssociation{},
			},
		},
		Associations: map[string]*inputAssociation{},
	}
}

// t_buildValidModelTree creates a complete valid model tree for testing.
func t_buildValidModelTree() *inputModel {
	toState := "confirmed"
	fromState := "pending"
	guardKey := "has_items"
	actionKey := "calculate_total"

	return &inputModel{
		Name: "Valid Model",
		Actors: map[string]*inputActor{
			"customer": {Name: "Customer", Type: "person"},
		},
		Domains: map[string]*inputDomain{
			"orders": {
				Name: "Orders",
				Subdomains: map[string]*inputSubdomain{
					"default": {
						Name: "Default",
						Classes: map[string]*inputClass{
							"order": {
								Name:     "Order",
								ActorKey: "customer",
								Attributes: map[string]*inputAttribute{
									"id":     {Name: "ID", DataTypeRules: "int"},
									"status": {Name: "Status", DataTypeRules: "string"},
								},
								Indexes: [][]string{{"id"}, {"status"}},
								StateMachine: &inputStateMachine{
									States: map[string]*inputState{
										"pending":   {Name: "Pending"},
										"confirmed": {Name: "Confirmed"},
									},
									Events: map[string]*inputEvent{
										"confirm": {Name: "confirm"},
									},
									Guards: map[string]*inputGuard{
										"has_items": {Name: "hasItems", Details: "Order has items"},
									},
									Transitions: []inputTransition{
										{
											FromStateKey: &fromState,
											ToStateKey:   &toState,
											EventKey:     "confirm",
											GuardKey:     &guardKey,
											ActionKey:    &actionKey,
										},
									},
								},
								Actions: map[string]*inputAction{
									"calculate_total": {Name: "Calculate Total"},
								},
								Queries: map[string]*inputQuery{},
							},
							"line_item": {
								Name:       "Line Item",
								Attributes: map[string]*inputAttribute{},
							},
							"product": {
								Name:       "Product",
								Attributes: map[string]*inputAttribute{},
							},
							"book": {
								Name:       "Book",
								Attributes: map[string]*inputAttribute{},
							},
						},
						Generalizations: map[string]*inputGeneralization{
							"product_type": {
								Name:          "Product Type",
								SuperclassKey: "product",
								SubclassKeys:  []string{"book"},
							},
						},
						Associations: map[string]*inputAssociation{
							"order_lines": {
								Name:             "Order Lines",
								FromClassKey:     "order",
								FromMultiplicity: "1",
								ToClassKey:       "line_item",
								ToMultiplicity:   "1..*",
							},
						},
					},
				},
				Associations: map[string]*inputAssociation{},
			},
		},
		Associations: map[string]*inputAssociation{},
	}
}

// t_buildCompleteModelTree creates a complete model tree that passes all completeness validations.
// This differs from t_buildValidModelTree in that every class has attributes, state machines, and transitions.
func t_buildCompleteModelTree() *inputModel {
	return &inputModel{
		Name: "Complete Model",
		Actors: map[string]*inputActor{
			"customer": {Name: "Customer", Type: "person"},
		},
		Domains: map[string]*inputDomain{
			"orders": {
				Name: "Orders",
				Subdomains: map[string]*inputSubdomain{
					"default": {
						Name: "Default",
						Classes: map[string]*inputClass{
							"order":     t_buildCompleteClass(),
							"line_item": t_buildCompleteClass(),
						},
						Generalizations: map[string]*inputGeneralization{},
						Associations: map[string]*inputAssociation{
							"order_lines": {
								Name:             "Order Lines",
								FromClassKey:     "order",
								FromMultiplicity: "1",
								ToClassKey:       "line_item",
								ToMultiplicity:   "1..*",
							},
						},
					},
				},
				Associations: map[string]*inputAssociation{},
			},
		},
		Associations: map[string]*inputAssociation{},
	}
}

// t_buildCompleteClass creates a complete class with all required elements.
func t_buildCompleteClass() *inputClass {
	toState := "active"
	return &inputClass{
		Name: "Complete Class",
		Attributes: map[string]*inputAttribute{
			"id": {Name: "ID", DataTypeRules: "int"},
		},
		StateMachine: &inputStateMachine{
			States: map[string]*inputState{
				"active": {Name: "Active"},
			},
			Events: map[string]*inputEvent{
				"create": {Name: "create"},
			},
			Guards:      map[string]*inputGuard{},
			Transitions: []inputTransition{
				{
					FromStateKey: nil, // Initial transition
					ToStateKey:   &toState,
					EventKey:     "create",
				},
			},
		},
		Actions: map[string]*inputAction{},
		Queries: map[string]*inputQuery{},
	}
}
