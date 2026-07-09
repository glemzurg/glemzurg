package model_class

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
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
			errstr: "key type is required",
		},
		{
			testName: "error wrong key type",
			class: Class{
				Key:  domainKey,
				Name: "Name",
			},
			errstr: "key: invalid key type 'domain' for class",
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
		suite.Run(tt.testName, func() {
			ctx := coreerr.NewContext("test", "")
			err := tt.class.Validate(ctx)
			if tt.errstr == "" {
				suite.Require().NoError(err)
			} else {
				suite.Require().ErrorContains(err, tt.errstr)
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
	class := NewClass(key, ClassLinks{ActorKey: &actorKey, SuperclassOfKey: &superclassOfKey, SubclassOfKey: &subclassOfKey}, ClassDetails{Name: "Name", Details: "Details", UnfinishedNotes: "", UmlComment: "UmlComment"})
	suite.Equal(Class{
		Key:             key,
		Name:            "Name",
		Details:         "Details",
		ActorKey:        &actorKey,
		SuperclassOfKey: &superclassOfKey,
		SubclassOfKey:   &subclassOfKey,
		UmlComment:      "UmlComment",
	}, class)
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *ClassSuite) TestValidateWithParent() {
	ctx := coreerr.NewContext("test", "")
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	validKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	otherSubdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "other_subdomain"))

	// Test that Validate is called.
	class := Class{
		Key:  validKey,
		Name: "", // Invalid
	}
	err := class.ValidateWithParent(ctx, &subdomainKey, nil, nil)
	suite.Require().ErrorContains(err, "Name", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - class key has subdomain1 as parent, but we pass other_subdomain.
	class = Class{
		Key:  validKey,
		Name: "Name",
	}
	err = class.ValidateWithParent(ctx, &otherSubdomainKey, nil, nil)
	suite.Require().ErrorContains(err, "does not match expected parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = class.ValidateWithParent(ctx, &subdomainKey, nil, nil)
	suite.Require().NoError(err)

	// Test child Invariant validation propagates error.
	class = Class{
		Key:  validKey,
		Name: "Name",
		Invariants: []model_logic.Logic{
			{Key: identity.Key{}, Type: model_logic.LogicTypeAssessment, Description: "Desc.", Spec: logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}}, // Invalid: empty key
		},
	}
	err = class.ValidateWithParent(ctx, &subdomainKey, nil, nil)
	suite.Require().ErrorContains(err, "invariant 0", "Should validate child Invariants")

	// Test child Invariant with wrong parent key is caught.
	otherClassKey := helper.Must(identity.NewClassKey(subdomainKey, "other_class"))
	wrongInvKey := helper.Must(identity.NewClassInvariantKey(otherClassKey, "0"))
	class = Class{
		Key:  validKey,
		Name: "Name",
		Invariants: []model_logic.Logic{
			model_logic.NewLogic(wrongInvKey, model_logic.LogicTypeAssessment, "Desc.", "", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil),
		},
	}
	err = class.ValidateWithParent(ctx, &subdomainKey, nil, nil)
	suite.Require().ErrorContains(err, "invariant 0", "Should catch invariant with wrong parent key")

	// Test valid class with let in invariants.
	letInvKey1 := helper.Must(identity.NewClassInvariantKey(validKey, "0"))
	letInvKey2 := helper.Must(identity.NewClassInvariantKey(validKey, "1"))
	class = Class{
		Key:  validKey,
		Name: "Name",
		Invariants: []model_logic.Logic{
			model_logic.NewLogic(letInvKey1, model_logic.LogicTypeLet, "Local total.", "total", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "1 + 2"}, nil),
			model_logic.NewLogic(letInvKey2, model_logic.LogicTypeAssessment, "Must be positive.", "", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil),
		},
	}
	err = class.ValidateWithParent(ctx, &subdomainKey, nil, nil)
	suite.Require().NoError(err, "Class with let in invariants should be valid")

	// Test duplicate let target in class invariants.
	class = Class{
		Key:  validKey,
		Name: "Name",
		Invariants: []model_logic.Logic{
			model_logic.NewLogic(letInvKey1, model_logic.LogicTypeLet, "Local a.", "a", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "1"}, nil),
			model_logic.NewLogic(letInvKey2, model_logic.LogicTypeLet, "Local a again.", "a", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus, Specification: "2"}, nil),
		},
	}
	err = class.ValidateWithParent(ctx, &subdomainKey, nil, nil)
	suite.Require().ErrorContains(err, "duplicate let target \"a\"", "Should catch duplicate let target in class invariants")

	// Test child Attribute validation propagates error.
	attrKey := helper.Must(identity.NewAttributeKey(validKey, "attr1"))
	class = Class{
		Key:  validKey,
		Name: "Name",
		Attributes: []Attribute{
			{Key: attrKey, Name: ""}, // Invalid: blank name
		},
	}
	err = class.ValidateWithParent(ctx, &subdomainKey, nil, nil)
	suite.Require().ErrorContains(err, "Name", "Should validate child Attributes")

	// Test child Action validation propagates error.
	actionKey := helper.Must(identity.NewActionKey(validKey, "action1"))
	class = Class{
		Key:  validKey,
		Name: "Name",
		Actions: map[identity.Key]model_state.Action{
			actionKey: {Key: actionKey, Name: ""}, // Invalid: blank name
		},
	}
	err = class.ValidateWithParent(ctx, &subdomainKey, nil, nil)
	suite.Require().ErrorContains(err, "Name", "Should validate child Actions")

	// Test child Query validation propagates error.
	queryKey := helper.Must(identity.NewQueryKey(validKey, "query1"))
	class = Class{
		Key:  validKey,
		Name: "Name",
		Queries: map[identity.Key]model_state.Query{
			queryKey: {Key: queryKey, Name: ""}, // Invalid: blank name
		},
	}
	err = class.ValidateWithParent(ctx, &subdomainKey, nil, nil)
	suite.Require().ErrorContains(err, "Name", "Should validate child Queries")

	// Test child Event validation propagates error.
	eventKey := helper.Must(identity.NewEventKey(validKey, "event1"))
	class = Class{
		Key:  validKey,
		Name: "Name",
		Events: map[identity.Key]model_state.Event{
			eventKey: {Key: eventKey, Name: ""}, // Invalid: blank name
		},
	}
	err = class.ValidateWithParent(ctx, &subdomainKey, nil, nil)
	suite.Require().ErrorContains(err, "Name", "Should validate child Events")

	// Test child Guard validation propagates error.
	guardKey := helper.Must(identity.NewGuardKey(validKey, "guard1"))
	class = Class{
		Key:  validKey,
		Name: "Name",
		Guards: map[identity.Key]model_state.Guard{
			guardKey: {Key: guardKey, Name: ""}, // Invalid: blank name
		},
	}
	err = class.ValidateWithParent(ctx, &subdomainKey, nil, nil)
	suite.Require().ErrorContains(err, "Name", "Should validate child Guards")

	// Test child State validation propagates error.
	stateKey := helper.Must(identity.NewStateKey(validKey, "state1"))
	class = Class{
		Key:  validKey,
		Name: "Name",
		States: map[identity.Key]model_state.State{
			stateKey: {Key: stateKey, Name: ""}, // Invalid: blank name
		},
	}
	err = class.ValidateWithParent(ctx, &subdomainKey, nil, nil)
	suite.Require().ErrorContains(err, "Name", "Should validate child States")

	// Test child Transition validation propagates error (bad event key).
	transitionKey := helper.Must(identity.NewTransitionKey(validKey, "state1", "event1", "", "", "state2"))
	class = Class{
		Key:  validKey,
		Name: "Name",
		Transitions: map[identity.Key]model_state.Transition{
			transitionKey: {Key: transitionKey, EventKey: identity.Key{}}, // Invalid: empty event key
		},
	}
	err = class.ValidateWithParent(ctx, &subdomainKey, nil, nil)
	suite.Require().Error(err, "Should validate child Transitions")

	// Test valid class with all child types.
	invKey := helper.Must(identity.NewClassInvariantKey(validKey, "0"))
	validInvariant := model_logic.NewLogic(invKey, model_logic.LogicTypeAssessment, "Desc.", "", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil)
	validLogic := model_logic.NewLogic(guardKey, model_logic.LogicTypeAssessment, "Desc.", "", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil)
	validAction := model_state.NewAction(actionKey, model_state.ActionDetails{Name: "Action", Details: ""}, nil, nil, nil, nil)
	validEvent := model_state.NewEvent(eventKey, "Event", "", nil)
	validState := model_state.NewState(stateKey, "State", "", "")
	validGuard := model_state.NewGuard(guardKey, "Guard", validLogic)
	validQuery := model_state.NewQuery(queryKey, "Query", "", nil, nil, nil)
	validAttr := Attribute{Key: attrKey, Name: "Attr"}
	validTransition := model_state.NewTransition(transitionKey, eventKey, model_state.TransitionStateKeys{FromStateKey: &stateKey, ToStateKey: &stateKey}, model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: nil}, "")
	class = Class{
		Key:         validKey,
		Name:        "Name",
		Invariants:  []model_logic.Logic{validInvariant},
		Attributes:  []Attribute{validAttr},
		States:      map[identity.Key]model_state.State{stateKey: validState},
		Events:      map[identity.Key]model_state.Event{eventKey: validEvent},
		Guards:      map[identity.Key]model_state.Guard{guardKey: validGuard},
		Actions:     map[identity.Key]model_state.Action{actionKey: validAction},
		Queries:     map[identity.Key]model_state.Query{queryKey: validQuery},
		Transitions: map[identity.Key]model_state.Transition{transitionKey: validTransition},
	}
	err = class.ValidateWithParent(ctx, &subdomainKey, nil, nil)
	suite.Require().NoError(err, "Valid class with all children should pass")

	// Test guard logic key mismatch is caught through class validation.
	otherGuardKey := helper.Must(identity.NewGuardKey(validKey, "other_guard"))
	mismatchedLogic := model_logic.NewLogic(otherGuardKey, model_logic.LogicTypeAssessment, "Desc.", "", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil)
	class = Class{
		Key:  validKey,
		Name: "Name",
		Guards: map[identity.Key]model_state.Guard{
			guardKey: model_state.NewGuard(guardKey, "Guard", mismatchedLogic),
		},
	}
	err = class.ValidateWithParent(ctx, &subdomainKey, nil, nil)
	suite.Require().ErrorContains(err, "does not match guard key", "Should catch guard logic key mismatch")

	// Test action require key with wrong parent is caught.
	otherActionKey := helper.Must(identity.NewActionKey(validKey, "other_action"))
	wrongReqKey := helper.Must(identity.NewActionRequireKey(otherActionKey, "req_1"))
	wrongReqLogic := model_logic.NewLogic(wrongReqKey, model_logic.LogicTypeAssessment, "Precondition.", "", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil)
	class = Class{
		Key:  validKey,
		Name: "Name",
		Actions: map[identity.Key]model_state.Action{
			actionKey: model_state.NewAction(actionKey, model_state.ActionDetails{Name: "Action", Details: ""}, []model_logic.Logic{wrongReqLogic}, nil, nil, nil),
		},
	}
	err = class.ValidateWithParent(ctx, &subdomainKey, nil, nil)
	suite.Require().ErrorContains(err, "requires[0]", "Should catch action require key with wrong parent")

	// Test query guarantee key with wrong parent is caught.
	otherQueryKey := helper.Must(identity.NewQueryKey(validKey, "other_query"))
	wrongGuarKey := helper.Must(identity.NewQueryGuaranteeKey(otherQueryKey, "guar_1"))
	wrongGuarLogic := model_logic.NewLogic(wrongGuarKey, model_logic.LogicTypeQuery, "Guarantee.", "result", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil)
	class = Class{
		Key:  validKey,
		Name: "Name",
		Queries: map[identity.Key]model_state.Query{
			queryKey: model_state.NewQuery(queryKey, "Query", "", nil, []model_logic.Logic{wrongGuarLogic}, nil),
		},
	}
	err = class.ValidateWithParent(ctx, &subdomainKey, nil, nil)
	suite.Require().ErrorContains(err, "guarantees[0]", "Should catch query guarantee key with wrong parent")

	// Test attribute derivation policy key with wrong parent is caught.
	otherAttrKey := helper.Must(identity.NewAttributeKey(validKey, "other_attr"))
	wrongDerivKey := helper.Must(identity.NewAttributeDerivationKey(otherAttrKey, "deriv1"))
	wrongDerivLogic := model_logic.NewLogic(wrongDerivKey, model_logic.LogicTypeStateChange, "Computed.", "field", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil)
	class = Class{
		Key:  validKey,
		Name: "Name",
		Attributes: []Attribute{
			{Key: attrKey, Name: "Attr", DerivationPolicy: &wrongDerivLogic},
		},
	}
	err = class.ValidateWithParent(ctx, &subdomainKey, nil, nil)
	suite.Require().ErrorContains(err, "DerivationPolicy", "Should catch attribute derivation policy key with wrong parent")
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
	invariants := []model_logic.Logic{model_logic.NewLogic(invKey, model_logic.LogicTypeAssessment, "Desc.", "", logic_spec.ExpressionSpec{Notation: model_logic.NotationTLAPlus}, nil)}
	class.SetInvariants(invariants)
	suite.Equal(invariants, class.Invariants)

	attrs := []Attribute{{Key: attrKey, Name: "Attr"}}
	class.SetAttributes(attrs)
	suite.Equal(attrs, class.Attributes)

	states := map[identity.Key]model_state.State{stateKey: model_state.NewState(stateKey, "State", "", "")}
	class.SetStates(states)
	suite.Equal(states, class.States)

	events := map[identity.Key]model_state.Event{eventKey: model_state.NewEvent(eventKey, "Event", "", nil)}
	class.SetEvents(events)
	suite.Equal(events, class.Events)

	guards := map[identity.Key]model_state.Guard{guardKey: {Key: guardKey, Name: "Guard"}}
	class.SetGuards(guards)
	suite.Equal(guards, class.Guards)

	actions := map[identity.Key]model_state.Action{actionKey: model_state.NewAction(actionKey, model_state.ActionDetails{Name: "Action", Details: ""}, nil, nil, nil, nil)}
	class.SetActions(actions)
	suite.Equal(actions, class.Actions)

	queries := map[identity.Key]model_state.Query{queryKey: model_state.NewQuery(queryKey, "Query", "", nil, nil, nil)}
	class.SetQueries(queries)
	suite.Equal(queries, class.Queries)

	transitions := map[identity.Key]model_state.Transition{transitionKey: {Key: transitionKey, EventKey: eventKey}}
	class.SetTransitions(transitions)
	suite.Equal(transitions, class.Transitions)
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
	localGeneralizations := map[identity.Key]bool{
		genKey: true,
	}
	allGeneralizations := map[identity.Key]bool{
		genKey:              true,
		genInOtherSubdomain: true,
	}

	tests := []struct {
		testName             string
		class                Class
		actors               map[identity.Key]bool
		localGeneralizations map[identity.Key]bool
		allGeneralizations   map[identity.Key]bool
		errstr               string
	}{
		{
			testName: "valid class with no references",
			class: Class{
				Key:  classKey,
				Name: "Name",
			},
			actors:               actors,
			localGeneralizations: localGeneralizations,
			allGeneralizations:   allGeneralizations,
		},
		{
			testName: "valid class with ActorKey reference",
			class: Class{
				Key:      classKey,
				Name:     "Name",
				ActorKey: &actorKey,
			},
			actors:               actors,
			localGeneralizations: localGeneralizations,
			allGeneralizations:   allGeneralizations,
		},
		{
			testName: "error ActorKey references non-existent actor",
			class: Class{
				Key:      classKey,
				Name:     "Name",
				ActorKey: &nonExistentActorKey,
			},
			actors:               actors,
			localGeneralizations: localGeneralizations,
			allGeneralizations:   allGeneralizations,
			errstr:               "references non-existent actor",
		},
		{
			testName: "valid class with SuperclassOfKey reference",
			class: Class{
				Key:             classKey,
				Name:            "Name",
				SuperclassOfKey: &genKey,
			},
			actors:               actors,
			localGeneralizations: localGeneralizations,
			allGeneralizations:   allGeneralizations,
		},
		{
			testName: "error SuperclassOfKey references non-existent generalization",
			class: Class{
				Key:             classKey,
				Name:            "Name",
				SuperclassOfKey: &nonExistentGenKey,
			},
			actors:               actors,
			localGeneralizations: localGeneralizations,
			allGeneralizations:   allGeneralizations,
			errstr:               "references non-existent generalization",
		},
		{
			testName: "error SuperclassOfKey references generalization in different subdomain",
			class: Class{
				Key:             classKey,
				Name:            "Name",
				SuperclassOfKey: &genInOtherSubdomain,
			},
			actors:               actors,
			localGeneralizations: allGeneralizations,
			allGeneralizations:   allGeneralizations,
			errstr:               "must be in the same subdomain",
		},
		{
			testName: "valid class with SubclassOfKey in same subdomain",
			class: Class{
				Key:           classKey,
				Name:          "Name",
				SubclassOfKey: &genKey,
			},
			actors:               actors,
			localGeneralizations: localGeneralizations,
			allGeneralizations:   allGeneralizations,
		},
		{
			testName: "valid class with SubclassOfKey in different subdomain",
			class: Class{
				Key:           classKey,
				Name:          "Name",
				SubclassOfKey: &genInOtherSubdomain,
			},
			actors:               actors,
			localGeneralizations: localGeneralizations,
			allGeneralizations:   allGeneralizations,
		},
		{
			testName: "error SubclassOfKey references non-existent generalization",
			class: Class{
				Key:           classKey,
				Name:          "Name",
				SubclassOfKey: &nonExistentGenKey,
			},
			actors:               actors,
			localGeneralizations: localGeneralizations,
			allGeneralizations:   allGeneralizations,
			errstr:               "references non-existent generalization",
		},
	}
	for _, tt := range tests {
		suite.Run(tt.testName, func() {
			ctx := coreerr.NewContext("test", "")
			err := tt.class.ValidateReferences(ctx, tt.actors, tt.localGeneralizations, tt.allGeneralizations)
			if tt.errstr == "" {
				suite.Require().NoError(err)
			} else {
				suite.Require().ErrorContains(err, tt.errstr)
			}
		})
	}
}

func (suite *ClassSuite) TestValidateTransitionSystemEvents() {
	ctx := coreerr.NewContext("test", "")
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "class1"))
	stateActiveKey := helper.Must(identity.NewStateKey(classKey, "active"))
	eventNewKey := helper.Must(identity.NewEventKey(classKey, "_new"))
	eventAddKey := helper.Must(identity.NewEventKey(classKey, "add"))
	eventDeleteKey := helper.Must(identity.NewEventKey(classKey, "_destroy"))
	eventOtherDeleteKey := helper.Must(identity.NewEventKey(classKey, "delete"))
	transCreateKey := helper.Must(identity.NewTransitionKey(classKey, "", "_new", "", "", "active"))
	transCreateBadKey := helper.Must(identity.NewTransitionKey(classKey, "", "add", "", "", "active"))
	transFinalKey := helper.Must(identity.NewTransitionKey(classKey, "active", "_destroy", "", "", ""))
	transFinalBadKey := helper.Must(identity.NewTransitionKey(classKey, "active", "delete", "", "", ""))

	stateActive := model_state.NewState(stateActiveKey, "Active", "", "")
	eventNew := model_state.NewEvent(eventNewKey, model_state.EventNameNew, "", nil)
	eventAdd := model_state.NewEvent(eventAddKey, "Add", "", nil)
	eventDelete := model_state.NewEvent(eventDeleteKey, model_state.EventNameDestroy, "", nil)
	eventOtherDelete := model_state.NewEvent(eventOtherDeleteKey, "Delete", "", nil)

	tests := []struct {
		testName string
		class    Class
		errstr   string
	}{
		{
			testName: "valid initial transition with «new»",
			class: Class{
				Key:  classKey,
				Name: "Name",
				States: map[identity.Key]model_state.State{
					stateActiveKey: stateActive,
				},
				Events: map[identity.Key]model_state.Event{
					eventNewKey: eventNew,
				},
				Transitions: map[identity.Key]model_state.Transition{
					transCreateKey: model_state.NewTransition(
						transCreateKey,
						eventNewKey,
						model_state.TransitionStateKeys{FromStateKey: nil, ToStateKey: &stateActiveKey},
						model_state.TransitionLogicKeys{},
						"",
					),
				},
			},
		},
		{
			testName: "error initial transition without «new»",
			class: Class{
				Key:  classKey,
				Name: "Name",
				States: map[identity.Key]model_state.State{
					stateActiveKey: stateActive,
				},
				Events: map[identity.Key]model_state.Event{
					eventAddKey: eventAdd,
				},
				Transitions: map[identity.Key]model_state.Transition{
					transCreateBadKey: model_state.NewTransition(
						transCreateBadKey,
						eventAddKey,
						model_state.TransitionStateKeys{FromStateKey: nil, ToStateKey: &stateActiveKey},
						model_state.TransitionLogicKeys{},
						"",
					),
				},
			},
			errstr: "TRANSITION_INITIAL_EVENT_INVALID",
		},
		{
			testName: "valid final transition with _destroy",
			class: Class{
				Key:  classKey,
				Name: "Name",
				States: map[identity.Key]model_state.State{
					stateActiveKey: stateActive,
				},
				Events: map[identity.Key]model_state.Event{
					eventDeleteKey: eventDelete,
				},
				Transitions: map[identity.Key]model_state.Transition{
					transFinalKey: model_state.NewTransition(
						transFinalKey,
						eventDeleteKey,
						model_state.TransitionStateKeys{FromStateKey: &stateActiveKey, ToStateKey: nil},
						model_state.TransitionLogicKeys{},
						"",
					),
				},
			},
		},
		{
			testName: "error final transition without _destroy",
			class: Class{
				Key:  classKey,
				Name: "Name",
				States: map[identity.Key]model_state.State{
					stateActiveKey: stateActive,
				},
				Events: map[identity.Key]model_state.Event{
					eventOtherDeleteKey: eventOtherDelete,
				},
				Transitions: map[identity.Key]model_state.Transition{
					transFinalBadKey: model_state.NewTransition(
						transFinalBadKey,
						eventOtherDeleteKey,
						model_state.TransitionStateKeys{FromStateKey: &stateActiveKey, ToStateKey: nil},
						model_state.TransitionLogicKeys{},
						"",
					),
				},
			},
			errstr: "TRANSITION_FINAL_EVENT_INVALID",
		},
	}
	for _, tt := range tests {
		suite.Run(tt.testName, func() {
			err := tt.class.ValidateWithParent(ctx, &subdomainKey, nil, nil)
			if tt.errstr == "" {
				suite.Require().NoError(err)
			} else {
				suite.Require().ErrorContains(err, tt.errstr)
			}
		})
	}
}
