package invariants

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/model_bridge"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

// _EXPRESSION_RETURNED_NIL is the error message used when an expression evaluates to nil.
const _EXPRESSION_RETURNED_NIL = "expression returned nil"

// InvariantChecker evaluates invariants against simulation state.
// It checks:
//   - Model-level invariants (Model.Invariants)
//   - Action post-condition guarantees
//   - Query post-condition guarantees
type InvariantChecker struct {
	// model is the requirements model containing invariant definitions
	model *core.Model

	// parsedInvariantItems caches pre-lowered model invariant items (both let and assessment).
	parsedInvariantItems []parsedInvariantItem

	// actionPostConditions maps action key to post-condition expressions
	actionPostConditions map[identity.Key][]parsedGuarantee

	// queryPostConditions maps query key to post-condition expressions
	queryPostConditions map[identity.Key][]parsedGuarantee

	// classNameMap maps class keys to class names for bindings
	classNameMap map[identity.Key]string
}

// parsedInvariantItem holds a pre-lowered invariant or let expression with metadata.
type parsedInvariantItem struct {
	isLet         bool          // True if this is a LogicTypeLet item.
	target        string        // Only set if isLet is true.
	expression    me.Expression // The lowered expression.
	originalIndex int           // Index in the original Model.Invariants slice.
	spec          string        // Original specification string for error messages.
}

// parsedGuarantee holds a lowered guarantee expression with its metadata.
type parsedGuarantee struct {
	expression me.Expression
	spec       string // original specification string for error messages
	index      int    // Index in the original guarantees array
}

// NewInvariantChecker creates a new invariant checker from a model.
// The model's ExpressionSpec.Expression fields must be populated
// (via parse functions passed to constructors).
func NewInvariantChecker(model *core.Model) (*InvariantChecker, error) {
	checker := &InvariantChecker{
		model:                model,
		parsedInvariantItems: make([]parsedInvariantItem, 0, len(model.Invariants)),
		actionPostConditions: make(map[identity.Key][]parsedGuarantee),
		queryPostConditions:  make(map[identity.Key][]parsedGuarantee),
		classNameMap:         make(map[identity.Key]string),
	}

	// Load model invariants from pre-parsed expressions.
	// Invariants with nil Expression (unparsed or empty) are silently skipped.
	for i, inv := range model.Invariants {
		expr := inv.Spec.Expression
		if expr == nil {
			continue // Skip unparsed or empty specs
		}
		isLet := inv.Type == model_logic.LogicTypeLet
		// Only non-let invariants are checked for primed variables
		if !isLet && model_bridge.ContainsAnyPrimedME(expr) {
			return nil, fmt.Errorf("model invariant %d must not contain primed variables: %s", i, inv.Spec.Specification)
		}
		checker.parsedInvariantItems = append(checker.parsedInvariantItems, parsedInvariantItem{
			isLet:         isLet,
			target:        inv.Target,
			expression:    expr,
			originalIndex: i,
			spec:          inv.Spec.Specification,
		})
	}

	// Iterate through all classes to collect class names
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				checker.classNameMap[class.Key] = class.Name
			}
		}
	}

	return checker, nil
}

// CheckModelInvariants evaluates all model-level invariants against the current state.
// Returns violations for any invariant that evaluates to FALSE.
func (c *InvariantChecker) CheckModelInvariants(
	_ *state.SimulationState,
	bindingsBuilder *state.BindingsBuilder,
) ViolationErrors {
	var violations ViolationErrors

	bindings := bindingsBuilder.BuildWithClassInstances(c.classNameMap)

	// Pass 1: Evaluate all let items in order, setting their targets in bindings.
	for _, item := range c.parsedInvariantItems {
		if !item.isLet {
			continue
		}
		result := evaluator.Eval(item.expression, bindings)
		if result.IsError() {
			violations = append(violations, NewModelInvariantViolation(
				item.originalIndex,
				item.spec,
				fmt.Sprintf("let evaluation error: %s", result.Error.Inspect()),
			))
			continue
		}
		bindings.Set(item.target, result.Value, evaluator.NamespaceLocal)
	}

	// Pass 2: Evaluate all non-let (assessment) items with let bindings available.
	for _, item := range c.parsedInvariantItems {
		if item.isLet {
			continue
		}
		result := evaluator.Eval(item.expression, bindings)

		if result.Error != nil {
			violations = append(violations, NewModelInvariantViolation(
				item.originalIndex,
				item.spec,
				fmt.Sprintf("evaluation error: %s", result.Error.Inspect()),
			))
			continue
		}

		// Check if result is TRUE
		if !isTrueBoolean(result.Value) {
			var message string
			if result.Value == nil {
				message = _EXPRESSION_RETURNED_NIL
			} else {
				message = fmt.Sprintf("expression returned %s", result.Value.Inspect())
			}
			violations = append(violations, NewModelInvariantViolation(
				item.originalIndex,
				item.spec,
				message,
			))
		}
	}

	return violations
}

// CheckActionPostConditions evaluates post-condition guarantees for an action.
// This should be called after the action's state changes have been applied.
// Returns violations for any post-condition that evaluates to FALSE.
func (c *InvariantChecker) CheckActionPostConditions(
	actionKey identity.Key,
	actionName string,
	instance *state.ClassInstance,
	bindingsBuilder *state.BindingsBuilder,
	additionalBindings map[string]object.Object,
) ViolationErrors {
	guarantees, ok := c.actionPostConditions[actionKey]
	if !ok {
		return nil // No post-conditions for this action
	}

	var violations ViolationErrors

	// Build bindings with self and any additional variables
	var bindings *evaluator.Bindings
	if len(additionalBindings) > 0 {
		bindings = bindingsBuilder.BuildForInstanceWithVariables(instance, additionalBindings)
	} else {
		bindings = bindingsBuilder.BuildForInstance(instance)
	}

	for _, g := range guarantees {
		result := evaluator.Eval(g.expression, bindings)

		if result.Error != nil {
			violations = append(violations, NewActionGuaranteeViolation(
				actionKey,
				actionName,
				g.index,
				g.spec,
				instance.ID,
				fmt.Sprintf("evaluation error: %s", result.Error.Inspect()),
			))
			continue
		}

		// Check if result is TRUE
		if !isTrueBoolean(result.Value) {
			var message string
			if result.Value == nil {
				message = _EXPRESSION_RETURNED_NIL
			} else {
				message = fmt.Sprintf("expression returned %s", result.Value.Inspect())
			}
			violations = append(violations, NewActionGuaranteeViolation(
				actionKey,
				actionName,
				g.index,
				g.spec,
				instance.ID,
				message,
			))
		}
	}

	return violations
}

// CheckQueryPostConditions evaluates post-condition guarantees for a query.
// This should be called after the query has been executed.
// Returns violations for any post-condition that evaluates to FALSE.
func (c *InvariantChecker) CheckQueryPostConditions(
	queryKey identity.Key,
	queryName string,
	instance *state.ClassInstance,
	bindingsBuilder *state.BindingsBuilder,
	additionalBindings map[string]object.Object,
) ViolationErrors {
	guarantees, ok := c.queryPostConditions[queryKey]
	if !ok {
		return nil // No post-conditions for this query
	}

	var violations ViolationErrors

	// Build bindings with self and any additional variables
	var bindings *evaluator.Bindings
	if len(additionalBindings) > 0 {
		bindings = bindingsBuilder.BuildForInstanceWithVariables(instance, additionalBindings)
	} else {
		bindings = bindingsBuilder.BuildForInstance(instance)
	}

	for _, g := range guarantees {
		result := evaluator.Eval(g.expression, bindings)

		if result.Error != nil {
			violations = append(violations, NewQueryGuaranteeViolation(
				queryKey,
				queryName,
				g.index,
				g.spec,
				instance.ID,
				fmt.Sprintf("evaluation error: %s", result.Error.Inspect()),
			))
			continue
		}

		// Check if result is TRUE
		if !isTrueBoolean(result.Value) {
			var message string
			if result.Value == nil {
				message = _EXPRESSION_RETURNED_NIL
			} else {
				message = fmt.Sprintf("expression returned %s", result.Value.Inspect())
			}
			violations = append(violations, NewQueryGuaranteeViolation(
				queryKey,
				queryName,
				g.index,
				g.spec,
				instance.ID,
				message,
			))
		}
	}

	return violations
}

// CheckAllInvariants is a convenience method that checks:
//   - Model invariants
//   - Data type constraints (requires a DataTypeChecker)
//
// This is typically called after each state change.
func (c *InvariantChecker) CheckAllInvariants(
	simState *state.SimulationState,
	bindingsBuilder *state.BindingsBuilder,
	dataTypeChecker *DataTypeChecker,
	indexChecker *IndexUniquenessChecker,
) ViolationErrors {
	var violations ViolationErrors

	// Check model invariants
	modelViolations := c.CheckModelInvariants(simState, bindingsBuilder)
	violations = append(violations, modelViolations...)

	// Check data type constraints
	if dataTypeChecker != nil {
		dataTypeViolations := dataTypeChecker.CheckState(simState)
		violations = append(violations, dataTypeViolations...)
	}

	// Check index uniqueness constraints
	if indexChecker != nil {
		indexViolations := indexChecker.CheckState(simState)
		violations = append(violations, indexViolations...)
	}

	return violations
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

// GetActionPostConditionCount returns the number of post-conditions for an action.
func (c *InvariantChecker) GetActionPostConditionCount(actionKey identity.Key) int {
	guarantees, ok := c.actionPostConditions[actionKey]
	if !ok {
		return 0
	}
	return len(guarantees)
}

// GetQueryPostConditionCount returns the number of post-conditions for a query.
func (c *InvariantChecker) GetQueryPostConditionCount(queryKey identity.Key) int {
	guarantees, ok := c.queryPostConditions[queryKey]
	if !ok {
		return 0
	}
	return len(guarantees)
}

// GetModelInvariantCount returns the number of model invariants (excluding let items).
func (c *InvariantChecker) GetModelInvariantCount() int {
	count := 0
	for _, item := range c.parsedInvariantItems {
		if !item.isLet {
			count++
		}
	}
	return count
}
