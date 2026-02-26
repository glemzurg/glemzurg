package engine

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/actions"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
	"github.com/stretchr/testify/suite"
)

type StateActionExecutorSuite struct {
	suite.Suite
}

func TestStateActionExecutorSuite(t *testing.T) {
	suite.Run(t, new(StateActionExecutorSuite))
}

// buildStateActionTestExecutor creates an ActionExecutor suitable for state action tests.
func buildStateActionTestExecutor(simState *state.SimulationState) *actions.ActionExecutor {
	bb := state.NewBindingsBuilder(simState)
	ge := actions.NewGuardEvaluator(bb)
	return actions.NewActionExecutor(bb, nil, nil, nil, ge, nil)
}

func (s *StateActionExecutorSuite) TestExitActionsFireOnTransition() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	stateOpenKey := mustKey("domain/d/subdomain/s/class/order/state/open")
	stateClosedKey := mustKey("domain/d/subdomain/s/class/order/state/closed")
	actionExitKey := mustKey("domain/d/subdomain/s/class/order/action/on_exit")
	stateActionKey := mustKey("domain/d/subdomain/s/class/order/state/open/saction/exit/on_exit")

	guaranteeKey := helper.Must(identity.NewActionGuaranteeKey(actionExitKey, "0"))
	guaranteeLogic := helper.Must(model_logic.NewLogic(guaranteeKey, model_logic.LogicTypeStateChange, "Postcondition.", "exit_count", model_logic.NotationTLAPlus, "self.exit_count + 1", nil))
	actionExit := helper.Must(model_state.NewAction(actionExitKey, "OnExit", "", nil, []model_logic.Logic{guaranteeLogic}, nil, nil))

	stateActionExit := helper.Must(model_state.NewStateAction(stateActionKey, actionExitKey, "exit"))

	stateOpen := helper.Must(model_state.NewState(stateOpenKey, "Open", "", ""))
	stateOpen.SetActions([]model_state.StateAction{stateActionExit})
	stateClosed := helper.Must(model_state.NewState(stateClosedKey, "Closed", "", ""))

	class := helper.Must(model_class.NewClass(classKey, "Order", "", nil, nil, nil, ""))
	class.SetAttributes(map[identity.Key]model_class.Attribute{})
	class.SetStates(map[identity.Key]model_state.State{
		stateOpenKey:   stateOpen,
		stateClosedKey: stateClosed,
	})
	class.SetEvents(map[identity.Key]model_state.Event{})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{
		actionExitKey: actionExit,
	})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{})

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("exit_count", object.NewInteger(0))
	attrs.Set("_state", object.NewString("Open"))
	instance := simState.CreateInstance(classKey, attrs)

	ae := buildStateActionTestExecutor(simState)
	sae := NewStateActionExecutor(ae)

	violations, err := sae.ExecuteExitActions(class, stateOpenKey, instance)
	s.NoError(err)
	s.Empty(violations)

	// The exit action should have incremented exit_count.
	updated := simState.GetInstance(instance.ID)
	s.Equal("1", updated.GetAttribute("exit_count").Inspect())
}

func (s *StateActionExecutorSuite) TestEntryActionsFireOnTransition() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	stateOpenKey := mustKey("domain/d/subdomain/s/class/order/state/open")
	actionEntryKey := mustKey("domain/d/subdomain/s/class/order/action/on_entry")
	stateActionKey := mustKey("domain/d/subdomain/s/class/order/state/open/saction/entry/on_entry")

	guaranteeKey := helper.Must(identity.NewActionGuaranteeKey(actionEntryKey, "0"))
	guaranteeLogic := helper.Must(model_logic.NewLogic(guaranteeKey, model_logic.LogicTypeStateChange, "Postcondition.", "entry_count", model_logic.NotationTLAPlus, "self.entry_count + 1", nil))
	actionEntry := helper.Must(model_state.NewAction(actionEntryKey, "OnEntry", "", nil, []model_logic.Logic{guaranteeLogic}, nil, nil))

	stateActionEntry := helper.Must(model_state.NewStateAction(stateActionKey, actionEntryKey, "entry"))

	stateOpen := helper.Must(model_state.NewState(stateOpenKey, "Open", "", ""))
	stateOpen.SetActions([]model_state.StateAction{stateActionEntry})

	class := helper.Must(model_class.NewClass(classKey, "Order", "", nil, nil, nil, ""))
	class.SetAttributes(map[identity.Key]model_class.Attribute{})
	class.SetStates(map[identity.Key]model_state.State{
		stateOpenKey: stateOpen,
	})
	class.SetEvents(map[identity.Key]model_state.Event{})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{
		actionEntryKey: actionEntry,
	})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{})

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("entry_count", object.NewInteger(0))
	attrs.Set("_state", object.NewString("Open"))
	instance := simState.CreateInstance(classKey, attrs)

	ae := buildStateActionTestExecutor(simState)
	sae := NewStateActionExecutor(ae)

	violations, err := sae.ExecuteEntryActions(class, stateOpenKey, instance)
	s.NoError(err)
	s.Empty(violations)

	updated := simState.GetInstance(instance.ID)
	s.Equal("1", updated.GetAttribute("entry_count").Inspect())
}

func (s *StateActionExecutorSuite) TestNoStateActionsReturnsEmpty() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	stateOpenKey := mustKey("domain/d/subdomain/s/class/order/state/open")

	stateOpen := helper.Must(model_state.NewState(stateOpenKey, "Open", "", ""))

	class := helper.Must(model_class.NewClass(classKey, "Order", "", nil, nil, nil, ""))
	class.SetAttributes(map[identity.Key]model_class.Attribute{})
	class.SetStates(map[identity.Key]model_state.State{
		stateOpenKey: stateOpen,
	})
	class.SetEvents(map[identity.Key]model_state.Event{})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{})

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("_state", object.NewString("Open"))
	instance := simState.CreateInstance(classKey, attrs)

	ae := buildStateActionTestExecutor(simState)
	sae := NewStateActionExecutor(ae)

	violations, err := sae.ExecuteExitActions(class, stateOpenKey, instance)
	s.NoError(err)
	s.Empty(violations)
}

func (s *StateActionExecutorSuite) TestStateNotFoundReturnsError() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	bogusStateKey := mustKey("domain/d/subdomain/s/class/order/state/bogus")

	class := helper.Must(model_class.NewClass(classKey, "Order", "", nil, nil, nil, ""))
	class.SetAttributes(map[identity.Key]model_class.Attribute{})
	class.SetStates(map[identity.Key]model_state.State{})
	class.SetEvents(map[identity.Key]model_state.Event{})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{})

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	instance := simState.CreateInstance(classKey, attrs)

	ae := buildStateActionTestExecutor(simState)
	sae := NewStateActionExecutor(ae)

	_, err := sae.ExecuteEntryActions(class, bogusStateKey, instance)
	s.Error(err)
	s.Contains(err.Error(), "not found")
}
