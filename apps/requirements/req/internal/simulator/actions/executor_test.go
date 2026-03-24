package actions

import (
	"math/rand"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
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

func buildTestExecutor(simState *state.SimulationState) *ActionExecutor {
	bb := state.NewBindingsBuilder(simState)
	ge := NewGuardEvaluator(bb)

	return NewActionExecutor(bb, nil, nil, nil, ge, nil)
}

// parsedSpec creates a TLA+ ExpressionSpec with the expression parsed via the convert pipeline.
// Uses nil LowerContext — suitable for context-free expressions (literals, arithmetic on params).
func parsedSpec(tla string) logic_spec.ExpressionSpec {
	pf := convert.NewExpressionParseFunc(nil)
	spec := helper.Must(logic_spec.NewExpressionSpec("tla_plus", tla, pf))
	return spec
}

// parsedSpecCtx creates a TLA+ ExpressionSpec parsed with a LowerContext built from the given
// class key, attribute names, and optional parameter names.
func parsedSpecCtx(tla string, classKey identity.Key, attrNames []string, paramNames []string) logic_spec.ExpressionSpec {
	ctx := &convert.LowerContext{
		ClassKey:       classKey,
		AttributeNames: make(map[string]identity.Key),
	}
	for _, name := range attrNames {
		ctx.AttributeNames[name] = helper.Must(identity.NewAttributeKey(classKey, name))
	}
	if len(paramNames) > 0 {
		ctx.Parameters = make(map[string]bool)
		for _, name := range paramNames {
			ctx.Parameters[name] = true
		}
	}
	pf := convert.NewExpressionParseFunc(ctx)
	spec := helper.Must(logic_spec.NewExpressionSpec("tla_plus", tla, pf))
	return spec
}

// orderSpec parses a TLA+ expression in the context of the standard Order class
// with attributes: amount, status.
func orderSpec(tla string) logic_spec.ExpressionSpec {
	return parsedSpecCtx(tla, mustKey("domain/d/subdomain/s/class/order"), []string{"amount", "status"}, nil)
}

// orderSpecWithParams parses a TLA+ expression in the Order class context with parameters.
func orderSpecWithParams(tla string, params []string) logic_spec.ExpressionSpec {
	return parsedSpecCtx(tla, mustKey("domain/d/subdomain/s/class/order"), []string{"amount", "status"}, params)
}

// counterSpec parses a TLA+ expression in the context of a class with attribute: count.
func counterSpec(tla string) logic_spec.ExpressionSpec {
	return parsedSpecCtx(tla, mustKey("domain/d/subdomain/s/class/c"), []string{"count"}, nil)
}

// lowerAction returns the action as-is since expressions are now parsed at construction time.
func lowerAction(action model_state.Action, _ identity.Key) model_state.Action {
	return action
}

// lowerQuery returns the query as-is since expressions are now parsed at construction time.
func lowerQuery(query model_state.Query, _ identity.Key) model_state.Query {
	return query
}

// lowerGuard returns the guard as-is since expressions are now parsed at construction time.
func lowerGuard(guard model_state.Guard, _ identity.Key) model_state.Guard {
	return guard
}

// lowerClass returns the class as-is since expressions are now parsed at construction time.
func lowerClass(class model_class.Class, _ identity.Key) model_class.Class {
	return class
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

	guaranteeLogic := model_logic.NewLogic(
		helper.Must(identity.NewActionGuaranteeKey(actionCloseKey, "0")),
		model_logic.LogicTypeStateChange, "Postcondition.", "amount", orderSpec("self.amount + 10"),
		nil,
	)

	action := model_state.NewAction(actionCloseKey, "DoClose", "", nil, []model_logic.Logic{guaranteeLogic}, nil, nil)
	event := model_state.NewEvent(eventCloseKey, "close", "", nil)

	transition := model_state.NewTransition(transKey, &stateOpenKey, eventCloseKey, nil, &actionCloseKey, &stateClosedKey, "")

	class := model_class.NewClass(classKey, "Order", "", nil, nil, nil, "")
	class.SetAttributes(map[identity.Key]model_class.Attribute{})
	class.SetStates(map[identity.Key]model_state.State{
		stateOpenKey:   model_state.NewState(stateOpenKey, "Open", "", ""),
		stateClosedKey: model_state.NewState(stateClosedKey, "Closed", "", ""),
	})
	class.SetEvents(map[identity.Key]model_state.Event{
		eventCloseKey: event,
	})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{
		actionCloseKey: action,
	})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{
		transKey: transition,
	})

	return class, classKey
}

// ========================================================================
// ExecutionContext tests
// ========================================================================

func (s *ActionsSuite) TestExecutionContextRecordPrimed() {
	ctx := NewExecutionContext()

	err := ctx.RecordPrimedAssignment(1, "count", object.NewInteger(42))
	s.Require().NoError(err)

	all := ctx.GetAllPrimedAssignments()
	s.Len(all, 1)
	s.Equal("42", all[1]["count"].Inspect())
}

func (s *ActionsSuite) TestExecutionContextRejectsStateField() {
	ctx := NewExecutionContext()

	err := ctx.RecordPrimedAssignment(1, "_state", object.NewString("Open"))
	s.Require().Error(err)
	s.Contains(err.Error(), "_state")
}

func (s *ActionsSuite) TestExecutionContextReentrancyGuard() {
	ctx := NewExecutionContext()

	// First mutation is fine
	s.True(ctx.CanMutate(1))
	err := ctx.RecordPrimedAssignment(1, "count", object.NewInteger(1))
	s.Require().NoError(err)

	// After mutation, instance 1 is locked
	s.False(ctx.CanMutate(1))

	// Instance 2 is still available
	s.True(ctx.CanMutate(2))
}

func (s *ActionsSuite) TestExecutionContextDepthLimit() {
	ctx := NewExecutionContext()

	for range 100 {
		err := ctx.IncrementDepth()
		s.Require().NoError(err)
	}

	// 101st should fail
	err := ctx.IncrementDepth()
	s.Require().Error(err)
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

	guaranteeLogic := model_logic.NewLogic(
		helper.Must(identity.NewActionGuaranteeKey(actionKey, "0")),
		model_logic.LogicTypeStateChange, "Postcondition.", "count", counterSpec("self.count + 1"),
		nil,
	)

	action := model_state.NewAction(actionKey, "increment", "", nil, []model_logic.Logic{guaranteeLogic}, nil, nil)
	action = lowerAction(action, classKey)

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("count", object.NewInteger(10))
	instance := simState.CreateInstance(classKey, attrs)

	exec := buildTestExecutor(simState)

	result, err := exec.ExecuteAction(action, instance, nil)
	s.Require().NoError(err)
	s.NotNil(result)
	s.True(result.Success)

	// Verify state was updated
	updated := simState.GetInstance(instance.ID)
	s.Equal("11", updated.GetAttribute("count").Inspect())
}

func (s *ActionsSuite) TestExecuteActionPreconditionPasses() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	actionKey := mustKey("domain/d/subdomain/s/class/order/action/close")

	requireLogic := model_logic.NewLogic(
		helper.Must(identity.NewActionRequireKey(actionKey, "0")),
		model_logic.LogicTypeAssessment, "Precondition.", "", orderSpec("self.status = \"open\""),
		nil,
	)
	guaranteeLogic := model_logic.NewLogic(
		helper.Must(identity.NewActionGuaranteeKey(actionKey, "0")),
		model_logic.LogicTypeStateChange, "Postcondition.", "status", parsedSpec("\"closed\""),
		nil,
	)

	action := model_state.NewAction(actionKey, "close", "", []model_logic.Logic{requireLogic}, []model_logic.Logic{guaranteeLogic}, nil, nil)
	action = lowerAction(action, classKey)

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("status", object.NewString("open"))
	instance := simState.CreateInstance(classKey, attrs)

	exec := buildTestExecutor(simState)

	result, err := exec.ExecuteAction(action, instance, nil)
	s.Require().NoError(err)
	s.True(result.Success)

	updated := simState.GetInstance(instance.ID)
	s.Equal("closed", updated.GetAttribute("status").(*object.String).Value())
}

func (s *ActionsSuite) TestExecuteActionPreconditionFails() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	actionKey := mustKey("domain/d/subdomain/s/class/order/action/close")

	requireLogic := model_logic.NewLogic(
		helper.Must(identity.NewActionRequireKey(actionKey, "0")),
		model_logic.LogicTypeAssessment, "Precondition.", "", orderSpec("self.status = \"open\""),
		nil,
	)
	guaranteeLogic := model_logic.NewLogic(
		helper.Must(identity.NewActionGuaranteeKey(actionKey, "0")),
		model_logic.LogicTypeStateChange, "Postcondition.", "status", parsedSpec("\"closed\""),
		nil,
	)

	action := model_state.NewAction(actionKey, "close", "", []model_logic.Logic{requireLogic}, []model_logic.Logic{guaranteeLogic}, nil, nil)
	action = lowerAction(action, classKey)

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("status", object.NewString("closed")) // already closed
	instance := simState.CreateInstance(classKey, attrs)

	exec := buildTestExecutor(simState)

	_, err := exec.ExecuteAction(action, instance, nil)
	s.Require().Error(err)
	s.Contains(err.Error(), "precondition failed")
}

func (s *ActionsSuite) TestExecuteActionWithParameters() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	actionKey := mustKey("domain/d/subdomain/s/class/order/action/set_amount")

	guaranteeLogic := model_logic.NewLogic(
		helper.Must(identity.NewActionGuaranteeKey(actionKey, "0")),
		model_logic.LogicTypeStateChange, "Postcondition.", "amount", orderSpecWithParams("amount", []string{"amount"}),
		nil,
	)

	actionParams := []model_state.Parameter{helper.Must(model_state.NewParameter("amount", "[0,1000]"))}
	action := model_state.NewAction(actionKey, "set_amount", "", nil, []model_logic.Logic{guaranteeLogic}, nil, actionParams)
	action = lowerAction(action, classKey)

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("amount", object.NewInteger(0))
	instance := simState.CreateInstance(classKey, attrs)

	exec := buildTestExecutor(simState)

	params := map[string]object.Object{
		"amount": object.NewInteger(500),
	}

	result, err := exec.ExecuteAction(action, instance, params)
	s.Require().NoError(err)
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

	guaranteeLogic := model_logic.NewLogic(
		helper.Must(identity.NewQueryGuaranteeKey(queryKey, "0")),
		model_logic.LogicTypeQuery, "Postcondition.", "result", orderSpec("self.amount * 2"),
		nil,
	)

	query := model_state.NewQuery(queryKey, "get_total", "", nil, []model_logic.Logic{guaranteeLogic}, nil)
	query = lowerQuery(query, classKey)

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("amount", object.NewInteger(50))
	instance := simState.CreateInstance(classKey, attrs)

	exec := buildTestExecutor(simState)

	result, err := exec.ExecuteQuery(query, instance, nil)
	s.Require().NoError(err)
	s.True(result.Success)
	s.NotNil(result.Outputs["result"])
	s.Equal("100", result.Outputs["result"].Inspect())
}

func (s *ActionsSuite) TestExecuteQueryDoesNotModifyState() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	queryKey := mustKey("domain/d/subdomain/s/class/order/query/get_total")

	guaranteeLogic := model_logic.NewLogic(
		helper.Must(identity.NewQueryGuaranteeKey(queryKey, "0")),
		model_logic.LogicTypeQuery, "Postcondition.", "result", orderSpec("self.amount"),
		nil,
	)

	query := model_state.NewQuery(queryKey, "get_total", "", nil, []model_logic.Logic{guaranteeLogic}, nil)
	query = lowerQuery(query, classKey)

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("amount", object.NewInteger(50))
	instance := simState.CreateInstance(classKey, attrs)

	exec := buildTestExecutor(simState)

	_, err := exec.ExecuteQuery(query, instance, nil)
	s.Require().NoError(err)

	// State should be unchanged
	unchanged := simState.GetInstance(instance.ID)
	s.Equal("50", unchanged.GetAttribute("amount").Inspect())
}

func (s *ActionsSuite) TestExecuteQueryPreconditionFails() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	queryKey := mustKey("domain/d/subdomain/s/class/order/query/get_total")

	requireLogic := model_logic.NewLogic(
		helper.Must(identity.NewQueryRequireKey(queryKey, "0")),
		model_logic.LogicTypeAssessment, "Precondition.", "", orderSpec("self.amount > 100"),
		nil,
	)
	guaranteeLogic := model_logic.NewLogic(
		helper.Must(identity.NewQueryGuaranteeKey(queryKey, "0")),
		model_logic.LogicTypeQuery, "Postcondition.", "result", orderSpec("self.amount"),
		nil,
	)

	query := model_state.NewQuery(queryKey, "get_total", "", []model_logic.Logic{requireLogic}, []model_logic.Logic{guaranteeLogic}, nil)
	query = lowerQuery(query, classKey)

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("amount", object.NewInteger(50))
	instance := simState.CreateInstance(classKey, attrs)

	exec := buildTestExecutor(simState)

	_, err := exec.ExecuteQuery(query, instance, nil)
	s.Require().Error(err)
	s.Contains(err.Error(), "precondition failed")
}

// ========================================================================
// GuardEvaluator tests
// ========================================================================

func (s *ActionsSuite) TestGuardEvaluatorAllTrue() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	guardKey := mustKey("domain/d/subdomain/s/class/order/guard/is_open")

	guardLogic := model_logic.NewLogic(
		guardKey,
		model_logic.LogicTypeAssessment, "Guard for open status and positive amount.", "", orderSpec("self.status = \"open\" /\\ self.amount > 0"),
		nil,
	)

	guard := model_state.NewGuard(guardKey, "is_open", guardLogic)
	guard = lowerGuard(guard, classKey)

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("status", object.NewString("open"))
	attrs.Set("amount", object.NewInteger(100))
	instance := simState.CreateInstance(classKey, attrs)

	bb := state.NewBindingsBuilder(simState)
	ge := NewGuardEvaluator(bb)

	passes, err := ge.EvaluateGuard(guard, instance)
	s.Require().NoError(err)
	s.True(passes)
}

func (s *ActionsSuite) TestGuardEvaluatorOneFalse() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	guardKey := mustKey("domain/d/subdomain/s/class/order/guard/is_open")

	guardLogic := model_logic.NewLogic(
		guardKey,
		model_logic.LogicTypeAssessment, "Guard for open status and positive amount.", "", orderSpec("self.status = \"open\" /\\ self.amount > 0"),
		nil,
	)

	guard := model_state.NewGuard(guardKey, "is_open", guardLogic)
	guard = lowerGuard(guard, classKey)

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("status", object.NewString("closed")) // guard fails here
	attrs.Set("amount", object.NewInteger(100))
	instance := simState.CreateInstance(classKey, attrs)

	bb := state.NewBindingsBuilder(simState)
	ge := NewGuardEvaluator(bb)

	passes, err := ge.EvaluateGuard(guard, instance)
	s.Require().NoError(err)
	s.False(passes)
}

// ========================================================================
// ExecuteTransition tests
// ========================================================================

func (s *ActionsSuite) TestExecuteTransitionNormal() {
	class, classKey := testOrderClass()
	class = lowerClass(class, classKey)

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("amount", object.NewInteger(100))
	attrs.Set("_state", object.NewString("Open"))
	instance := simState.CreateInstance(classKey, attrs)

	// Set state machine state
	stateOpenKey := mustKey("domain/d/subdomain/s/class/order/state/open")
	_ = simState.SetStateMachineState(instance.ID, stateOpenKey)

	exec := buildTestExecutor(simState)

	eventCloseKey := mustKey("domain/d/subdomain/s/class/order/event/close")
	event := class.Events[eventCloseKey]

	result, err := exec.ExecuteTransition(class, event, instance, nil, nil, nil)
	s.Require().NoError(err)
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

	event := model_state.NewEvent(eventCreateKey, "create", "", nil)

	transition := model_state.NewTransition(transKey, nil, eventCreateKey, nil, nil, &stateOpenKey, "")

	class := model_class.NewClass(classKey, "Order", "", nil, nil, nil, "")
	class.SetAttributes(map[identity.Key]model_class.Attribute{})
	class.SetStates(map[identity.Key]model_state.State{
		stateOpenKey: model_state.NewState(stateOpenKey, "Open", "", ""),
	})
	class.SetEvents(map[identity.Key]model_state.Event{
		eventCreateKey: event,
	})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{
		transKey: transition,
	})

	simState := state.NewSimulationState()
	exec := buildTestExecutor(simState)

	eventObj := class.Events[eventCreateKey]

	result, err := exec.ExecuteTransition(class, eventObj, nil, nil, nil, nil)
	s.Require().NoError(err)
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

	event := model_state.NewEvent(eventDeleteKey, "delete", "", nil)

	transition := model_state.NewTransition(transKey, &stateOpenKey, eventDeleteKey, nil, nil, nil, "")

	class := model_class.NewClass(classKey, "Order", "", nil, nil, nil, "")
	class.SetAttributes(map[identity.Key]model_class.Attribute{})
	class.SetStates(map[identity.Key]model_state.State{
		stateOpenKey: model_state.NewState(stateOpenKey, "Open", "", ""),
	})
	class.SetEvents(map[identity.Key]model_state.Event{
		eventDeleteKey: event,
	})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{
		transKey: transition,
	})

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("_state", object.NewString("Open"))
	instance := simState.CreateInstance(classKey, attrs)
	_ = simState.SetStateMachineState(instance.ID, stateOpenKey)

	exec := buildTestExecutor(simState)

	eventObj := class.Events[eventDeleteKey]

	result, err := exec.ExecuteTransition(class, eventObj, instance, nil, nil, nil)
	s.Require().NoError(err)
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

	exec := buildTestExecutor(simState)

	eventCloseKey := mustKey("domain/d/subdomain/s/class/order/event/close")
	event := class.Events[eventCloseKey]

	_, err := exec.ExecuteTransition(class, event, instance, nil, nil, nil)
	s.Require().Error(err)
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

	guardHighLogic := model_logic.NewLogic(
		guardHighKey,
		model_logic.LogicTypeAssessment, "High value guard.", "", orderSpec("self.amount >= 100"),
		nil,
	)
	guardLowLogic := model_logic.NewLogic(
		guardLowKey,
		model_logic.LogicTypeAssessment, "Low value guard.", "", orderSpec("self.amount < 100"),
		nil,
	)

	guardHigh := model_state.NewGuard(guardHighKey, "high_value", guardHighLogic)
	guardLow := model_state.NewGuard(guardLowKey, "low_value", guardLowLogic)
	eventReview := model_state.NewEvent(eventReviewKey, "review", "", nil)

	transApprove := model_state.NewTransition(transApproveKey, &stateOpenKey, eventReviewKey, &guardHighKey, nil, &stateApprovedKey, "")
	transReject := model_state.NewTransition(transRejectKey, &stateOpenKey, eventReviewKey, &guardLowKey, nil, &stateRejectedKey, "")

	class := model_class.NewClass(classKey, "Order", "", nil, nil, nil, "")
	class.SetAttributes(map[identity.Key]model_class.Attribute{})
	class.SetStates(map[identity.Key]model_state.State{
		stateOpenKey:     model_state.NewState(stateOpenKey, "Open", "", ""),
		stateApprovedKey: model_state.NewState(stateApprovedKey, "Approved", "", ""),
		stateRejectedKey: model_state.NewState(stateRejectedKey, "Rejected", "", ""),
	})
	class.SetEvents(map[identity.Key]model_state.Event{
		eventReviewKey: eventReview,
	})
	class.SetGuards(map[identity.Key]model_state.Guard{
		guardHighKey: guardHigh,
		guardLowKey:  guardLow,
	})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{
		transApproveKey: transApprove,
		transRejectKey:  transReject,
	})
	class = lowerClass(class, classKey)

	simState := state.NewSimulationState()

	// Case 1: High value order -> should go to Approved
	attrs := object.NewRecord()
	attrs.Set("amount", object.NewInteger(200))
	attrs.Set("_state", object.NewString("Open"))
	instance := simState.CreateInstance(classKey, attrs)
	_ = simState.SetStateMachineState(instance.ID, stateOpenKey)

	exec := buildTestExecutor(simState)

	event := class.Events[eventReviewKey]
	result, err := exec.ExecuteTransition(class, event, instance, nil, nil, nil)
	s.Require().NoError(err)
	s.Equal("Approved", result.ToState)

	// Case 2: Low value order -> should go to Rejected
	attrs2 := object.NewRecord()
	attrs2.Set("amount", object.NewInteger(50))
	attrs2.Set("_state", object.NewString("Open"))
	instance2 := simState.CreateInstance(classKey, attrs2)
	_ = simState.SetStateMachineState(instance2.ID, stateOpenKey)

	result2, err := exec.ExecuteTransition(class, event, instance2, nil, nil, nil)
	s.Require().NoError(err)
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

	guardAlways1Logic := model_logic.NewLogic(guardAlwaysKey1, model_logic.LogicTypeAssessment, "Always true guard.", "", parsedSpec("TRUE"), nil)
	guardAlways2Logic := model_logic.NewLogic(guardAlwaysKey2, model_logic.LogicTypeAssessment, "Always true guard.", "", parsedSpec("TRUE"), nil)

	guardAlways1 := model_state.NewGuard(guardAlwaysKey1, "always1", guardAlways1Logic)
	guardAlways2 := model_state.NewGuard(guardAlwaysKey2, "always2", guardAlways2Logic)
	eventGo := model_state.NewEvent(eventKey, "go", "", nil)

	trans1 := model_state.NewTransition(trans1Key, &stateOpenKey, eventKey, &guardAlwaysKey1, nil, &stateAKey, "")
	trans2 := model_state.NewTransition(trans2Key, &stateOpenKey, eventKey, &guardAlwaysKey2, nil, &stateBKey, "")

	class := model_class.NewClass(classKey, "Order", "", nil, nil, nil, "")
	class.SetAttributes(map[identity.Key]model_class.Attribute{})
	class.SetStates(map[identity.Key]model_state.State{
		stateOpenKey: model_state.NewState(stateOpenKey, "Open", "", ""),
		stateAKey:    model_state.NewState(stateAKey, "A", "", ""),
		stateBKey:    model_state.NewState(stateBKey, "B", "", ""),
	})
	class.SetEvents(map[identity.Key]model_state.Event{
		eventKey: eventGo,
	})
	class.SetGuards(map[identity.Key]model_state.Guard{
		guardAlwaysKey1: guardAlways1,
		guardAlwaysKey2: guardAlways2,
	})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{
		trans1Key: trans1,
		trans2Key: trans2,
	})
	class = lowerClass(class, classKey)

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("_state", object.NewString("Open"))
	instance := simState.CreateInstance(classKey, attrs)
	_ = simState.SetStateMachineState(instance.ID, stateOpenKey)

	exec := buildTestExecutor(simState)

	event := class.Events[eventKey]
	_, err := exec.ExecuteTransition(class, event, instance, nil, nil, nil)
	s.Require().Error(err)
	s.Contains(err.Error(), "non-determinism")
}

func (s *ActionsSuite) TestTransitionNoGuardsTrue() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	stateOpenKey := mustKey("domain/d/subdomain/s/class/order/state/open")
	stateAKey := mustKey("domain/d/subdomain/s/class/order/state/a")
	eventKey := mustKey("domain/d/subdomain/s/class/order/event/go")
	guardNeverKey := mustKey("domain/d/subdomain/s/class/order/guard/never")
	transKey := mustKey("domain/d/subdomain/s/class/order/transition/t1")

	guardNeverLogic := model_logic.NewLogic(guardNeverKey, model_logic.LogicTypeAssessment, "Never true guard.", "", parsedSpec("FALSE"), nil)
	guardNever := model_state.NewGuard(guardNeverKey, "never", guardNeverLogic)
	eventGo := model_state.NewEvent(eventKey, "go", "", nil)

	trans := model_state.NewTransition(transKey, &stateOpenKey, eventKey, &guardNeverKey, nil, &stateAKey, "")

	class := model_class.NewClass(classKey, "Order", "", nil, nil, nil, "")
	class.SetAttributes(map[identity.Key]model_class.Attribute{})
	class.SetStates(map[identity.Key]model_state.State{
		stateOpenKey: model_state.NewState(stateOpenKey, "Open", "", ""),
		stateAKey:    model_state.NewState(stateAKey, "A", "", ""),
	})
	class.SetEvents(map[identity.Key]model_state.Event{
		eventKey: eventGo,
	})
	class.SetGuards(map[identity.Key]model_state.Guard{
		guardNeverKey: guardNever,
	})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{
		transKey: trans,
	})
	class = lowerClass(class, classKey)

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("_state", object.NewString("Open"))
	instance := simState.CreateInstance(classKey, attrs)
	_ = simState.SetStateMachineState(instance.ID, stateOpenKey)

	exec := buildTestExecutor(simState)

	event := class.Events[eventKey]
	_, err := exec.ExecuteTransition(class, event, instance, nil, nil, nil)
	s.Require().Error(err)
	s.Contains(err.Error(), "deadlock")
}

// ========================================================================
// ValidateClassForSimulation test
// ========================================================================

func (s *ActionsSuite) TestValidateClassForSimulationNoStates() {
	class := model_class.NewClass(mustKey("domain/d/subdomain/s/class/empty"), "Empty", "", nil, nil, nil, "")
	class.SetStates(map[identity.Key]model_state.State{})

	err := ValidateClassForSimulation(class)
	s.Require().Error(err)
	s.Contains(err.Error(), "no states")
}

func (s *ActionsSuite) TestValidateClassForSimulationWithStates() {
	stateKey := mustKey("domain/d/subdomain/s/class/c/state/s1")

	class := model_class.NewClass(mustKey("domain/d/subdomain/s/class/c"), "C", "", nil, nil, nil, "")
	class.SetStates(map[identity.Key]model_state.State{
		stateKey: model_state.NewState(stateKey, "S1", "", ""),
	})

	err := ValidateClassForSimulation(class)
	s.Require().NoError(err)
}

// ========================================================================
// GetStateEnumValues test
// ========================================================================

func (s *ActionsSuite) TestGetStateEnumValues() {
	stateOpenKey := mustKey("domain/d/subdomain/s/class/c/state/open")
	stateClosedKey := mustKey("domain/d/subdomain/s/class/c/state/closed")

	class := model_class.NewClass(mustKey("domain/d/subdomain/s/class/c"), "C", "", nil, nil, nil, "")
	class.SetStates(map[identity.Key]model_state.State{
		stateOpenKey:   model_state.NewState(stateOpenKey, "Open", "", ""),
		stateClosedKey: model_state.NewState(stateClosedKey, "Closed", "", ""),
	})

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
		helper.Must(model_state.NewParameter("amount", "[0,100]")),
		helper.Must(model_state.NewParameter("name", "string")),
	}

	values := map[string]object.Object{
		"amount": object.NewInteger(50),
		"name":   object.NewString("test"),
	}

	result, err := binder.BindParameters(paramDefs, values)
	s.Require().NoError(err)
	s.Len(result, 2)
	s.Equal("50", result["amount"].Inspect())
	s.Equal("test", result["name"].(*object.String).Value())
}

func (s *ActionsSuite) TestBindParametersMissing() {
	binder := NewParameterBinder()

	paramDefs := []model_state.Parameter{
		helper.Must(model_state.NewParameter("amount", "[0,100]")),
	}

	values := map[string]object.Object{} // missing amount

	_, err := binder.BindParameters(paramDefs, values)
	s.Require().Error(err)
	s.Contains(err.Error(), "missing required parameter")
}

func (s *ActionsSuite) TestGenerateRandomParametersSpan() {
	binder := NewParameterBinder()
	rng := rand.New(rand.NewSource(42)) //nolint:gosec // deterministic seed intentional for test reproducibility

	lowerValue := 10
	higherValue := 20

	countParam := helper.Must(model_state.NewParameter("count", "[10, 20]"))
	countParam.DataType = &model_data_type.DataType{
		Key:            "count",
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
	}

	paramDefs := []model_state.Parameter{countParam}

	for range 100 {
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
	rng := rand.New(rand.NewSource(42)) //nolint:gosec // deterministic seed intentional for test reproducibility

	colorParam := helper.Must(model_state.NewParameter("color", "{red, green, blue}"))
	colorParam.DataType = &model_data_type.DataType{
		Key:            "color",
		CollectionType: "atomic",
		Atomic: &model_data_type.Atomic{
			ConstraintType: "enumeration",
			Enums: []model_data_type.AtomicEnum{
				{Value: "red", SortOrder: 0},
				{Value: "green", SortOrder: 1},
				{Value: "blue", SortOrder: 2},
			},
		},
	}

	paramDefs := []model_state.Parameter{colorParam}

	allowedValues := map[string]bool{"red": true, "green": true, "blue": true}

	for range 100 {
		result := binder.GenerateRandomParameters(paramDefs, rng)
		str, ok := result["color"].(*object.String)
		s.True(ok)
		s.True(allowedValues[str.Value()], "Generated value %s should be in {red, green, blue}", str.Value())
	}
}

func (s *ActionsSuite) TestGenerateRandomParametersNoType() {
	binder := NewParameterBinder()
	rng := rand.New(rand.NewSource(42)) //nolint:gosec // deterministic seed intentional for test reproducibility

	paramDefs := []model_state.Parameter{
		helper.Must(model_state.NewParameter("x", "unknown")),
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

	requireLogic := model_logic.NewLogic(
		helper.Must(identity.NewActionRequireKey(actionKey, "0")),
		model_logic.LogicTypeAssessment, "Precondition.", "", counterSpec("self.count' > 0"),
		nil,
	)

	action := model_state.NewAction(actionKey, "BadRequires", "", []model_logic.Logic{requireLogic}, nil, nil, nil)
	action = lowerAction(action, classKey)

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("count", object.NewInteger(5))
	instance := simState.CreateInstance(classKey, attrs)

	executor := buildTestExecutor(simState)

	_, err := executor.ExecuteAction(action, instance, nil)
	s.Require().Error(err)
	s.Contains(err.Error(), "Requires must not contain primed variables")
}

func (s *ActionsSuite) TestActionSafetyRulesMustHavePrime() {
	classKey := mustKey("domain/d/subdomain/s/class/c")
	actionKey := mustKey("domain/d/subdomain/s/class/c/action/a")

	safetyLogic := model_logic.NewLogic(
		helper.Must(identity.NewActionSafetyKey(actionKey, "0")),
		model_logic.LogicTypeSafetyRule, "Safety rule.", "", counterSpec("self.count > 0"),
		nil,
	)

	action := model_state.NewAction(actionKey, "BadSafety", "", nil, nil, []model_logic.Logic{safetyLogic}, nil)
	action = lowerAction(action, classKey)

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("count", object.NewInteger(5))
	instance := simState.CreateInstance(classKey, attrs)

	executor := buildTestExecutor(simState)

	_, err := executor.ExecuteAction(action, instance, nil)
	s.Require().Error(err)
	s.Contains(err.Error(), "SafetyRules must reference primed variables")
}

// ========================================================================
// Safety rule pass / violation tests
// ========================================================================

func (s *ActionsSuite) TestActionSafetyRulesPass() {
	classKey := mustKey("domain/d/subdomain/s/class/c")
	actionKey := mustKey("domain/d/subdomain/s/class/c/action/a")

	guaranteeLogic := model_logic.NewLogic(
		helper.Must(identity.NewActionGuaranteeKey(actionKey, "0")),
		model_logic.LogicTypeStateChange, "Postcondition.", "count", counterSpec("self.count + 1"),
		nil,
	)
	safetyLogic := model_logic.NewLogic(
		helper.Must(identity.NewActionSafetyKey(actionKey, "0")),
		model_logic.LogicTypeSafetyRule, "Safety rule.", "", counterSpec("self.count' >= 1"),
		nil,
	)

	action := model_state.NewAction(actionKey, "GoodAction", "", nil, []model_logic.Logic{guaranteeLogic}, []model_logic.Logic{safetyLogic}, nil)
	action = lowerAction(action, classKey)

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("count", object.NewInteger(5))
	instance := simState.CreateInstance(classKey, attrs)

	executor := buildTestExecutor(simState)

	result, err := executor.ExecuteAction(action, instance, nil)
	s.Require().NoError(err)
	s.True(result.Success)
	s.Empty(result.Violations)
}

func (s *ActionsSuite) TestActionSafetyRuleViolation() {
	classKey := mustKey("domain/d/subdomain/s/class/c")
	actionKey := mustKey("domain/d/subdomain/s/class/c/action/a")

	guaranteeLogic := model_logic.NewLogic(
		helper.Must(identity.NewActionGuaranteeKey(actionKey, "0")),
		model_logic.LogicTypeStateChange, "Postcondition.", "count", counterSpec("self.count + 1"),
		nil,
	)
	safetyLogic := model_logic.NewLogic(
		helper.Must(identity.NewActionSafetyKey(actionKey, "0")),
		model_logic.LogicTypeSafetyRule, "Safety rule.", "", counterSpec("self.count' < 0"),
		nil,
	)

	action := model_state.NewAction(actionKey, "ViolatingAction", "", nil, []model_logic.Logic{guaranteeLogic}, []model_logic.Logic{safetyLogic}, nil)
	action = lowerAction(action, classKey)

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("count", object.NewInteger(5))
	instance := simState.CreateInstance(classKey, attrs)

	executor := buildTestExecutor(simState)

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

	guardLogic := model_logic.NewLogic(
		guardKey,
		model_logic.LogicTypeAssessment, "Guard with primed variable.", "", counterSpec("self.count' > 0"),
		nil,
	)

	guard := model_state.NewGuard(guardKey, "BadGuard", guardLogic)
	guard = lowerGuard(guard, classKey)

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("count", object.NewInteger(5))
	instance := simState.CreateInstance(classKey, attrs)

	bb := state.NewBindingsBuilder(simState)
	ge := NewGuardEvaluator(bb)

	_, err := ge.EvaluateGuard(guard, instance)
	s.Require().Error(err)
	s.Contains(err.Error(), "guards must not contain primed variables")
}

// ========================================================================
// Query rejects primed variables in Requires test
// ========================================================================

func (s *ActionsSuite) TestQueryRejectsRequiresWithPrime() {
	classKey := mustKey("domain/d/subdomain/s/class/c")
	queryKey := mustKey("domain/d/subdomain/s/class/c/query/q")

	requireLogic := model_logic.NewLogic(
		helper.Must(identity.NewQueryRequireKey(queryKey, "0")),
		model_logic.LogicTypeAssessment, "Precondition.", "", counterSpec("self.count' > 0"),
		nil,
	)
	guaranteeLogic := model_logic.NewLogic(
		helper.Must(identity.NewQueryGuaranteeKey(queryKey, "0")),
		model_logic.LogicTypeQuery, "Postcondition.", "result", counterSpec("self.count"),
		nil,
	)

	query := model_state.NewQuery(queryKey, "BadQuery", "", []model_logic.Logic{requireLogic}, []model_logic.Logic{guaranteeLogic}, nil)
	query = lowerQuery(query, classKey)

	simState := state.NewSimulationState()
	attrs := object.NewRecord()
	attrs.Set("count", object.NewInteger(5))
	instance := simState.CreateInstance(classKey, attrs)

	executor := buildTestExecutor(simState)

	_, err := executor.ExecuteQuery(query, instance, nil)
	s.Require().Error(err)
	s.Contains(err.Error(), "Requires must not contain primed variables")
}
