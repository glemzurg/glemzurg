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
	guaranteeLogic := helper.Must(model_logic.NewLogic(guaranteeKey, model_logic.LogicTypeStateChange, "Postcondition.", model_logic.NotationTLAPlus, "self.exit_count' = self.exit_count + 1"))
	actionExit := helper.Must(model_state.NewAction(actionExitKey, "OnExit", "", nil, []model_logic.Logic{guaranteeLogic}, nil, nil))

	class := helper.Must(model_class.NewClass(classKey, "Order", "", nil, nil, nil, ""))
	class.Attributes = map[identity.Key]model_class.Attribute{}
	class.States = map[identity.Key]model_state.State{
		stateOpenKey: {
			Key:  stateOpenKey,
			Name: "Open",
			Actions: []model_state.StateAction{
				{Key: stateActionKey, ActionKey: actionExitKey, When: "exit"},
			},
		},
		stateClosedKey: {Key: stateClosedKey, Name: "Closed"},
	}
	class.Events = map[identity.Key]model_state.Event{}
	class.Guards = map[identity.Key]model_state.Guard{}
	class.Actions = map[identity.Key]model_state.Action{
		actionExitKey: actionExit,
	}
	class.Queries = map[identity.Key]model_state.Query{}
	class.Transitions = map[identity.Key]model_state.Transition{}

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
	guaranteeLogic := helper.Must(model_logic.NewLogic(guaranteeKey, model_logic.LogicTypeStateChange, "Postcondition.", model_logic.NotationTLAPlus, "self.entry_count' = self.entry_count + 1"))
	actionEntry := helper.Must(model_state.NewAction(actionEntryKey, "OnEntry", "", nil, []model_logic.Logic{guaranteeLogic}, nil, nil))

	class := helper.Must(model_class.NewClass(classKey, "Order", "", nil, nil, nil, ""))
	class.Attributes = map[identity.Key]model_class.Attribute{}
	class.States = map[identity.Key]model_state.State{
		stateOpenKey: {
			Key:  stateOpenKey,
			Name: "Open",
			Actions: []model_state.StateAction{
				{Key: stateActionKey, ActionKey: actionEntryKey, When: "entry"},
			},
		},
	}
	class.Events = map[identity.Key]model_state.Event{}
	class.Guards = map[identity.Key]model_state.Guard{}
	class.Actions = map[identity.Key]model_state.Action{
		actionEntryKey: actionEntry,
	}
	class.Queries = map[identity.Key]model_state.Query{}
	class.Transitions = map[identity.Key]model_state.Transition{}

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

	class := helper.Must(model_class.NewClass(classKey, "Order", "", nil, nil, nil, ""))
	class.Attributes = map[identity.Key]model_class.Attribute{}
	class.States = map[identity.Key]model_state.State{
		stateOpenKey: {Key: stateOpenKey, Name: "Open"},
	}
	class.Events = map[identity.Key]model_state.Event{}
	class.Guards = map[identity.Key]model_state.Guard{}
	class.Actions = map[identity.Key]model_state.Action{}
	class.Queries = map[identity.Key]model_state.Query{}
	class.Transitions = map[identity.Key]model_state.Transition{}

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
	class.Attributes = map[identity.Key]model_class.Attribute{}
	class.States = map[identity.Key]model_state.State{}
	class.Events = map[identity.Key]model_state.Event{}
	class.Guards = map[identity.Key]model_state.Guard{}
	class.Actions = map[identity.Key]model_state.Action{}
	class.Queries = map[identity.Key]model_state.Query{}
	class.Transitions = map[identity.Key]model_state.Transition{}

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	instance := simState.CreateInstance(classKey, attrs)

	ae := buildStateActionTestExecutor(simState)
	sae := NewStateActionExecutor(ae)

	_, err := sae.ExecuteEntryActions(class, bogusStateKey, instance)
	s.Error(err)
	s.Contains(err.Error(), "not found")
}
