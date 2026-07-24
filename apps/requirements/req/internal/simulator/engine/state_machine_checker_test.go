package engine

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/schema"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/invariants"
	"github.com/stretchr/testify/suite"
)

func TestStateMachineCheckerSuite(t *testing.T) {
	suite.Run(t, new(StateMachineCheckerSuite))
}

type StateMachineCheckerSuite struct {
	suite.Suite
}

func (s *StateMachineCheckerSuite) TestNoStateMachine_NoViolation() {
	classKey := mustKey("domain/d/subdomain/s/class/empty")
	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Empty", Details: ""})

	catalog := NewClassCatalog(schema.New(testModel(classEntry(class, classKey))))
	checker := NewStateMachineChecker(catalog)

	violations := checker.Check()
	s.Empty(violations.ByType(invariants.ViolationTypeStateMachineIncomplete))
}

func (s *StateMachineCheckerSuite) TestStateMachineWithoutNewEvent_Violation() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	stateOpenKey := mustKey("domain/d/subdomain/s/class/order/state/open")
	eventCreateKey := mustKey("domain/d/subdomain/s/class/order/event/create")

	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Order", Details: ""})
	class.SetStates(map[identity.Key]model_state.State{
		stateOpenKey: model_state.NewState(stateOpenKey, "Open", "", ""),
	})
	class.SetEvents(map[identity.Key]model_state.Event{
		eventCreateKey: model_state.NewEvent(eventCreateKey, "create", "", nil),
	})

	catalog := NewClassCatalog(schema.New(testModel(classEntry(class, classKey))))
	checker := NewStateMachineChecker(catalog)

	violations := checker.Check()
	incomplete := violations.ByType(invariants.ViolationTypeStateMachineIncomplete)
	s.Len(incomplete, 1)
	s.Equal(classKey, incomplete[0].ClassKey)
	s.Contains(incomplete[0].Message, "Order")
	s.Contains(incomplete[0].Message, "«new»")
}

func (s *StateMachineCheckerSuite) TestStateMachineWithNewEvent_NoViolation() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	stateOpenKey := mustKey("domain/d/subdomain/s/class/order/state/open")
	eventNewKey := mustKey("domain/d/subdomain/s/class/order/event/_new")

	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Order", Details: ""})
	class.SetStates(map[identity.Key]model_state.State{
		stateOpenKey: model_state.NewState(stateOpenKey, "Open", "", ""),
	})
	class.SetEvents(map[identity.Key]model_state.Event{
		eventNewKey: model_state.NewEvent(eventNewKey, model_state.EventNameNew, "", nil),
	})

	catalog := NewClassCatalog(schema.New(testModel(classEntry(class, classKey))))
	checker := NewStateMachineChecker(catalog)

	violations := checker.Check()
	s.Empty(violations.ByType(invariants.ViolationTypeStateMachineIncomplete))
}

func (s *StateMachineCheckerSuite) TestTransitionsOnlyWithoutNewEvent_Violation() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	stateOpenKey := mustKey("domain/d/subdomain/s/class/order/state/open")
	eventCreateKey := mustKey("domain/d/subdomain/s/class/order/event/create")
	transCreateKey := mustKey("domain/d/subdomain/s/class/order/transition/create")

	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Order", Details: ""})
	class.SetEvents(map[identity.Key]model_state.Event{
		eventCreateKey: model_state.NewEvent(eventCreateKey, "create", "", nil),
	})
	class.SetTransitions(map[identity.Key]model_state.Transition{
		transCreateKey: model_state.NewTransition(
			transCreateKey,
			eventCreateKey,
			model_state.TransitionStateKeys{FromStateKey: nil, ToStateKey: &stateOpenKey},
			model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: nil},
			"",
		),
	})

	catalog := NewClassCatalog(schema.New(testModel(classEntry(class, classKey))))
	checker := NewStateMachineChecker(catalog)

	violations := checker.Check()
	incomplete := violations.ByType(invariants.ViolationTypeStateMachineIncomplete)
	s.Len(incomplete, 1)
}

func (s *StateMachineCheckerSuite) TestMultipleClasses_OnlyIncompleteReported() {
	orderClass, orderKey := simpleOrderClass()

	itemKey := mustKey("domain/d/subdomain/s/class/item")
	stateActiveKey := mustKey("domain/d/subdomain/s/class/item/state/active")
	eventNewKey := mustKey("domain/d/subdomain/s/class/item/event/_new")
	itemClass := model_class.NewClass(itemKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Item", Details: ""})
	itemClass.SetStates(map[identity.Key]model_state.State{
		stateActiveKey: model_state.NewState(stateActiveKey, "Active", "", ""),
	})
	itemClass.SetEvents(map[identity.Key]model_state.Event{
		eventNewKey: model_state.NewEvent(eventNewKey, model_state.EventNameNew, "", nil),
	})

	model := testModel(classEntry(orderClass, orderKey), classEntry(itemClass, itemKey))
	catalog := NewClassCatalog(schema.New(model))
	checker := NewStateMachineChecker(catalog)

	violations := checker.Check()
	incomplete := violations.ByType(invariants.ViolationTypeStateMachineIncomplete)
	s.Len(incomplete, 1)
	s.Contains(incomplete[0].Message, "Order")
}

func (s *StateMachineCheckerSuite) TestEngineRunReportsIncompleteStateMachine() {
	orderClass, orderKey := simpleOrderClass()
	model := testModel(classEntry(orderClass, orderKey))

	engine, err := NewSimulationEngine(model, SimulationConfig{MaxSteps: 1, RandomSeed: 1})
	s.Require().NoError(err)

	result, err := engine.Run()
	s.Require().NoError(err)

	incomplete := result.Violations.ByType(invariants.ViolationTypeStateMachineIncomplete)
	s.Len(incomplete, 1)
	s.Contains(incomplete[0].Message, "Order")
}
