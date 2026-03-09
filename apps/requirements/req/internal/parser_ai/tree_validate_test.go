package parser_ai

import (
	"errors"
	"testing"

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
	model := t_buildValidModelTree()
	err := validateModelTree(model)
	suite.Require().NoError(err)
}

// TestClassActorNotFound verifies error when class references missing actor.
func (suite *TreeValidateSuite) TestClassActorNotFound() {
	model := t_buildMinimalModelTree()
	// Add class with invalid actor reference
	model.Domains["domain1"].Subdomains["subdomain1"].Classes["class1"].ActorKey = "nonexistent_actor"

	err := validateModelTree(model)
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeClassActorNotFound, parseErr.Code)
	suite.Equal("actor_key", parseErr.Field)
}

// TestClassActorValid verifies valid actor reference passes.
func (suite *TreeValidateSuite) TestClassActorValid() {
	model := t_buildMinimalModelTree()
	model.Actors["customer"] = &inputActor{Name: "Customer", Type: "person"}
	model.Domains["domain1"].Subdomains["subdomain1"].Classes["class1"].ActorKey = "customer"

	err := validateModelTree(model)
	suite.Require().NoError(err)
}

// TestClassIndexAttrNotFound verifies error when index references missing attribute.
func (suite *TreeValidateSuite) TestClassIndexAttrNotFound() {
	model := t_buildMinimalModelTree()
	// Add index referencing non-existent attribute
	model.Domains["domain1"].Subdomains["subdomain1"].Classes["class1"].Indexes = [][]string{{"missing_attr"}}

	err := validateModelTree(model)
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeClassIndexAttrNotFound, parseErr.Code)
	suite.Equal("indexes[0][0]", parseErr.Field)
}

// TestClassIndexDuplicateAttr verifies error when index has duplicate attributes.
func (suite *TreeValidateSuite) TestClassIndexDuplicateAttr() {
	model := t_buildMinimalModelTree()
	// Add attribute and index with duplicate
	model.Domains["domain1"].Subdomains["subdomain1"].Classes["class1"].Attributes = map[string]*inputAttribute{
		"id": {Name: "ID", DataTypeRules: "unconstrained"},
	}
	model.Domains["domain1"].Subdomains["subdomain1"].Classes["class1"].Indexes = [][]string{{"id", "id"}}

	err := validateModelTree(model)
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeClassIndexAttrNotFound, parseErr.Code)
	suite.Contains(parseErr.Message, "duplicate")
}

// TestStateMachineStateActionNotFound verifies error when state action references missing action.
func (suite *TreeValidateSuite) TestStateMachineStateActionNotFound() {
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
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeStateMachineActionNotFound, parseErr.Code)
	suite.Equal("states.pending.actions[0].action_key", parseErr.Field)
}

// TestTransitionNoStates verifies error when transition has neither from nor to state.
func (suite *TreeValidateSuite) TestTransitionNoStates() {
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
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeTransitionNoStates, parseErr.Code)
	suite.Equal("transitions[0]", parseErr.Field)
}

// TestTransitionFromStateNotFound verifies error when transition from_state_key doesn't exist.
func (suite *TreeValidateSuite) TestTransitionFromStateNotFound() {
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
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeStateMachineStateNotFound, parseErr.Code)
	suite.Equal("transitions[0].from_state_key", parseErr.Field)
}

// TestTransitionToStateNotFound verifies error when transition to_state_key doesn't exist.
func (suite *TreeValidateSuite) TestTransitionToStateNotFound() {
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
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeStateMachineStateNotFound, parseErr.Code)
	suite.Equal("transitions[0].to_state_key", parseErr.Field)
}

// TestTransitionEventNotFound verifies error when transition event_key doesn't exist.
func (suite *TreeValidateSuite) TestTransitionEventNotFound() {
	model := t_buildMinimalModelTree()
	class := model.Domains["domain1"].Subdomains["subdomain1"].Classes["class1"]
	toState := "pending"
	class.StateMachine = &inputStateMachine{
		States: map[string]*inputState{
			"pending": {Name: "Pending"},
		},
		Events: map[string]*inputEvent{},
		Guards: map[string]*inputGuard{},
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
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeStateMachineEventNotFound, parseErr.Code)
	suite.Equal("transitions[0].event_key", parseErr.Field)
}

// TestTransitionGuardNotFound verifies error when transition guard_key doesn't exist.
func (suite *TreeValidateSuite) TestTransitionGuardNotFound() {
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
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeStateMachineGuardNotFound, parseErr.Code)
	suite.Equal("transitions[0].guard_key", parseErr.Field)
}

// TestTransitionActionNotFound verifies error when transition action_key doesn't exist.
func (suite *TreeValidateSuite) TestTransitionActionNotFound() {
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
		Guards: map[string]*inputGuard{},
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
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeStateMachineActionNotFound, parseErr.Code)
	suite.Equal("transitions[0].action_key", parseErr.Field)
}

// TestActionUnreferenced verifies error when an action exists but is not referenced.
func (suite *TreeValidateSuite) TestActionUnreferenced() {
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
		Guards: map[string]*inputGuard{},
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
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeActionUnreferenced, parseErr.Code)
	suite.Equal("action_key", parseErr.Field)
	suite.Contains(parseErr.Message, "unreferenced_action")
	suite.Contains(parseErr.Message, "not referenced")
}

// TestActionReferencedByStateAction verifies that action referenced by state action passes.
func (suite *TreeValidateSuite) TestActionReferencedByStateAction() {
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
		Guards: map[string]*inputGuard{},
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
	suite.Require().NoError(err)
}

// TestActionReferencedByTransition verifies that action referenced by transition passes.
func (suite *TreeValidateSuite) TestActionReferencedByTransition() {
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
		Guards: map[string]*inputGuard{},
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
	suite.Require().NoError(err)
}

// TestGenSuperclassNotFound verifies error when generalization superclass doesn't exist.
func (suite *TreeValidateSuite) TestGenSuperclassNotFound() {
	model := t_buildMinimalModelTree()
	subdomain := model.Domains["domain1"].Subdomains["subdomain1"]
	subdomain.Classes["book"] = &inputClass{Name: "Book"}
	subdomain.ClassGeneralizations = map[string]*inputClassGeneralization{
		"medium": {
			Name:          "Medium",
			SuperclassKey: "missing_class",
			SubclassKeys:  []string{"book"},
		},
	}

	err := validateModelTree(model)
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeClassGenSuperclassNotFound, parseErr.Code)
	suite.Equal("superclass_key", parseErr.Field)
}

// TestGenSubclassNotFound verifies error when generalization subclass doesn't exist.
func (suite *TreeValidateSuite) TestGenSubclassNotFound() {
	model := t_buildMinimalModelTree()
	subdomain := model.Domains["domain1"].Subdomains["subdomain1"]
	subdomain.Classes["product"] = &inputClass{Name: "Product"}
	subdomain.ClassGeneralizations = map[string]*inputClassGeneralization{
		"medium": {
			Name:          "Medium",
			SuperclassKey: "product",
			SubclassKeys:  []string{"missing_class"},
		},
	}

	err := validateModelTree(model)
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeClassGenSubclassNotFound, parseErr.Code)
	suite.Equal("subclass_keys[0]", parseErr.Field)
}

// TestGenSubclassDuplicate verifies error when generalization has duplicate subclass.
func (suite *TreeValidateSuite) TestGenSubclassDuplicate() {
	model := t_buildMinimalModelTree()
	subdomain := model.Domains["domain1"].Subdomains["subdomain1"]
	subdomain.Classes["product"] = &inputClass{Name: "Product"}
	subdomain.Classes["book"] = &inputClass{Name: "Book"}
	subdomain.ClassGeneralizations = map[string]*inputClassGeneralization{
		"medium": {
			Name:          "Medium",
			SuperclassKey: "product",
			SubclassKeys:  []string{"book", "book"},
		},
	}

	err := validateModelTree(model)
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeClassGenSubclassDuplicate, parseErr.Code)
	suite.Equal("subclass_keys[1]", parseErr.Field)
}

// TestGenSuperclassIsSubclass verifies error when superclass is also listed as subclass.
func (suite *TreeValidateSuite) TestGenSuperclassIsSubclass() {
	model := t_buildMinimalModelTree()
	subdomain := model.Domains["domain1"].Subdomains["subdomain1"]
	subdomain.Classes["product"] = &inputClass{Name: "Product"}
	subdomain.Classes["book"] = &inputClass{Name: "Book"}
	subdomain.ClassGeneralizations = map[string]*inputClassGeneralization{
		"medium": {
			Name:          "Medium",
			SuperclassKey: "product",
			SubclassKeys:  []string{"book", "product"},
		},
	}

	err := validateModelTree(model)
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeClassGenSuperclassIsSubclass, parseErr.Code)
}

// TestSubdomainAssocFromClassNotFound verifies error when subdomain association from_class_key doesn't exist.
func (suite *TreeValidateSuite) TestSubdomainAssocFromClassNotFound() {
	model := t_buildMinimalModelTree()
	subdomain := model.Domains["domain1"].Subdomains["subdomain1"]
	subdomain.ClassAssociations = map[string]*inputClassAssociation{
		"test_assoc": {
			Name:             "Test Association",
			FromClassKey:     "missing_class",
			FromMultiplicity: "1",
			ToClassKey:       "class1",
			ToMultiplicity:   "*",
		},
	}

	err := validateModelTree(model)
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeAssocFromClassNotFound, parseErr.Code)
	suite.Equal("from_class_key", parseErr.Field)
}

// TestSubdomainAssocToClassNotFound verifies error when subdomain association to_class_key doesn't exist.
func (suite *TreeValidateSuite) TestSubdomainAssocToClassNotFound() {
	model := t_buildMinimalModelTree()
	subdomain := model.Domains["domain1"].Subdomains["subdomain1"]
	subdomain.ClassAssociations = map[string]*inputClassAssociation{
		"test_assoc": {
			Name:             "Test Association",
			FromClassKey:     "class1",
			FromMultiplicity: "1",
			ToClassKey:       "missing_class",
			ToMultiplicity:   "*",
		},
	}

	err := validateModelTree(model)
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeAssocToClassNotFound, parseErr.Code)
	suite.Equal("to_class_key", parseErr.Field)
}

// TestSubdomainAssocClassNotFound verifies error when subdomain association association_class_key doesn't exist.
func (suite *TreeValidateSuite) TestSubdomainAssocClassNotFound() {
	model := t_buildMinimalModelTree()
	subdomain := model.Domains["domain1"].Subdomains["subdomain1"]
	subdomain.Classes["class2"] = &inputClass{Name: "Class 2"}
	missingClass := "missing_class"
	subdomain.ClassAssociations = map[string]*inputClassAssociation{
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
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeAssocClassNotFound, parseErr.Code)
	suite.Equal("association_class_key", parseErr.Field)
}

// TestSubdomainAssocClassSameAsFromClass verifies error when association_class_key equals from_class_key.
func (suite *TreeValidateSuite) TestSubdomainAssocClassSameAsFromClass() {
	model := t_buildMinimalModelTree()
	subdomain := model.Domains["domain1"].Subdomains["subdomain1"]
	subdomain.Classes["class2"] = &inputClass{Name: "Class 2"}
	sameAsFrom := "class1"
	subdomain.ClassAssociations = map[string]*inputClassAssociation{
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
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeAssocClassSameAsEndpoint, parseErr.Code)
	suite.Equal("association_class_key", parseErr.Field)
	suite.Contains(parseErr.Message, "from_class_key")
}

// TestSubdomainAssocClassSameAsToClass verifies error when association_class_key equals to_class_key.
func (suite *TreeValidateSuite) TestSubdomainAssocClassSameAsToClass() {
	model := t_buildMinimalModelTree()
	subdomain := model.Domains["domain1"].Subdomains["subdomain1"]
	subdomain.Classes["class2"] = &inputClass{Name: "Class 2"}
	sameAsTo := "class2"
	subdomain.ClassAssociations = map[string]*inputClassAssociation{
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
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeAssocClassSameAsEndpoint, parseErr.Code)
	suite.Equal("association_class_key", parseErr.Field)
	suite.Contains(parseErr.Message, "to_class_key")
}

// TestSubdomainAssocMultiplicityInvalid verifies error when multiplicity format is invalid.
func (suite *TreeValidateSuite) TestSubdomainAssocMultiplicityInvalid() {
	model := t_buildMinimalModelTree()
	subdomain := model.Domains["domain1"].Subdomains["subdomain1"]
	subdomain.Classes["class2"] = &inputClass{Name: "Class 2"}
	subdomain.ClassAssociations = map[string]*inputClassAssociation{
		"test_assoc": {
			Name:             "Test Association",
			FromClassKey:     "class1",
			FromMultiplicity: "invalid",
			ToClassKey:       "class2",
			ToMultiplicity:   "*",
		},
	}

	err := validateModelTree(model)
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeAssocMultiplicityInvalid, parseErr.Code)
	suite.Equal("from_multiplicity", parseErr.Field)
}

// TestDomainAssocFromClassNotFound verifies error when domain association from_class_key doesn't exist.
func (suite *TreeValidateSuite) TestDomainAssocFromClassNotFound() {
	model := t_buildMinimalModelTree()
	// Add another subdomain with a class
	model.Domains["domain1"].Subdomains["subdomain2"] = &inputSubdomain{
		Name:                 "Subdomain 2",
		Classes:              map[string]*inputClass{"class2": {Name: "Class 2"}},
		ClassGeneralizations: map[string]*inputClassGeneralization{},
		ClassAssociations:    map[string]*inputClassAssociation{},
	}
	model.Domains["domain1"].ClassAssociations = map[string]*inputClassAssociation{
		"test_assoc": {
			Name:             "Test Association",
			FromClassKey:     "subdomain1/missing_class",
			FromMultiplicity: "1",
			ToClassKey:       "subdomain2/class2",
			ToMultiplicity:   "*",
		},
	}

	err := validateModelTree(model)
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeAssocFromClassNotFound, parseErr.Code)
	suite.Equal("from_class_key", parseErr.Field)
}

// TestDomainAssocInvalidKeyFormat verifies error when domain association has invalid key format.
func (suite *TreeValidateSuite) TestDomainAssocInvalidKeyFormat() {
	model := t_buildMinimalModelTree()
	model.Domains["domain1"].ClassAssociations = map[string]*inputClassAssociation{
		"test_assoc": {
			Name:             "Test Association",
			FromClassKey:     "class1", // Wrong format - should be subdomain/class
			FromMultiplicity: "1",
			ToClassKey:       "subdomain1/class1",
			ToMultiplicity:   "*",
		},
	}

	err := validateModelTree(model)
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeAssocFromClassNotFound, parseErr.Code)
}

// TestModelAssocFromClassNotFound verifies error when model association from_class_key doesn't exist.
func (suite *TreeValidateSuite) TestModelAssocFromClassNotFound() {
	model := t_buildMinimalModelTree()
	// Add another domain with subdomain and class
	model.Domains["domain2"] = &inputDomain{
		Name: "Domain 2",
		Subdomains: map[string]*inputSubdomain{
			"subdomain1": {
				Name:                 "Subdomain 1",
				Classes:              map[string]*inputClass{"class1": {Name: "Class 1"}},
				ClassGeneralizations: map[string]*inputClassGeneralization{},
				ClassAssociations:    map[string]*inputClassAssociation{},
			},
		},
		ClassAssociations: map[string]*inputClassAssociation{},
	}
	model.ClassAssociations = map[string]*inputClassAssociation{
		"test_assoc": {
			Name:             "Test Association",
			FromClassKey:     "domain1/subdomain1/missing_class",
			FromMultiplicity: "1",
			ToClassKey:       "domain2/subdomain1/class1",
			ToMultiplicity:   "*",
		},
	}

	err := validateModelTree(model)
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeAssocFromClassNotFound, parseErr.Code)
	suite.Equal("from_class_key", parseErr.Field)
}

// TestModelAssocToClassNotFound verifies error when model association to_class_key doesn't exist.
func (suite *TreeValidateSuite) TestModelAssocToClassNotFound() {
	model := t_buildMinimalModelTree()
	model.Domains["domain2"] = &inputDomain{
		Name: "Domain 2",
		Subdomains: map[string]*inputSubdomain{
			"subdomain1": {
				Name:                 "Subdomain 1",
				Classes:              map[string]*inputClass{},
				ClassGeneralizations: map[string]*inputClassGeneralization{},
				ClassAssociations:    map[string]*inputClassAssociation{},
			},
		},
		ClassAssociations: map[string]*inputClassAssociation{},
	}
	model.ClassAssociations = map[string]*inputClassAssociation{
		"test_assoc": {
			Name:             "Test Association",
			FromClassKey:     "domain1/subdomain1/class1",
			FromMultiplicity: "1",
			ToClassKey:       "domain2/subdomain1/missing_class",
			ToMultiplicity:   "*",
		},
	}

	err := validateModelTree(model)
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeAssocToClassNotFound, parseErr.Code)
	suite.Equal("to_class_key", parseErr.Field)
}

// TestModelAssocInvalidKeyFormat verifies error when model association has invalid key format.
func (suite *TreeValidateSuite) TestModelAssocInvalidKeyFormat() {
	model := t_buildMinimalModelTree()
	model.ClassAssociations = map[string]*inputClassAssociation{
		"test_assoc": {
			Name:             "Test Association",
			FromClassKey:     "subdomain1/class1", // Wrong format - should be domain/subdomain/class
			FromMultiplicity: "1",
			ToClassKey:       "domain1/subdomain1/class1",
			ToMultiplicity:   "*",
		},
	}

	err := validateModelTree(model)
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeAssocFromClassNotFound, parseErr.Code)
}

// TestMultiplicityValidation tests various multiplicity formats.
func (suite *TreeValidateSuite) TestMultiplicityValidation() {
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
		suite.Run(tt.name, func() {
			err := validateMultiplicity(tt.mult)
			if tt.expected {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// ======================================
// Completeness Validation Tests
// ======================================

// TestCompletenessValidModel verifies that a complete valid model passes completeness validation.
func (suite *TreeValidateSuite) TestCompletenessValidModel() {
	model := t_buildCompleteModelTree()
	err := validateModelCompleteness(model)
	suite.Require().NoError(err)
}

// TestCompletenessModelNoActors verifies error when model has no actors.
func (suite *TreeValidateSuite) TestCompletenessModelNoActors() {
	model := t_buildCompleteModelTree()
	model.Actors = map[string]*inputActor{} // Remove all actors

	err := validateModelCompleteness(model)
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeModelNoActors, parseErr.Code)
	suite.Equal("actors", parseErr.Field)
	suite.Contains(parseErr.Message, "at least one actor")
	suite.Contains(parseErr.Message, "actors/") // Check for guidance about file location
}

// TestCompletenessModelNoDomains verifies error when model has no domains.
func (suite *TreeValidateSuite) TestCompletenessModelNoDomains() {
	model := t_buildCompleteModelTree()
	model.Domains = map[string]*inputDomain{} // Remove all domains

	err := validateModelCompleteness(model)
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeModelNoDomains, parseErr.Code)
	suite.Equal("domains", parseErr.Field)
	suite.Contains(parseErr.Message, "at least one domain")
	suite.Contains(parseErr.Message, "domains/") // Check for guidance about file location
}

// TestCompletenessDomainNoSubdomains verifies error when domain has no subdomains.
func (suite *TreeValidateSuite) TestCompletenessDomainNoSubdomains() {
	model := t_buildCompleteModelTree()
	model.Domains["orders"].Subdomains = map[string]*inputSubdomain{} // Remove all subdomains

	err := validateModelCompleteness(model)
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeDomainNoSubdomains, parseErr.Code)
	suite.Equal("subdomains", parseErr.Field)
	suite.Contains(parseErr.Message, "at least one subdomain")
	suite.Contains(parseErr.Message, "orders") // Check for specific domain name
}

// TestSingleSubdomainNotDefault verifies error when a single subdomain is not named "default".
func (suite *TreeValidateSuite) TestSingleSubdomainNotDefault() {
	model := t_buildCompleteModelTree()
	// Rename "default" subdomain to something else (single subdomain case)
	model.Domains["orders"].Subdomains["core"] = model.Domains["orders"].Subdomains["default"]
	delete(model.Domains["orders"].Subdomains, "default")

	err := validateModelCompleteness(model)
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeSingleSubdomainNotDefault, parseErr.Code)
	suite.Equal("subdomain_key", parseErr.Field)
	suite.Contains(parseErr.Message, "core")
	suite.Contains(parseErr.Message, "must be renamed to 'default'")
	suite.Contains(parseErr.Message, "domains/orders/subdomains/core")
}

// TestMultipleSubdomainsHasDefault verifies error when multiple subdomains include one named "default".
func (suite *TreeValidateSuite) TestMultipleSubdomainsHasDefault() {
	model := t_buildCompleteModelTree()
	// Add a second subdomain while keeping "default" (multiple subdomains case)
	model.Domains["orders"].Subdomains["shipping"] = &inputSubdomain{
		Name:    "Shipping",
		Details: "Shipping subdomain",
		Classes: map[string]*inputClass{
			"shipment": t_buildCompleteClass(),
			"tracking": t_buildCompleteClass(),
		},
		ClassAssociations: map[string]*inputClassAssociation{
			"shipment_tracking": {
				Name:             "Shipment Tracking",
				FromClassKey:     "shipment",
				FromMultiplicity: "1",
				ToClassKey:       "tracking",
				ToMultiplicity:   "*",
			},
		},
	}

	err := validateModelCompleteness(model)
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeMultipleSubdomainsHasDefault, parseErr.Code)
	suite.Equal("subdomain_key", parseErr.Field)
	suite.Contains(parseErr.Message, "multiple subdomains")
	suite.Contains(parseErr.Message, "named 'default'")
	suite.Contains(parseErr.Message, "domains/orders/subdomains/default")
}

// TestMultipleSubdomainsValid verifies that multiple subdomains without "default" is valid.
func (suite *TreeValidateSuite) TestMultipleSubdomainsValid() {
	model := t_buildCompleteModelTree()
	// Rename "default" to "ordering" and add "shipping" subdomain
	model.Domains["orders"].Subdomains["ordering"] = model.Domains["orders"].Subdomains["default"]
	delete(model.Domains["orders"].Subdomains, "default")
	model.Domains["orders"].Subdomains["shipping"] = &inputSubdomain{
		Name:    "Shipping",
		Details: "Shipping subdomain",
		Classes: map[string]*inputClass{
			"shipment": t_buildCompleteClass(),
			"tracking": t_buildCompleteClass(),
		},
		ClassAssociations: map[string]*inputClassAssociation{
			"shipment_tracking": {
				Name:             "Shipment Tracking",
				FromClassKey:     "shipment",
				FromMultiplicity: "1",
				ToClassKey:       "tracking",
				ToMultiplicity:   "*",
			},
		},
	}

	err := validateModelCompleteness(model)
	// Should pass - multiple subdomains without "default" is valid
	suite.Require().NoError(err)
}

// TestCompletenessSubdomainTooFewClasses verifies error when subdomain has less than 2 classes.
func (suite *TreeValidateSuite) TestCompletenessSubdomainTooFewClasses() {
	model := t_buildCompleteModelTree()
	// Keep only one class
	model.Domains["orders"].Subdomains["default"].Classes = map[string]*inputClass{
		"order": t_buildCompleteClass(),
	}
	// Update association to use remaining classes
	model.Domains["orders"].Subdomains["default"].ClassAssociations = map[string]*inputClassAssociation{
		"self_ref": {
			Name:             "Self Ref",
			FromClassKey:     "order",
			FromMultiplicity: "1",
			ToClassKey:       "order",
			ToMultiplicity:   "*",
		},
	}

	err := validateModelCompleteness(model)
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeSubdomainTooFewClasses, parseErr.Code)
	suite.Equal("classes", parseErr.Field)
	suite.Contains(parseErr.Message, "at least 2 classes")
	suite.Contains(parseErr.Message, "has 1")
}

// TestCompletenessSubdomainNoAssociations verifies error when subdomain has no associations.
func (suite *TreeValidateSuite) TestCompletenessSubdomainNoAssociations() {
	model := t_buildCompleteModelTree()
	model.Domains["orders"].Subdomains["default"].ClassAssociations = map[string]*inputClassAssociation{} // Remove all associations

	err := validateModelCompleteness(model)
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeSubdomainNoAssociations, parseErr.Code)
	suite.Equal("associations", parseErr.Field)
	suite.Contains(parseErr.Message, "at least one association")
	suite.Contains(parseErr.Message, "associations/") // Check for guidance about file location
}

// TestCompletenessClassNoAttributes verifies error when class has no attributes.
func (suite *TreeValidateSuite) TestCompletenessClassNoAttributes() {
	model := t_buildCompleteModelTree()
	model.Domains["orders"].Subdomains["default"].Classes["order"].Attributes = map[string]*inputAttribute{} // Remove all attributes

	err := validateModelCompleteness(model)
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeClassNoAttributes, parseErr.Code)
	suite.Equal("attributes", parseErr.Field)
	suite.Contains(parseErr.Message, "at least one attribute")
	suite.Contains(parseErr.Message, "order") // Check for specific class name
}

// TestCompletenessClassNoStateMachine verifies error when class has no state machine.
func (suite *TreeValidateSuite) TestCompletenessClassNoStateMachine() {
	model := t_buildCompleteModelTree()
	model.Domains["orders"].Subdomains["default"].Classes["order"].StateMachine = nil // Remove state machine

	err := validateModelCompleteness(model)
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeClassNoStateMachine, parseErr.Code)
	suite.Equal("state_machine", parseErr.Field)
	suite.Contains(parseErr.Message, "must have a state machine")
	suite.Contains(parseErr.Message, "state_machine.json") // Check for guidance about file
}

// TestCompletenessStateMachineNoTransitions verifies error when state machine has no transitions.
func (suite *TreeValidateSuite) TestCompletenessStateMachineNoTransitions() {
	model := t_buildCompleteModelTree()
	model.Domains["orders"].Subdomains["default"].Classes["order"].StateMachine.Transitions = []inputTransition{} // Remove all transitions

	err := validateModelCompleteness(model)
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeStateMachineNoTransitions, parseErr.Code)
	suite.Equal("transitions", parseErr.Field)
	suite.Contains(parseErr.Message, "at least one transition")
}

// TestCompletenessAllErrorsProvideGuidance verifies all completeness errors provide helpful guidance.
func (suite *TreeValidateSuite) TestCompletenessAllErrorsProvideGuidance() {
	// Test each error type and verify it contains actionable guidance
	tests := []struct {
		name          string
		buildModel    func() *inputModel
		expectedCode  int
		shouldContain []string
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
				m.Domains["orders"].Subdomains["default"].ClassAssociations = map[string]*inputClassAssociation{
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
				m.Domains["orders"].Subdomains["default"].ClassAssociations = map[string]*inputClassAssociation{}
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
		suite.Run(tt.name, func() {
			model := tt.buildModel()
			err := validateModelCompleteness(model)
			suite.Require().Error(err)

			var parseErr *ParseError
			ok := errors.As(err, &parseErr)
			suite.True(ok, "error should be a ParseError")
			suite.Equal(tt.expectedCode, parseErr.Code, "error code should match")

			// Verify all expected guidance strings are present
			for _, s := range tt.shouldContain {
				suite.Contains(parseErr.Message, s,
					"error message should contain guidance: %s", s)
			}
		})
	}
}

// New tests for tree-level validations added in tree_validate.go.
func (suite *TreeValidateSuite) TestModelDomainAssocDomainNotFound() {
	model := t_buildMinimalModelTree()
	model.DomainAssociations = map[string]*inputDomainAssociation{
		"bad_da": {
			ProblemDomainKey:  "missing_domain",
			SolutionDomainKey: "domain1",
		},
	}

	err := validateModelTree(model)
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeDomainAssocDomainNotFound, parseErr.Code)
	suite.Equal("problem_domain_key", parseErr.Field)
}

func (suite *TreeValidateSuite) TestActorGenActorNotFound() {
	model := t_buildMinimalModelTree()
	model.ActorGeneralizations = map[string]*inputActorGeneralization{
		"ag1": {
			Name:          "AG1",
			SuperclassKey: "missing_actor",
			SubclassKeys:  []string{"sub_a"},
		},
	}

	err := validateModelTree(model)
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeActorGenActorNotFound, parseErr.Code)
	suite.Equal("superclass_key", parseErr.Field)
}

func (suite *TreeValidateSuite) TestScenarioStepReferencesValidation() {
	model := t_buildValidModelTree()

	// Prepare use case and scenario containers
	sub := model.Domains["orders"].Subdomains["default"]
	if sub.UseCases == nil {
		sub.UseCases = make(map[string]*inputUseCase)
	}

	// 1) object missing
	scMissingObj := &inputScenario{
		Name:    "ScMissingObj",
		Objects: map[string]*inputObject{},
		Steps: &inputStep{
			StepType:      "action",
			FromObjectKey: ptrString("missing_obj"),
		},
	}
	uc1 := &inputUseCase{Name: "UC1", Level: "mud", Scenarios: map[string]*inputScenario{"sc1": scMissingObj}}
	sub.UseCases = map[string]*inputUseCase{"uc_missing_obj": uc1}

	err := validateModelTree(model)
	suite.Require().Error(err)
	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeScenarioStepObjectNotFound, parseErr.Code)
	suite.Equal("steps.from_object_key", parseErr.Field)

	// 2) event missing on class
	// reuse same model but add a scenario with an object referencing class 'order'
	scEvent := &inputScenario{
		Name: "ScEvent",
		Objects: map[string]*inputObject{
			"o1": {ObjectNumber: 1, Name: "O1", ClassKey: "order"},
		},
		Steps: &inputStep{
			StepType:      "action",
			FromObjectKey: ptrString("o1"),
			EventKey:      ptrString("missing_event"),
		},
	}
	uc2 := &inputUseCase{Name: "UC2", Level: "mud", Scenarios: map[string]*inputScenario{"sc_event": scEvent}}
	sub.UseCases = map[string]*inputUseCase{"uc_event": uc2}

	err = validateModelTree(model)
	suite.Require().Error(err)
	ok = errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeScenarioStepEventNotFound, parseErr.Code)
	suite.Equal("steps.event_key", parseErr.Field)

	// 3) query missing on class
	scQuery := &inputScenario{
		Name: "ScQuery",
		Objects: map[string]*inputObject{
			"o1": {ObjectNumber: 1, Name: "O1", ClassKey: "order"},
		},
		Steps: &inputStep{
			StepType:      "action",
			FromObjectKey: ptrString("o1"),
			QueryKey:      ptrString("missing_query"),
		},
	}
	uc3 := &inputUseCase{Name: "UC3", Level: "mud", Scenarios: map[string]*inputScenario{"sc_query": scQuery}}
	sub.UseCases = map[string]*inputUseCase{"uc_query": uc3}

	err = validateModelTree(model)
	suite.Require().Error(err)
	ok = errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrTreeScenarioStepQueryNotFound, parseErr.Code)
	suite.Equal("steps.query_key", parseErr.Field)
}

// helper to get *string.
func ptrString(s string) *string { return &s }

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
						ClassGeneralizations: map[string]*inputClassGeneralization{},
						ClassAssociations:    map[string]*inputClassAssociation{},
					},
				},
				ClassAssociations: map[string]*inputClassAssociation{},
			},
		},
		ClassAssociations: map[string]*inputClassAssociation{},
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
									"id":     {Name: "ID", DataTypeRules: "unconstrained"},
									"status": {Name: "Status", DataTypeRules: "enum of active, pending, completed"},
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
										"has_items": {Name: "hasItems", Logic: inputLogic{Description: "Order has items", Notation: "tla_plus"}},
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
						ClassGeneralizations: map[string]*inputClassGeneralization{
							"product_type": {
								Name:          "Product Type",
								SuperclassKey: "product",
								SubclassKeys:  []string{"book"},
							},
						},
						ClassAssociations: map[string]*inputClassAssociation{
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
				ClassAssociations: map[string]*inputClassAssociation{},
			},
		},
		ClassAssociations: map[string]*inputClassAssociation{},
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
						ClassGeneralizations: map[string]*inputClassGeneralization{},
						ClassAssociations: map[string]*inputClassAssociation{
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
				ClassAssociations: map[string]*inputClassAssociation{},
			},
		},
		ClassAssociations: map[string]*inputClassAssociation{},
	}
}

// t_buildCompleteClass creates a complete class with all required elements.
func t_buildCompleteClass() *inputClass {
	toState := "active"
	return &inputClass{
		Name: "Complete Class",
		Attributes: map[string]*inputAttribute{
			"id": {Name: "ID", DataTypeRules: "unconstrained"},
		},
		StateMachine: &inputStateMachine{
			States: map[string]*inputState{
				"active": {Name: "Active"},
			},
			Events: map[string]*inputEvent{
				"create": {Name: "create"},
			},
			Guards: map[string]*inputGuard{},
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
