package parser_ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TreeValidateSuite tests the ValidateModelTree function.
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
	err := ValidateModelTree(model)
	assert.NoError(t, err)
}

// TestClassActorNotFound verifies error when class references missing actor.
func (suite *TreeValidateSuite) TestClassActorNotFound() {
	t := suite.T()

	model := t_buildMinimalModelTree()
	// Add class with invalid actor reference
	model.Domains["domain1"].Subdomains["subdomain1"].Classes["class1"].ActorKey = "nonexistent_actor"

	err := ValidateModelTree(model)
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

	err := ValidateModelTree(model)
	assert.NoError(t, err)
}

// TestClassIndexAttrNotFound verifies error when index references missing attribute.
func (suite *TreeValidateSuite) TestClassIndexAttrNotFound() {
	t := suite.T()

	model := t_buildMinimalModelTree()
	// Add index referencing non-existent attribute
	model.Domains["domain1"].Subdomains["subdomain1"].Classes["class1"].Indexes = [][]string{{"missing_attr"}}

	err := ValidateModelTree(model)
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

	err := ValidateModelTree(model)
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

	err := ValidateModelTree(model)
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

	err := ValidateModelTree(model)
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

	err := ValidateModelTree(model)
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

	err := ValidateModelTree(model)
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

	err := ValidateModelTree(model)
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

	err := ValidateModelTree(model)
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

	err := ValidateModelTree(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeStateMachineActionNotFound, parseErr.Code)
	assert.Equal(t, "transitions[0].action_key", parseErr.Field)
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

	err := ValidateModelTree(model)
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

	err := ValidateModelTree(model)
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

	err := ValidateModelTree(model)
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

	err := ValidateModelTree(model)
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

	err := ValidateModelTree(model)
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

	err := ValidateModelTree(model)
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

	err := ValidateModelTree(model)
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrTreeAssocClassNotFound, parseErr.Code)
	assert.Equal(t, "association_class_key", parseErr.Field)
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

	err := ValidateModelTree(model)
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

	err := ValidateModelTree(model)
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

	err := ValidateModelTree(model)
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

	err := ValidateModelTree(model)
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

	err := ValidateModelTree(model)
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

	err := ValidateModelTree(model)
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
