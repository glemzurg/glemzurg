package actions

import (
	"math/rand"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
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

	guaranteeLogic := helper.Must(model_logic.NewLogic(
		helper.Must(identity.NewActionGuaranteeKey(actionCloseKey, "0")),
		model_logic.LogicTypeStateChange, "Postcondition.", "amount", model_logic.NotationTLAPlus, "self.amount + 10",
	))

	action := helper.Must(model_state.NewAction(actionCloseKey, "DoClose", "", nil, []model_logic.Logic{guaranteeLogic}, nil, nil))
	event := helper.Must(model_state.NewEvent(eventCloseKey, "close", "", nil))

	class := helper.Must(model_class.NewClass(classKey, "Order", "", nil, nil, nil, ""))
	class.Attributes = map[identity.Key]model_class.Attribute{}
	class.States = map[identity.Key]model_state.State{
		stateOpenKey:   {Key: stateOpenKey, Name: "Open"},
		stateClosedKey: {Key: stateClosedKey, Name: "Closed"},
	}
	class.Events = map[identity.Key]model_state.Event{
		eventCloseKey: event,
	}
	class.Guards = map[identity.Key]model_state.Guard{}
	class.Actions = map[identity.Key]model_state.Action{
		actionCloseKey: action,
	}
	class.Queries = map[identity.Key]model_state.Query{}
	class.Transitions = map[identity.Key]model_state.Transition{
		transKey: {
			Key:          transKey,
			FromStateKey: &stateOpenKey,
			EventKey:     eventCloseKey,
			ActionKey:    &actionCloseKey,
			ToStateKey:   &stateClosedKey,
		},
	}

	return class, classKey
}

// testModel wraps a class into a minimal model.
func testModel(class model_class.Class, classKey identity.Key) *req_model.Model {
	subdomainKey := mustKey("domain/d/subdomain/s")
	domainKey := mustKey("domain/d")

	subdomain := helper.Must(model_domain.NewSubdomain(subdomainKey, "S", "", ""))
	subdomain.Classes = map[identity.Key]model_class.Class{
		classKey: class,
	}

	domain := helper.Must(model_domain.NewDomain(domainKey, "D", "", false, ""))
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomainKey: subdomain,
	}

	model := helper.Must(req_model.NewModel("test", "Test", "", nil, nil))
	model.Domains = map[identity.Key]model_domain.Domain{
		domainKey: domain,
	}

	return &model
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

	guaranteeLogic := helper.Must(model_logic.NewLogic(
		helper.Must(identity.NewActionGuaranteeKey(actionKey, "0")),
		model_logic.LogicTypeStateChange, "Postcondition.", "count", model_logic.NotationTLAPlus, "self.count + 1",
	))

	action := helper.Must(model_state.NewAction(actionKey, "increment", "", nil, []model_logic.Logic{guaranteeLogic}, nil, nil))

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

	requireLogic := helper.Must(model_logic.NewLogic(
		helper.Must(identity.NewActionRequireKey(actionKey, "0")),
		model_logic.LogicTypeAssessment, "Precondition.", "", model_logic.NotationTLAPlus, "self.status = \"open\"",
	))
	guaranteeLogic := helper.Must(model_logic.NewLogic(
		helper.Must(identity.NewActionGuaranteeKey(actionKey, "0")),
		model_logic.LogicTypeStateChange, "Postcondition.", "status", model_logic.NotationTLAPlus, "\"closed\"",
	))

	action := helper.Must(model_state.NewAction(actionKey, "close", "", []model_logic.Logic{requireLogic}, []model_logic.Logic{guaranteeLogic}, nil, nil))

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

	requireLogic := helper.Must(model_logic.NewLogic(
		helper.Must(identity.NewActionRequireKey(actionKey, "0")),
		model_logic.LogicTypeAssessment, "Precondition.", "", model_logic.NotationTLAPlus, "self.status = \"open\"",
	))
	guaranteeLogic := helper.Must(model_logic.NewLogic(
		helper.Must(identity.NewActionGuaranteeKey(actionKey, "0")),
		model_logic.LogicTypeStateChange, "Postcondition.", "status", model_logic.NotationTLAPlus, "\"closed\"",
	))

	action := helper.Must(model_state.NewAction(actionKey, "close", "", []model_logic.Logic{requireLogic}, []model_logic.Logic{guaranteeLogic}, nil, nil))

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

	guaranteeLogic := helper.Must(model_logic.NewLogic(
		helper.Must(identity.NewActionGuaranteeKey(actionKey, "0")),
		model_logic.LogicTypeStateChange, "Postcondition.", "amount", model_logic.NotationTLAPlus, "amount",
	))

	action := helper.Must(model_state.NewAction(actionKey, "set_amount", "", nil, []model_logic.Logic{guaranteeLogic}, nil, nil))

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

	guaranteeLogic := helper.Must(model_logic.NewLogic(
		helper.Must(identity.NewQueryGuaranteeKey(queryKey, "0")),
		model_logic.LogicTypeQuery, "Postcondition.", "result", model_logic.NotationTLAPlus, "self.amount * 2",
	))

	query := helper.Must(model_state.NewQuery(queryKey, "get_total", "", nil, []model_logic.Logic{guaranteeLogic}, nil))

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

	guaranteeLogic := helper.Must(model_logic.NewLogic(
		helper.Must(identity.NewQueryGuaranteeKey(queryKey, "0")),
		model_logic.LogicTypeQuery, "Postcondition.", "result", model_logic.NotationTLAPlus, "self.amount",
	))

	query := helper.Must(model_state.NewQuery(queryKey, "get_total", "", nil, []model_logic.Logic{guaranteeLogic}, nil))

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

	requireLogic := helper.Must(model_logic.NewLogic(
		helper.Must(identity.NewQueryRequireKey(queryKey, "0")),
		model_logic.LogicTypeAssessment, "Precondition.", "", model_logic.NotationTLAPlus, "self.amount > 100",
	))
	guaranteeLogic := helper.Must(model_logic.NewLogic(
		helper.Must(identity.NewQueryGuaranteeKey(queryKey, "0")),
		model_logic.LogicTypeQuery, "Postcondition.", "result", model_logic.NotationTLAPlus, "self.amount",
	))

	query := helper.Must(model_state.NewQuery(queryKey, "get_total", "", []model_logic.Logic{requireLogic}, []model_logic.Logic{guaranteeLogic}, nil))

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
	guardKey := mustKey("domain/d/subdomain/s/class/order/guard/is_open")

	guardLogic := helper.Must(model_logic.NewLogic(
		guardKey,
		model_logic.LogicTypeAssessment, "Guard for open status and positive amount.", "", model_logic.NotationTLAPlus, "self.status = \"open\" /\\ self.amount > 0",
	))

	guard := helper.Must(model_state.NewGuard(guardKey, "is_open", guardLogic))

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
	guardKey := mustKey("domain/d/subdomain/s/class/order/guard/is_open")

	guardLogic := helper.Must(model_logic.NewLogic(
		guardKey,
		model_logic.LogicTypeAssessment, "Guard for open status and positive amount.", "", model_logic.NotationTLAPlus, "self.status = \"open\" /\\ self.amount > 0",
	))

	guard := helper.Must(model_state.NewGuard(guardKey, "is_open", guardLogic))

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

	event := helper.Must(model_state.NewEvent(eventCreateKey, "create", "", nil))

	class := helper.Must(model_class.NewClass(classKey, "Order", "", nil, nil, nil, ""))
	class.Attributes = map[identity.Key]model_class.Attribute{}
	class.States = map[identity.Key]model_state.State{
		stateOpenKey: {Key: stateOpenKey, Name: "Open"},
	}
	class.Events = map[identity.Key]model_state.Event{
		eventCreateKey: event,
	}
	class.Guards = map[identity.Key]model_state.Guard{}
	class.Actions = map[identity.Key]model_state.Action{}
	class.Queries = map[identity.Key]model_state.Query{}
	class.Transitions = map[identity.Key]model_state.Transition{
		transKey: {
			Key:          transKey,
			FromStateKey: nil, // Initial state -> creation
			EventKey:     eventCreateKey,
			ToStateKey:   &stateOpenKey,
		},
	}

	simState := state.NewSimulationState()
	exec, err := buildTestExecutor(simState, nil)
	s.NoError(err)

	eventObj := class.Events[eventCreateKey]

	result, err := exec.ExecuteTransition(class, eventObj, nil, nil, nil, nil)
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

	event := helper.Must(model_state.NewEvent(eventDeleteKey, "delete", "", nil))

	class := helper.Must(model_class.NewClass(classKey, "Order", "", nil, nil, nil, ""))
	class.Attributes = map[identity.Key]model_class.Attribute{}
	class.States = map[identity.Key]model_state.State{
		stateOpenKey: {Key: stateOpenKey, Name: "Open"},
	}
	class.Events = map[identity.Key]model_state.Event{
		eventDeleteKey: event,
	}
	class.Guards = map[identity.Key]model_state.Guard{}
	class.Actions = map[identity.Key]model_state.Action{}
	class.Queries = map[identity.Key]model_state.Query{}
	class.Transitions = map[identity.Key]model_state.Transition{
		transKey: {
			Key:          transKey,
			FromStateKey: &stateOpenKey,
			EventKey:     eventDeleteKey,
			ToStateKey:   nil, // Final state -> deletion
		},
	}

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("_state", object.NewString("Open"))
	instance := simState.CreateInstance(classKey, attrs)
	_ = simState.SetStateMachineState(instance.ID, stateOpenKey)

	exec, err := buildTestExecutor(simState, nil)
	s.NoError(err)

	eventObj := class.Events[eventDeleteKey]

	result, err := exec.ExecuteTransition(class, eventObj, instance, nil, nil, nil)
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

	guardHighLogic := helper.Must(model_logic.NewLogic(
		guardHighKey,
		model_logic.LogicTypeAssessment, "High value guard.", "", model_logic.NotationTLAPlus, "self.amount >= 100",
	))
	guardLowLogic := helper.Must(model_logic.NewLogic(
		guardLowKey,
		model_logic.LogicTypeAssessment, "Low value guard.", "", model_logic.NotationTLAPlus, "self.amount < 100",
	))

	guardHigh := helper.Must(model_state.NewGuard(guardHighKey, "high_value", guardHighLogic))
	guardLow := helper.Must(model_state.NewGuard(guardLowKey, "low_value", guardLowLogic))
	eventReview := helper.Must(model_state.NewEvent(eventReviewKey, "review", "", nil))

	class := helper.Must(model_class.NewClass(classKey, "Order", "", nil, nil, nil, ""))
	class.Attributes = map[identity.Key]model_class.Attribute{}
	class.States = map[identity.Key]model_state.State{
		stateOpenKey:     {Key: stateOpenKey, Name: "Open"},
		stateApprovedKey: {Key: stateApprovedKey, Name: "Approved"},
		stateRejectedKey: {Key: stateRejectedKey, Name: "Rejected"},
	}
	class.Events = map[identity.Key]model_state.Event{
		eventReviewKey: eventReview,
	}
	class.Guards = map[identity.Key]model_state.Guard{
		guardHighKey: guardHigh,
		guardLowKey:  guardLow,
	}
	class.Actions = map[identity.Key]model_state.Action{}
	class.Queries = map[identity.Key]model_state.Query{}
	class.Transitions = map[identity.Key]model_state.Transition{
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
	}

	simState := state.NewSimulationState()

	// Case 1: High value order -> should go to Approved
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

	// Case 2: Low value order -> should go to Rejected
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

	guardAlways1Logic := helper.Must(model_logic.NewLogic(guardAlwaysKey1, model_logic.LogicTypeAssessment, "Always true guard.", "", model_logic.NotationTLAPlus, "TRUE"))
	guardAlways2Logic := helper.Must(model_logic.NewLogic(guardAlwaysKey2, model_logic.LogicTypeAssessment, "Always true guard.", "", model_logic.NotationTLAPlus, "TRUE"))

	guardAlways1 := helper.Must(model_state.NewGuard(guardAlwaysKey1, "always1", guardAlways1Logic))
	guardAlways2 := helper.Must(model_state.NewGuard(guardAlwaysKey2, "always2", guardAlways2Logic))
	eventGo := helper.Must(model_state.NewEvent(eventKey, "go", "", nil))

	class := helper.Must(model_class.NewClass(classKey, "Order", "", nil, nil, nil, ""))
	class.Attributes = map[identity.Key]model_class.Attribute{}
	class.States = map[identity.Key]model_state.State{
		stateOpenKey: {Key: stateOpenKey, Name: "Open"},
		stateAKey:    {Key: stateAKey, Name: "A"},
		stateBKey:    {Key: stateBKey, Name: "B"},
	}
	class.Events = map[identity.Key]model_state.Event{
		eventKey: eventGo,
	}
	class.Guards = map[identity.Key]model_state.Guard{
		guardAlwaysKey1: guardAlways1,
		guardAlwaysKey2: guardAlways2,
	}
	class.Actions = map[identity.Key]model_state.Action{}
	class.Queries = map[identity.Key]model_state.Query{}
	class.Transitions = map[identity.Key]model_state.Transition{
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

	guardNeverLogic := helper.Must(model_logic.NewLogic(guardNeverKey, model_logic.LogicTypeAssessment, "Never true guard.", "", model_logic.NotationTLAPlus, "FALSE"))
	guardNever := helper.Must(model_state.NewGuard(guardNeverKey, "never", guardNeverLogic))
	eventGo := helper.Must(model_state.NewEvent(eventKey, "go", "", nil))

	class := helper.Must(model_class.NewClass(classKey, "Order", "", nil, nil, nil, ""))
	class.Attributes = map[identity.Key]model_class.Attribute{}
	class.States = map[identity.Key]model_state.State{
		stateOpenKey: {Key: stateOpenKey, Name: "Open"},
		stateAKey:    {Key: stateAKey, Name: "A"},
	}
	class.Events = map[identity.Key]model_state.Event{
		eventKey: eventGo,
	}
	class.Guards = map[identity.Key]model_state.Guard{
		guardNeverKey: guardNever,
	}
	class.Actions = map[identity.Key]model_state.Action{}
	class.Queries = map[identity.Key]model_state.Query{}
	class.Transitions = map[identity.Key]model_state.Transition{
		transKey: {
			Key:          transKey,
			FromStateKey: &stateOpenKey,
			EventKey:     eventKey,
			GuardKey:     &guardNeverKey,
			ToStateKey:   &stateAKey,
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
	class := helper.Must(model_class.NewClass(mustKey("domain/d/subdomain/s/class/empty"), "Empty", "", nil, nil, nil, ""))
	class.States = map[identity.Key]model_state.State{}

	err := ValidateClassForSimulation(class)
	s.Error(err)
	s.Contains(err.Error(), "no states")
}

func (s *ActionsSuite) TestValidateClassForSimulationWithStates() {
	stateKey := mustKey("domain/d/subdomain/s/class/c/state/s1")

	class := helper.Must(model_class.NewClass(mustKey("domain/d/subdomain/s/class/c"), "C", "", nil, nil, nil, ""))
	class.States = map[identity.Key]model_state.State{
		stateKey: {Key: stateKey, Name: "S1"},
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

	class := helper.Must(model_class.NewClass(mustKey("domain/d/subdomain/s/class/c"), "C", "", nil, nil, nil, ""))
	class.States = map[identity.Key]model_state.State{
		stateOpenKey:   {Key: stateOpenKey, Name: "Open"},
		stateClosedKey: {Key: stateClosedKey, Name: "Closed"},
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
				CollectionType: "atomic",
				Atomic: &model_data_type.Atomic{
					ConstraintType: "span",
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
				CollectionType: "atomic",
				Atomic: &model_data_type.Atomic{
					ConstraintType: "enumeration",
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

// TestActionRejectsGuaranteesNonPrimed is no longer applicable because
// NewLogic now requires a non-empty Target for LogicTypeStateChange,
// making it impossible to construct a state_change Logic without a target.
// The executor's legacy ClassifyGuarantee path is unreachable for state_change logics.

func (s *ActionsSuite) TestActionRejectsRequiresWithPrime() {
	classKey := mustKey("domain/d/subdomain/s/class/c")
	actionKey := mustKey("domain/d/subdomain/s/class/c/action/a")

	requireLogic := helper.Must(model_logic.NewLogic(
		helper.Must(identity.NewActionRequireKey(actionKey, "0")),
		model_logic.LogicTypeAssessment, "Precondition.", "", model_logic.NotationTLAPlus, "self.count' > 0",
	))

	action := helper.Must(model_state.NewAction(actionKey, "BadRequires", "", []model_logic.Logic{requireLogic}, nil, nil, nil))

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

	safetyLogic := helper.Must(model_logic.NewLogic(
		helper.Must(identity.NewActionSafetyKey(actionKey, "0")),
		model_logic.LogicTypeSafetyRule, "Safety rule.", "", model_logic.NotationTLAPlus, "self.count > 0",
	))

	action := helper.Must(model_state.NewAction(actionKey, "BadSafety", "", nil, nil, []model_logic.Logic{safetyLogic}, nil))

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

	guaranteeLogic := helper.Must(model_logic.NewLogic(
		helper.Must(identity.NewActionGuaranteeKey(actionKey, "0")),
		model_logic.LogicTypeStateChange, "Postcondition.", "count", model_logic.NotationTLAPlus, "self.count + 1",
	))
	safetyLogic := helper.Must(model_logic.NewLogic(
		helper.Must(identity.NewActionSafetyKey(actionKey, "0")),
		model_logic.LogicTypeSafetyRule, "Safety rule.", "", model_logic.NotationTLAPlus, "self.count' >= 1",
	))

	action := helper.Must(model_state.NewAction(actionKey, "GoodAction", "", nil, []model_logic.Logic{guaranteeLogic}, []model_logic.Logic{safetyLogic}, nil))

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

	guaranteeLogic := helper.Must(model_logic.NewLogic(
		helper.Must(identity.NewActionGuaranteeKey(actionKey, "0")),
		model_logic.LogicTypeStateChange, "Postcondition.", "count", model_logic.NotationTLAPlus, "self.count + 1",
	))
	safetyLogic := helper.Must(model_logic.NewLogic(
		helper.Must(identity.NewActionSafetyKey(actionKey, "0")),
		model_logic.LogicTypeSafetyRule, "Safety rule.", "", model_logic.NotationTLAPlus, "self.count' < 0",
	))

	action := helper.Must(model_state.NewAction(actionKey, "ViolatingAction", "", nil, []model_logic.Logic{guaranteeLogic}, []model_logic.Logic{safetyLogic}, nil))

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
	guardKey := mustKey("domain/d/subdomain/s/class/c/guard/g")

	guardLogic := helper.Must(model_logic.NewLogic(
		guardKey,
		model_logic.LogicTypeAssessment, "Guard with primed variable.", "", model_logic.NotationTLAPlus, "self.count' > 0",
	))

	guard := helper.Must(model_state.NewGuard(guardKey, "BadGuard", guardLogic))

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

	requireLogic := helper.Must(model_logic.NewLogic(
		helper.Must(identity.NewQueryRequireKey(queryKey, "0")),
		model_logic.LogicTypeAssessment, "Precondition.", "", model_logic.NotationTLAPlus, "self.count' > 0",
	))
	guaranteeLogic := helper.Must(model_logic.NewLogic(
		helper.Must(identity.NewQueryGuaranteeKey(queryKey, "0")),
		model_logic.LogicTypeQuery, "Postcondition.", "result", model_logic.NotationTLAPlus, "result' = self.count",
	))

	query := helper.Must(model_state.NewQuery(queryKey, "BadQuery", "", []model_logic.Logic{requireLogic}, []model_logic.Logic{guaranteeLogic}, nil))

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
