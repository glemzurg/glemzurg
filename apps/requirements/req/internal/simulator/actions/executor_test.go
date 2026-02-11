package actions

import (
	"math/rand"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/invariants"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
	"github.com/stretchr/testify/suite"
)

type ActionsSuite struct {
	suite.Suite
}

func TestActionsSuite(t *testing.T) {
	suite.Run(t, new(ActionsSuite))
}

// mustKey parses a key string or panics.
func mustKey(s string) identity.Key {
	k, err := identity.ParseKey(s)
	if err != nil {
		panic(err)
	}
	return k
}

// --- Helper: build a minimal executor for tests ---

func buildTestExecutor(simState *state.SimulationState, model *req_model.Model) (*ActionExecutor, error) {
	bb := state.NewBindingsBuilder(simState)
	ge := NewGuardEvaluator(bb)

	var ic *invariants.InvariantChecker
	if model != nil {
		var err error
		ic, err = invariants.NewInvariantChecker(model)
		if err != nil {
			return nil, err
		}
	}

	var dc *invariants.DataTypeChecker
	if model != nil {
		dc, _ = invariants.NewDataTypeChecker(model)
	}

	return NewActionExecutor(bb, ic, dc, nil, ge, nil), nil
}

// --- Helper: create a simple class with states, actions, and transitions ---

// testOrderClass creates an Order class with states (Open, Closed), events (close),
// an action (DoClose) that sets self.amount' = self.amount + 10, and a transition.
func testOrderClass() (model_class.Class, identity.Key) {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	stateOpenKey := mustKey("domain/d/subdomain/s/class/order/state/open")
	stateClosedKey := mustKey("domain/d/subdomain/s/class/order/state/closed")
	eventCloseKey := mustKey("domain/d/subdomain/s/class/order/event/close")
	actionCloseKey := mustKey("domain/d/subdomain/s/class/order/action/do_close")
	transKey := mustKey("domain/d/subdomain/s/class/order/transition/close")

	class := model_class.Class{
		Key:        classKey,
		Name:       "Order",
		Attributes: map[identity.Key]model_class.Attribute{},
		States: map[identity.Key]model_state.State{
			stateOpenKey:   {Key: stateOpenKey, Name: "Open"},
			stateClosedKey: {Key: stateClosedKey, Name: "Closed"},
		},
		Events: map[identity.Key]model_state.Event{
			eventCloseKey: {Key: eventCloseKey, Name: "close"},
		},
		Guards: map[identity.Key]model_state.Guard{},
		Actions: map[identity.Key]model_state.Action{
			actionCloseKey: {
				Key:  actionCloseKey,
				Name: "DoClose",
				Guarantees: []model_logic.Logic{
					{Key: "guar_1", Description: "Postcondition.", Notation: model_logic.NotationTLAPlus, Specification: "self.amount' = self.amount + 10"},
				},
			},
		},
		Queries: map[identity.Key]model_state.Query{},
		Transitions: map[identity.Key]model_state.Transition{
			transKey: {
				Key:          transKey,
				FromStateKey: &stateOpenKey,
				EventKey:     eventCloseKey,
				ActionKey:    &actionCloseKey,
				ToStateKey:   &stateClosedKey,
			},
		},
	}

	return class, classKey
}

// testModel wraps a class into a minimal model.
func testModel(class model_class.Class, classKey identity.Key) *req_model.Model {
	subdomainKey := mustKey("domain/d/subdomain/s")
	domainKey := mustKey("domain/d")

	return &req_model.Model{
		Key:  "test",
		Name: "Test",
		Domains: map[identity.Key]model_domain.Domain{
			domainKey: {
				Key:  domainKey,
				Name: "D",
				Subdomains: map[identity.Key]model_domain.Subdomain{
					subdomainKey: {
						Key:  subdomainKey,
						Name: "S",
						Classes: map[identity.Key]model_class.Class{
							classKey: class,
						},
					},
				},
			},
		},
	}
}

// ========================================================================
// ExecutionContext tests
// ========================================================================

func (s *ActionsSuite) TestExecutionContextRecordPrimed() {
	ctx := NewExecutionContext()

	err := ctx.RecordPrimedAssignment(1, "count", object.NewInteger(42))
	s.NoError(err)

	all := ctx.GetAllPrimedAssignments()
	s.Len(all, 1)
	s.Equal("42", all[1]["count"].Inspect())
}

func (s *ActionsSuite) TestExecutionContextRejectsStateField() {
	ctx := NewExecutionContext()

	err := ctx.RecordPrimedAssignment(1, "_state", object.NewString("Open"))
	s.Error(err)
	s.Contains(err.Error(), "_state")
}

func (s *ActionsSuite) TestExecutionContextReentrancyGuard() {
	ctx := NewExecutionContext()

	// First mutation is fine
	s.True(ctx.CanMutate(1))
	err := ctx.RecordPrimedAssignment(1, "count", object.NewInteger(1))
	s.NoError(err)

	// After mutation, instance 1 is locked
	s.False(ctx.CanMutate(1))

	// Instance 2 is still available
	s.True(ctx.CanMutate(2))
}

func (s *ActionsSuite) TestExecutionContextDepthLimit() {
	ctx := NewExecutionContext()

	for i := 0; i < 100; i++ {
		err := ctx.IncrementDepth()
		s.NoError(err)
	}

	// 101st should fail
	err := ctx.IncrementDepth()
	s.Error(err)
	s.Contains(err.Error(), "depth exceeded")
}

func (s *ActionsSuite) TestExecutionContextPostConditions() {
	ctx := NewExecutionContext()
	ctx.AddPostCondition(DeferredPostCondition{
		SourceName: "testAction",
		SourceType: "action",
		Index:      0,
	})
	s.Len(ctx.GetAllPostConditions(), 1)
}

// ========================================================================
// ExecuteAction tests
// ========================================================================

func (s *ActionsSuite) TestExecuteActionWithPrimedAssignment() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	actionKey := mustKey("domain/d/subdomain/s/class/order/action/increment")

	action := model_state.Action{
		Key:  actionKey,
		Name: "increment",
		Guarantees: []model_logic.Logic{
			{Key: "guar_1", Description: "Postcondition.", Notation: model_logic.NotationTLAPlus, Specification: "self.count' = self.count + 1"},
		},
	}

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("count", object.NewInteger(10))
	instance := simState.CreateInstance(classKey, attrs)

	exec, err := buildTestExecutor(simState, nil)
	s.NoError(err)

	result, err := exec.ExecuteAction(action, instance, nil)
	s.NoError(err)
	s.NotNil(result)
	s.True(result.Success)

	// Verify state was updated
	updated := simState.GetInstance(instance.ID)
	s.Equal("11", updated.GetAttribute("count").Inspect())
}

func (s *ActionsSuite) TestExecuteActionPreconditionPasses() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	actionKey := mustKey("domain/d/subdomain/s/class/order/action/close")

	action := model_state.Action{
		Key:  actionKey,
		Name: "close",
		Requires: []model_logic.Logic{
			{Key: "req_1", Description: "Precondition.", Notation: model_logic.NotationTLAPlus, Specification: "self.status = \"open\""},
		},
		Guarantees: []model_logic.Logic{
			{Key: "guar_1", Description: "Postcondition.", Notation: model_logic.NotationTLAPlus, Specification: "self.status' = \"closed\""},
		},
	}

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("status", object.NewString("open"))
	instance := simState.CreateInstance(classKey, attrs)

	exec, err := buildTestExecutor(simState, nil)
	s.NoError(err)

	result, err := exec.ExecuteAction(action, instance, nil)
	s.NoError(err)
	s.True(result.Success)

	updated := simState.GetInstance(instance.ID)
	s.Equal("closed", updated.GetAttribute("status").(*object.String).Value())
}

func (s *ActionsSuite) TestExecuteActionPreconditionFails() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	actionKey := mustKey("domain/d/subdomain/s/class/order/action/close")

	action := model_state.Action{
		Key:  actionKey,
		Name: "close",
		Requires: []model_logic.Logic{
			{Key: "req_1", Description: "Precondition.", Notation: model_logic.NotationTLAPlus, Specification: "self.status = \"open\""},
		},
		Guarantees: []model_logic.Logic{
			{Key: "guar_1", Description: "Postcondition.", Notation: model_logic.NotationTLAPlus, Specification: "self.status' = \"closed\""},
		},
	}

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("status", object.NewString("closed")) // already closed
	instance := simState.CreateInstance(classKey, attrs)

	exec, err := buildTestExecutor(simState, nil)
	s.NoError(err)

	_, err = exec.ExecuteAction(action, instance, nil)
	s.Error(err)
	s.Contains(err.Error(), "precondition failed")
}

func (s *ActionsSuite) TestExecuteActionWithParameters() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	actionKey := mustKey("domain/d/subdomain/s/class/order/action/set_amount")

	action := model_state.Action{
		Key:  actionKey,
		Name: "set_amount",
		Guarantees: []model_logic.Logic{
			{Key: "guar_1", Description: "Postcondition.", Notation: model_logic.NotationTLAPlus, Specification: "self.amount' = amount"},
		},
	}

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("amount", object.NewInteger(0))
	instance := simState.CreateInstance(classKey, attrs)

	exec, err := buildTestExecutor(simState, nil)
	s.NoError(err)

	params := map[string]object.Object{
		"amount": object.NewInteger(500),
	}

	result, err := exec.ExecuteAction(action, instance, params)
	s.NoError(err)
	s.True(result.Success)

	updated := simState.GetInstance(instance.ID)
	s.Equal("500", updated.GetAttribute("amount").Inspect())
}

// ========================================================================
// ExecuteQuery tests
// ========================================================================

func (s *ActionsSuite) TestExecuteQueryReturnsOutput() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	queryKey := mustKey("domain/d/subdomain/s/class/order/query/get_total")

	query := model_state.Query{
		Key:  queryKey,
		Name: "get_total",
		Guarantees: []model_logic.Logic{
			{Key: "guar_1", Description: "Postcondition.", Notation: model_logic.NotationTLAPlus, Specification: "result' = self.amount * 2"},
		},
	}

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("amount", object.NewInteger(50))
	instance := simState.CreateInstance(classKey, attrs)

	exec, err := buildTestExecutor(simState, nil)
	s.NoError(err)

	result, err := exec.ExecuteQuery(query, instance, nil)
	s.NoError(err)
	s.True(result.Success)
	s.NotNil(result.Outputs["result"])
	s.Equal("100", result.Outputs["result"].Inspect())
}

func (s *ActionsSuite) TestExecuteQueryDoesNotModifyState() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	queryKey := mustKey("domain/d/subdomain/s/class/order/query/get_total")

	query := model_state.Query{
		Key:  queryKey,
		Name: "get_total",
		Guarantees: []model_logic.Logic{
			{Key: "guar_1", Description: "Postcondition.", Notation: model_logic.NotationTLAPlus, Specification: "result' = self.amount"},
		},
	}

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("amount", object.NewInteger(50))
	instance := simState.CreateInstance(classKey, attrs)

	exec, err := buildTestExecutor(simState, nil)
	s.NoError(err)

	_, err = exec.ExecuteQuery(query, instance, nil)
	s.NoError(err)

	// State should be unchanged
	unchanged := simState.GetInstance(instance.ID)
	s.Equal("50", unchanged.GetAttribute("amount").Inspect())
}

func (s *ActionsSuite) TestExecuteQueryPreconditionFails() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	queryKey := mustKey("domain/d/subdomain/s/class/order/query/get_total")

	query := model_state.Query{
		Key:  queryKey,
		Name: "get_total",
		Requires: []model_logic.Logic{
			{Key: "req_1", Description: "Precondition.", Notation: model_logic.NotationTLAPlus, Specification: "self.amount > 100"},
		},
		Guarantees: []model_logic.Logic{
			{Key: "guar_1", Description: "Postcondition.", Notation: model_logic.NotationTLAPlus, Specification: "result' = self.amount"},
		},
	}

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("amount", object.NewInteger(50))
	instance := simState.CreateInstance(classKey, attrs)

	exec, err := buildTestExecutor(simState, nil)
	s.NoError(err)

	_, err = exec.ExecuteQuery(query, instance, nil)
	s.Error(err)
	s.Contains(err.Error(), "precondition failed")
}

// ========================================================================
// GuardEvaluator tests
// ========================================================================

func (s *ActionsSuite) TestGuardEvaluatorAllTrue() {
	classKey := mustKey("domain/d/subdomain/s/class/order")

	guard := model_state.Guard{
		Key:  mustKey("domain/d/subdomain/s/class/order/guard/is_open"),
		Name: "is_open",
		Logic: model_logic.Logic{
			Key:           "guard_logic_1",
			Description:   "Guard for open status and positive amount.",
			Notation:      model_logic.NotationTLAPlus,
			Specification: "self.status = \"open\" /\\ self.amount > 0",
		},
	}

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("status", object.NewString("open"))
	attrs.Set("amount", object.NewInteger(100))
	instance := simState.CreateInstance(classKey, attrs)

	bb := state.NewBindingsBuilder(simState)
	ge := NewGuardEvaluator(bb)

	passes, err := ge.EvaluateGuard(guard, instance)
	s.NoError(err)
	s.True(passes)
}

func (s *ActionsSuite) TestGuardEvaluatorOneFalse() {
	classKey := mustKey("domain/d/subdomain/s/class/order")

	guard := model_state.Guard{
		Key:  mustKey("domain/d/subdomain/s/class/order/guard/is_open"),
		Name: "is_open",
		Logic: model_logic.Logic{
			Key:           "guard_logic_2",
			Description:   "Guard for open status and positive amount.",
			Notation:      model_logic.NotationTLAPlus,
			Specification: "self.status = \"open\" /\\ self.amount > 0",
		},
	}

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("status", object.NewString("closed")) // guard fails here
	attrs.Set("amount", object.NewInteger(100))
	instance := simState.CreateInstance(classKey, attrs)

	bb := state.NewBindingsBuilder(simState)
	ge := NewGuardEvaluator(bb)

	passes, err := ge.EvaluateGuard(guard, instance)
	s.NoError(err)
	s.False(passes)
}

// ========================================================================
// ExecuteTransition tests
// ========================================================================

func (s *ActionsSuite) TestExecuteTransitionNormal() {
	class, classKey := testOrderClass()

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("amount", object.NewInteger(100))
	attrs.Set("_state", object.NewString("Open"))
	instance := simState.CreateInstance(classKey, attrs)

	// Set state machine state
	stateOpenKey := mustKey("domain/d/subdomain/s/class/order/state/open")
	_ = simState.SetStateMachineState(instance.ID, stateOpenKey)

	exec, err := buildTestExecutor(simState, nil)
	s.NoError(err)

	eventCloseKey := mustKey("domain/d/subdomain/s/class/order/event/close")
	event := class.Events[eventCloseKey]

	result, err := exec.ExecuteTransition(class, event, instance, nil, nil, nil)
	s.NoError(err)
	s.NotNil(result)

	s.Equal("Open", result.FromState)
	s.Equal("Closed", result.ToState)
	s.False(result.WasCreation)
	s.False(result.WasDeletion)

	// Check action result
	s.NotNil(result.ActionResult)
	s.True(result.ActionResult.Success)

	// Check state was updated (amount went from 100 to 110)
	updated := simState.GetInstance(instance.ID)
	s.Equal("110", updated.GetAttribute("amount").Inspect())
	s.Equal("Closed", updated.GetAttribute("_state").(*object.String).Value())
}

func (s *ActionsSuite) TestExecuteTransitionCreation() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	stateOpenKey := mustKey("domain/d/subdomain/s/class/order/state/open")
	eventCreateKey := mustKey("domain/d/subdomain/s/class/order/event/create")
	transKey := mustKey("domain/d/subdomain/s/class/order/transition/create")

	class := model_class.Class{
		Key:        classKey,
		Name:       "Order",
		Attributes: map[identity.Key]model_class.Attribute{},
		States: map[identity.Key]model_state.State{
			stateOpenKey: {Key: stateOpenKey, Name: "Open"},
		},
		Events: map[identity.Key]model_state.Event{
			eventCreateKey: {Key: eventCreateKey, Name: "create"},
		},
		Guards:  map[identity.Key]model_state.Guard{},
		Actions: map[identity.Key]model_state.Action{},
		Queries: map[identity.Key]model_state.Query{},
		Transitions: map[identity.Key]model_state.Transition{
			transKey: {
				Key:          transKey,
				FromStateKey: nil, // Initial state → creation
				EventKey:     eventCreateKey,
				ToStateKey:   &stateOpenKey,
			},
		},
	}

	simState := state.NewSimulationState()
	exec, err := buildTestExecutor(simState, nil)
	s.NoError(err)

	event := class.Events[eventCreateKey]

	result, err := exec.ExecuteTransition(class, event, nil, nil, nil, nil)
	s.NoError(err)
	s.True(result.WasCreation)
	s.Equal("Open", result.ToState)

	// Verify instance was created
	s.Equal(1, simState.InstanceCount())
	created := simState.GetInstance(result.InstanceID)
	s.NotNil(created)
	s.Equal("Open", created.GetAttribute("_state").(*object.String).Value())
}

func (s *ActionsSuite) TestExecuteTransitionDeletion() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	stateOpenKey := mustKey("domain/d/subdomain/s/class/order/state/open")
	eventDeleteKey := mustKey("domain/d/subdomain/s/class/order/event/delete")
	transKey := mustKey("domain/d/subdomain/s/class/order/transition/delete")

	class := model_class.Class{
		Key:        classKey,
		Name:       "Order",
		Attributes: map[identity.Key]model_class.Attribute{},
		States: map[identity.Key]model_state.State{
			stateOpenKey: {Key: stateOpenKey, Name: "Open"},
		},
		Events: map[identity.Key]model_state.Event{
			eventDeleteKey: {Key: eventDeleteKey, Name: "delete"},
		},
		Guards:  map[identity.Key]model_state.Guard{},
		Actions: map[identity.Key]model_state.Action{},
		Queries: map[identity.Key]model_state.Query{},
		Transitions: map[identity.Key]model_state.Transition{
			transKey: {
				Key:          transKey,
				FromStateKey: &stateOpenKey,
				EventKey:     eventDeleteKey,
				ToStateKey:   nil, // Final state → deletion
			},
		},
	}

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("_state", object.NewString("Open"))
	instance := simState.CreateInstance(classKey, attrs)
	_ = simState.SetStateMachineState(instance.ID, stateOpenKey)

	exec, err := buildTestExecutor(simState, nil)
	s.NoError(err)

	event := class.Events[eventDeleteKey]

	result, err := exec.ExecuteTransition(class, event, instance, nil, nil, nil)
	s.NoError(err)
	s.True(result.WasDeletion)

	// Instance should be deleted
	s.Nil(simState.GetInstance(instance.ID))
	s.Equal(0, simState.InstanceCount())
}

func (s *ActionsSuite) TestExecuteTransitionNoMatchingTransition() {
	class, classKey := testOrderClass()

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("_state", object.NewString("Closed")) // No transition from Closed
	instance := simState.CreateInstance(classKey, attrs)

	stateClosedKey := mustKey("domain/d/subdomain/s/class/order/state/closed")
	_ = simState.SetStateMachineState(instance.ID, stateClosedKey)

	exec, err := buildTestExecutor(simState, nil)
	s.NoError(err)

	eventCloseKey := mustKey("domain/d/subdomain/s/class/order/event/close")
	event := class.Events[eventCloseKey]

	_, err = exec.ExecuteTransition(class, event, instance, nil, nil, nil)
	s.Error(err)
	s.Contains(err.Error(), "no transitions")
}

// ========================================================================
// Guard determinism tests (transition with guards)
// ========================================================================

func (s *ActionsSuite) TestTransitionGuardDeterminism() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	stateOpenKey := mustKey("domain/d/subdomain/s/class/order/state/open")
	stateApprovedKey := mustKey("domain/d/subdomain/s/class/order/state/approved")
	stateRejectedKey := mustKey("domain/d/subdomain/s/class/order/state/rejected")
	eventReviewKey := mustKey("domain/d/subdomain/s/class/order/event/review")
	guardHighKey := mustKey("domain/d/subdomain/s/class/order/guard/high_value")
	guardLowKey := mustKey("domain/d/subdomain/s/class/order/guard/low_value")
	transApproveKey := mustKey("domain/d/subdomain/s/class/order/transition/approve")
	transRejectKey := mustKey("domain/d/subdomain/s/class/order/transition/reject")

	class := model_class.Class{
		Key:        classKey,
		Name:       "Order",
		Attributes: map[identity.Key]model_class.Attribute{},
		States: map[identity.Key]model_state.State{
			stateOpenKey:     {Key: stateOpenKey, Name: "Open"},
			stateApprovedKey: {Key: stateApprovedKey, Name: "Approved"},
			stateRejectedKey: {Key: stateRejectedKey, Name: "Rejected"},
		},
		Events: map[identity.Key]model_state.Event{
			eventReviewKey: {Key: eventReviewKey, Name: "review"},
		},
		Guards: map[identity.Key]model_state.Guard{
			guardHighKey: {
				Key:  guardHighKey,
				Name: "high_value",
				Logic: model_logic.Logic{
					Key:           "guard_logic_3",
					Description:   "High value guard.",
					Notation:      model_logic.NotationTLAPlus,
					Specification: "self.amount >= 100",
				},
			},
			guardLowKey: {
				Key:  guardLowKey,
				Name: "low_value",
				Logic: model_logic.Logic{
					Key:           "guard_logic_4",
					Description:   "Low value guard.",
					Notation:      model_logic.NotationTLAPlus,
					Specification: "self.amount < 100",
				},
			},
		},
		Actions: map[identity.Key]model_state.Action{},
		Queries: map[identity.Key]model_state.Query{},
		Transitions: map[identity.Key]model_state.Transition{
			transApproveKey: {
				Key:          transApproveKey,
				FromStateKey: &stateOpenKey,
				EventKey:     eventReviewKey,
				GuardKey:     &guardHighKey,
				ToStateKey:   &stateApprovedKey,
			},
			transRejectKey: {
				Key:          transRejectKey,
				FromStateKey: &stateOpenKey,
				EventKey:     eventReviewKey,
				GuardKey:     &guardLowKey,
				ToStateKey:   &stateRejectedKey,
			},
		},
	}

	simState := state.NewSimulationState()

	// Case 1: High value order → should go to Approved
	attrs := object.NewRecord()
	attrs.Set("amount", object.NewInteger(200))
	attrs.Set("_state", object.NewString("Open"))
	instance := simState.CreateInstance(classKey, attrs)
	_ = simState.SetStateMachineState(instance.ID, stateOpenKey)

	exec, err := buildTestExecutor(simState, nil)
	s.NoError(err)

	event := class.Events[eventReviewKey]
	result, err := exec.ExecuteTransition(class, event, instance, nil, nil, nil)
	s.NoError(err)
	s.Equal("Approved", result.ToState)

	// Case 2: Low value order → should go to Rejected
	attrs2 := object.NewRecord()
	attrs2.Set("amount", object.NewInteger(50))
	attrs2.Set("_state", object.NewString("Open"))
	instance2 := simState.CreateInstance(classKey, attrs2)
	_ = simState.SetStateMachineState(instance2.ID, stateOpenKey)

	result2, err := exec.ExecuteTransition(class, event, instance2, nil, nil, nil)
	s.NoError(err)
	s.Equal("Rejected", result2.ToState)
}

func (s *ActionsSuite) TestTransitionMultipleGuardsTrue() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	stateOpenKey := mustKey("domain/d/subdomain/s/class/order/state/open")
	stateAKey := mustKey("domain/d/subdomain/s/class/order/state/a")
	stateBKey := mustKey("domain/d/subdomain/s/class/order/state/b")
	eventKey := mustKey("domain/d/subdomain/s/class/order/event/go")
	guardAlwaysKey1 := mustKey("domain/d/subdomain/s/class/order/guard/always1")
	guardAlwaysKey2 := mustKey("domain/d/subdomain/s/class/order/guard/always2")
	trans1Key := mustKey("domain/d/subdomain/s/class/order/transition/t1")
	trans2Key := mustKey("domain/d/subdomain/s/class/order/transition/t2")

	class := model_class.Class{
		Key:        classKey,
		Name:       "Order",
		Attributes: map[identity.Key]model_class.Attribute{},
		States: map[identity.Key]model_state.State{
			stateOpenKey: {Key: stateOpenKey, Name: "Open"},
			stateAKey:    {Key: stateAKey, Name: "A"},
			stateBKey:    {Key: stateBKey, Name: "B"},
		},
		Events: map[identity.Key]model_state.Event{
			eventKey: {Key: eventKey, Name: "go"},
		},
		Guards: map[identity.Key]model_state.Guard{
			guardAlwaysKey1: {Key: guardAlwaysKey1, Name: "always1", Logic: model_logic.Logic{Key: "guard_logic_5", Description: "Always true guard.", Notation: model_logic.NotationTLAPlus, Specification: "TRUE"}},
			guardAlwaysKey2: {Key: guardAlwaysKey2, Name: "always2", Logic: model_logic.Logic{Key: "guard_logic_6", Description: "Always true guard.", Notation: model_logic.NotationTLAPlus, Specification: "TRUE"}},
		},
		Actions: map[identity.Key]model_state.Action{},
		Queries: map[identity.Key]model_state.Query{},
		Transitions: map[identity.Key]model_state.Transition{
			trans1Key: {
				Key:          trans1Key,
				FromStateKey: &stateOpenKey,
				EventKey:     eventKey,
				GuardKey:     &guardAlwaysKey1,
				ToStateKey:   &stateAKey,
			},
			trans2Key: {
				Key:          trans2Key,
				FromStateKey: &stateOpenKey,
				EventKey:     eventKey,
				GuardKey:     &guardAlwaysKey2,
				ToStateKey:   &stateBKey,
			},
		},
	}

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("_state", object.NewString("Open"))
	instance := simState.CreateInstance(classKey, attrs)
	_ = simState.SetStateMachineState(instance.ID, stateOpenKey)

	exec, err := buildTestExecutor(simState, nil)
	s.NoError(err)

	event := class.Events[eventKey]
	_, err = exec.ExecuteTransition(class, event, instance, nil, nil, nil)
	s.Error(err)
	s.Contains(err.Error(), "non-determinism")
}

func (s *ActionsSuite) TestTransitionNoGuardsTrue() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	stateOpenKey := mustKey("domain/d/subdomain/s/class/order/state/open")
	stateAKey := mustKey("domain/d/subdomain/s/class/order/state/a")
	eventKey := mustKey("domain/d/subdomain/s/class/order/event/go")
	guardNeverKey := mustKey("domain/d/subdomain/s/class/order/guard/never")
	transKey := mustKey("domain/d/subdomain/s/class/order/transition/t1")

	class := model_class.Class{
		Key:        classKey,
		Name:       "Order",
		Attributes: map[identity.Key]model_class.Attribute{},
		States: map[identity.Key]model_state.State{
			stateOpenKey: {Key: stateOpenKey, Name: "Open"},
			stateAKey:    {Key: stateAKey, Name: "A"},
		},
		Events: map[identity.Key]model_state.Event{
			eventKey: {Key: eventKey, Name: "go"},
		},
		Guards: map[identity.Key]model_state.Guard{
			guardNeverKey: {Key: guardNeverKey, Name: "never", Logic: model_logic.Logic{Key: "guard_logic_7", Description: "Never true guard.", Notation: model_logic.NotationTLAPlus, Specification: "FALSE"}},
		},
		Actions: map[identity.Key]model_state.Action{},
		Queries: map[identity.Key]model_state.Query{},
		Transitions: map[identity.Key]model_state.Transition{
			transKey: {
				Key:          transKey,
				FromStateKey: &stateOpenKey,
				EventKey:     eventKey,
				GuardKey:     &guardNeverKey,
				ToStateKey:   &stateAKey,
			},
		},
	}

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("_state", object.NewString("Open"))
	instance := simState.CreateInstance(classKey, attrs)
	_ = simState.SetStateMachineState(instance.ID, stateOpenKey)

	exec, err := buildTestExecutor(simState, nil)
	s.NoError(err)

	event := class.Events[eventKey]
	_, err = exec.ExecuteTransition(class, event, instance, nil, nil, nil)
	s.Error(err)
	s.Contains(err.Error(), "deadlock")
}

// ========================================================================
// ValidateClassForSimulation test
// ========================================================================

func (s *ActionsSuite) TestValidateClassForSimulationNoStates() {
	class := model_class.Class{
		Key:    mustKey("domain/d/subdomain/s/class/empty"),
		Name:   "Empty",
		States: map[identity.Key]model_state.State{},
	}

	err := ValidateClassForSimulation(class)
	s.Error(err)
	s.Contains(err.Error(), "no states")
}

func (s *ActionsSuite) TestValidateClassForSimulationWithStates() {
	stateKey := mustKey("domain/d/subdomain/s/class/c/state/s1")
	class := model_class.Class{
		Key:  mustKey("domain/d/subdomain/s/class/c"),
		Name: "C",
		States: map[identity.Key]model_state.State{
			stateKey: {Key: stateKey, Name: "S1"},
		},
	}

	err := ValidateClassForSimulation(class)
	s.NoError(err)
}

// ========================================================================
// GetStateEnumValues test
// ========================================================================

func (s *ActionsSuite) TestGetStateEnumValues() {
	stateOpenKey := mustKey("domain/d/subdomain/s/class/c/state/open")
	stateClosedKey := mustKey("domain/d/subdomain/s/class/c/state/closed")

	class := model_class.Class{
		Key:  mustKey("domain/d/subdomain/s/class/c"),
		Name: "C",
		States: map[identity.Key]model_state.State{
			stateOpenKey:   {Key: stateOpenKey, Name: "Open"},
			stateClosedKey: {Key: stateClosedKey, Name: "Closed"},
		},
	}

	values := GetStateEnumValues(class)
	s.Len(values, 2)
	s.Contains(values, "Open")
	s.Contains(values, "Closed")
}

// ========================================================================
// ParameterBinder tests
// ========================================================================

func (s *ActionsSuite) TestBindParametersSuccess() {
	binder := NewParameterBinder()

	paramDefs := []model_state.Parameter{
		{Name: "amount", DataTypeRules: "[0,100]"},
		{Name: "name", DataTypeRules: "string"},
	}

	values := map[string]object.Object{
		"amount": object.NewInteger(50),
		"name":   object.NewString("test"),
	}

	result, err := binder.BindParameters(paramDefs, values)
	s.NoError(err)
	s.Len(result, 2)
	s.Equal("50", result["amount"].Inspect())
	s.Equal("test", result["name"].(*object.String).Value())
}

func (s *ActionsSuite) TestBindParametersMissing() {
	binder := NewParameterBinder()

	paramDefs := []model_state.Parameter{
		{Name: "amount", DataTypeRules: "[0,100]"},
	}

	values := map[string]object.Object{} // missing amount

	_, err := binder.BindParameters(paramDefs, values)
	s.Error(err)
	s.Contains(err.Error(), "missing required parameter")
}

func (s *ActionsSuite) TestGenerateRandomParametersSpan() {
	binder := NewParameterBinder()
	rng := rand.New(rand.NewSource(42))

	lowerValue := 10
	higherValue := 20

	paramDefs := []model_state.Parameter{
		{
			Name:          "count",
			DataTypeRules: "[10, 20]",
			DataType: &model_data_type.DataType{
				CollectionType: model_data_type.CollectionTypeAtomic,
				Atomic: &model_data_type.Atomic{
					ConstraintType: model_data_type.ConstraintTypeSpan,
					Span: &model_data_type.AtomicSpan{
						LowerType:   "closed",
						LowerValue:  &lowerValue,
						HigherType:  "closed",
						HigherValue: &higherValue,
					},
				},
			},
		},
	}

	for i := 0; i < 100; i++ {
		result := binder.GenerateRandomParameters(paramDefs, rng)
		s.Contains(result, "count")
		num, ok := result["count"].(*object.Number)
		s.True(ok)
		val := num.Rat().Num().Int64()
		s.True(val >= 10 && val <= 20, "Generated value %d should be in [10,20]", val)
	}
}

func (s *ActionsSuite) TestGenerateRandomParametersEnum() {
	binder := NewParameterBinder()
	rng := rand.New(rand.NewSource(42))

	paramDefs := []model_state.Parameter{
		{
			Name:          "color",
			DataTypeRules: "{red, green, blue}",
			DataType: &model_data_type.DataType{
				CollectionType: model_data_type.CollectionTypeAtomic,
				Atomic: &model_data_type.Atomic{
					ConstraintType: model_data_type.ConstraintTypeEnumeration,
					Enums: []model_data_type.AtomicEnum{
						{Value: "red", SortOrder: 0},
						{Value: "green", SortOrder: 1},
						{Value: "blue", SortOrder: 2},
					},
				},
			},
		},
	}

	allowedValues := map[string]bool{"red": true, "green": true, "blue": true}

	for i := 0; i < 100; i++ {
		result := binder.GenerateRandomParameters(paramDefs, rng)
		str, ok := result["color"].(*object.String)
		s.True(ok)
		s.True(allowedValues[str.Value()], "Generated value %s should be in {red, green, blue}", str.Value())
	}
}

func (s *ActionsSuite) TestGenerateRandomParametersNoType() {
	binder := NewParameterBinder()
	rng := rand.New(rand.NewSource(42))

	paramDefs := []model_state.Parameter{
		{Name: "x", DataTypeRules: "unknown"},
	}

	result := binder.GenerateRandomParameters(paramDefs, rng)
	s.Contains(result, "x")
	_, ok := result["x"].(*object.Number)
	s.True(ok, "Should generate a number as default")
}

// ========================================================================
// ExecutionContext safety rules tests
// ========================================================================

func (s *ActionsSuite) TestExecutionContextSafetyRules() {
	ctx := NewExecutionContext()
	ctx.AddSafetyRule(DeferredSafetyRule{
		SourceName: "testAction",
		Index:      0,
	})
	s.Len(ctx.GetAllSafetyRules(), 1)
}

// ========================================================================
// Action validation tests (primed / non-primed constraints)
// ========================================================================

func (s *ActionsSuite) TestActionRejectsGuaranteesNonPrimed() {
	classKey := mustKey("domain/d/subdomain/s/class/c")
	actionKey := mustKey("domain/d/subdomain/s/class/c/action/a")

	action := model_state.Action{
		Key:  actionKey,
		Name: "BadAction",
		Guarantees: []model_logic.Logic{
			{Key: "guar_1", Description: "Postcondition.", Notation: model_logic.NotationTLAPlus, Specification: "self.count > 0"},
		},
	}

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("count", object.NewInteger(5))
	instance := simState.CreateInstance(classKey, attrs)

	executor, err := buildTestExecutor(simState, nil)
	s.Require().NoError(err)

	_, err = executor.ExecuteAction(action, instance, nil)
	s.Error(err)
	s.Contains(err.Error(), "Guarantees must be primed assignments only")
}

func (s *ActionsSuite) TestActionRejectsRequiresWithPrime() {
	classKey := mustKey("domain/d/subdomain/s/class/c")
	actionKey := mustKey("domain/d/subdomain/s/class/c/action/a")

	action := model_state.Action{
		Key:  actionKey,
		Name: "BadRequires",
		Requires: []model_logic.Logic{
			{Key: "req_1", Description: "Precondition.", Notation: model_logic.NotationTLAPlus, Specification: "self.count' > 0"},
		},
	}

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("count", object.NewInteger(5))
	instance := simState.CreateInstance(classKey, attrs)

	executor, err := buildTestExecutor(simState, nil)
	s.Require().NoError(err)

	_, err = executor.ExecuteAction(action, instance, nil)
	s.Error(err)
	s.Contains(err.Error(), "Requires must not contain primed variables")
}

func (s *ActionsSuite) TestActionSafetyRulesMustHavePrime() {
	classKey := mustKey("domain/d/subdomain/s/class/c")
	actionKey := mustKey("domain/d/subdomain/s/class/c/action/a")

	action := model_state.Action{
		Key:  actionKey,
		Name: "BadSafety",
		SafetyRules: []model_logic.Logic{
			{Key: "safety_1", Description: "Safety rule.", Notation: model_logic.NotationTLAPlus, Specification: "self.count > 0"},
		},
	}

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("count", object.NewInteger(5))
	instance := simState.CreateInstance(classKey, attrs)

	executor, err := buildTestExecutor(simState, nil)
	s.Require().NoError(err)

	_, err = executor.ExecuteAction(action, instance, nil)
	s.Error(err)
	s.Contains(err.Error(), "SafetyRules must reference primed variables")
}

// ========================================================================
// Safety rule pass / violation tests
// ========================================================================

func (s *ActionsSuite) TestActionSafetyRulesPass() {
	classKey := mustKey("domain/d/subdomain/s/class/c")
	actionKey := mustKey("domain/d/subdomain/s/class/c/action/a")

	action := model_state.Action{
		Key:  actionKey,
		Name: "GoodAction",
		Guarantees: []model_logic.Logic{
			{Key: "guar_1", Description: "Postcondition.", Notation: model_logic.NotationTLAPlus, Specification: "self.count' = self.count + 1"},
		},
		SafetyRules: []model_logic.Logic{
			{Key: "safety_1", Description: "Safety rule.", Notation: model_logic.NotationTLAPlus, Specification: "self.count' >= 1"},
		},
	}

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("count", object.NewInteger(5))
	instance := simState.CreateInstance(classKey, attrs)

	executor, err := buildTestExecutor(simState, nil)
	s.Require().NoError(err)

	result, err := executor.ExecuteAction(action, instance, nil)
	s.Require().NoError(err)
	s.True(result.Success)
	s.Empty(result.Violations)
}

func (s *ActionsSuite) TestActionSafetyRuleViolation() {
	classKey := mustKey("domain/d/subdomain/s/class/c")
	actionKey := mustKey("domain/d/subdomain/s/class/c/action/a")

	action := model_state.Action{
		Key:  actionKey,
		Name: "ViolatingAction",
		Guarantees: []model_logic.Logic{
			{Key: "guar_1", Description: "Postcondition.", Notation: model_logic.NotationTLAPlus, Specification: "self.count' = self.count + 1"},
		},
		SafetyRules: []model_logic.Logic{
			{Key: "safety_1", Description: "Safety rule.", Notation: model_logic.NotationTLAPlus, Specification: "self.count' < 0"}, // Will be FALSE (6 < 0 is false)
		},
	}

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("count", object.NewInteger(5))
	instance := simState.CreateInstance(classKey, attrs)

	executor, err := buildTestExecutor(simState, nil)
	s.Require().NoError(err)

	result, err := executor.ExecuteAction(action, instance, nil)
	s.Require().NoError(err)
	s.False(result.Success)
	s.Len(result.Violations, 1)
	s.Equal(invariants.ViolationTypeSafetyRule, result.Violations[0].Type)
}

// ========================================================================
// Guard rejects primed variables test
// ========================================================================

func (s *ActionsSuite) TestGuardRejectsPrimedVariables() {
	classKey := mustKey("domain/d/subdomain/s/class/c")

	guard := model_state.Guard{
		Key:  mustKey("domain/d/subdomain/s/class/c/guard/g"),
		Name: "BadGuard",
		Logic: model_logic.Logic{
			Key:           "guard_logic_8",
			Description:   "Guard with primed variable.",
			Notation:      model_logic.NotationTLAPlus,
			Specification: "self.count' > 0",
		},
	}

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("count", object.NewInteger(5))
	instance := simState.CreateInstance(classKey, attrs)

	bb := state.NewBindingsBuilder(simState)
	ge := NewGuardEvaluator(bb)

	_, err := ge.EvaluateGuard(guard, instance)
	s.Error(err)
	s.Contains(err.Error(), "guards must not contain primed variables")
}

// ========================================================================
// Query rejects primed variables in Requires test
// ========================================================================

func (s *ActionsSuite) TestQueryRejectsRequiresWithPrime() {
	classKey := mustKey("domain/d/subdomain/s/class/c")
	queryKey := mustKey("domain/d/subdomain/s/class/c/query/q")

	query := model_state.Query{
		Key:  queryKey,
		Name: "BadQuery",
		Requires: []model_logic.Logic{
			{Key: "req_1", Description: "Precondition.", Notation: model_logic.NotationTLAPlus, Specification: "self.count' > 0"},
		},
		Guarantees: []model_logic.Logic{
			{Key: "guar_1", Description: "Postcondition.", Notation: model_logic.NotationTLAPlus, Specification: "result' = self.count"},
		},
	}

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("count", object.NewInteger(5))
	instance := simState.CreateInstance(classKey, attrs)

	executor, err := buildTestExecutor(simState, nil)
	s.Require().NoError(err)

	_, err = executor.ExecuteQuery(query, instance, nil)
	s.Error(err)
	s.Contains(err.Error(), "Requires must not contain primed variables")
}
