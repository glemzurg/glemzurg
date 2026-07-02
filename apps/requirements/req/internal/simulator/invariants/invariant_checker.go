package invariants

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
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
//   - Class-level invariants (Class.Invariants, per instance)
//   - Attribute invariants (per instance; skipped when a nullable attribute is unset)
//   - Action post-condition guarantees
//   - Query post-condition guarantees
type InvariantChecker struct {
	// model is the requirements model containing invariant definitions
	model *core.Model

	// parsedInvariantItems caches pre-lowered model invariant items (both let and assessment).
	parsedInvariantItems []parsedInvariantItem

	// parsedClassInvariants maps class key to pre-lowered class invariant items.
	parsedClassInvariants map[identity.Key][]parsedClassInvariantItem

	// parsedAttributeInvariants maps class key to per-attribute invariant items.
	parsedAttributeInvariants map[identity.Key][]parsedAttributeInvariantItem

	// actionPostConditions maps action key to post-condition expressions
	actionPostConditions map[identity.Key][]parsedGuarantee

	// queryPostConditions maps query key to post-condition expressions
	queryPostConditions map[identity.Key][]parsedGuarantee

	// classNameMap maps class keys to class names for bindings
	classNameMap map[identity.Key]string

	// classAttributes maps class keys to attribute definitions for nullable checks.
	classAttributes map[identity.Key][]model_class.Attribute
}

// parsedAttributeInvariantItem holds a pre-lowered attribute invariant with metadata.
type parsedAttributeInvariantItem struct {
	attributeFieldKey string // YAML field key (attribute SubKey) for instance lookup.
	attributeName     string // Display name for violation messages.
	isLet             bool
	target            string
	expression        me.Expression
	originalIndex     int
	spec              string
}

// parsedClassInvariantItem holds a pre-lowered class invariant with metadata.
type parsedClassInvariantItem struct {
	isLet         bool
	target        string
	expression    me.Expression
	originalIndex int
	spec          string
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

// ClassNameMap returns class keys mapped to display names for class-set bindings.
func (c *InvariantChecker) ClassNameMap() map[identity.Key]string {
	return c.classNameMap
}

// NewInvariantChecker creates a new invariant checker from a model.
// The model's ExpressionSpec.Expression fields must be populated
// (via parse functions passed to constructors).
func NewInvariantChecker(model *core.Model) (*InvariantChecker, error) {
	checker := &InvariantChecker{
		model:                     model,
		parsedInvariantItems:      make([]parsedInvariantItem, 0, len(model.Invariants)),
		parsedClassInvariants:     make(map[identity.Key][]parsedClassInvariantItem),
		parsedAttributeInvariants: make(map[identity.Key][]parsedAttributeInvariantItem),
		actionPostConditions:      make(map[identity.Key][]parsedGuarantee),
		queryPostConditions:       make(map[identity.Key][]parsedGuarantee),
		classNameMap:              make(map[identity.Key]string),
		classAttributes:           make(map[identity.Key][]model_class.Attribute),
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

	// Iterate through all classes to collect class names and class invariants.
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				checker.classNameMap[class.Key] = class.Name
				checker.classAttributes[class.Key] = class.Attributes
				if err := checker.loadClassInvariants(class); err != nil {
					return nil, err
				}
				if err := checker.loadAttributeInvariants(class); err != nil {
					return nil, err
				}
			}
		}
	}

	return checker, nil
}

func (c *InvariantChecker) loadClassInvariants(class model_class.Class) error {
	if len(class.Invariants) == 0 {
		return nil
	}

	items := make([]parsedClassInvariantItem, 0, len(class.Invariants))
	for i, inv := range class.Invariants {
		expr := inv.Spec.Expression
		if expr == nil {
			continue
		}
		isLet := inv.Type == model_logic.LogicTypeLet
		if !isLet && model_bridge.ContainsAnyPrimedME(expr) {
			return fmt.Errorf("class %s invariant %d must not contain primed variables: %s", class.Name, i, inv.Spec.Specification)
		}
		items = append(items, parsedClassInvariantItem{
			isLet:         isLet,
			target:        inv.Target,
			expression:    expr,
			originalIndex: i,
			spec:          inv.Spec.Specification,
		})
	}
	if len(items) > 0 {
		c.parsedClassInvariants[class.Key] = items
	}
	return nil
}

func (c *InvariantChecker) loadAttributeInvariants(class model_class.Class) error {
	var items []parsedAttributeInvariantItem
	for _, attr := range class.Attributes {
		for i, inv := range attr.Invariants {
			expr := inv.Spec.Expression
			if expr == nil {
				continue
			}
			isLet := inv.Type == model_logic.LogicTypeLet
			if !isLet && model_bridge.ContainsAnyPrimedME(expr) {
				return fmt.Errorf("class %s attribute %q invariant %d must not contain primed variables: %s", class.Name, attr.Name, i, inv.Spec.Specification)
			}
			items = append(items, parsedAttributeInvariantItem{
				attributeFieldKey: attr.Key.SubKey,
				attributeName:     attr.Name,
				isLet:             isLet,
				target:            inv.Target,
				expression:        expr,
				originalIndex:     i,
				spec:              inv.Spec.Specification,
			})
		}
	}
	if len(items) > 0 {
		c.parsedAttributeInvariants[class.Key] = items
	}
	return nil
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

// CheckClassInvariants evaluates class-level invariants for every instance in state.
func (c *InvariantChecker) CheckClassInvariants(
	simState *state.SimulationState,
	bindingsBuilder *state.BindingsBuilder,
) ViolationErrors {
	var violations ViolationErrors

	for _, instance := range simState.AllInstances() {
		items, ok := c.parsedClassInvariants[instance.ClassKey]
		if !ok {
			continue
		}
		violations = append(violations, c.checkClassInvariantsForInstance(instance, items, bindingsBuilder)...)
	}

	return violations
}

func (c *InvariantChecker) checkClassInvariantsForInstance(
	instance *state.ClassInstance,
	items []parsedClassInvariantItem,
	bindingsBuilder *state.BindingsBuilder,
) ViolationErrors {
	var violations ViolationErrors
	bindings := bindingsBuilder.BuildForInstance(instance)

	for _, item := range items {
		if !item.isLet {
			continue
		}
		result := evaluator.Eval(item.expression, bindings)
		if result.IsError() {
			violations = append(violations, NewClassInvariantViolation(
				instance.ClassKey, instance.ID, item.originalIndex, item.spec,
				fmt.Sprintf("let evaluation error: %s", result.Error.Inspect()),
			))
			continue
		}
		bindings.Set(item.target, result.Value, evaluator.NamespaceLocal)
	}

	for _, item := range items {
		if item.isLet {
			continue
		}
		result := evaluator.Eval(item.expression, bindings)
		if result.Error != nil {
			violations = append(violations, NewClassInvariantViolation(
				instance.ClassKey, instance.ID, item.originalIndex, item.spec,
				fmt.Sprintf("evaluation error: %s", result.Error.Inspect()),
			))
			continue
		}
		if !isTrueBoolean(result.Value) {
			var message string
			if result.Value == nil {
				message = _EXPRESSION_RETURNED_NIL
			} else {
				message = fmt.Sprintf("expression returned %s", result.Value.Inspect())
			}
			violations = append(violations, NewClassInvariantViolation(
				instance.ClassKey, instance.ID, item.originalIndex, item.spec, message,
			))
		}
	}

	return violations
}

// CheckAttributeInvariants evaluates attribute invariants for every instance in state.
// Nullable attributes with no value skip invariant checks.
func (c *InvariantChecker) CheckAttributeInvariants(
	simState *state.SimulationState,
	bindingsBuilder *state.BindingsBuilder,
) ViolationErrors {
	var violations ViolationErrors

	for _, instance := range simState.AllInstances() {
		items, ok := c.parsedAttributeInvariants[instance.ClassKey]
		if !ok {
			continue
		}
		nullableByFieldKey := attributeNullableByFieldKey(c.classAttributes[instance.ClassKey])
		violations = append(violations, checkAttributeInvariantsForInstance(instance, items, nullableByFieldKey, bindingsBuilder)...)
	}

	return violations
}

func attributeNullableByFieldKey(attrs []model_class.Attribute) map[string]bool {
	nullable := make(map[string]bool, len(attrs))
	for _, attr := range attrs {
		nullable[attr.Key.SubKey] = attr.Nullable
	}
	return nullable
}

func skipNullableUnsetAttribute(
	nullableByFieldKey map[string]bool,
	instance *state.ClassInstance,
	attributeFieldKey string,
) bool {
	return nullableByFieldKey[attributeFieldKey] && object.IsNull(instance.GetAttribute(attributeFieldKey))
}

func checkAttributeInvariantsForInstance(
	instance *state.ClassInstance,
	items []parsedAttributeInvariantItem,
	nullableByFieldKey map[string]bool,
	bindingsBuilder *state.BindingsBuilder,
) ViolationErrors {
	var violations ViolationErrors
	bindings := bindingsBuilder.BuildForInstance(instance)

	for _, item := range items {
		if skipNullableUnsetAttribute(nullableByFieldKey, instance, item.attributeFieldKey) || !item.isLet {
			continue
		}
		violations = append(violations, evalAttributeInvariantLet(instance, item, bindings)...)
	}

	for _, item := range items {
		if skipNullableUnsetAttribute(nullableByFieldKey, instance, item.attributeFieldKey) || item.isLet {
			continue
		}
		violations = append(violations, evalAttributeInvariantAssessment(instance, item, bindings)...)
	}

	return violations
}

func evalAttributeInvariantLet(
	instance *state.ClassInstance,
	item parsedAttributeInvariantItem,
	bindings *evaluator.Bindings,
) ViolationErrors {
	result := evaluator.Eval(item.expression, bindings)
	if result.IsError() {
		return ViolationErrors{NewAttributeInvariantViolation(
			instance.ClassKey, instance.ID, item.attributeName, item.originalIndex, item.spec,
			fmt.Sprintf("let evaluation error: %s", result.Error.Inspect()),
		)}
	}
	bindings.Set(item.target, result.Value, evaluator.NamespaceLocal)
	return nil
}

func evalAttributeInvariantAssessment(
	instance *state.ClassInstance,
	item parsedAttributeInvariantItem,
	bindings *evaluator.Bindings,
) ViolationErrors {
	result := evaluator.Eval(item.expression, bindings)
	if result.Error != nil {
		return ViolationErrors{NewAttributeInvariantViolation(
			instance.ClassKey, instance.ID, item.attributeName, item.originalIndex, item.spec,
			fmt.Sprintf("evaluation error: %s", result.Error.Inspect()),
		)}
	}
	if isTrueBoolean(result.Value) {
		return nil
	}
	return ViolationErrors{NewAttributeInvariantViolation(
		instance.ClassKey, instance.ID, item.attributeName, item.originalIndex, item.spec,
		invariantAssessmentFailureMessage(result.Value),
	)}
}

func invariantAssessmentFailureMessage(value object.Object) string {
	if value == nil {
		return _EXPRESSION_RETURNED_NIL
	}
	return fmt.Sprintf("expression returned %s", value.Inspect())
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

	// Check class invariants
	classViolations := c.CheckClassInvariants(simState, bindingsBuilder)
	violations = append(violations, classViolations...)

	// Check attribute invariants
	attrViolations := c.CheckAttributeInvariants(simState, bindingsBuilder)
	violations = append(violations, attrViolations...)

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
