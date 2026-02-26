package engine

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
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

	eventCreate := helper.Must(model_state.NewEvent(eventCreateKey, "create", "", nil))

	attrAmount := helper.Must(model_class.NewAttribute(attrAmountKey, "amount", "", "", nil, false, "", nil))
	stateOpen := helper.Must(model_state.NewState(stateOpenKey, "Open", "", ""))
	transCreate := helper.Must(model_state.NewTransition(transCreateKey, nil, eventCreateKey, nil, nil, &stateOpenKey, ""))

	class := helper.Must(model_class.NewClass(classKey, "Order", "", nil, nil, nil, ""))
	class.SetAttributes(map[identity.Key]model_class.Attribute{
		attrAmountKey: attrAmount,
	})
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

	eventCreate := helper.Must(model_state.NewEvent(eventCreateKey, "create", "", nil))

	attrName := helper.Must(model_class.NewAttribute(attrNameKey, "name", "", "", nil, false, "", nil))
	stateActive := helper.Must(model_state.NewState(stateActiveKey, "Active", "", ""))
	transCreate := helper.Must(model_state.NewTransition(transCreateKey, nil, eventCreateKey, nil, nil, &stateActiveKey, ""))

	class := helper.Must(model_class.NewClass(classKey, "Item", "", nil, nil, nil, ""))
	class.SetAttributes(map[identity.Key]model_class.Attribute{
		attrNameKey: attrName,
	})
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
	s.Len(classViolations, 0)
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
	s.Len(classViolations, 0)
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
	s.Len(attrViolations, 0)
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

	eventCreate := helper.Must(model_state.NewEvent(eventCreateKey, "create", "", nil))
	derivationLogic := helper.Must(model_logic.NewLogic(mustKey("invariant/20"), model_logic.LogicTypeValue, "Sum of items.", "", model_logic.NotationTLAPlus, "self.amount * 2"))

	attrDerived := helper.Must(model_class.NewAttribute(attrDerivedKey, "total", "", "", &derivationLogic, false, "", nil))
	stateOpen := helper.Must(model_state.NewState(stateOpenKey, "Open", "", ""))
	transCreate := helper.Must(model_state.NewTransition(transCreateKey, nil, eventCreateKey, nil, nil, &stateOpenKey, ""))

	class := helper.Must(model_class.NewClass(classKey, "Order", "", nil, nil, nil, ""))
	class.SetAttributes(map[identity.Key]model_class.Attribute{
		attrDerivedKey: attrDerived,
	})
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
	s.Len(attrViolations, 0)
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
	s.Len(attrViolations, 0)
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
	s.Len(attrViolations, 0)
}

func (s *LivenessCheckerSuite) TestAllAssociationsLinked_NoViolations() {
	orderClass, orderKey := livenessOrderClass()
	itemClass, itemKey := livenessItemClass()

	assocKey := testAssocKey(orderKey, itemKey, "order_items")
	fromMult := helper.Must(model_class.NewMultiplicity("1"))
	toMult := helper.Must(model_class.NewMultiplicity("any"))
	assoc := helper.Must(model_class.NewAssociation(assocKey, "order_items", "", orderKey, fromMult, itemKey, toMult, nil, ""))

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
	s.Len(assocViolations, 0)
}

func (s *LivenessCheckerSuite) TestAssociationNotLinked_Violation() {
	orderClass, orderKey := livenessOrderClass()
	itemClass, itemKey := livenessItemClass()

	assocKey := testAssocKey(orderKey, itemKey, "order_items")
	fromMult := helper.Must(model_class.NewMultiplicity("any"))
	toMult := helper.Must(model_class.NewMultiplicity("any"))
	assoc := helper.Must(model_class.NewAssociation(assocKey, "order_items", "", orderKey, fromMult, itemKey, toMult, nil, ""))

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

func (s *LivenessCheckerSuite) TestNoSimulatableClasses_NoViolations() {
	// Empty model with no simulatable classes — the catalog would be empty.
	// But NewSimulationEngine would fail before reaching liveness.
	// Test the checker directly with an empty catalog.
	statelessKey := mustKey("domain/d/subdomain/s/class/stateless")

	statelessClass := helper.Must(model_class.NewClass(statelessKey, "Stateless", "", nil, nil, nil, ""))
	statelessClass.SetAttributes(map[identity.Key]model_class.Attribute{})
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
	s.Len(violations, 0)
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
	// 2 classes not instantiated + 2 attributes not written (amount + name).
	classViolations := violations.ByType(invariants.ViolationTypeLivenessClassNotInstantiated)
	s.Len(classViolations, 2)

	attrViolations := violations.ByType(invariants.ViolationTypeLivenessAttributeNotWritten)
	s.Len(attrViolations, 2)

	// LivenessViolations filter should return all of them.
	liveness := violations.LivenessViolations()
	s.Len(liveness, 4)
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
