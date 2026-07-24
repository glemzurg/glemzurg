package actions

import (
	"math"
	"math/big"
	"math/rand"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/schema"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"
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

func buildTestExecutor(simState *instance.State) *ActionExecutor {
	bb := state.NewBindingsBuilder(simState)
	ge := NewGuardEvaluator(bb)

	return NewActionExecutor(bb, InvariantRuntimeCheckers{Checker: nil, DataType: nil}, nil, ge, nil, nil)
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

func paramWithNatTypeSpec(parentKey identity.Key, name, rules string) model_state.Parameter {
	param := helper.Must(model_state.NewParameter(parentKey, name, rules, false))
	natTypeSpec := helper.Must(logic_spec.NewTypeSpec(model_logic.NotationTLAPlus, "Nat", nil))
	if param.DataType == nil {
		param.DataType = &model_data_type.DataType{
			CollectionType: "atomic",
			Atomic:         &model_data_type.Atomic{ConstraintType: model_data_type.CONSTRAINT_TYPE_UNCONSTRAINED},
		}
	}
	param.DataType.TypeSpec = &natTypeSpec
	return param
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

	action := model_state.NewAction(actionCloseKey, model_state.ActionDetails{Name: "DoClose", Details: ""}, nil, []model_logic.Logic{guaranteeLogic}, nil, nil)
	event := model_state.NewEvent(eventCloseKey, "close", "", nil)

	transition := model_state.NewTransition(transKey, eventCloseKey, model_state.TransitionStateKeys{FromStateKey: &stateOpenKey, ToStateKey: &stateClosedKey}, model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: &actionCloseKey}, "")

	class := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Order", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetAttributes(nil)
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
	actionA := mustKey("domain/d/subdomain/s/class/order/action/a")
	actionB := mustKey("domain/d/subdomain/s/class/order/action/b")

	s.True(ctx.ClaimInstanceForAction(1, actionA))
	err := ctx.RecordPrimedAssignment(1, "count", object.NewInteger(1))
	s.Require().NoError(err)
	err = ctx.RecordPrimedAssignment(1, "status", object.NewString("open"))
	s.Require().NoError(err)

	s.False(ctx.ClaimInstanceForAction(1, actionB))
	s.True(ctx.ClaimInstanceForAction(2, actionB))
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

	action := model_state.NewAction(actionKey, model_state.ActionDetails{Name: "increment", Details: ""}, nil, []model_logic.Logic{guaranteeLogic}, nil, nil)
	action = lowerAction(action, classKey)

	simState := instance.NewState(emptySchema())
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

	action := model_state.NewAction(actionKey, model_state.ActionDetails{Name: "close", Details: ""}, []model_logic.Logic{requireLogic}, []model_logic.Logic{guaranteeLogic}, nil, nil)
	action = lowerAction(action, classKey)

	simState := instance.NewState(emptySchema())
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

	action := model_state.NewAction(actionKey, model_state.ActionDetails{Name: "close", Details: ""}, []model_logic.Logic{requireLogic}, []model_logic.Logic{guaranteeLogic}, nil, nil)
	action = lowerAction(action, classKey)

	simState := instance.NewState(emptySchema())
	attrs := object.NewRecord()
	attrs.Set("status", object.NewString("closed")) // already closed
	instance := simState.CreateInstance(classKey, attrs)

	exec := buildTestExecutor(simState)

	result, err := exec.ExecuteAction(action, instance, nil)
	s.Require().NoError(err)
	s.False(result.Success)
	s.Require().Len(result.Violations, 1)
	s.Equal(invariants.ViolationTypeActionRequires, result.Violations[0].Type)
	s.Contains(result.Violations[0].Message, "requires[0] failed")
}

func (s *ActionsSuite) TestExecuteActionWithParameters() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	actionKey := mustKey("domain/d/subdomain/s/class/order/action/set_amount")

	guaranteeLogic := model_logic.NewLogic(
		helper.Must(identity.NewActionGuaranteeKey(actionKey, "0")),
		model_logic.LogicTypeStateChange, "Postcondition.", "amount", orderSpecWithParams("amount", []string{"amount"}),
		nil,
	)

	actionParams := []model_state.Parameter{paramWithNatTypeSpec(actionKey, "amount", "[0,1000]")}
	action := model_state.NewAction(actionKey, model_state.ActionDetails{Name: "set_amount", Details: ""}, nil, []model_logic.Logic{guaranteeLogic}, nil, actionParams)
	action = lowerAction(action, classKey)

	simState := instance.NewState(emptySchema())
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

func (s *ActionsSuite) TestExecuteActionReportsUnparsedParameterDataType() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	actionKey := mustKey("domain/d/subdomain/s/class/order/action/set_amount")
	param := helper.Must(model_state.NewParameter(actionKey, "amount", "not a valid rule", false))
	param.DataType = nil
	action := model_state.NewAction(actionKey, model_state.ActionDetails{Name: "set_amount", Details: ""}, nil, nil, nil, []model_state.Parameter{param})

	simState := instance.NewState(emptySchema())
	instance := simState.CreateInstance(classKey, object.NewRecord())
	exec := buildTestExecutor(simState)

	result, err := exec.ExecuteAction(action, instance, map[string]object.Object{
		"amount": object.NewInteger(1),
	})
	s.Require().NoError(err)
	s.False(result.Success)
	s.Require().Len(result.Violations, 1)
	s.Equal(invariants.ViolationTypeUnparsedDataType, result.Violations[0].Type)
}

func (s *ActionsSuite) TestExecuteActionReportsMissingParameterTypeSpec() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	actionKey := mustKey("domain/d/subdomain/s/class/order/action/set_amount")
	actionParams := []model_state.Parameter{
		helper.Must(model_state.NewParameter(actionKey, "amount", "unconstrained", false)),
	}
	action := model_state.NewAction(actionKey, model_state.ActionDetails{Name: "set_amount", Details: ""}, nil, nil, nil, actionParams)

	simState := instance.NewState(emptySchema())
	instance := simState.CreateInstance(classKey, object.NewRecord())
	exec := buildTestExecutor(simState)

	result, err := exec.ExecuteAction(action, instance, map[string]object.Object{
		"amount": object.NewInteger(1),
	})
	s.Require().NoError(err)
	s.False(result.Success)
	s.Require().Len(result.Violations, 1)
	s.Equal(invariants.ViolationTypeMissingParameterTypeSpec, result.Violations[0].Type)
	s.Equal("amount", result.Violations[0].AttributeName)
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

	simState := instance.NewState(emptySchema())
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

func (s *ActionsSuite) TestExecuteQueryReportsMissingParameterTypeSpec() {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	queryKey := mustKey("domain/d/subdomain/s/class/order/query/filter")
	queryParams := []model_state.Parameter{
		helper.Must(model_state.NewParameter(queryKey, "limit", "unconstrained", false)),
	}
	query := model_state.NewQuery(queryKey, "filter", "", nil, nil, queryParams)

	simState := instance.NewState(emptySchema())
	instance := simState.CreateInstance(classKey, object.NewRecord())
	exec := buildTestExecutor(simState)

	result, err := exec.ExecuteQuery(query, instance, map[string]object.Object{
		"limit": object.NewInteger(5),
	})
	s.Require().NoError(err)
	s.False(result.Success)
	s.Require().Len(result.Violations, 1)
	s.Equal(invariants.ViolationTypeMissingParameterTypeSpec, result.Violations[0].Type)
	s.Equal("limit", result.Violations[0].AttributeName)
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

	simState := instance.NewState(emptySchema())
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

	simState := instance.NewState(emptySchema())
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

	simState := instance.NewState(emptySchema())
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

	simState := instance.NewState(emptySchema())
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

	simState := instance.NewState(emptySchema())
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

	result, err := exec.ExecuteTransition(class, event, instance, nil, CreationLinkSource{SourceAssocKey: nil, SourceID: nil}, nil)
	s.Require().NoError(err)
	s.NotNil(result)

	s.Equal("Open", result.FromState)
	s.Equal("Closed", result.ToState)
	s.False(result.WasCreation)
	s.False(result.WasDestroy)

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

	transition := model_state.NewTransition(transKey, eventCreateKey, model_state.TransitionStateKeys{FromStateKey: nil, ToStateKey: &stateOpenKey}, model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: nil}, "")

	class := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Order", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetAttributes(nil)
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

	simState := instance.NewState(emptySchema())
	exec := buildTestExecutor(simState)

	eventObj := class.Events[eventCreateKey]

	result, err := exec.ExecuteTransition(class, eventObj, nil, nil, CreationLinkSource{SourceAssocKey: nil, SourceID: nil}, nil)
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

	transition := model_state.NewTransition(transKey, eventDeleteKey, model_state.TransitionStateKeys{FromStateKey: &stateOpenKey, ToStateKey: nil}, model_state.TransitionLogicKeys{GuardKey: nil, ActionKey: nil}, "")

	class := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Order", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetAttributes(nil)
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

	simState := instance.NewState(emptySchema())
	attrs := object.NewRecord()
	attrs.Set("_state", object.NewString("Open"))
	instance := simState.CreateInstance(classKey, attrs)
	_ = simState.SetStateMachineState(instance.ID, stateOpenKey)

	exec := buildTestExecutor(simState)

	eventObj := class.Events[eventDeleteKey]

	result, err := exec.ExecuteTransition(class, eventObj, instance, nil, CreationLinkSource{SourceAssocKey: nil, SourceID: nil}, nil)
	s.Require().NoError(err)
	s.True(result.WasDestroy)

	// Instance should be deleted
	s.Nil(simState.GetInstance(instance.ID))
	s.Equal(0, simState.InstanceCount())
}

func (s *ActionsSuite) TestExecuteTransitionNoMatchingTransition() {
	class, classKey := testOrderClass()

	simState := instance.NewState(emptySchema())
	attrs := object.NewRecord()
	attrs.Set("_state", object.NewString("Closed")) // No transition from Closed
	instance := simState.CreateInstance(classKey, attrs)

	stateClosedKey := mustKey("domain/d/subdomain/s/class/order/state/closed")
	_ = simState.SetStateMachineState(instance.ID, stateClosedKey)

	exec := buildTestExecutor(simState)

	eventCloseKey := mustKey("domain/d/subdomain/s/class/order/event/close")
	event := class.Events[eventCloseKey]

	_, err := exec.ExecuteTransition(class, event, instance, nil, CreationLinkSource{SourceAssocKey: nil, SourceID: nil}, nil)
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

	transApprove := model_state.NewTransition(transApproveKey, eventReviewKey, model_state.TransitionStateKeys{FromStateKey: &stateOpenKey, ToStateKey: &stateApprovedKey}, model_state.TransitionLogicKeys{GuardKey: &guardHighKey, ActionKey: nil}, "")
	transReject := model_state.NewTransition(transRejectKey, eventReviewKey, model_state.TransitionStateKeys{FromStateKey: &stateOpenKey, ToStateKey: &stateRejectedKey}, model_state.TransitionLogicKeys{GuardKey: &guardLowKey, ActionKey: nil}, "")

	class := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Order", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetAttributes(nil)
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

	simState := instance.NewState(emptySchema())

	// Case 1: High value order -> should go to Approved
	attrs := object.NewRecord()
	attrs.Set("amount", object.NewInteger(200))
	attrs.Set("_state", object.NewString("Open"))
	instance := simState.CreateInstance(classKey, attrs)
	_ = simState.SetStateMachineState(instance.ID, stateOpenKey)

	exec := buildTestExecutor(simState)

	event := class.Events[eventReviewKey]
	result, err := exec.ExecuteTransition(class, event, instance, nil, CreationLinkSource{SourceAssocKey: nil, SourceID: nil}, nil)
	s.Require().NoError(err)
	s.Equal("Approved", result.ToState)

	// Case 2: Low value order -> should go to Rejected
	attrs2 := object.NewRecord()
	attrs2.Set("amount", object.NewInteger(50))
	attrs2.Set("_state", object.NewString("Open"))
	instance2 := simState.CreateInstance(classKey, attrs2)
	_ = simState.SetStateMachineState(instance2.ID, stateOpenKey)

	result2, err := exec.ExecuteTransition(class, event, instance2, nil, CreationLinkSource{SourceAssocKey: nil, SourceID: nil}, nil)
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

	trans1 := model_state.NewTransition(trans1Key, eventKey, model_state.TransitionStateKeys{FromStateKey: &stateOpenKey, ToStateKey: &stateAKey}, model_state.TransitionLogicKeys{GuardKey: &guardAlwaysKey1, ActionKey: nil}, "")
	trans2 := model_state.NewTransition(trans2Key, eventKey, model_state.TransitionStateKeys{FromStateKey: &stateOpenKey, ToStateKey: &stateBKey}, model_state.TransitionLogicKeys{GuardKey: &guardAlwaysKey2, ActionKey: nil}, "")

	class := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Order", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetAttributes(nil)
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

	simState := instance.NewState(emptySchema())
	attrs := object.NewRecord()
	attrs.Set("_state", object.NewString("Open"))
	instance := simState.CreateInstance(classKey, attrs)
	_ = simState.SetStateMachineState(instance.ID, stateOpenKey)

	exec := buildTestExecutor(simState)

	event := class.Events[eventKey]
	_, err := exec.ExecuteTransition(class, event, instance, nil, CreationLinkSource{SourceAssocKey: nil, SourceID: nil}, nil)
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

	trans := model_state.NewTransition(transKey, eventKey, model_state.TransitionStateKeys{FromStateKey: &stateOpenKey, ToStateKey: &stateAKey}, model_state.TransitionLogicKeys{GuardKey: &guardNeverKey, ActionKey: nil}, "")

	class := model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Order", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetAttributes(nil)
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

	simState := instance.NewState(emptySchema())
	attrs := object.NewRecord()
	attrs.Set("_state", object.NewString("Open"))
	instance := simState.CreateInstance(classKey, attrs)
	_ = simState.SetStateMachineState(instance.ID, stateOpenKey)

	exec := buildTestExecutor(simState)

	event := class.Events[eventKey]
	_, err := exec.ExecuteTransition(class, event, instance, nil, CreationLinkSource{SourceAssocKey: nil, SourceID: nil}, nil)
	s.Require().Error(err)
	s.Contains(err.Error(), "deadlock")
}

// ========================================================================
// ValidateClassForSimulation test
// ========================================================================

func (s *ActionsSuite) TestValidateClassForSimulationNoStates() {
	class := model_class.NewClass(mustKey("domain/d/subdomain/s/class/empty"), model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Empty", Details: "", UnfinishedNotes: "", UmlComment: ""})
	class.SetStates(map[identity.Key]model_state.State{})

	err := ValidateClassForSimulation(class)
	s.Require().Error(err)
	s.Contains(err.Error(), "no states")
}

func (s *ActionsSuite) TestValidateClassForSimulationWithStates() {
	stateKey := mustKey("domain/d/subdomain/s/class/c/state/s1")

	class := model_class.NewClass(mustKey("domain/d/subdomain/s/class/c"), model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "C", Details: "", UnfinishedNotes: "", UmlComment: ""})
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

	class := model_class.NewClass(mustKey("domain/d/subdomain/s/class/c"), model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "C", Details: "", UnfinishedNotes: "", UmlComment: ""})
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

	ak := mustKey("domain/d/subdomain/s/class/c/action/a")
	paramDefs := []model_state.Parameter{
		helper.Must(model_state.NewParameter(ak, "amount", "[0,100]", false)),
		helper.Must(model_state.NewParameter(ak, "name", "string", false)),
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

	ak := mustKey("domain/d/subdomain/s/class/c/action/a")
	paramDefs := []model_state.Parameter{
		helper.Must(model_state.NewParameter(ak, "amount", "[0,100]", false)),
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

	ak := mustKey("domain/d/subdomain/s/class/c/action/a")
	countParam := helper.Must(model_state.NewParameter(ak, "count", "[10, 20]", false))
	countParam.DataType = &model_data_type.DataType{
		Key:            helper.Must(identity.NewDataTypeKey(countParam.Key, "")),
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
		s.Equal(object.KindReal, num.Kind())
		val := num.Float64()
		s.True(val >= 10 && val <= 20, "Generated value %g should be in [10,20]", val)
	}
}

func (s *ActionsSuite) TestSpanSamplingIntervalUnconstrainedScalesWithPrecision() {
	span := &model_data_type.AtomicSpan{
		LowerType:  "unconstrained",
		HigherType: "unconstrained",
		Precision:  0.1,
	}

	lower, upper := spanSamplingInterval(span)
	s.InDelta(-10.0, lower, 1e-9)
	s.InDelta(10.0, upper, 1e-9)

	span.Precision = 1
	lower, upper = spanSamplingInterval(span)
	s.InDelta(-100.0, lower, 1e-9)
	s.InDelta(100.0, upper, 1e-9)
}

func (s *ActionsSuite) TestGenerateRandomParametersSpanUnconstrainedPrecisionGrid() {
	binder := NewParameterBinder()
	rng := rand.New(rand.NewSource(42)) //nolint:gosec // deterministic seed intentional for test reproducibility

	ak := mustKey("domain/d/subdomain/s/class/c/action/a")
	amountParam := helper.Must(model_state.NewParameter(ak, "amount", "span", false))
	amountParam.DataType = &model_data_type.DataType{
		Key:            helper.Must(identity.NewDataTypeKey(amountParam.Key, "")),
		CollectionType: "atomic",
		Atomic: &model_data_type.Atomic{
			ConstraintType: "span",
			Span: &model_data_type.AtomicSpan{
				LowerType:  "unconstrained",
				HigherType: "unconstrained",
				Precision:  0.1,
			},
		},
	}

	paramDefs := []model_state.Parameter{amountParam}
	seen := make(map[float64]struct{})

	for range 5000 {
		result := binder.GenerateRandomParameters(paramDefs, rng)
		num, ok := result["amount"].(*object.Number)
		s.Require().True(ok)
		s.Equal(object.KindReal, num.Kind())
		val := num.Float64()
		s.GreaterOrEqual(val, -10.0)
		s.LessOrEqual(val, 10.0)
		quotient := val / 0.1
		s.InDelta(math.Round(quotient), quotient, 1e-6, "value %g should align to 0.1 grid", val)
		seen[val] = struct{}{}
	}

	s.Len(seen, 201)
}

func (s *ActionsSuite) TestGenerateRandomParametersSpanFractional() {
	binder := NewParameterBinder()
	rng := rand.New(rand.NewSource(42)) //nolint:gosec // deterministic seed intentional for test reproducibility

	lowerValue := 3
	lowerDenom := 4
	higherValue := 5
	higherDenom := 6

	ak := mustKey("domain/d/subdomain/s/class/c/action/a")
	lengthParam := helper.Must(model_state.NewParameter(ak, "length", "(3/4 .. 5/6]", false))
	lengthParam.DataType = &model_data_type.DataType{
		Key:            helper.Must(identity.NewDataTypeKey(lengthParam.Key, "")),
		CollectionType: "atomic",
		Atomic: &model_data_type.Atomic{
			ConstraintType: "span",
			Span: &model_data_type.AtomicSpan{
				LowerType:         "open",
				LowerValue:        &lowerValue,
				LowerDenominator:  &lowerDenom,
				HigherType:        "closed",
				HigherValue:       &higherValue,
				HigherDenominator: &higherDenom,
				Precision:         0.01,
			},
		},
	}

	paramDefs := []model_state.Parameter{lengthParam}

	for range 100 {
		result := binder.GenerateRandomParameters(paramDefs, rng)
		num, ok := result["length"].(*object.Number)
		s.Require().True(ok)
		s.Equal(object.KindReal, num.Kind())
		val := num.Float64()
		s.Greater(val, 0.75, "Generated value %g should be above open lower bound 3/4", val)
		s.LessOrEqual(val, 5.0/6.0+1e-9, "Generated value %g should be at most closed upper bound 5/6", val)
	}
}

func (s *ActionsSuite) TestGenerateRandomParametersEnum() {
	binder := NewParameterBinder()
	rng := rand.New(rand.NewSource(42)) //nolint:gosec // deterministic seed intentional for test reproducibility

	ak := mustKey("domain/d/subdomain/s/class/c/action/a")
	colorParam := helper.Must(model_state.NewParameter(ak, "color", "{red, green, blue}", false))
	colorParam.DataType = &model_data_type.DataType{
		Key:            helper.Must(identity.NewDataTypeKey(colorParam.Key, "")),
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

	ak := mustKey("domain/d/subdomain/s/class/c/action/a")
	paramDefs := []model_state.Parameter{
		helper.Must(model_state.NewParameter(ak, "x", "unknown", false)),
	}

	result := binder.GenerateRandomParameters(paramDefs, rng)
	s.Contains(result, "x")
	_, ok := result["x"].(*object.Number)
	s.True(ok, "Should generate a number as default")
}

func (s *ActionsSuite) TestGenerateRandomParametersDateTime() {
	binder := NewParameterBinder()
	rng := rand.New(rand.NewSource(99)) //nolint:gosec // deterministic seed intentional for test reproducibility

	ak := mustKey("domain/d/subdomain/s/class/event/action/record")
	paramDefs := []model_state.Parameter{
		helper.Must(model_state.NewParameter(ak, "when", "datetime", false)),
	}

	for range 50 {
		result := binder.GenerateRandomParameters(paramDefs, rng)
		num, ok := result["when"].(*object.Number)
		s.Require().True(ok)
		rat := num.Rat()
		s.True(rat.IsInt())
		s.GreaterOrEqual(rat.Cmp(big.NewRat(model_data_type.DateTimeValueMin, 1)), 0)
		s.LessOrEqual(rat.Cmp(big.NewRat(model_data_type.DateTimeValueMax, 1)), 0)
	}
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

	action := model_state.NewAction(actionKey, model_state.ActionDetails{Name: "BadRequires", Details: ""}, []model_logic.Logic{requireLogic}, nil, nil, nil)
	action = lowerAction(action, classKey)

	simState := instance.NewState(emptySchema())
	attrs := object.NewRecord()
	attrs.Set("count", object.NewInteger(5))
	instance := simState.CreateInstance(classKey, attrs)

	executor := buildTestExecutor(simState)

	_, err := executor.ExecuteAction(action, instance, nil)
	s.Require().Error(err)
	s.Contains(err.Error(), "requires[0]: must not contain primed variables")
}

func (s *ActionsSuite) TestActionSafetyRulesMustHavePrime() {
	classKey := mustKey("domain/d/subdomain/s/class/c")
	actionKey := mustKey("domain/d/subdomain/s/class/c/action/a")

	safetyLogic := model_logic.NewLogic(
		helper.Must(identity.NewActionSafetyKey(actionKey, "0")),
		model_logic.LogicTypeSafetyRule, "Safety rule.", "", counterSpec("self.count > 0"),
		nil,
	)

	action := model_state.NewAction(actionKey, model_state.ActionDetails{Name: "BadSafety", Details: ""}, nil, nil, []model_logic.Logic{safetyLogic}, nil)
	action = lowerAction(action, classKey)

	simState := instance.NewState(emptySchema())
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

	action := model_state.NewAction(actionKey, model_state.ActionDetails{Name: "GoodAction", Details: ""}, nil, []model_logic.Logic{guaranteeLogic}, []model_logic.Logic{safetyLogic}, nil)
	action = lowerAction(action, classKey)

	simState := instance.NewState(emptySchema())
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

	action := model_state.NewAction(actionKey, model_state.ActionDetails{Name: "ViolatingAction", Details: ""}, nil, []model_logic.Logic{guaranteeLogic}, []model_logic.Logic{safetyLogic}, nil)
	action = lowerAction(action, classKey)

	simState := instance.NewState(emptySchema())
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

	simState := instance.NewState(emptySchema())
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

	simState := instance.NewState(emptySchema())
	attrs := object.NewRecord()
	attrs.Set("count", object.NewInteger(5))
	instance := simState.CreateInstance(classKey, attrs)

	executor := buildTestExecutor(simState)

	_, err := executor.ExecuteQuery(query, instance, nil)
	s.Require().Error(err)
	s.Contains(err.Error(), "requires[0]: must not contain primed variables")
}

func (s *ActionsSuite) TestExecuteTransitionReportsMultiplicityViolation() {
	orderClass, orderKey := testOrderClass()
	itemClass, itemKey := multiplicityItemClass()

	assocKey := multiplicityAssocKey(orderKey, itemKey, "OrderItem")
	fromMult := helper.Must(model_class.NewMultiplicity("1"))
	toMult := helper.Must(model_class.NewMultiplicity("2..many"))
	assoc := model_class.NewAssociation(assocKey, model_class.AssociationDetails{Name: "OrderItem", Details: ""}, model_class.AssociationEnd{ClassKey: orderKey, Multiplicity: fromMult}, model_class.AssociationEnd{ClassKey: itemKey, Multiplicity: toMult}, model_class.AssociationOptions{AssociationClassKey: nil, UmlComment: ""})

	model := multiplicityTestModel(orderClass, orderKey, itemClass, itemKey)
	model.ClassAssociations = map[identity.Key]model_class.Association{assocKey: assoc}

	simState := instance.NewState(emptySchema())
	bb := state.NewBindingsBuilder(simState)
	multChecker := invariants.NewMultiplicityChecker(schema.New(model))
	ge := NewGuardEvaluator(bb)
	exec := NewActionExecutor(bb, InvariantRuntimeCheckers{Checker: nil, DataType: nil}, &invariants.StructuralInvariantCheckers{
		Multiplicity: multChecker,
	}, ge, nil, nil)

	orderAttrs := object.NewRecord()
	orderAttrs.Set("_state", object.NewString("Open"))
	orderAttrs.Set("amount", object.NewInteger(0))
	order := simState.CreateInstance(orderKey, orderAttrs)
	item := simState.CreateInstance(itemKey, object.NewRecord())
	s.Require().NoError(simState.AddLink(assocKey, order.ID, item.ID))

	event := orderClass.Events[mustKey("domain/d/subdomain/s/class/order/event/close")]
	instance := simState.GetInstance(order.ID)
	result, err := exec.ExecuteTransition(orderClass, event, instance, nil, CreationLinkSource{SourceAssocKey: nil, SourceID: nil}, nil)
	s.Require().NoError(err)

	multViolations := result.Violations.ByType(invariants.ViolationTypeMultiplicity)
	s.Require().Len(multViolations, 1)
	s.Contains(multViolations[0].Message, "at least 2")
}

func multiplicityItemClass() (model_class.Class, identity.Key) {
	classKey := mustKey("domain/d/subdomain/s/class/item")
	stateActiveKey := mustKey("domain/d/subdomain/s/class/item/state/active")
	class := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Item"})
	class.SetStates(map[identity.Key]model_state.State{
		stateActiveKey: model_state.NewState(stateActiveKey, "Active", "", ""),
	})
	return class, classKey
}

func multiplicityTestModel(orderClass model_class.Class, orderKey identity.Key, itemClass model_class.Class, itemKey identity.Key) *core.Model {
	subdomainKey := mustKey("domain/d/subdomain/s")
	domainKey := mustKey("domain/d")
	subdomain := model_domain.NewSubdomain(subdomainKey, "S", "", "", "")
	subdomain.Classes = map[identity.Key]model_class.Class{
		orderKey: orderClass,
		itemKey:  itemClass,
	}
	domain := model_domain.NewDomain(domainKey, "D", "", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}
	model := core.NewModel("test", core.ModelDetails{Name: "Test", Details: ""}, "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{domainKey: domain}
	return &model
}

func multiplicityAssocKey(fromClassKey, toClassKey identity.Key, name string) identity.Key {
	parentKey := mustKey("domain/d/subdomain/s")
	k, err := identity.NewClassAssociationKey(parentKey, fromClassKey, toClassKey, name)
	if err != nil {
		panic(err)
	}
	return k
}
