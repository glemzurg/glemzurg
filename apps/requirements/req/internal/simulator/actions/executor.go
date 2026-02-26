package actions

import (
	"fmt"
	"math/rand"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/invariants"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/model_bridge"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/parser"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

// ActionResult holds the result of executing an action.
type ActionResult struct {
	// InstanceID is the primary instance the action was executed on.
	InstanceID state.InstanceID

	// PrimedAssignments contains all state changes grouped by instance ID.
	PrimedAssignments map[state.InstanceID]map[string]object.Object

	// Violations contains any invariant violations detected after state changes.
	Violations invariants.ViolationList

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
	Violations invariants.ViolationList

	// Success is true if there are no violations.
	Success bool
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

	// ActionResult is the result of the action executed during the transition (if any).
	ActionResult *ActionResult

	// Violations contains any violations from the transition.
	Violations invariants.ViolationList
}

// ActionExecutor executes actions, queries, and transitions against simulation state.
type ActionExecutor struct {
	bindingsBuilder  *state.BindingsBuilder
	invariantChecker *invariants.InvariantChecker
	dataTypeChecker  *invariants.DataTypeChecker
	indexChecker     *invariants.IndexUniquenessChecker
	guardEvaluator   *GuardEvaluator
	rng              *rand.Rand
}

// NewActionExecutor creates a new action executor.
func NewActionExecutor(
	bindingsBuilder *state.BindingsBuilder,
	invariantChecker *invariants.InvariantChecker,
	dataTypeChecker *invariants.DataTypeChecker,
	indexChecker *invariants.IndexUniquenessChecker,
	guardEvaluator *GuardEvaluator,
	rng *rand.Rand,
) *ActionExecutor {
	return &ActionExecutor{
		bindingsBuilder:  bindingsBuilder,
		invariantChecker: invariantChecker,
		dataTypeChecker:  dataTypeChecker,
		indexChecker:     indexChecker,
		guardEvaluator:   guardEvaluator,
		rng:              rng,
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

	// Phase A: Execute the action chain (collecting primed values and post-conditions)
	if err := e.executeActionInContext(ctx, action, instance, parameters); err != nil {
		return nil, err
	}

	// Phase B: Apply ALL primed assignments from the entire chain
	simState := e.bindingsBuilder.State()
	for instanceID, primedFields := range ctx.GetAllPrimedAssignments() {
		for fieldName, value := range primedFields {
			if err := simState.UpdateInstanceField(instanceID, fieldName, value); err != nil {
				return nil, fmt.Errorf("failed to apply primed assignment %s on instance %d: %w", fieldName, instanceID, err)
			}
		}
	}

	// Phase C: Check ALL post-conditions from the entire chain
	var allViolations invariants.ViolationList

	for _, pc := range ctx.GetAllPostConditions() {
		targetInstance := simState.GetInstance(pc.InstanceID)
		if targetInstance == nil {
			continue // Instance may have been deleted
		}
		postBindings := e.bindingsBuilder.BuildForInstance(targetInstance)
		result := evaluator.Eval(pc.Expression, postBindings)

		if result.IsError() {
			v := createPostConditionViolation(pc, fmt.Sprintf("evaluation error: %s", result.Error.Inspect()))
			allViolations = append(allViolations, v)
			continue
		}

		if !isTrueBoolean(result.Value) {
			msg := "expression returned nil"
			if result.Value != nil {
				msg = fmt.Sprintf("expression returned %s", result.Value.Inspect())
			}
			v := createPostConditionViolation(pc, msg)
			allViolations = append(allViolations, v)
		}
	}

	// Phase C2: Check ALL safety rules from the entire chain
	for _, sr := range ctx.GetAllSafetyRules() {
		targetInstance := simState.GetInstance(sr.InstanceID)
		if targetInstance == nil {
			continue
		}
		safetyBindings := e.bindingsBuilder.BuildForInstance(targetInstance)
		result := evaluator.Eval(sr.Expression, safetyBindings)

		if result.IsError() {
			allViolations = append(allViolations, invariants.NewSafetyRuleViolation(
				sr.SourceKey, sr.SourceName, sr.Index, sr.OriginalExpression,
				sr.InstanceID, fmt.Sprintf("evaluation error: %s", result.Error.Inspect()),
			))
			continue
		}

		if !isTrueBoolean(result.Value) {
			msg := "expression returned nil"
			if result.Value != nil {
				msg = fmt.Sprintf("expression returned %s", result.Value.Inspect())
			}
			allViolations = append(allViolations, invariants.NewSafetyRuleViolation(
				sr.SourceKey, sr.SourceName, sr.Index, sr.OriginalExpression,
				sr.InstanceID, msg,
			))
		}
	}

	// Phase D: Check data type constraints on all mutated instances
	if e.dataTypeChecker != nil {
		for _, instanceID := range ctx.MutatedInstanceIDs() {
			targetInstance := simState.GetInstance(instanceID)
			if targetInstance == nil {
				continue
			}
			allViolations = append(allViolations, e.dataTypeChecker.CheckInstance(targetInstance)...)
		}
	}

	// Phase E: Check model invariants
	if e.invariantChecker != nil {
		allViolations = append(allViolations, e.invariantChecker.CheckModelInvariants(simState, e.bindingsBuilder)...)
	}

	// Phase F: Check index uniqueness
	if e.indexChecker != nil {
		allViolations = append(allViolations, e.indexChecker.CheckState(simState)...)
	}

	return &ActionResult{
		InstanceID:        instance.ID,
		PrimedAssignments: ctx.GetAllPrimedAssignments(),
		Violations:        allViolations,
		Success:           !allViolations.HasViolations(),
	}, nil
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

	// Step 1: Check preconditions (Requires)
	bindings := e.bindingsBuilder.BuildForInstanceWithVariables(instance, parameters)

	for i, req := range action.Requires {
		expr, err := parser.ParseExpression(req.Specification)
		if err != nil {
			return fmt.Errorf("action %s requires[%d] parse error: %w", action.Name, i, err)
		}

		if model_bridge.ContainsAnyPrimed(expr) {
			return fmt.Errorf("action %s requires[%d]: Requires must not contain primed variables", action.Name, i)
		}

		result := evaluator.Eval(expr, bindings)
		if result.IsError() {
			return fmt.Errorf("action %s requires[%d] evaluation error: %s", action.Name, i, result.Error.Inspect())
		}
		if !isTrueBoolean(result.Value) {
			return fmt.Errorf("action %s precondition failed: requires[%d] = %s", action.Name, i, req.Specification)
		}
	}

	// Step 2: Evaluate guarantees.
	for i, guar := range action.Guarantees {
		// Check re-entrancy constraint.
		if !ctx.CanMutate(instance.ID) {
			return fmt.Errorf("re-entrant mutation on instance %d in action %s: instance already has primed values from another action", instance.ID, action.Name)
		}

		if guar.Target != "" {
			// Target is set: use it directly as the field name.
			// The specification is the RHS value expression only.
			if guar.Specification == "" {
				return fmt.Errorf("action %s guarantee[%d]: target %q is set but specification is empty", action.Name, i, guar.Target)
			}
			expr, err := parser.ParseExpression(guar.Specification)
			if err != nil {
				return fmt.Errorf("action %s guarantee[%d] parse error: %w", action.Name, i, err)
			}
			rhsValue := evaluator.Eval(expr, bindings)
			if rhsValue.IsError() {
				return fmt.Errorf("action %s guarantee[%d] evaluation error: %s", action.Name, i, rhsValue.Error.Inspect())
			}
			if err := ctx.RecordPrimedAssignment(instance.ID, guar.Target, rhsValue.Value); err != nil {
				return fmt.Errorf("action %s guarantee[%d]: %w", action.Name, i, err)
			}
		} else {
			// Legacy: extract field name from primed TLA+ expression.
			expr, err := parser.ParseExpression(guar.Specification)
			if err != nil {
				return fmt.Errorf("action %s guarantee[%d] parse error: %w", action.Name, i, err)
			}
			kind := model_bridge.ClassifyGuarantee(expr)
			if kind != model_bridge.GuaranteePrimedAssignment {
				return fmt.Errorf("action %s guarantee[%d]: Guarantees must be primed assignments only (e.g., self.field' = expr); use SafetyRules for boolean assertions", action.Name, i)
			}
			fieldName, rhsValue, err := evaluatePrimedAssignment(expr, bindings)
			if err != nil {
				return fmt.Errorf("action %s guarantee[%d]: %w", action.Name, i, err)
			}
			if err := ctx.RecordPrimedAssignment(instance.ID, fieldName, rhsValue); err != nil {
				return fmt.Errorf("action %s guarantee[%d]: %w", action.Name, i, err)
			}
		}
	}

	// Step 3: Collect safety rules (must contain primed variables).
	for i, rule := range action.SafetyRules {
		expr, err := parser.ParseExpression(rule.Specification)
		if err != nil {
			return fmt.Errorf("action %s safety_rule[%d] parse error: %w", action.Name, i, err)
		}

		if !model_bridge.ContainsAnyPrimed(expr) {
			return fmt.Errorf("action %s safety_rule[%d]: SafetyRules must reference primed variables", action.Name, i)
		}

		ctx.AddSafetyRule(DeferredSafetyRule{
			Expression:         expr,
			InstanceID:         instance.ID,
			SourceKey:          action.Key,
			SourceName:         action.Name,
			Index:              i,
			OriginalExpression: rule.Specification,
		})
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

	outputs, err := e.executeQueryInContext(ctx, query, instance, parameters)
	if err != nil {
		return nil, err
	}

	// Check post-conditions
	var allViolations invariants.ViolationList
	simState := e.bindingsBuilder.State()

	for _, pc := range ctx.GetAllPostConditions() {
		targetInstance := simState.GetInstance(pc.InstanceID)
		if targetInstance == nil {
			continue
		}
		postBindings := e.bindingsBuilder.BuildForInstance(targetInstance)
		result := evaluator.Eval(pc.Expression, postBindings)

		if result.IsError() {
			v := createPostConditionViolation(pc, fmt.Sprintf("evaluation error: %s", result.Error.Inspect()))
			allViolations = append(allViolations, v)
			continue
		}

		if !isTrueBoolean(result.Value) {
			msg := "expression returned nil"
			if result.Value != nil {
				msg = fmt.Sprintf("expression returned %s", result.Value.Inspect())
			}
			v := createPostConditionViolation(pc, msg)
			allViolations = append(allViolations, v)
		}
	}

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

	// Step 1: Check preconditions
	bindings := e.bindingsBuilder.BuildForInstanceWithVariables(instance, parameters)

	for i, req := range query.Requires {
		expr, err := parser.ParseExpression(req.Specification)
		if err != nil {
			return nil, fmt.Errorf("query %s requires[%d] parse error: %w", query.Name, i, err)
		}

		if model_bridge.ContainsAnyPrimed(expr) {
			return nil, fmt.Errorf("query %s requires[%d]: Requires must not contain primed variables", query.Name, i)
		}

		result := evaluator.Eval(expr, bindings)
		if result.IsError() {
			return nil, fmt.Errorf("query %s requires[%d] evaluation error: %s", query.Name, i, result.Error.Inspect())
		}
		if !isTrueBoolean(result.Value) {
			return nil, fmt.Errorf("query %s precondition failed: requires[%d] = %s", query.Name, i, req.Specification)
		}
	}

	// Step 2: Evaluate guarantees
	outputs := make(map[string]object.Object)

	for i, guar := range query.Guarantees {
		if guar.Target != "" {
			// Target is set: use it directly as the output name.
			// The specification is the value expression only.
			if guar.Specification == "" {
				return nil, fmt.Errorf("query %s guarantee[%d]: target %q is set but specification is empty", query.Name, i, guar.Target)
			}
			expr, err := parser.ParseExpression(guar.Specification)
			if err != nil {
				return nil, fmt.Errorf("query %s guarantee[%d] parse error: %w", query.Name, i, err)
			}
			rhsValue := evaluator.Eval(expr, bindings)
			if rhsValue.IsError() {
				return nil, fmt.Errorf("query %s guarantee[%d] evaluation error: %s", query.Name, i, rhsValue.Error.Inspect())
			}
			outputs[guar.Target] = rhsValue.Value
		} else {
			// Legacy: extract output name from primed TLA+ expression.
			expr, err := parser.ParseExpression(guar.Specification)
			if err != nil {
				return nil, fmt.Errorf("query %s guarantee[%d] parse error: %w", query.Name, i, err)
			}
			kind := model_bridge.ClassifyGuarantee(expr)
			if kind == model_bridge.GuaranteePrimedAssignment {
				outputName, rhsValue, err := evaluatePrimedAssignment(expr, bindings)
				if err != nil {
					return nil, fmt.Errorf("query %s guarantee[%d]: %w", query.Name, i, err)
				}
				outputs[outputName] = rhsValue
			} else if kind == model_bridge.GuaranteePostCondition {
				ctx.AddPostCondition(DeferredPostCondition{
					Expression:         expr,
					InstanceID:         instance.ID,
					SourceKey:          query.Key,
					SourceName:         query.Name,
					SourceType:         "query",
					Index:              i,
					OriginalExpression: guar.Specification,
				})
			}
		}
	}

	return outputs, nil
}

// ExecuteTransition handles an event arriving at an instance.
// It finds matching transitions, evaluates guards, runs the action (if any),
// and sets the _state attribute. Handles creation and deletion.
func (e *ActionExecutor) ExecuteTransition(
	class model_class.Class,
	event model_state.Event,
	instance *state.ClassInstance, // nil for creation (from initial state)
	eventParams map[string]object.Object,
	sourceAssocKey *identity.Key, // association for creation linking
	sourceID *state.InstanceID, // parent instance for creation linking
) (*TransitionResult, error) {
	simState := e.bindingsBuilder.State()

	var currentStateName string
	if instance != nil {
		stateAttr := instance.GetAttribute("_state")
		if stateAttr != nil {
			if strObj, ok := stateAttr.(*object.String); ok {
				currentStateName = strObj.Value()
			}
		}
	}

	// Step 1: Find candidate transitions
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

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no transitions for event %s from state %s on class %s", event.Name, currentStateName, class.Name)
	}

	// Step 2: Evaluate guards to pick exactly one transition
	var chosen *model_state.Transition

	if len(candidates) == 1 && candidates[0].GuardKey == nil {
		chosen = &candidates[0]
	} else {
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
		chosen = &trueGuards[0]
	}

	// Step 3: Handle creation (FromStateKey == nil)
	if chosen.FromStateKey == nil {
		newAttrs := object.NewRecord()

		// Generate index-safe values if the class has indexes
		if e.indexChecker != nil && e.rng != nil {
			if indexInfo := e.indexChecker.GetClassIndexInfo(class.Key); indexInfo != nil {
				existingInstances := simState.InstancesByClass(class.Key)
				if err := generateIndexSafeValues(newAttrs, indexInfo, existingInstances, class, e.rng); err != nil {
					return nil, fmt.Errorf("failed to generate index-safe values for class %s: %w", class.Name, err)
				}
			}
		}

		instance = simState.CreateInstance(class.Key, newAttrs)

		// Link to parent over the association
		if sourceAssocKey != nil && sourceID != nil {
			simState.AddLink(*sourceAssocKey, *sourceID, instance.ID)
		}
	}

	// Step 4: Execute the action (if any)
	var actionResult *ActionResult
	if chosen.ActionKey != nil {
		action, ok := class.Actions[*chosen.ActionKey]
		if !ok {
			return nil, fmt.Errorf("action %s not found in class %s", chosen.ActionKey.String(), class.Name)
		}
		var err error
		actionResult, err = e.ExecuteAction(action, instance, eventParams)
		if err != nil {
			return nil, fmt.Errorf("transition action error: %w", err)
		}
	}

	// Step 5: Apply state transition
	var toStateName string
	var violations invariants.ViolationList

	if actionResult != nil {
		violations = actionResult.Violations
	}

	if chosen.ToStateKey == nil {
		// To final state = object deletion
		if err := simState.DeleteInstance(instance.ID); err != nil {
			return nil, fmt.Errorf("failed to delete instance %d: %w", instance.ID, err)
		}
	} else {
		toStateName = stateKeyToName(*chosen.ToStateKey, class)
		instance.SetAttribute("_state", object.NewString(toStateName))
		if err := simState.SetStateMachineState(instance.ID, *chosen.ToStateKey); err != nil {
			return nil, fmt.Errorf("failed to set state machine state: %w", err)
		}
	}

	return &TransitionResult{
		InstanceID:    instance.ID,
		FromState:     currentStateName,
		ToState:       toStateName,
		EventKey:      event.Key,
		TransitionKey: chosen.Key,
		WasCreation:   chosen.FromStateKey == nil,
		WasDeletion:   chosen.ToStateKey == nil,
		ActionResult:  actionResult,
		Violations:    violations,
	}, nil
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

// evaluatePrimedAssignment handles a guarantee expression that the classifier
// identified as a primed assignment (e.g., "self.count' = self.count + 1").
// The parser produces a BinaryEquality node for these. We extract the field name
// from the primed LHS and evaluate the RHS to get the new value.
//
// Supported LHS patterns:
//   - Primed(FieldAccess(Identifier("self"), "field")) → field name is "field"
//   - Primed(Identifier("name")) → field name is "name" (for query outputs like result')
func evaluatePrimedAssignment(
	expr ast.Expression,
	bindings *evaluator.Bindings,
) (string, object.Object, error) {
	eq, ok := expr.(*ast.BinaryEquality)
	if !ok || eq.Operator != "=" {
		return "", nil, fmt.Errorf("expected primed assignment (x' = expr), got %T", expr)
	}

	// Extract the field name from the primed LHS
	fieldName, err := extractPrimedFieldName(eq.Left)
	if err != nil {
		return "", nil, fmt.Errorf("invalid primed LHS: %w", err)
	}

	// Evaluate the RHS
	result := evaluator.Eval(eq.Right, bindings)
	if result.IsError() {
		return "", nil, fmt.Errorf("evaluation error: %s", result.Error.Inspect())
	}

	return fieldName, result.Value, nil
}

// extractPrimedFieldName extracts the target field name from a primed LHS expression.
// For "self.count'" the AST is Primed(FieldAccess(Identifier("self"), "count")) → returns "count".
// For "result'" the AST is Primed(Identifier("result")) → returns "result".
func extractPrimedFieldName(expr ast.Expression) (string, error) {
	primed, ok := expr.(*ast.Primed)
	if !ok {
		return "", fmt.Errorf("expected Primed expression, got %T", expr)
	}

	switch base := primed.Base.(type) {
	case *ast.FieldAccess:
		// self.field' → field name is the Member
		return base.Member, nil

	case *ast.Identifier:
		// result' → field name is the identifier value
		return base.Value, nil

	default:
		return "", fmt.Errorf("unsupported primed base type: %T", primed.Base)
	}
}

// createPostConditionViolation creates a violation from a deferred post-condition.
func createPostConditionViolation(pc DeferredPostCondition, message string) *invariants.Violation {
	if pc.SourceType == "action" {
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
