package engine

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/actions"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/invariants"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
	"github.com/stretchr/testify/suite"
)

func TestLivenessCheckerSuite(t *testing.T) {
	suite.Run(t, new(LivenessCheckerSuite))
}

type LivenessCheckerSuite struct {
	suite.Suite
}

// --- helpers ---

// livenessOrderClass creates an Order class with an "amount" attribute.
func livenessOrderClass() (model_class.Class, identity.Key) {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	stateOpenKey := mustKey("domain/d/subdomain/s/class/order/state/open")
	eventCreateKey := mustKey("domain/d/subdomain/s/class/order/event/create")
	transCreateKey := mustKey("domain/d/subdomain/s/class/order/transition/create")
	attrAmountKey := mustKey("domain/d/subdomain/s/class/order/attribute/amount")

	eventCreate := model_state.NewEvent(eventCreateKey, "create", "", nil)

	attrAmount := helper.Must(model_class.NewAttribute(attrAmountKey, model_class.AttributeDetails{Name: "amount", Details: ""}, "", nil, false, model_class.AttributeAnnotations{}))
	stateOpen := model_state.NewState(stateOpenKey, "Open", "", "")
	transCreate := model_state.NewTransition(transCreateKey, eventCreateKey, model_state.TransitionStateKeys{FromStateKey: nil, ToStateKey: &stateOpenKey}, model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: nil}, "")

	class := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Order", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetAttributes([]model_class.Attribute{attrAmount})
	class.SetStates(map[identity.Key]model_state.State{
		stateOpenKey: stateOpen,
	})
	class.SetEvents(map[identity.Key]model_state.Event{
		eventCreateKey: eventCreate,
	})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{
		transCreateKey: transCreate,
	})
	return class, classKey
}

// livenessItemClass creates an Item class with a "name" attribute.
func livenessItemClass() (model_class.Class, identity.Key) {
	classKey := mustKey("domain/d/subdomain/s/class/item")
	stateActiveKey := mustKey("domain/d/subdomain/s/class/item/state/active")
	eventCreateKey := mustKey("domain/d/subdomain/s/class/item/event/create")
	transCreateKey := mustKey("domain/d/subdomain/s/class/item/transition/create")
	attrNameKey := mustKey("domain/d/subdomain/s/class/item/attribute/name")

	eventCreate := model_state.NewEvent(eventCreateKey, "create", "", nil)

	attrName := helper.Must(model_class.NewAttribute(attrNameKey, model_class.AttributeDetails{Name: "name", Details: ""}, "", nil, false, model_class.AttributeAnnotations{}))
	stateActive := model_state.NewState(stateActiveKey, "Active", "", "")
	transCreate := model_state.NewTransition(transCreateKey, eventCreateKey, model_state.TransitionStateKeys{FromStateKey: nil, ToStateKey: &stateActiveKey}, model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: nil}, "")

	class := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Item", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetAttributes([]model_class.Attribute{attrName})
	class.SetStates(map[identity.Key]model_state.State{
		stateActiveKey: stateActive,
	})
	class.SetEvents(map[identity.Key]model_state.Event{
		eventCreateKey: eventCreate,
	})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{
		transCreateKey: transCreate,
	})
	return class, classKey
}

// makeCreationStep creates a SimulationStep representing a creation event.
func makeCreationStep(classKey identity.Key, className string, instanceID state.InstanceID) *SimulationStep {
	return &SimulationStep{
		Kind:       StepKindCreation,
		ClassKey:   classKey,
		ClassName:  className,
		InstanceID: instanceID,
	}
}

// makeStepWithWrite creates a step with a primed assignment (attribute write).
func makeStepWithWrite(classKey identity.Key, className string, instanceID state.InstanceID, attrName string, value object.Object) *SimulationStep {
	return &SimulationStep{
		Kind:      StepKindCreation,
		ClassKey:  classKey,
		ClassName: className,
		TransitionResult: &actions.TransitionResult{
			InstanceID: instanceID,
			ActionResult: &actions.ActionResult{
				InstanceID: instanceID,
				PrimedAssignments: map[state.InstanceID]map[string]object.Object{
					instanceID: {attrName: value},
				},
			},
		},
	}
}

// makeDoStepWithWrite creates a step with a DoActionResult primed assignment.
func makeDoStepWithWrite(classKey identity.Key, className string, instanceID state.InstanceID, attrName string, value object.Object) *SimulationStep {
	return &SimulationStep{
		Kind:      StepKindNormal,
		ClassKey:  classKey,
		ClassName: className,
		DoActionResult: &actions.ActionResult{
			InstanceID: instanceID,
			PrimedAssignments: map[state.InstanceID]map[string]object.Object{
				instanceID: {attrName: value},
			},
		},
	}
}

// makeFinalState creates a SimulationState for test results.
func makeFinalState() *state.SimulationState {
	return state.NewSimulationState()
}

// --- Tests ---

func (s *LivenessCheckerSuite) TestAllClassesInstantiated_NoViolations() {
	orderClass, orderKey := livenessOrderClass()
	model := testModel(classEntry(orderClass, orderKey))
	catalog := NewClassCatalog(model)
	checker := NewLivenessChecker(catalog)

	result := &SimulationResult{
		Steps:      []*SimulationStep{makeCreationStep(orderKey, "Order", 1)},
		FinalState: makeFinalState(),
	}

	violations := checker.Check(result)
	classViolations := violations.ByType(invariants.ViolationTypeLivenessClassNotInstantiated)
	s.Empty(classViolations)
}

func (s *LivenessCheckerSuite) TestClassNotInstantiated_Violation() {
	orderClass, orderKey := livenessOrderClass()
	itemClass, itemKey := livenessItemClass()
	model := testModel(classEntry(orderClass, orderKey), classEntry(itemClass, itemKey))
	catalog := NewClassCatalog(model)
	checker := NewLivenessChecker(catalog)

	// Only Order is created, not Item.
	result := &SimulationResult{
		Steps:      []*SimulationStep{makeCreationStep(orderKey, "Order", 1)},
		FinalState: makeFinalState(),
	}

	violations := checker.Check(result)
	classViolations := violations.ByType(invariants.ViolationTypeLivenessClassNotInstantiated)
	s.Len(classViolations, 1)
	s.Contains(classViolations[0].Message, "Item")
}

func (s *LivenessCheckerSuite) TestCascadedCreationStepsCounted() {
	orderClass, orderKey := livenessOrderClass()
	itemClass, itemKey := livenessItemClass()
	model := testModel(classEntry(orderClass, orderKey), classEntry(itemClass, itemKey))
	catalog := NewClassCatalog(model)
	checker := NewLivenessChecker(catalog)

	// Item is created as a cascaded step from Order's creation.
	result := &SimulationResult{
		Steps: []*SimulationStep{
			{
				Kind:      StepKindCreation,
				ClassKey:  orderKey,
				ClassName: "Order",
				CascadedSteps: []*SimulationStep{
					makeCreationStep(itemKey, "Item", 2),
				},
			},
		},
		FinalState: makeFinalState(),
	}

	violations := checker.Check(result)
	classViolations := violations.ByType(invariants.ViolationTypeLivenessClassNotInstantiated)
	s.Empty(classViolations)
}

// livenessJurisdictionClass creates a Jurisdiction-like class whose attribute
// display names differ from the subKeys recorded in primed assignments.
func livenessJurisdictionClass() (model_class.Class, identity.Key) {
	classKey := mustKey("domain/d/subdomain/s/class/jurisdiction")
	stateActiveKey := mustKey("domain/d/subdomain/s/class/jurisdiction/state/active")
	eventCreateKey := mustKey("domain/d/subdomain/s/class/jurisdiction/event/create")
	transCreateKey := mustKey("domain/d/subdomain/s/class/jurisdiction/transition/create")
	attrCountryCodeKey := mustKey("domain/d/subdomain/s/class/jurisdiction/attribute/country_code")

	eventCreate := model_state.NewEvent(eventCreateKey, "create", "", nil)

	attrCountryCode := helper.Must(model_class.NewAttribute(attrCountryCodeKey, model_class.AttributeDetails{Name: "Country Code", Details: ""}, "", nil, false, model_class.AttributeAnnotations{}))
	stateActive := model_state.NewState(stateActiveKey, "Active", "", "")
	transCreate := model_state.NewTransition(transCreateKey, eventCreateKey, model_state.TransitionStateKeys{FromStateKey: nil, ToStateKey: &stateActiveKey}, model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: nil}, "")

	class := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Jurisdiction", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetAttributes([]model_class.Attribute{attrCountryCode})
	class.SetStates(map[identity.Key]model_state.State{
		stateActiveKey: stateActive,
	})
	class.SetEvents(map[identity.Key]model_state.Event{
		eventCreateKey: eventCreate,
	})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{
		transCreateKey: transCreate,
	})
	return class, classKey
}

func (s *LivenessCheckerSuite) TestAllAttributesWritten_NoViolations() {
	orderClass, orderKey := livenessOrderClass()
	model := testModel(classEntry(orderClass, orderKey))
	catalog := NewClassCatalog(model)
	checker := NewLivenessChecker(catalog)

	result := &SimulationResult{
		Steps: []*SimulationStep{
			makeStepWithWrite(orderKey, "Order", 1, "amount", object.NewInteger(100)),
		},
		FinalState: makeFinalState(),
	}

	violations := checker.Check(result)
	attrViolations := violations.ByType(invariants.ViolationTypeLivenessAttributeNotWritten)
	s.Empty(attrViolations)
}

func (s *LivenessCheckerSuite) TestAttributeWrittenBySubKey_MatchesDisplayName() {
	jurisdictionClass, jurisdictionKey := livenessJurisdictionClass()
	model := testModel(classEntry(jurisdictionClass, jurisdictionKey))
	catalog := NewClassCatalog(model)
	checker := NewLivenessChecker(catalog)

	// Primed assignments use the attribute subKey, not the display name.
	result := &SimulationResult{
		Steps: []*SimulationStep{
			makeStepWithWrite(jurisdictionKey, "Jurisdiction", 1, "country_code", object.NewString("US")),
		},
		FinalState: makeFinalState(),
	}

	violations := checker.Check(result)
	attrViolations := violations.ByType(invariants.ViolationTypeLivenessAttributeNotWritten)
	s.Empty(attrViolations)
}

func (s *LivenessCheckerSuite) TestAttributeNotWritten_Violation() {
	orderClass, orderKey := livenessOrderClass()
	model := testModel(classEntry(orderClass, orderKey))
	catalog := NewClassCatalog(model)
	checker := NewLivenessChecker(catalog)

	// No steps at all — amount was never written.
	result := &SimulationResult{
		Steps:      []*SimulationStep{},
		FinalState: makeFinalState(),
	}

	violations := checker.Check(result)
	attrViolations := violations.ByType(invariants.ViolationTypeLivenessAttributeNotWritten)
	s.Len(attrViolations, 1)
	s.Contains(attrViolations[0].Message, "amount")
	s.Contains(attrViolations[0].Message, "Order")
}

func (s *LivenessCheckerSuite) TestDerivedAttributesExcluded() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	stateOpenKey := mustKey("domain/d/subdomain/s/class/order/state/open")
	eventCreateKey := mustKey("domain/d/subdomain/s/class/order/event/create")
	transCreateKey := mustKey("domain/d/subdomain/s/class/order/transition/create")
	attrDerivedKey := mustKey("domain/d/subdomain/s/class/order/attribute/total")

	eventCreate := model_state.NewEvent(eventCreateKey, "create", "", nil)
	derivationLogic := model_logic.NewLogic(mustKey("invariant/20"), model_logic.LogicTypeValue, "Sum of items.", "", orderSpec("self.amount * 2"), nil)

	attrDerived := helper.Must(model_class.NewAttribute(attrDerivedKey, model_class.AttributeDetails{Name: "total", Details: ""}, "", &derivationLogic, false, model_class.AttributeAnnotations{}))
	stateOpen := model_state.NewState(stateOpenKey, "Open", "", "")
	transCreate := model_state.NewTransition(transCreateKey, eventCreateKey, model_state.TransitionStateKeys{FromStateKey: nil, ToStateKey: &stateOpenKey}, model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: nil}, "")

	class := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Order", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetAttributes([]model_class.Attribute{attrDerived})
	class.SetStates(map[identity.Key]model_state.State{
		stateOpenKey: stateOpen,
	})
	class.SetEvents(map[identity.Key]model_state.Event{
		eventCreateKey: eventCreate,
	})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{
		transCreateKey: transCreate,
	})

	model := testModel(classEntry(class, classKey))
	catalog := NewClassCatalog(model)
	checker := NewLivenessChecker(catalog)

	// No writes — but the only attribute is derived, so no violation.
	result := &SimulationResult{
		Steps:      []*SimulationStep{},
		FinalState: makeFinalState(),
	}

	violations := checker.Check(result)
	attrViolations := violations.ByType(invariants.ViolationTypeLivenessAttributeNotWritten)
	s.Empty(attrViolations)
}

func (s *LivenessCheckerSuite) TestDoActionWritesCounted() {
	orderClass, orderKey := livenessOrderClass()
	model := testModel(classEntry(orderClass, orderKey))
	catalog := NewClassCatalog(model)
	checker := NewLivenessChecker(catalog)

	// Write happens via a "do" action.
	result := &SimulationResult{
		Steps: []*SimulationStep{
			makeDoStepWithWrite(orderKey, "Order", 1, "amount", object.NewInteger(42)),
		},
		FinalState: makeFinalState(),
	}

	violations := checker.Check(result)
	attrViolations := violations.ByType(invariants.ViolationTypeLivenessAttributeNotWritten)
	s.Empty(attrViolations)
}

func (s *LivenessCheckerSuite) TestCascadedStepWritesCounted() {
	orderClass, orderKey := livenessOrderClass()
	model := testModel(classEntry(orderClass, orderKey))
	catalog := NewClassCatalog(model)
	checker := NewLivenessChecker(catalog)

	// Write happens in a cascaded step.
	result := &SimulationResult{
		Steps: []*SimulationStep{
			{
				Kind:     StepKindCreation,
				ClassKey: orderKey,
				CascadedSteps: []*SimulationStep{
					makeStepWithWrite(orderKey, "Order", 1, "amount", object.NewInteger(10)),
				},
			},
		},
		FinalState: makeFinalState(),
	}

	violations := checker.Check(result)
	attrViolations := violations.ByType(invariants.ViolationTypeLivenessAttributeNotWritten)
	s.Empty(attrViolations)
}

func (s *LivenessCheckerSuite) TestAllAssociationsLinked_NoViolations() {
	orderClass, orderKey := livenessOrderClass()
	itemClass, itemKey := livenessItemClass()

	assocKey := testAssocKey(orderKey, itemKey, "order_items")
	fromMult := helper.Must(model_class.NewMultiplicity("1"))
	toMult := helper.Must(model_class.NewMultiplicity("any"))
	assoc := model_class.NewAssociation(assocKey, model_class.AssociationDetails{Name: "order_items", Details: ""}, model_class.AssociationEnd{ClassKey: orderKey, Multiplicity: fromMult}, model_class.AssociationEnd{ClassKey: itemKey, Multiplicity: toMult}, model_class.AssociationOptions{AssociationClassKey: nil, UmlComment: ""})

	model := testModel(classEntry(orderClass, orderKey), classEntry(itemClass, itemKey))
	model.ClassAssociations = map[identity.Key]model_class.Association{
		assocKey: assoc,
	}
	catalog := NewClassCatalog(model)
	checker := NewLivenessChecker(catalog)

	// Create a state with a link.
	finalState := makeFinalState()
	finalState.Links().AddLink(
		evaluator.AssociationKey(assocKey.String()),
		evaluator.ObjectID(1),
		evaluator.ObjectID(2),
	)

	result := &SimulationResult{
		Steps:      []*SimulationStep{},
		FinalState: finalState,
	}

	violations := checker.Check(result)
	assocViolations := violations.ByType(invariants.ViolationTypeLivenessAssociationNotLinked)
	s.Empty(assocViolations)
}

func (s *LivenessCheckerSuite) TestAssociationNotLinked_Violation() {
	orderClass, orderKey := livenessOrderClass()
	itemClass, itemKey := livenessItemClass()

	assocKey := testAssocKey(orderKey, itemKey, "order_items")
	fromMult := helper.Must(model_class.NewMultiplicity("any"))
	toMult := helper.Must(model_class.NewMultiplicity("any"))
	assoc := model_class.NewAssociation(assocKey, model_class.AssociationDetails{Name: "order_items", Details: ""}, model_class.AssociationEnd{ClassKey: orderKey, Multiplicity: fromMult}, model_class.AssociationEnd{ClassKey: itemKey, Multiplicity: toMult}, model_class.AssociationOptions{AssociationClassKey: nil, UmlComment: ""})

	model := testModel(classEntry(orderClass, orderKey), classEntry(itemClass, itemKey))
	model.ClassAssociations = map[identity.Key]model_class.Association{
		assocKey: assoc,
	}
	catalog := NewClassCatalog(model)
	checker := NewLivenessChecker(catalog)

	// No links in final state.
	result := &SimulationResult{
		Steps:      []*SimulationStep{},
		FinalState: makeFinalState(),
	}

	violations := checker.Check(result)
	assocViolations := violations.ByType(invariants.ViolationTypeLivenessAssociationNotLinked)
	s.Len(assocViolations, 1)
	s.Contains(assocViolations[0].Message, "order_items")
}

func (s *LivenessCheckerSuite) TestStatelessClass_InstantiationViolation() {
	statelessKey := mustKey("domain/d/subdomain/s/class/stateless")

	statelessClass := model_class.NewClass(statelessKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Stateless", Details: "", UnfinishedNotes: "", UmlComment: ""})
	statelessClass.SetAttributes(nil)
	statelessClass.SetStates(map[identity.Key]model_state.State{})
	statelessClass.SetEvents(map[identity.Key]model_state.Event{})
	statelessClass.SetGuards(map[identity.Key]model_state.Guard{})
	statelessClass.SetActions(map[identity.Key]model_state.Action{})
	statelessClass.SetQueries(map[identity.Key]model_state.Query{})
	statelessClass.SetTransitions(map[identity.Key]model_state.Transition{})

	model := testModel(classEntry(statelessClass, statelessKey))
	catalog := NewClassCatalog(model)
	checker := NewLivenessChecker(catalog)

	result := &SimulationResult{
		Steps:      []*SimulationStep{},
		FinalState: makeFinalState(),
	}

	violations := checker.Check(result)
	classViolations := violations.ByType(invariants.ViolationTypeLivenessClassNotInstantiated)
	s.Len(classViolations, 1)
	s.Contains(classViolations[0].Message, "Stateless")
}

func (s *LivenessCheckerSuite) TestMultipleViolationsCombined() {
	orderClass, orderKey := livenessOrderClass()
	itemClass, itemKey := livenessItemClass()
	model := testModel(classEntry(orderClass, orderKey), classEntry(itemClass, itemKey))
	catalog := NewClassCatalog(model)
	checker := NewLivenessChecker(catalog)

	// No steps at all — both classes not instantiated, attributes not written.
	result := &SimulationResult{
		Steps:      []*SimulationStep{},
		FinalState: makeFinalState(),
	}

	violations := checker.Check(result)
	classViolations := violations.ByType(invariants.ViolationTypeLivenessClassNotInstantiated)
	s.Len(classViolations, 2)

	attrViolations := violations.ByType(invariants.ViolationTypeLivenessAttributeNotWritten)
	s.Len(attrViolations, 2)

	eventViolations := violations.ByType(invariants.ViolationTypeLivenessEventNotSent)
	s.NotEmpty(eventViolations)

	s.NotEmpty(violations.LivenessViolations())
}

func (s *LivenessCheckerSuite) TestEventNotSent_Violation() {
	orderClass, orderKey := livenessOrderClass()
	model := testModel(classEntry(orderClass, orderKey))
	catalog := NewClassCatalog(model)
	checker := NewLivenessChecker(catalog)

	result := &SimulationResult{
		Steps:      []*SimulationStep{},
		FinalState: makeFinalState(),
	}

	violations := checker.Check(result)
	eventViolations := violations.ByType(invariants.ViolationTypeLivenessEventNotSent)
	s.NotEmpty(eventViolations)
	s.Contains(eventViolations[0].Message, "create")
}

func (s *LivenessCheckerSuite) TestNilFinalState_NoAssociationPanic() {
	orderClass, orderKey := livenessOrderClass()
	model := testModel(classEntry(orderClass, orderKey))
	catalog := NewClassCatalog(model)
	checker := NewLivenessChecker(catalog)

	result := &SimulationResult{
		Steps:      []*SimulationStep{},
		FinalState: nil,
	}

	// Should not panic on nil FinalState.
	violations := checker.Check(result)
	s.NotNil(violations) // Will have class/attr violations, but no panic.
}
