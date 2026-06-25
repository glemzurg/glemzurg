package actions

import (
	"fmt"
	"maps"
	"math/rand"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/invariants"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/model_bridge"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

// _EXPRESSION_RETURNED_NIL is the error message used when an expression evaluates to nil.
const _EXPRESSION_RETURNED_NIL = "expression returned nil"

// ActionResult holds the result of executing an action.
type ActionResult struct {
	// InstanceID is the primary instance the action was executed on.
	InstanceID state.InstanceID

	// PrimedAssignments contains all state changes grouped by instance ID.
	PrimedAssignments map[state.InstanceID]map[string]object.Object

	// Violations contains any invariant violations detected after state changes.
	Violations invariants.ViolationErrors

	// Success is true if there are no violations.
	Success bool
}

// QueryResult holds the result of executing a query.
type QueryResult struct {
	// InstanceID is the instance the query was executed on.
	InstanceID state.InstanceID

	// Outputs contains the query's output values from primed assignments (e.g., result' = ...).
	Outputs map[string]object.Object

	// Violations contains any post-condition violations.
	Violations invariants.ViolationErrors

	// Success is true if there are no violations.
	Success bool
}

// AssociationMaterialization records a host association row created via an association class.
type AssociationMaterialization struct {
	HostAssociationName string
	HostAssociationKey  identity.Key
	FromClassName       string
	FromClassKey        identity.Key
	ToClassName         string
	ToClassKey          identity.Key
	FromInstanceID      state.InstanceID
	ToInstanceID        state.InstanceID
}

// TransitionResult holds the result of executing a state machine transition.
type TransitionResult struct {
	// InstanceID is the instance that transitioned.
	InstanceID state.InstanceID

	// FromState is the name of the state before the transition (empty for creation).
	FromState string

	// ToState is the name of the state after the transition (empty for deletion).
	ToState string

	// EventKey is the event that triggered the transition.
	EventKey identity.Key

	// TransitionKey is the key of the transition that was taken.
	TransitionKey identity.Key

	// WasCreation is true if the transition was from the initial state (object creation).
	WasCreation bool

	// WasDeletion is true if the transition was to the final state (object deletion).
	WasDeletion bool

	// AssociationMaterialization is set when creation materializes a host association via an association class.
	AssociationMaterialization *AssociationMaterialization

	// ActionResult is the result of the action executed during the transition (if any).
	ActionResult *ActionResult

	// Violations contains any violations from the transition.
	Violations invariants.ViolationErrors
}

// AssociationClassIndex resolves association-class metadata for creation linking.
type AssociationClassIndex interface {
	GetAssociationClassInfo(classKey identity.Key) AssociationClassLinkInfo
	IsAssociationClass(classKey identity.Key) bool
	IsAssociationClassHost(assocKey identity.Key) bool
}

// AssociationClassLinkInfo holds host-association and endpoint metadata for one AC row.
type AssociationClassLinkInfo struct {
	Found               bool
	HostAssocKey        identity.Key
	HostAssociationName string
	FromClassKey        identity.Key
	FromClassName       string
	ToClassKey          identity.Key
	ToClassName         string
}

// ActionExecutor executes actions, queries, and transitions against simulation state.
type ActionExecutor struct {
	bindingsBuilder    *state.BindingsBuilder
	invariantChecker   *invariants.InvariantChecker
	dataTypeChecker    *invariants.DataTypeChecker
	structuralCheckers *invariants.StructuralInvariantCheckers
	guardEvaluator     *GuardEvaluator
	acIndex            AssociationClassIndex
	rng                *rand.Rand

	// deferMultiplicityInActionCheck skips implicit multiplicity checks inside
	// checkAllInvariants while a transition is mid-flight; ExecuteTransition
	// runs them after links and _state are fully applied.
	deferMultiplicityInActionCheck bool

	// classNameMap binds class display names to instance sets for action requires.
	classNameMap map[identity.Key]string
}

// InvariantRuntimeCheckers pairs the primary invariant and data-type checkers used during action execution.
type InvariantRuntimeCheckers struct {
	Checker  *invariants.InvariantChecker
	DataType *invariants.DataTypeChecker
}

// NewActionExecutor creates a new action executor.
func NewActionExecutor(
	bindingsBuilder *state.BindingsBuilder,
	runtime InvariantRuntimeCheckers,
	structuralCheckers *invariants.StructuralInvariantCheckers,
	guardEvaluator *GuardEvaluator,
	acIndex AssociationClassIndex,
	rng *rand.Rand,
) *ActionExecutor {
	var classNameMap map[identity.Key]string
	if runtime.Checker != nil {
		classNameMap = runtime.Checker.ClassNameMap()
	}
	return &ActionExecutor{
		bindingsBuilder:    bindingsBuilder,
		invariantChecker:   runtime.Checker,
		dataTypeChecker:    runtime.DataType,
		structuralCheckers: structuralCheckers,
		guardEvaluator:     guardEvaluator,
		acIndex:            acIndex,
		rng:                rng,
		classNameMap:       classNameMap,
	}
}

// ExecuteAction is the top-level entry point for action execution.
// It creates an ExecutionContext, runs the action (which may chain to others),
// then applies all primed assignments and checks all invariants.
func (e *ActionExecutor) ExecuteAction(
	action model_state.Action,
	instance *state.ClassInstance,
	parameters map[string]object.Object,
) (*ActionResult, error) {
	ctx := NewExecutionContext()
	paramViolations := invariants.CheckParameterTypeSpecs(
		action.Parameters, action.Key, action.Name, "action", instance.ID, instance.ClassKey,
	)

	// Phase A: Execute the action chain (collecting primed values and post-conditions)
	if err := e.executeActionInContext(ctx, action, instance, parameters); err != nil {
		return nil, err
	}

	if ctx.RequiresViolations().HasViolations() {
		return &ActionResult{
			InstanceID: instance.ID,
			Violations: append(ctx.RequiresViolations(), paramViolations...),
			Success:    false,
		}, nil
	}

	// Phase B: Apply ALL primed assignments from the entire chain
	if err := e.applyPrimedAssignments(ctx); err != nil {
		return nil, err
	}

	// Phases C-F: Check all post-conditions and invariants
	allViolations := e.checkAllInvariants(ctx)
	allViolations = append(allViolations, paramViolations...)

	return &ActionResult{
		InstanceID:        instance.ID,
		PrimedAssignments: ctx.GetAllPrimedAssignments(),
		Violations:        allViolations,
		Success:           !allViolations.HasViolations(),
	}, nil
}

// applyPrimedAssignments applies all primed assignments from the context to simulation state.
func (e *ActionExecutor) applyPrimedAssignments(ctx *ExecutionContext) error {
	simState := e.bindingsBuilder.State()
	for instanceID, primedFields := range ctx.GetAllPrimedAssignments() {
		instance := simState.GetInstance(instanceID)
		for fieldName, value := range primedFields {
			if instance != nil && e.dataTypeChecker != nil {
				if attr := e.dataTypeChecker.AttributeDef(instance.ClassKey, fieldName); attr != nil {
					value = CoerceValueForDataType(attr.DataType, value)
				}
			}
			if err := simState.UpdateInstanceField(instanceID, fieldName, value); err != nil {
				return fmt.Errorf("failed to apply primed assignment %s on instance %d: %w", fieldName, instanceID, err)
			}
		}
	}
	return nil
}

// checkAllInvariants runs all post-condition and invariant checks and returns combined violations.
func (e *ActionExecutor) checkAllInvariants(ctx *ExecutionContext) invariants.ViolationErrors {
	var allViolations invariants.ViolationErrors

	allViolations = append(allViolations, e.checkPostConditions(ctx)...)
	allViolations = append(allViolations, e.checkSafetyRules(ctx)...)
	allViolations = append(allViolations, e.checkDataTypeConstraints(ctx)...)
	allViolations = append(allViolations, e.checkModelInvariants()...)
	allViolations = append(allViolations, e.checkIndexUniqueness()...)
	if !e.deferMultiplicityInActionCheck {
		allViolations = append(allViolations, e.checkAssociationStructuralInvariants()...)
	}

	return allViolations
}

// checkPostConditions evaluates all deferred post-conditions from the execution context.
func (e *ActionExecutor) checkPostConditions(ctx *ExecutionContext) invariants.ViolationErrors {
	var violations invariants.ViolationErrors
	simState := e.bindingsBuilder.State()

	for _, pc := range ctx.GetAllPostConditions() {
		targetInstance := simState.GetInstance(pc.InstanceID)
		if targetInstance == nil {
			continue
		}
		postBindings := e.bindingsBuilder.BuildForInstance(targetInstance)
		if msg := evalBooleanCheck(pc.Expression, postBindings); msg != "" {
			violations = append(violations, createPostConditionViolation(pc, msg))
		}
	}

	return violations
}

// checkSafetyRules evaluates all deferred safety rules from the execution context.
func (e *ActionExecutor) checkSafetyRules(ctx *ExecutionContext) invariants.ViolationErrors {
	var violations invariants.ViolationErrors
	simState := e.bindingsBuilder.State()

	for _, sr := range ctx.GetAllSafetyRules() {
		targetInstance := simState.GetInstance(sr.InstanceID)
		if targetInstance == nil {
			continue
		}
		safetyBindings := e.bindingsBuilder.BuildForInstance(targetInstance)
		for name, value := range sr.LetBindings {
			safetyBindings.Set(name, value, evaluator.NamespaceLocal)
		}
		if msg := evalBooleanCheck(sr.Expression, safetyBindings); msg != "" {
			violations = append(violations, invariants.NewSafetyRuleViolation(
				sr.SourceKey, sr.SourceName, sr.Index, sr.OriginalExpression,
				sr.InstanceID, msg,
			))
		}
	}

	return violations
}

// evalBooleanCheck evaluates an expression and returns an error message if it
// doesn't evaluate to TRUE. Returns empty string on success.
func evalBooleanCheck(expr me.Expression, bindings *evaluator.Bindings) string {
	result := evaluator.Eval(expr, bindings)
	if result.IsError() {
		return fmt.Sprintf("evaluation error: %s", result.Error.Inspect())
	}
	if isTrueBoolean(result.Value) {
		return ""
	}
	if result.Value == nil {
		return _EXPRESSION_RETURNED_NIL
	}
	return fmt.Sprintf("expression returned %s", result.Value.Inspect())
}

// checkDataTypeConstraints checks data type constraints on all mutated instances.
func (e *ActionExecutor) checkDataTypeConstraints(ctx *ExecutionContext) invariants.ViolationErrors {
	if e.dataTypeChecker == nil {
		return nil
	}
	var violations invariants.ViolationErrors
	simState := e.bindingsBuilder.State()

	for _, instanceID := range ctx.MutatedInstanceIDs() {
		targetInstance := simState.GetInstance(instanceID)
		if targetInstance == nil {
			continue
		}
		violations = append(violations, e.dataTypeChecker.CheckInstance(targetInstance)...)
	}

	return violations
}

// checkModelInvariants checks model-level invariants.
func (e *ActionExecutor) checkModelInvariants() invariants.ViolationErrors {
	if e.invariantChecker == nil {
		return nil
	}
	return e.invariantChecker.CheckModelInvariants(e.bindingsBuilder.State(), e.bindingsBuilder)
}

// checkIndexUniqueness checks index uniqueness constraints.
func (e *ActionExecutor) checkIndexUniqueness() invariants.ViolationErrors {
	if e.structuralCheckers == nil || e.structuralCheckers.Index == nil {
		return nil
	}
	return e.structuralCheckers.Index.CheckState(e.bindingsBuilder.State())
}

// checkAssociationStructuralInvariants checks association multiplicities and association invariants.
func (e *ActionExecutor) checkAssociationStructuralInvariants() invariants.ViolationErrors {
	if e.structuralCheckers == nil {
		return nil
	}
	var violations invariants.ViolationErrors
	if e.structuralCheckers.Multiplicity != nil {
		violations = append(violations, e.structuralCheckers.Multiplicity.CheckState(e.bindingsBuilder.State())...)
	}
	if e.structuralCheckers.AssociationInvariants != nil {
		violations = append(violations, e.structuralCheckers.AssociationInvariants.CheckState(
			e.bindingsBuilder.State(),
			e.bindingsBuilder,
		)...)
	}
	return violations
}

func (e *ActionExecutor) buildRequiresBindings(
	instance *state.ClassInstance,
	parameters map[string]object.Object,
) *evaluator.Bindings {
	if len(e.classNameMap) == 0 {
		return e.bindingsBuilder.BuildForInstanceWithVariables(instance, parameters)
	}
	return e.bindingsBuilder.BuildWithClassInstancesForInstanceWithVariables(
		e.classNameMap, instance, parameters,
	)
}

// executeActionInContext runs a single action within an existing context.
// This is called both for top-level actions and for chained actions.
func (e *ActionExecutor) executeActionInContext(
	ctx *ExecutionContext,
	action model_state.Action,
	instance *state.ClassInstance,
	parameters map[string]object.Object,
) error {
	if err := ctx.IncrementDepth(); err != nil {
		return err
	}
	defer ctx.DecrementDepth()

	if !ctx.ClaimInstanceForAction(instance.ID, action.Key) {
		return fmt.Errorf(
			"re-entrant mutation on instance %d in action %s: instance already mutated by another action in the chain",
			instance.ID, action.Name,
		)
	}

	bindings := e.buildRequiresBindings(instance, parameters)

	reqViolations, err := e.evaluateActionRequires(action, instance.ID, bindings)
	if err != nil {
		return err
	}
	if reqViolations.HasViolations() {
		ctx.SetRequiresViolations(reqViolations)
		return nil
	}

	if err := e.evaluateActionGuarantees(ctx, action, instance, bindings); err != nil {
		return err
	}

	return e.collectActionSafetyRules(ctx, action, instance, bindings)
}

// evaluateActionRequires evaluates the preconditions (Requires) for an action.
func (e *ActionExecutor) evaluateActionRequires(
	action model_state.Action,
	instanceID state.InstanceID,
	bindings *evaluator.Bindings,
) (invariants.ViolationErrors, error) {
	owner := ParameterOwnerFromAction(action)
	reqFailures, err := owner.AssessRequires(action.Parameters, bindings)
	if err != nil {
		return nil, err
	}
	invFailures, err := owner.AssessParameterInvariants(action.Parameters, bindings)
	if err != nil {
		return nil, err
	}
	violations := owner.ActionRequiresViolations(reqFailures, instanceID)
	violations = append(violations, owner.ParameterInvariantViolations(invFailures, instanceID)...)
	return violations, nil
}

// evaluateActionGuarantees evaluates the guarantees for an action and records primed assignments.
func (e *ActionExecutor) evaluateActionGuarantees(
	ctx *ExecutionContext,
	action model_state.Action,
	instance *state.ClassInstance,
	bindings *evaluator.Bindings,
) error {
	// Pass 1: Evaluate all let bindings in guarantees (in order).
	if err := evalLetBindings(action.Guarantees, bindings, "action", action.Name, "guarantee"); err != nil {
		return err
	}
	// Pass 2: Evaluate non-let state_change items.
	for i, guar := range action.Guarantees {
		if guar.Type == model_logic.LogicTypeLet {
			continue
		}

		if guar.Target == "" {
			return fmt.Errorf("action %s guarantee[%d]: target must be set", action.Name, i)
		}
		expr := guar.Spec.Expression
		if expr == nil {
			return fmt.Errorf("action %s guarantee[%d]: expression not lowered", action.Name, i)
		}
		rhsValue := evaluator.Eval(expr, bindings)
		if rhsValue.IsError() {
			return fmt.Errorf("action %s guarantee[%d] evaluation error: %s", action.Name, i, rhsValue.Error.Inspect())
		}
		if err := ctx.RecordPrimedAssignment(instance.ID, guar.Target, rhsValue.Value); err != nil {
			return fmt.Errorf("action %s guarantee[%d]: %w", action.Name, i, err)
		}
	}
	return nil
}

// collectActionSafetyRules collects safety rules for deferred evaluation after state changes.
func (e *ActionExecutor) collectActionSafetyRules(
	ctx *ExecutionContext,
	action model_state.Action,
	instance *state.ClassInstance,
	bindings *evaluator.Bindings,
) error {
	// Pass 1: Evaluate all let bindings in safety rules and capture them.
	letBindings := make(map[string]object.Object)
	for i, rule := range action.SafetyRules {
		if rule.Type != model_logic.LogicTypeLet {
			continue
		}
		expr := rule.Spec.Expression
		if expr == nil {
			return fmt.Errorf("action %s safety_rule[%d] (let): expression not lowered", action.Name, i)
		}
		result := evaluator.Eval(expr, bindings)
		if result.IsError() {
			return fmt.Errorf("action %s safety_rule[%d] (let %q) evaluation error: %s", action.Name, i, rule.Target, result.Error.Inspect())
		}
		bindings.Set(rule.Target, result.Value, evaluator.NamespaceLocal)
		letBindings[rule.Target] = result.Value
	}
	// Pass 2: Collect non-let safety rules with let bindings snapshot.
	for i, rule := range action.SafetyRules {
		if rule.Type == model_logic.LogicTypeLet {
			continue
		}
		expr := rule.Spec.Expression
		if expr == nil {
			return fmt.Errorf("action %s safety_rule[%d]: expression not lowered", action.Name, i)
		}

		if !model_bridge.ContainsAnyPrimedME(expr) {
			return fmt.Errorf("action %s safety_rule[%d]: SafetyRules must reference primed variables", action.Name, i)
		}

		// Copy current letBindings for deferred evaluation.
		var capturedLetBindings map[string]object.Object
		if len(letBindings) > 0 {
			capturedLetBindings = make(map[string]object.Object, len(letBindings))
			maps.Copy(capturedLetBindings, letBindings)
		}

		ctx.AddSafetyRule(DeferredSafetyRule{
			Expression:         expr,
			InstanceID:         instance.ID,
			SourceKey:          action.Key,
			SourceName:         action.Name,
			Index:              i,
			OriginalExpression: rule.Spec.Specification,
			LetBindings:        capturedLetBindings,
		})
	}

	return nil
}

// evalLetBindings evaluates all LogicTypeLet items in a logic list in order,
// adding each result to bindings. Non-let items are skipped.
func evalLetBindings(logics []model_logic.Logic, bindings *evaluator.Bindings, ownerType, ownerName, listName string) error {
	for i, logic := range logics {
		if logic.Type != model_logic.LogicTypeLet {
			continue
		}
		expr := logic.Spec.Expression
		if expr == nil {
			return fmt.Errorf("%s %s %s[%d] (let): expression not lowered", ownerType, ownerName, listName, i)
		}
		result := evaluator.Eval(expr, bindings)
		if result.IsError() {
			return fmt.Errorf("%s %s %s[%d] (let %q) evaluation error: %s", ownerType, ownerName, listName, i, logic.Target, result.Error.Inspect())
		}
		bindings.Set(logic.Target, result.Value, evaluator.NamespaceLocal)
	}
	return nil
}

// ExecuteQuery runs a query on an instance. Queries do not modify state.
// Query primed assignments produce output values, not state changes.
func (e *ActionExecutor) ExecuteQuery(
	query model_state.Query,
	instance *state.ClassInstance,
	parameters map[string]object.Object,
) (*QueryResult, error) {
	ctx := NewExecutionContext()
	paramViolations := invariants.CheckParameterTypeSpecs(
		query.Parameters, query.Key, query.Name, "query", instance.ID, instance.ClassKey,
	)

	outputs, err := e.executeQueryInContext(ctx, query, instance, parameters)
	if err != nil {
		return nil, err
	}

	// Check post-conditions (reuse shared helper).
	allViolations := e.checkPostConditions(ctx)
	allViolations = append(allViolations, paramViolations...)

	return &QueryResult{
		InstanceID: instance.ID,
		Outputs:    outputs,
		Violations: allViolations,
		Success:    !allViolations.HasViolations(),
	}, nil
}

// executeQueryInContext runs a query within an existing execution context.
func (e *ActionExecutor) executeQueryInContext(
	ctx *ExecutionContext,
	query model_state.Query,
	instance *state.ClassInstance,
	parameters map[string]object.Object,
) (map[string]object.Object, error) {
	if err := ctx.IncrementDepth(); err != nil {
		return nil, err
	}
	defer ctx.DecrementDepth()

	bindings := e.buildRequiresBindings(instance, parameters)
	if err := checkQueryRequires(query, bindings); err != nil {
		return nil, err
	}
	return evaluateQueryGuarantees(query, bindings)
}

func checkQueryRequires(query model_state.Query, bindings *evaluator.Bindings) error {
	owner := ParameterOwnerFromQuery(query)
	reqFailures, err := owner.AssessRequires(query.Parameters, bindings)
	if err != nil {
		return err
	}
	if err := owner.RequireAssessmentError(reqFailures); err != nil {
		return err
	}
	invFailures, err := owner.AssessParameterInvariants(query.Parameters, bindings)
	if err != nil {
		return err
	}
	return owner.RequireAssessmentError(invFailures)
}

func evaluateQueryGuarantees(query model_state.Query, bindings *evaluator.Bindings) (map[string]object.Object, error) {
	outputs := make(map[string]object.Object)

	if err := evalLetBindings(query.Guarantees, bindings, logicOwnerKindQuery, query.Name, "guarantee"); err != nil {
		return nil, err
	}

	for i, guar := range query.Guarantees {
		if guar.Type == model_logic.LogicTypeLet {
			continue
		}
		if guar.Target == "" {
			return nil, fmt.Errorf("query %s guarantee[%d]: target must be set", query.Name, i)
		}
		expr := guar.Spec.Expression
		if expr == nil {
			return nil, fmt.Errorf("query %s guarantee[%d]: expression not lowered", query.Name, i)
		}
		rhsValue := evaluator.Eval(expr, bindings)
		if rhsValue.IsError() {
			return nil, fmt.Errorf("query %s guarantee[%d] evaluation error: %s", query.Name, i, rhsValue.Error.Inspect())
		}
		outputs[guar.Target] = rhsValue.Value
	}

	return outputs, nil
}

// CreationLinkSource holds the association and parent instance for object creation linking.
type CreationLinkSource struct {
	SourceAssocKey *identity.Key
	SourceID       *state.InstanceID
}

// ExecuteTransition handles an event arriving at an instance.
// It finds matching transitions, evaluates guards, runs the action (if any),
// and sets the _state attribute. Handles creation and deletion.
func (e *ActionExecutor) ExecuteTransition(
	class model_class.Class,
	event model_state.Event,
	instance *state.ClassInstance, // nil for creation (from initial state)
	eventParams map[string]object.Object,
	source CreationLinkSource,
	targetID *state.InstanceID,
) (*TransitionResult, error) {
	currentStateName := getInstanceCurrentState(instance)

	// Step 1: Find candidate transitions
	candidates := e.findCandidateTransitions(class, event, instance, currentStateName)
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no transitions for event %s from state %s on class %s", event.Name, currentStateName, class.Name)
	}

	// Step 2: Evaluate guards to pick exactly one transition
	chosen, err := e.evaluateGuards(candidates, class, instance, event, currentStateName)
	if err != nil {
		return nil, err
	}

	// Step 3: Handle creation (FromStateKey == nil)
	var associationMaterialization *AssociationMaterialization
	if chosen.FromStateKey == nil {
		instance, err = e.handleCreation(class, instance, source.SourceAssocKey, source.SourceID, targetID)
		if err != nil {
			return nil, err
		}
		associationMaterialization = e.associationMaterializationForCreation(class, source, targetID)
	}

	// Step 4: Execute the action (if any). Multiplicity is deferred until after _state applies.
	e.deferMultiplicityInActionCheck = true
	actionResult, err := e.executeTransitionAction(chosen, class, instance, eventParams)
	e.deferMultiplicityInActionCheck = false
	if err != nil {
		return nil, err
	}

	// Step 5: Apply state transition
	toStateName, err := e.applyStateTransition(chosen, class, instance)
	if err != nil {
		return nil, err
	}

	var violations invariants.ViolationErrors
	if actionResult != nil {
		violations = actionResult.Violations
	}
	violations = append(violations, e.checkAssociationStructuralInvariants()...)

	return &TransitionResult{
		InstanceID:                 instance.ID,
		FromState:                  currentStateName,
		ToState:                    toStateName,
		EventKey:                   event.Key,
		TransitionKey:              chosen.Key,
		WasCreation:                chosen.FromStateKey == nil,
		WasDeletion:                chosen.ToStateKey == nil,
		AssociationMaterialization: associationMaterialization,
		ActionResult:               actionResult,
		Violations:                 violations,
	}, nil
}

func (e *ActionExecutor) associationMaterializationForCreation(
	class model_class.Class,
	source CreationLinkSource,
	targetID *state.InstanceID,
) *AssociationMaterialization {
	if e.acIndex == nil || !e.acIndex.IsAssociationClass(class.Key) {
		return nil
	}
	if source.SourceID == nil || targetID == nil {
		return nil
	}

	linkInfo := e.acIndex.GetAssociationClassInfo(class.Key)
	if !linkInfo.Found {
		return nil
	}

	return &AssociationMaterialization{
		HostAssociationName: linkInfo.HostAssociationName,
		HostAssociationKey:  linkInfo.HostAssocKey,
		FromClassName:       linkInfo.FromClassName,
		FromClassKey:        linkInfo.FromClassKey,
		ToClassName:         linkInfo.ToClassName,
		ToClassKey:          linkInfo.ToClassKey,
		FromInstanceID:      *source.SourceID,
		ToInstanceID:        *targetID,
	}
}

// getInstanceCurrentState extracts the current state name from an instance.
func getInstanceCurrentState(instance *state.ClassInstance) string {
	if instance == nil {
		return ""
	}
	stateAttr := instance.GetAttribute("_state")
	if stateAttr == nil {
		return ""
	}
	if strObj, ok := stateAttr.(*object.String); ok {
		return strObj.Value()
	}
	return ""
}

// findCandidateTransitions finds transitions matching the event and current state.
func (e *ActionExecutor) findCandidateTransitions(
	class model_class.Class,
	event model_state.Event,
	instance *state.ClassInstance,
	currentStateName string,
) []model_state.Transition {
	var candidates []model_state.Transition
	for _, t := range class.Transitions {
		if t.EventKey != event.Key {
			continue
		}
		if instance == nil {
			// Creation: only transitions from initial state
			if t.FromStateKey == nil {
				candidates = append(candidates, t)
			}
		} else {
			// Normal: match current state
			if t.FromStateKey != nil {
				fromStateName := stateKeyToName(*t.FromStateKey, class)
				if fromStateName == currentStateName {
					candidates = append(candidates, t)
				}
			}
		}
	}
	return candidates
}

// evaluateGuards picks exactly one transition from candidates by evaluating guards.
func (e *ActionExecutor) evaluateGuards(
	candidates []model_state.Transition,
	class model_class.Class,
	instance *state.ClassInstance,
	event model_state.Event,
	currentStateName string,
) (*model_state.Transition, error) {
	if len(candidates) == 1 && candidates[0].GuardKey == nil {
		return &candidates[0], nil
	}

	var trueGuards []model_state.Transition
	for _, t := range candidates {
		if t.GuardKey == nil {
			trueGuards = append(trueGuards, t)
		} else {
			guard, ok := class.Guards[*t.GuardKey]
			if !ok {
				return nil, fmt.Errorf("guard %s not found in class %s", t.GuardKey.String(), class.Name)
			}
			passes, err := e.guardEvaluator.EvaluateGuard(guard, instance)
			if err != nil {
				return nil, fmt.Errorf("guard evaluation error: %w", err)
			}
			if passes {
				trueGuards = append(trueGuards, t)
			}
		}
	}

	if len(trueGuards) == 0 {
		return nil, fmt.Errorf("no guard is true for event %s from state %s on class %s (deadlock)", event.Name, currentStateName, class.Name)
	}
	if len(trueGuards) > 1 {
		return nil, fmt.Errorf("multiple guards true for event %s from state %s on class %s (non-determinism)", event.Name, currentStateName, class.Name)
	}
	return &trueGuards[0], nil
}

// handleCreation creates a new instance for a creation transition.
func (e *ActionExecutor) handleCreation(
	class model_class.Class,
	_ *state.ClassInstance,
	sourceAssocKey *identity.Key,
	sourceID *state.InstanceID,
	targetID *state.InstanceID,
) (*state.ClassInstance, error) {
	if e.acIndex != nil && e.acIndex.IsAssociationClass(class.Key) {
		return e.handleAssociationClassCreation(class, sourceID, targetID)
	}

	simState := e.bindingsBuilder.State()
	newAttrs, err := e.creationAttributes(class, simState)
	if err != nil {
		return nil, err
	}

	instance := simState.CreateInstance(class.Key, newAttrs)
	if err := e.linkPlainCreationOverAssociation(simState, sourceAssocKey, sourceID, instance.ID); err != nil {
		return nil, err
	}

	return instance, nil
}

func (e *ActionExecutor) creationAttributes(
	class model_class.Class,
	simState *state.SimulationState,
) (*object.Record, error) {
	newAttrs := object.NewRecord()
	if e.structuralCheckers == nil || e.structuralCheckers.Index == nil || e.rng == nil {
		return newAttrs, nil
	}

	indexInfo := e.structuralCheckers.Index.GetClassIndexInfo(class.Key)
	if indexInfo == nil {
		return newAttrs, nil
	}

	existingInstances := simState.InstancesByClass(class.Key)
	if err := generateIndexSafeValues(newAttrs, indexInfo, existingInstances, e.rng); err != nil {
		return nil, fmt.Errorf("failed to generate index-safe values for class %s: %w", class.Name, err)
	}
	return newAttrs, nil
}

func (e *ActionExecutor) linkPlainCreationOverAssociation(
	simState *state.SimulationState,
	sourceAssocKey *identity.Key,
	sourceID *state.InstanceID,
	newInstanceID state.InstanceID,
) error {
	if sourceAssocKey == nil || sourceID == nil {
		return nil
	}
	if e.acIndex != nil && e.acIndex.IsAssociationClassHost(*sourceAssocKey) {
		return fmt.Errorf(
			"host association %s requires an association-class instance; cannot link endpoints directly",
			sourceAssocKey.String(),
		)
	}
	simState.AddLink(*sourceAssocKey, *sourceID, newInstanceID)
	return nil
}

func (e *ActionExecutor) handleAssociationClassCreation(
	class model_class.Class,
	sourceID *state.InstanceID,
	targetID *state.InstanceID,
) (*state.ClassInstance, error) {
	if sourceID == nil || targetID == nil {
		return nil, fmt.Errorf(
			"association class %s Add requires both endpoint instances",
			class.Name,
		)
	}
	if e.acIndex == nil {
		return nil, fmt.Errorf("association class %s: no association-class index configured", class.Name)
	}

	linkInfo := e.acIndex.GetAssociationClassInfo(class.Key)
	if !linkInfo.Found {
		return nil, fmt.Errorf("association class %s: no host association metadata", class.Name)
	}

	simState := e.bindingsBuilder.State()
	fromInstance := simState.GetInstance(*sourceID)
	toInstance := simState.GetInstance(*targetID)
	if fromInstance == nil || toInstance == nil {
		return nil, fmt.Errorf("association class %s Add: endpoint instance not found", class.Name)
	}

	newAttrs := object.NewRecord()
	if e.structuralCheckers != nil && e.structuralCheckers.Index != nil && e.rng != nil {
		if indexInfo := e.structuralCheckers.Index.GetClassIndexInfo(class.Key); indexInfo != nil {
			existingInstances := simState.InstancesByClass(class.Key)
			if err := generateIndexSafeValues(newAttrs, indexInfo, existingInstances, e.rng); err != nil {
				return nil, fmt.Errorf("failed to generate index-safe values for class %s: %w", class.Name, err)
			}
		}
	}

	instance := simState.CreateInstance(class.Key, newAttrs)
	simState.AddAssociationLink(linkInfo.HostAssocKey, *sourceID, *targetID, instance.ID)

	return instance, nil
}

// executeTransitionAction executes the action associated with a transition (if any).
func (e *ActionExecutor) executeTransitionAction(
	chosen *model_state.Transition,
	class model_class.Class,
	instance *state.ClassInstance,
	eventParams map[string]object.Object,
) (*ActionResult, error) {
	if chosen.ActionKey == nil {
		return nil, nil //nolint:nilnil // no action to execute is a valid case
	}
	action, ok := class.Actions[*chosen.ActionKey]
	if !ok {
		return nil, fmt.Errorf("action %s not found in class %s", chosen.ActionKey.String(), class.Name)
	}
	result, err := e.ExecuteAction(action, instance, eventParams)
	if err != nil {
		return nil, fmt.Errorf("transition action error: %w", err)
	}
	return result, nil
}

// applyStateTransition applies the state change for a transition (including deletion).
func (e *ActionExecutor) applyStateTransition(
	chosen *model_state.Transition,
	class model_class.Class,
	instance *state.ClassInstance,
) (string, error) {
	simState := e.bindingsBuilder.State()

	if chosen.ToStateKey == nil {
		// To final state = object deletion
		if err := simState.DeleteInstance(instance.ID); err != nil {
			return "", fmt.Errorf("failed to delete instance %d: %w", instance.ID, err)
		}
		return "", nil
	}

	toStateName := stateKeyToName(*chosen.ToStateKey, class)
	instance.SetAttribute("_state", object.NewString(toStateName))
	if err := simState.SetStateMachineState(instance.ID, *chosen.ToStateKey); err != nil {
		return "", fmt.Errorf("failed to set state machine state: %w", err)
	}
	return toStateName, nil
}

// ValidateClassForSimulation checks that a class is valid for simulation.
// Every simulated class must have at least one defined state.
func ValidateClassForSimulation(class model_class.Class) error {
	if len(class.States) == 0 {
		return fmt.Errorf("class %s has no states defined; cannot simulate", class.Name)
	}
	return nil
}

// GetStateEnumValues returns the allowed _state values for a class.
func GetStateEnumValues(class model_class.Class) []string {
	values := make([]string, 0, len(class.States))
	for _, s := range class.States {
		values = append(values, s.Name)
	}
	return values
}

// stateKeyToName looks up a state key in the class and returns its name.
func stateKeyToName(stateKey identity.Key, class model_class.Class) string {
	if s, ok := class.States[stateKey]; ok {
		return s.Name
	}
	return stateKey.String()
}

// isTrueBoolean checks if an object is a TRUE boolean.
func isTrueBoolean(obj object.Object) bool {
	if obj == nil {
		return false
	}
	b, ok := obj.(*object.Boolean)
	if !ok {
		return false
	}
	return b.Value()
}

// createPostConditionViolation creates a violation from a deferred post-condition.
func createPostConditionViolation(pc DeferredPostCondition, message string) *invariants.ViolationError {
	if pc.SourceType == logicOwnerKindAction {
		return invariants.NewActionGuaranteeViolation(
			pc.SourceKey,
			pc.SourceName,
			pc.Index,
			pc.OriginalExpression,
			pc.InstanceID,
			message,
		)
	}
	return invariants.NewQueryGuaranteeViolation(
		pc.SourceKey,
		pc.SourceName,
		pc.Index,
		pc.OriginalExpression,
		pc.InstanceID,
		message,
	)
}
