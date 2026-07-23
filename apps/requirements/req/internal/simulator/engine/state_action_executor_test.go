package engine

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/schema"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/actions"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
	"github.com/stretchr/testify/suite"
)

// stateActionOrderSpec parses a TLA+ expression in the context of the Order class
// used by state action tests, with attributes: exit_count, entry_count.
func stateActionOrderSpec(tla string) logic_spec.ExpressionSpec {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	ctx := &convert.LowerContext{
		ClassKey: classKey,
		AttributeNames: map[string]identity.Key{
			"exit_count":  helper.Must(identity.NewAttributeKey(classKey, "exit_count")),
			"entry_count": helper.Must(identity.NewAttributeKey(classKey, "entry_count")),
		},
	}
	pf := convert.NewExpressionParseFunc(ctx)
	spec := helper.Must(logic_spec.NewExpressionSpec("tla_plus", tla, pf))
	return spec
}

type StateActionExecutorSuite struct {
	suite.Suite
}

func TestStateActionExecutorSuite(t *testing.T) {
	suite.Run(t, new(StateActionExecutorSuite))
}

// buildStateActionTestExecutor creates an ActionExecutor suitable for state action tests.
func buildStateActionTestExecutor(simState *instance.State) *actions.ActionExecutor {
	bb := state.NewBindingsBuilder(simState)
	ge := actions.NewGuardEvaluator(bb)
	return actions.NewActionExecutor(bb, actions.InvariantRuntimeCheckers{Checker: nil, DataType: nil}, nil, ge, nil, nil)
}

func (s *StateActionExecutorSuite) TestExitActionsFireOnTransition() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	stateOpenKey := mustKey("domain/d/subdomain/s/class/order/state/open")
	stateClosedKey := mustKey("domain/d/subdomain/s/class/order/state/closed")
	actionExitKey := mustKey("domain/d/subdomain/s/class/order/action/on_exit")
	stateActionKey := mustKey("domain/d/subdomain/s/class/order/state/open/saction/exit/on_exit")

	guaranteeKey := helper.Must(identity.NewActionGuaranteeKey(actionExitKey, "0"))
	guaranteeLogic := model_logic.NewLogic(guaranteeKey, model_logic.LogicTypeStateChange, "Postcondition.", "exit_count", stateActionOrderSpec("self.exit_count + 1"), nil)
	actionExit := model_state.NewAction(actionExitKey, model_state.ActionDetails{Name: "OnExit", Details: ""}, nil, []model_logic.Logic{guaranteeLogic}, nil, nil)

	stateActionExit := model_state.NewStateAction(stateActionKey, actionExitKey, "exit")

	stateOpen := model_state.NewState(stateOpenKey, "Open", "", "")
	stateOpen.SetActions([]model_state.StateAction{stateActionExit})
	stateClosed := model_state.NewState(stateClosedKey, "Closed", "", "")

	class := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Order", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetAttributes(nil)
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
	class = lowerClass(class, classKey)

	simState := instance.NewState(schema.NewFromModel(schema.EmptyModel()))
	attrs := object.NewRecord()
	attrs.Set("exit_count", object.NewInteger(0))
	attrs.Set("_state", object.NewString("Open"))
	instance := simState.CreateInstance(classKey, attrs)

	ae := buildStateActionTestExecutor(simState)
	sae := NewStateActionExecutor(ae)

	_, violations, err := sae.ExecuteExitActions(class, stateOpenKey, instance)
	s.Require().NoError(err)
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
	guaranteeLogic := model_logic.NewLogic(guaranteeKey, model_logic.LogicTypeStateChange, "Postcondition.", "entry_count", stateActionOrderSpec("self.entry_count + 1"), nil)
	actionEntry := model_state.NewAction(actionEntryKey, model_state.ActionDetails{Name: "OnEntry", Details: ""}, nil, []model_logic.Logic{guaranteeLogic}, nil, nil)

	stateActionEntry := model_state.NewStateAction(stateActionKey, actionEntryKey, "entry")

	stateOpen := model_state.NewState(stateOpenKey, "Open", "", "")
	stateOpen.SetActions([]model_state.StateAction{stateActionEntry})

	class := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Order", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetAttributes(nil)
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
	class = lowerClass(class, classKey)

	simState := instance.NewState(schema.NewFromModel(schema.EmptyModel()))
	attrs := object.NewRecord()
	attrs.Set("entry_count", object.NewInteger(0))
	attrs.Set("_state", object.NewString("Open"))
	instance := simState.CreateInstance(classKey, attrs)

	ae := buildStateActionTestExecutor(simState)
	sae := NewStateActionExecutor(ae)

	_, violations, err := sae.ExecuteEntryActions(class, stateOpenKey, instance)
	s.Require().NoError(err)
	s.Empty(violations)

	updated := simState.GetInstance(instance.ID)
	s.Equal("1", updated.GetAttribute("entry_count").Inspect())
}

func (s *StateActionExecutorSuite) TestNoStateActionsReturnsEmpty() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	stateOpenKey := mustKey("domain/d/subdomain/s/class/order/state/open")

	stateOpen := model_state.NewState(stateOpenKey, "Open", "", "")

	class := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Order", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetAttributes(nil)
	class.SetStates(map[identity.Key]model_state.State{
		stateOpenKey: stateOpen,
	})
	class.SetEvents(map[identity.Key]model_state.Event{})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{})

	simState := instance.NewState(schema.NewFromModel(schema.EmptyModel()))
	attrs := object.NewRecord()
	attrs.Set("_state", object.NewString("Open"))
	instance := simState.CreateInstance(classKey, attrs)

	ae := buildStateActionTestExecutor(simState)
	sae := NewStateActionExecutor(ae)

	_, violations, err := sae.ExecuteExitActions(class, stateOpenKey, instance)
	s.Require().NoError(err)
	s.Empty(violations)
}

func (s *StateActionExecutorSuite) TestStateNotFoundReturnsError() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	bogusStateKey := mustKey("domain/d/subdomain/s/class/order/state/bogus")

	class := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Order", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetAttributes(nil)
	class.SetStates(map[identity.Key]model_state.State{})
	class.SetEvents(map[identity.Key]model_state.Event{})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{})

	simState := instance.NewState(schema.NewFromModel(schema.EmptyModel()))
	attrs := object.NewRecord()
	instance := simState.CreateInstance(classKey, attrs)

	ae := buildStateActionTestExecutor(simState)
	sae := NewStateActionExecutor(ae)

	_, _, err := sae.ExecuteEntryActions(class, bogusStateKey, instance)
	s.Require().Error(err)
	s.Contains(err.Error(), "not found")
}
