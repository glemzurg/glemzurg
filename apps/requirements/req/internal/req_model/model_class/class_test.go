package model_class

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestClassSuite(t *testing.T) {
	suite.Run(t, new(ClassSuite))
}

type ClassSuite struct {
	suite.Suite
}

// TestValidate tests all validation rules for Class.
func (suite *ClassSuite) TestValidate() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	validKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	genKey := helper.Must(identity.NewGeneralizationKey(subdomainKey, "gen1"))

	tests := []struct {
		testName string
		class    Class
		errstr   string
	}{
		{
			testName: "valid class",
			class: Class{
				Key:  validKey,
				Name: "Name",
			},
		},
		{
			testName: "error empty key",
			class: Class{
				Key:  identity.Key{},
				Name: "Name",
			},
			errstr: "'KeyType' failed on the 'required' tag",
		},
		{
			testName: "error wrong key type",
			class: Class{
				Key:  domainKey,
				Name: "Name",
			},
			errstr: "Key: invalid key type 'domain' for class.",
		},
		{
			testName: "error blank name",
			class: Class{
				Key:  validKey,
				Name: "",
			},
			errstr: "Name",
		},
		{
			testName: "error SuperclassOfKey and SubclassOfKey are the same",
			class: func() Class {
				genKey := helper.Must(identity.NewGeneralizationKey(subdomainKey, "gen1"))
				return Class{
					Key:             validKey,
					Name:            "Name",
					SuperclassOfKey: &genKey,
					SubclassOfKey:   &genKey,
				}
			}(),
			errstr: "SuperclassOfKey and SubclassOfKey cannot be the same",
		},
		{
			testName: "valid class with SuperclassOfKey referencing a generalization",
			class: Class{
				Key:             validKey,
				Name:            "Name",
				SuperclassOfKey: &genKey,
			},
		},
		{
			testName: "error ActorKey wrong key type",
			class: func() Class {
				wrongKey := domainKey
				return Class{
					Key:      validKey,
					Name:     "Name",
					ActorKey: &wrongKey,
				}
			}(),
			errstr: "ActorKey: invalid key type 'domain' for actor",
		},
		{
			testName: "error SuperclassOfKey wrong key type",
			class: func() Class {
				wrongKey := domainKey
				return Class{
					Key:             validKey,
					Name:            "Name",
					SuperclassOfKey: &wrongKey,
				}
			}(),
			errstr: "SuperclassOfKey: invalid key type 'domain' for class generalization",
		},
		{
			testName: "error SubclassOfKey wrong key type",
			class: func() Class {
				wrongKey := domainKey
				return Class{
					Key:           validKey,
					Name:          "Name",
					SubclassOfKey: &wrongKey,
				}
			}(),
			errstr: "SubclassOfKey: invalid key type 'domain' for class generalization",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.class.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewClass maps parameters correctly and calls Validate.
func (suite *ClassSuite) TestNew() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	key := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	actorKey := helper.Must(identity.NewActorKey("actor1"))
	superclassOfKey := helper.Must(identity.NewGeneralizationKey(subdomainKey, "gen1"))
	subclassOfKey := helper.Must(identity.NewGeneralizationKey(subdomainKey, "gen2"))

	// Test parameters are mapped correctly.
	class, err := NewClass(key, "Name", "Details", &actorKey, &superclassOfKey, &subclassOfKey, "UmlComment")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), Class{
		Key:             key,
		Name:            "Name",
		Details:         "Details",
		ActorKey:        &actorKey,
		SuperclassOfKey: &superclassOfKey,
		SubclassOfKey:   &subclassOfKey,
		UmlComment:      "UmlComment",
	}, class)

	// Test that Validate is called (invalid data should fail).
	_, err = NewClass(key, "", "Details", nil, nil, nil, "UmlComment")
	assert.ErrorContains(suite.T(), err, "Name")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *ClassSuite) TestValidateWithParent() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	validKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	otherSubdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "other_subdomain"))

	// Test that Validate is called.
	class := Class{
		Key:  validKey,
		Name: "", // Invalid
	}
	err := class.ValidateWithParent(&subdomainKey)
	assert.ErrorContains(suite.T(), err, "Name", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - class key has subdomain1 as parent, but we pass other_subdomain.
	class = Class{
		Key:  validKey,
		Name: "Name",
	}
	err = class.ValidateWithParent(&otherSubdomainKey)
	assert.ErrorContains(suite.T(), err, "does not match expected parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = class.ValidateWithParent(&subdomainKey)
	assert.NoError(suite.T(), err)

	// Test child Invariant validation propagates error.
	class = Class{
		Key:  validKey,
		Name: "Name",
		Invariants: []model_logic.Logic{
			{Key: identity.Key{}, Type: model_logic.LogicTypeAssessment, Description: "Desc.", Notation: model_logic.NotationTLAPlus}, // Invalid: empty key
		},
	}
	err = class.ValidateWithParent(&subdomainKey)
	assert.ErrorContains(suite.T(), err, "invariant 0", "Should validate child Invariants")

	// Test child Invariant with wrong parent key is caught.
	otherClassKey := helper.Must(identity.NewClassKey(subdomainKey, "other_class"))
	wrongInvKey := helper.Must(identity.NewClassInvariantKey(otherClassKey, "0"))
	class = Class{
		Key:  validKey,
		Name: "Name",
		Invariants: []model_logic.Logic{
			helper.Must(model_logic.NewLogic(wrongInvKey, model_logic.LogicTypeAssessment, "Desc.", "", model_logic.NotationTLAPlus, "")),
		},
	}
	err = class.ValidateWithParent(&subdomainKey)
	assert.ErrorContains(suite.T(), err, "invariant 0", "Should catch invariant with wrong parent key")

	// Test child Attribute validation propagates error.
	attrKey := helper.Must(identity.NewAttributeKey(validKey, "attr1"))
	class = Class{
		Key:  validKey,
		Name: "Name",
		Attributes: map[identity.Key]Attribute{
			attrKey: {Key: attrKey, Name: ""}, // Invalid: blank name
		},
	}
	err = class.ValidateWithParent(&subdomainKey)
	assert.ErrorContains(suite.T(), err, "Name", "Should validate child Attributes")

	// Test child Action validation propagates error.
	actionKey := helper.Must(identity.NewActionKey(validKey, "action1"))
	class = Class{
		Key:  validKey,
		Name: "Name",
		Actions: map[identity.Key]model_state.Action{
			actionKey: {Key: actionKey, Name: ""}, // Invalid: blank name
		},
	}
	err = class.ValidateWithParent(&subdomainKey)
	assert.ErrorContains(suite.T(), err, "Name", "Should validate child Actions")

	// Test child Query validation propagates error.
	queryKey := helper.Must(identity.NewQueryKey(validKey, "query1"))
	class = Class{
		Key:  validKey,
		Name: "Name",
		Queries: map[identity.Key]model_state.Query{
			queryKey: {Key: queryKey, Name: ""}, // Invalid: blank name
		},
	}
	err = class.ValidateWithParent(&subdomainKey)
	assert.ErrorContains(suite.T(), err, "Name", "Should validate child Queries")

	// Test child Event validation propagates error.
	eventKey := helper.Must(identity.NewEventKey(validKey, "event1"))
	class = Class{
		Key:  validKey,
		Name: "Name",
		Events: map[identity.Key]model_state.Event{
			eventKey: {Key: eventKey, Name: ""}, // Invalid: blank name
		},
	}
	err = class.ValidateWithParent(&subdomainKey)
	assert.ErrorContains(suite.T(), err, "Name", "Should validate child Events")

	// Test child Guard validation propagates error.
	guardKey := helper.Must(identity.NewGuardKey(validKey, "guard1"))
	class = Class{
		Key:  validKey,
		Name: "Name",
		Guards: map[identity.Key]model_state.Guard{
			guardKey: {Key: guardKey, Name: ""}, // Invalid: blank name
		},
	}
	err = class.ValidateWithParent(&subdomainKey)
	assert.ErrorContains(suite.T(), err, "Name", "Should validate child Guards")

	// Test child State validation propagates error.
	stateKey := helper.Must(identity.NewStateKey(validKey, "state1"))
	class = Class{
		Key:  validKey,
		Name: "Name",
		States: map[identity.Key]model_state.State{
			stateKey: {Key: stateKey, Name: ""}, // Invalid: blank name
		},
	}
	err = class.ValidateWithParent(&subdomainKey)
	assert.ErrorContains(suite.T(), err, "Name", "Should validate child States")

	// Test child Transition validation propagates error (bad event key).
	transitionKey := helper.Must(identity.NewTransitionKey(validKey, "state1", "event1", "", "", "state2"))
	class = Class{
		Key:  validKey,
		Name: "Name",
		Transitions: map[identity.Key]model_state.Transition{
			transitionKey: {Key: transitionKey, EventKey: identity.Key{}}, // Invalid: empty event key
		},
	}
	err = class.ValidateWithParent(&subdomainKey)
	assert.Error(suite.T(), err, "Should validate child Transitions")

	// Test valid class with all child types.
	invKey := helper.Must(identity.NewClassInvariantKey(validKey, "0"))
	validInvariant := helper.Must(model_logic.NewLogic(invKey, model_logic.LogicTypeAssessment, "Desc.", "", model_logic.NotationTLAPlus, ""))
	validLogic := helper.Must(model_logic.NewLogic(guardKey, model_logic.LogicTypeAssessment, "Desc.", "", model_logic.NotationTLAPlus, ""))
	validAction := helper.Must(model_state.NewAction(actionKey, "Action", "", nil, nil, nil, nil))
	validEvent := helper.Must(model_state.NewEvent(eventKey, "Event", "", nil))
	validState := helper.Must(model_state.NewState(stateKey, "State", "", ""))
	validGuard := helper.Must(model_state.NewGuard(guardKey, "Guard", validLogic))
	validQuery := helper.Must(model_state.NewQuery(queryKey, "Query", "", nil, nil, nil))
	validAttr := Attribute{Key: attrKey, Name: "Attr"}
	validTransition := helper.Must(model_state.NewTransition(transitionKey, &stateKey, eventKey, nil, nil, &stateKey, ""))
	class = Class{
		Key:         validKey,
		Name:        "Name",
		Invariants:  []model_logic.Logic{validInvariant},
		Attributes:  map[identity.Key]Attribute{attrKey: validAttr},
		States:      map[identity.Key]model_state.State{stateKey: validState},
		Events:      map[identity.Key]model_state.Event{eventKey: validEvent},
		Guards:      map[identity.Key]model_state.Guard{guardKey: validGuard},
		Actions:     map[identity.Key]model_state.Action{actionKey: validAction},
		Queries:     map[identity.Key]model_state.Query{queryKey: validQuery},
		Transitions: map[identity.Key]model_state.Transition{transitionKey: validTransition},
	}
	err = class.ValidateWithParent(&subdomainKey)
	assert.NoError(suite.T(), err, "Valid class with all children should pass")

	// Test guard logic key mismatch is caught through class validation.
	otherGuardKey := helper.Must(identity.NewGuardKey(validKey, "other_guard"))
	mismatchedLogic := helper.Must(model_logic.NewLogic(otherGuardKey, model_logic.LogicTypeAssessment, "Desc.", "", model_logic.NotationTLAPlus, ""))
	class = Class{
		Key:  validKey,
		Name: "Name",
		Guards: map[identity.Key]model_state.Guard{
			guardKey: helper.Must(model_state.NewGuard(guardKey, "Guard", mismatchedLogic)),
		},
	}
	err = class.ValidateWithParent(&subdomainKey)
	assert.ErrorContains(suite.T(), err, "does not match guard key", "Should catch guard logic key mismatch")

	// Test action require key with wrong parent is caught.
	otherActionKey := helper.Must(identity.NewActionKey(validKey, "other_action"))
	wrongReqKey := helper.Must(identity.NewActionRequireKey(otherActionKey, "req_1"))
	wrongReqLogic := helper.Must(model_logic.NewLogic(wrongReqKey, model_logic.LogicTypeAssessment, "Precondition.", "", model_logic.NotationTLAPlus, ""))
	class = Class{
		Key:  validKey,
		Name: "Name",
		Actions: map[identity.Key]model_state.Action{
			actionKey: helper.Must(model_state.NewAction(actionKey, "Action", "", []model_logic.Logic{wrongReqLogic}, nil, nil, nil)),
		},
	}
	err = class.ValidateWithParent(&subdomainKey)
	assert.ErrorContains(suite.T(), err, "requires 0", "Should catch action require key with wrong parent")

	// Test query guarantee key with wrong parent is caught.
	otherQueryKey := helper.Must(identity.NewQueryKey(validKey, "other_query"))
	wrongGuarKey := helper.Must(identity.NewQueryGuaranteeKey(otherQueryKey, "guar_1"))
	wrongGuarLogic := helper.Must(model_logic.NewLogic(wrongGuarKey, model_logic.LogicTypeQuery, "Guarantee.", "result", model_logic.NotationTLAPlus, ""))
	class = Class{
		Key:  validKey,
		Name: "Name",
		Queries: map[identity.Key]model_state.Query{
			queryKey: helper.Must(model_state.NewQuery(queryKey, "Query", "", nil, []model_logic.Logic{wrongGuarLogic}, nil)),
		},
	}
	err = class.ValidateWithParent(&subdomainKey)
	assert.ErrorContains(suite.T(), err, "guarantee 0", "Should catch query guarantee key with wrong parent")

	// Test attribute derivation policy key with wrong parent is caught.
	otherAttrKey := helper.Must(identity.NewAttributeKey(validKey, "other_attr"))
	wrongDerivKey := helper.Must(identity.NewAttributeDerivationKey(otherAttrKey, "deriv1"))
	wrongDerivLogic := helper.Must(model_logic.NewLogic(wrongDerivKey, model_logic.LogicTypeStateChange, "Computed.", "field", model_logic.NotationTLAPlus, ""))
	class = Class{
		Key:  validKey,
		Name: "Name",
		Attributes: map[identity.Key]Attribute{
			attrKey: {Key: attrKey, Name: "Attr", DerivationPolicy: &wrongDerivLogic},
		},
	}
	err = class.ValidateWithParent(&subdomainKey)
	assert.ErrorContains(suite.T(), err, "DerivationPolicy", "Should catch attribute derivation policy key with wrong parent")
}

// TestSetters tests that all Set* methods correctly set their fields.
func (suite *ClassSuite) TestSetters() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	class := Class{Key: classKey, Name: "Name"}

	attrKey := helper.Must(identity.NewAttributeKey(classKey, "attr1"))
	stateKey := helper.Must(identity.NewStateKey(classKey, "state1"))
	eventKey := helper.Must(identity.NewEventKey(classKey, "event1"))
	guardKey := helper.Must(identity.NewGuardKey(classKey, "guard1"))
	actionKey := helper.Must(identity.NewActionKey(classKey, "action1"))
	queryKey := helper.Must(identity.NewQueryKey(classKey, "query1"))
	transitionKey := helper.Must(identity.NewTransitionKey(classKey, "state1", "event1", "", "", "state1"))

	invKey := helper.Must(identity.NewClassInvariantKey(classKey, "0"))
	invariants := []model_logic.Logic{helper.Must(model_logic.NewLogic(invKey, model_logic.LogicTypeAssessment, "Desc.", "", model_logic.NotationTLAPlus, ""))}
	class.SetInvariants(invariants)
	assert.Equal(suite.T(), invariants, class.Invariants)

	attrs := map[identity.Key]Attribute{attrKey: {Key: attrKey, Name: "Attr"}}
	class.SetAttributes(attrs)
	assert.Equal(suite.T(), attrs, class.Attributes)

	states := map[identity.Key]model_state.State{stateKey: helper.Must(model_state.NewState(stateKey, "State", "", ""))}
	class.SetStates(states)
	assert.Equal(suite.T(), states, class.States)

	events := map[identity.Key]model_state.Event{eventKey: helper.Must(model_state.NewEvent(eventKey, "Event", "", nil))}
	class.SetEvents(events)
	assert.Equal(suite.T(), events, class.Events)

	guards := map[identity.Key]model_state.Guard{guardKey: {Key: guardKey, Name: "Guard"}}
	class.SetGuards(guards)
	assert.Equal(suite.T(), guards, class.Guards)

	actions := map[identity.Key]model_state.Action{actionKey: helper.Must(model_state.NewAction(actionKey, "Action", "", nil, nil, nil, nil))}
	class.SetActions(actions)
	assert.Equal(suite.T(), actions, class.Actions)

	queries := map[identity.Key]model_state.Query{queryKey: helper.Must(model_state.NewQuery(queryKey, "Query", "", nil, nil, nil))}
	class.SetQueries(queries)
	assert.Equal(suite.T(), queries, class.Queries)

	transitions := map[identity.Key]model_state.Transition{transitionKey: {Key: transitionKey, EventKey: eventKey}}
	class.SetTransitions(transitions)
	assert.Equal(suite.T(), transitions, class.Transitions)
}

// TestValidateReferences tests that ValidateReferences validates cross-references correctly.
func (suite *ClassSuite) TestValidateReferences() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	otherSubdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "other_subdomain"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	genKey := helper.Must(identity.NewGeneralizationKey(subdomainKey, "gen1"))
	genInOtherSubdomain := helper.Must(identity.NewGeneralizationKey(otherSubdomainKey, "gen2"))
	actorKey := helper.Must(identity.NewActorKey("actor1"))
	nonExistentActorKey := helper.Must(identity.NewActorKey("nonexistent"))
	nonExistentGenKey := helper.Must(identity.NewGeneralizationKey(subdomainKey, "nonexistent"))

	// Build lookup maps
	actors := map[identity.Key]bool{
		actorKey: true,
	}
	generalizations := map[identity.Key]bool{
		genKey: true,
	}

	tests := []struct {
		testName        string
		class           Class
		actors          map[identity.Key]bool
		generalizations map[identity.Key]bool
		errstr          string
	}{
		{
			testName: "valid class with no references",
			class: Class{
				Key:  classKey,
				Name: "Name",
			},
			actors:          actors,
			generalizations: generalizations,
		},
		{
			testName: "valid class with ActorKey reference",
			class: Class{
				Key:      classKey,
				Name:     "Name",
				ActorKey: &actorKey,
			},
			actors:          actors,
			generalizations: generalizations,
		},
		{
			testName: "error ActorKey references non-existent actor",
			class: Class{
				Key:      classKey,
				Name:     "Name",
				ActorKey: &nonExistentActorKey,
			},
			actors:          actors,
			generalizations: generalizations,
			errstr:          "references non-existent actor",
		},
		{
			testName: "valid class with SuperclassOfKey reference",
			class: Class{
				Key:             classKey,
				Name:            "Name",
				SuperclassOfKey: &genKey,
			},
			actors:          actors,
			generalizations: generalizations,
		},
		{
			testName: "error SuperclassOfKey references non-existent generalization",
			class: Class{
				Key:             classKey,
				Name:            "Name",
				SuperclassOfKey: &nonExistentGenKey,
			},
			actors:          actors,
			generalizations: generalizations,
			errstr:          "references non-existent generalization",
		},
		{
			testName: "error SuperclassOfKey references generalization in different subdomain",
			class: Class{
				Key:             classKey,
				Name:            "Name",
				SuperclassOfKey: &genInOtherSubdomain,
			},
			actors: actors,
			generalizations: map[identity.Key]bool{
				genKey:              true,
				genInOtherSubdomain: true,
			},
			errstr: "must be in the same subdomain",
		},
		{
			testName: "valid class with SubclassOfKey reference",
			class: Class{
				Key:           classKey,
				Name:          "Name",
				SubclassOfKey: &genKey,
			},
			actors:          actors,
			generalizations: generalizations,
		},
		{
			testName: "error SubclassOfKey references non-existent generalization",
			class: Class{
				Key:           classKey,
				Name:          "Name",
				SubclassOfKey: &nonExistentGenKey,
			},
			actors:          actors,
			generalizations: generalizations,
			errstr:          "references non-existent generalization",
		},
		{
			testName: "error SubclassOfKey references generalization in different subdomain",
			class: Class{
				Key:           classKey,
				Name:          "Name",
				SubclassOfKey: &genInOtherSubdomain,
			},
			actors: actors,
			generalizations: map[identity.Key]bool{
				genKey:              true,
				genInOtherSubdomain: true,
			},
			errstr: "must be in the same subdomain",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.class.ValidateReferences(tt.actors, tt.generalizations)
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}
