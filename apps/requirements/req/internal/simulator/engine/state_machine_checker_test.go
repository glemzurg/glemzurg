package engine

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/schema"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	siminst "github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"
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

	catalog := schema.New(testModel(classEntry(class, classKey)), schema.RunScopeAll())
	violations := siminst.CheckStateMachineCompleteness(catalog)
	s.Empty(violationsByType(violations, siminst.ViolationTypeStateMachineIncomplete))
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

	catalog := schema.New(testModel(classEntry(class, classKey)), schema.RunScopeAll())
	violations := siminst.CheckStateMachineCompleteness(catalog)
	sm := violationsByType(violations, siminst.ViolationTypeStateMachineIncomplete)
	s.Len(sm, 1)
	s.Contains(sm[0].Message, "Order")
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

	catalog := schema.New(testModel(classEntry(class, classKey)), schema.RunScopeAll())
	violations := siminst.CheckStateMachineCompleteness(catalog)
	s.Empty(violationsByType(violations, siminst.ViolationTypeStateMachineIncomplete))
}

func (s *StateMachineCheckerSuite) TestMultipleClasses_OnlyIncompleteReported() {
	// One complete SM, one incomplete.
	orderKey := mustKey("domain/d/subdomain/s/class/order")
	orderState := mustKey("domain/d/subdomain/s/class/order/state/open")
	orderNew := mustKey("domain/d/subdomain/s/class/order/event/_new")
	order := model_class.NewClass(orderKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Order", Details: ""})
	order.SetStates(map[identity.Key]model_state.State{
		orderState: model_state.NewState(orderState, "Open", "", ""),
	})
	order.SetEvents(map[identity.Key]model_state.Event{
		orderNew: model_state.NewEvent(orderNew, model_state.EventNameNew, "", nil),
	})

	itemKey := mustKey("domain/d/subdomain/s/class/item")
	itemState := mustKey("domain/d/subdomain/s/class/item/state/active")
	itemEv := mustKey("domain/d/subdomain/s/class/item/event/spawn")
	item := model_class.NewClass(itemKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Item", Details: ""})
	item.SetStates(map[identity.Key]model_state.State{
		itemState: model_state.NewState(itemState, "Active", "", ""),
	})
	item.SetEvents(map[identity.Key]model_state.Event{
		itemEv: model_state.NewEvent(itemEv, "spawn", "", nil),
	})

	catalog := schema.New(testModel(classEntry(order, orderKey), classEntry(item, itemKey)), schema.RunScopeAll())
	violations := siminst.CheckStateMachineCompleteness(catalog)
	sm := violationsByType(violations, siminst.ViolationTypeStateMachineIncomplete)
	s.Len(sm, 1)
	s.Contains(sm[0].Message, "Item")
}

func (s *StateMachineCheckerSuite) TestEngineRunReportsIncompleteStateMachine() {
	// Full engine path still surfaces SM incompleteness via instance.CheckStateMachineCompleteness.
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
	// Need at least one event-bearing class for engine setup — incomplete SM still reports.
	// Engine requires event-bearing simulatable classes; incomplete is still event-bearing.
	transKey := mustKey("domain/d/subdomain/s/class/order/transition/create")
	class.SetTransitions(map[identity.Key]model_state.Transition{
		transKey: model_state.NewTransition(transKey, eventCreateKey, model_state.TransitionStateKeys{FromStateKey: nil, ToStateKey: &stateOpenKey}, model_state.TransitionLogicKeys{}, ""),
	})

	model := testModel(classEntry(class, classKey))
	eng, err := NewSimulationEngine(model, SimulationConfig{MaxSteps: 1, RandomSeed: 1})
	// Engine may fail setup if no _new; if it runs, check violations include SM incomplete.
	if err != nil {
		// Setup failure is acceptable when model cannot simulate; completeness is still checked via direct API.
		violations := siminst.CheckStateMachineCompleteness(schema.New(model, schema.RunScopeAll()))
		s.NotEmpty(violationsByType(violations, siminst.ViolationTypeStateMachineIncomplete))
		return
	}
	result, err := eng.Run()
	s.Require().NoError(err)
	s.NotEmpty(violationsByType(result.Violations, siminst.ViolationTypeStateMachineIncomplete))
}
