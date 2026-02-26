package engine

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/stretchr/testify/suite"
)

type EngineSuite struct {
	suite.Suite
}

func TestEngineSuite(t *testing.T) {
	suite.Run(t, new(EngineSuite))
}

// simpleOrderClass creates an Order class for engine tests.
// Unlike testOrderClass, the close transition has no action (avoiding attribute dependency).
func simpleOrderClass() (model_class.Class, identity.Key) {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	stateOpenKey := mustKey("domain/d/subdomain/s/class/order/state/open")
	stateClosedKey := mustKey("domain/d/subdomain/s/class/order/state/closed")
	eventCreateKey := mustKey("domain/d/subdomain/s/class/order/event/create")
	eventCloseKey := mustKey("domain/d/subdomain/s/class/order/event/close")
	transCreateKey := mustKey("domain/d/subdomain/s/class/order/transition/create")
	transCloseKey := mustKey("domain/d/subdomain/s/class/order/transition/close")

	eventCreate := helper.Must(model_state.NewEvent(eventCreateKey, "create", "", nil))
	eventClose := helper.Must(model_state.NewEvent(eventCloseKey, "close", "", nil))

	stateOpen := helper.Must(model_state.NewState(stateOpenKey, "Open", "", ""))
	stateClosed := helper.Must(model_state.NewState(stateClosedKey, "Closed", "", ""))

	transCreate := helper.Must(model_state.NewTransition(transCreateKey, nil, eventCreateKey, nil, nil, &stateOpenKey, ""))
	transClose := helper.Must(model_state.NewTransition(transCloseKey, &stateOpenKey, eventCloseKey, nil, nil, &stateClosedKey, ""))

	class := helper.Must(model_class.NewClass(classKey, "Order", "", nil, nil, nil, ""))
	class.SetAttributes(map[identity.Key]model_class.Attribute{})
	class.SetStates(map[identity.Key]model_state.State{
		stateOpenKey:   stateOpen,
		stateClosedKey: stateClosed,
	})
	class.SetEvents(map[identity.Key]model_state.Event{
		eventCreateKey: eventCreate,
		eventCloseKey:  eventClose,
	})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{
		transCreateKey: transCreate,
		transCloseKey:  transClose,
	})

	return class, classKey
}

func (s *EngineSuite) TestSimulationRunsToMaxSteps() {
	orderClass, orderKey := simpleOrderClass()
	model := testModel(classEntry(orderClass, orderKey))

	config := SimulationConfig{
		MaxSteps:   10,
		RandomSeed: 42,
	}

	engine, err := NewSimulationEngine(model, config)
	s.Require().NoError(err)

	result, err := engine.Run()
	s.Require().NoError(err)

	s.Equal("max_steps", result.TerminationReason)
	s.Equal(10, result.StepsTaken)
	s.Len(result.Steps, 10)
	s.NotNil(result.FinalState)
}

func (s *EngineSuite) TestSimulationCreatesInstances() {
	orderClass, orderKey := simpleOrderClass()
	model := testModel(classEntry(orderClass, orderKey))

	config := SimulationConfig{
		MaxSteps:   20,
		RandomSeed: 42,
	}

	engine, err := NewSimulationEngine(model, config)
	s.Require().NoError(err)

	result, err := engine.Run()
	s.Require().NoError(err)

	// At least some steps should be creation steps.
	foundCreation := false
	for _, step := range result.Steps {
		if step.Kind == StepKindCreation {
			foundCreation = true
			s.NotZero(step.InstanceID)
			s.Equal("Order", step.ClassName)
			break
		}
	}
	s.True(foundCreation, "should have at least one creation step")

	// Final state should have instances.
	allInstances := result.FinalState.AllInstances()
	s.NotEmpty(allInstances)
}

func (s *EngineSuite) TestSimulationTransitions() {
	orderClass, orderKey := simpleOrderClass()
	model := testModel(classEntry(orderClass, orderKey))

	config := SimulationConfig{
		MaxSteps:   100,
		RandomSeed: 42,
	}

	engine, err := NewSimulationEngine(model, config)
	s.Require().NoError(err)

	result, err := engine.Run()
	s.Require().NoError(err)

	// With enough steps, should find both creation and normal transition steps.
	foundCreation := false
	foundNormal := false
	for _, step := range result.Steps {
		if step.Kind == StepKindCreation {
			foundCreation = true
		}
		if step.Kind == StepKindNormal {
			foundNormal = true
		}
		if foundCreation && foundNormal {
			break
		}
	}
	s.True(foundCreation, "should have creation steps")
	s.True(foundNormal, "should have normal transition steps")
}

func (s *EngineSuite) TestDeadlockDetection() {
	// A class with no creation transitions → immediate deadlock.
	classKey := mustKey("domain/d/subdomain/s/class/stuck")
	stateActiveKey := mustKey("domain/d/subdomain/s/class/stuck/state/active")
	eventUpdateKey := mustKey("domain/d/subdomain/s/class/stuck/event/update")
	transUpdateKey := mustKey("domain/d/subdomain/s/class/stuck/transition/update")

	eventUpdate := helper.Must(model_state.NewEvent(eventUpdateKey, "update", "", nil))

	stateActive := helper.Must(model_state.NewState(stateActiveKey, "Active", "", ""))
	transUpdate := helper.Must(model_state.NewTransition(transUpdateKey, &stateActiveKey, eventUpdateKey, nil, nil, &stateActiveKey, ""))

	class := helper.Must(model_class.NewClass(classKey, "Stuck", "", nil, nil, nil, ""))
	class.SetAttributes(map[identity.Key]model_class.Attribute{})
	class.SetStates(map[identity.Key]model_state.State{
		stateActiveKey: stateActive,
	})
	class.SetEvents(map[identity.Key]model_state.Event{
		eventUpdateKey: eventUpdate,
	})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{
		transUpdateKey: transUpdate,
	})

	model := testModel(classEntry(class, classKey))

	config := SimulationConfig{
		MaxSteps:   10,
		RandomSeed: 42,
	}

	engine, err := NewSimulationEngine(model, config)
	s.Require().NoError(err)

	result, err := engine.Run()
	s.Require().NoError(err)
	s.Equal("deadlock", result.TerminationReason)
	s.Equal(0, result.StepsTaken)
}

func (s *EngineSuite) TestStopOnViolation() {
	orderClass, orderKey := simpleOrderClass()

	// Add an invariant that will fail: "FALSE".
	model := testModel(classEntry(orderClass, orderKey))
	invariantKey := helper.Must(identity.NewInvariantKey("0"))
	invariantLogic := helper.Must(model_logic.NewLogic(invariantKey, model_logic.LogicTypeAssessment, "Always false.", "", model_logic.NotationTLAPlus, "FALSE", nil))
	model.Invariants = []model_logic.Logic{invariantLogic}

	config := SimulationConfig{
		MaxSteps:        100,
		RandomSeed:      42,
		StopOnViolation: true,
	}

	engine, err := NewSimulationEngine(model, config)
	s.Require().NoError(err)

	result, err := engine.Run()
	s.Require().NoError(err)
	s.Equal("violation", result.TerminationReason)
	s.True(result.Violations.HasViolations())
	// Should stop early, not run all 100 steps.
	s.Less(result.StepsTaken, 100)
}

func (s *EngineSuite) TestReproducibility() {
	orderClass, orderKey := simpleOrderClass()
	model := testModel(classEntry(orderClass, orderKey))

	config := SimulationConfig{
		MaxSteps:   20,
		RandomSeed: 123,
	}

	// Run twice with same seed.
	engine1, err := NewSimulationEngine(model, config)
	s.Require().NoError(err)
	result1, err := engine1.Run()
	s.Require().NoError(err)

	engine2, err := NewSimulationEngine(model, config)
	s.Require().NoError(err)
	result2, err := engine2.Run()
	s.Require().NoError(err)

	// Same seed → same steps.
	s.Equal(result1.StepsTaken, result2.StepsTaken)
	s.Require().Len(result1.Steps, len(result2.Steps))
	for i := range result1.Steps {
		s.Equal(result1.Steps[i].Kind, result2.Steps[i].Kind)
		s.Equal(result1.Steps[i].ClassName, result2.Steps[i].ClassName)
		s.Equal(result1.Steps[i].EventName, result2.Steps[i].EventName)
		s.Equal(result1.Steps[i].InstanceID, result2.Steps[i].InstanceID)
	}
}

func (s *EngineSuite) TestNoSimulatableClassesReturnsError() {
	// A class with no states → not simulatable.
	classKey := mustKey("domain/d/subdomain/s/class/empty")

	class := helper.Must(model_class.NewClass(classKey, "Empty", "", nil, nil, nil, ""))
	class.SetAttributes(map[identity.Key]model_class.Attribute{})
	class.SetStates(map[identity.Key]model_state.State{})
	class.SetEvents(map[identity.Key]model_state.Event{})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{})

	model := testModel(classEntry(class, classKey))

	config := SimulationConfig{MaxSteps: 10, RandomSeed: 42}
	_, err := NewSimulationEngine(model, config)
	s.Error(err)
	s.Contains(err.Error(), "no simulatable classes")
}

func (s *EngineSuite) TestStepNumbersAreSequential() {
	orderClass, orderKey := simpleOrderClass()
	model := testModel(classEntry(orderClass, orderKey))

	config := SimulationConfig{
		MaxSteps:   15,
		RandomSeed: 42,
	}

	engine, err := NewSimulationEngine(model, config)
	s.Require().NoError(err)

	result, err := engine.Run()
	s.Require().NoError(err)

	for i, step := range result.Steps {
		s.Equal(i+1, step.StepNumber, "step numbers should be 1-based and sequential")
	}
}
