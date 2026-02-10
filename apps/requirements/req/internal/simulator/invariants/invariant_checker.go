package invariants

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/model_bridge"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/parser"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

// InvariantChecker evaluates TLA+ invariants against simulation state.
// It checks:
//   - Model-level invariants (Model.Invariants)
//   - Action post-condition guarantees
//   - Query post-condition guarantees
type InvariantChecker struct {
	// model is the requirements model containing invariant definitions
	model *req_model.Model

	// parsedInvariants caches parsed model invariant expressions
	parsedInvariants []ast.Expression

	// actionPostConditions maps action key to parsed post-condition expressions
	actionPostConditions map[identity.Key][]parsedGuarantee

	// queryPostConditions maps query key to parsed post-condition expressions
	queryPostConditions map[identity.Key][]parsedGuarantee

	// classNameMap maps class keys to class names for bindings
	classNameMap map[identity.Key]string
}

// parsedGuarantee holds a parsed guarantee expression with its metadata
type parsedGuarantee struct {
	expression ast.Expression
	index      int  // Index in the original guarantees array
	isPrimed   bool // True if this is a primed assignment, false if post-condition
}

// NewInvariantChecker creates a new invariant checker from a model.
// Returns an error if any TLA+ expression fails to parse.
func NewInvariantChecker(model *req_model.Model) (*InvariantChecker, error) {
	checker := &InvariantChecker{
		model:                model,
		parsedInvariants:     make([]ast.Expression, 0, len(model.Invariants)),
		actionPostConditions: make(map[identity.Key][]parsedGuarantee),
		queryPostConditions:  make(map[identity.Key][]parsedGuarantee),
		classNameMap:         make(map[identity.Key]string),
	}

	// Parse model invariants
	for i, inv := range model.Invariants {
		expr, err := parser.ParseExpression(inv.Specification)
		if err != nil {
			return nil, fmt.Errorf("failed to parse model invariant %d: %w", i, err)
		}
		if model_bridge.ContainsAnyPrimed(expr) {
			return nil, fmt.Errorf("model invariant %d must not contain primed variables: %s", i, inv.Specification)
		}
		checker.parsedInvariants = append(checker.parsedInvariants, expr)
	}

	// Iterate through all classes to collect actions, queries, and class names
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				checker.classNameMap[class.Key] = class.Name

				// Parse action post-conditions
				for _, action := range class.Actions {
					guarantees := make([]parsedGuarantee, 0)
					for i, gStr := range action.TlaGuarantees {
						expr, err := parser.ParseExpression(gStr)
						if err != nil {
							return nil, fmt.Errorf("failed to parse action %s guarantee %d: %w", action.Name, i, err)
						}

						kind := model_bridge.ClassifyGuarantee(expr)
						if kind == model_bridge.GuaranteePostCondition {
							guarantees = append(guarantees, parsedGuarantee{
								expression: expr,
								index:      i,
								isPrimed:   false,
							})
						}
					}
					if len(guarantees) > 0 {
						checker.actionPostConditions[action.Key] = guarantees
					}
				}

				// Parse query post-conditions
				for _, query := range class.Queries {
					guarantees := make([]parsedGuarantee, 0)
					for i, gStr := range query.TlaGuarantees {
						expr, err := parser.ParseExpression(gStr)
						if err != nil {
							return nil, fmt.Errorf("failed to parse query %s guarantee %d: %w", query.Name, i, err)
						}

						kind := model_bridge.ClassifyGuarantee(expr)
						if kind == model_bridge.GuaranteePostCondition {
							guarantees = append(guarantees, parsedGuarantee{
								expression: expr,
								index:      i,
								isPrimed:   false,
							})
						}
					}
					if len(guarantees) > 0 {
						checker.queryPostConditions[query.Key] = guarantees
					}
				}
			}
		}
	}

	return checker, nil
}

// CheckModelInvariants evaluates all model-level invariants against the current state.
// Returns violations for any invariant that evaluates to FALSE.
func (c *InvariantChecker) CheckModelInvariants(
	simState *state.SimulationState,
	bindingsBuilder *state.BindingsBuilder,
) ViolationList {
	var violations ViolationList

	bindings := bindingsBuilder.BuildWithClassInstances(c.classNameMap)

	for i, expr := range c.parsedInvariants {
		result := evaluator.Eval(expr, bindings)

		if result.Error != nil {
			violations = append(violations, NewModelInvariantViolation(
				i,
				c.model.Invariants[i].Specification,
				fmt.Sprintf("evaluation error: %s", result.Error.Inspect()),
			))
			continue
		}

		// Check if result is TRUE
		if !isTrueBoolean(result.Value) {
			var message string
			if result.Value == nil {
				message = "expression returned nil"
			} else {
				message = fmt.Sprintf("expression returned %s", result.Value.Inspect())
			}
			violations = append(violations, NewModelInvariantViolation(
				i,
				c.model.Invariants[i].Specification,
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
) ViolationList {
	guarantees, ok := c.actionPostConditions[actionKey]
	if !ok {
		return nil // No post-conditions for this action
	}

	var violations ViolationList

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
				expressionToString(g.expression),
				instance.ID,
				fmt.Sprintf("evaluation error: %s", result.Error.Inspect()),
			))
			continue
		}

		// Check if result is TRUE
		if !isTrueBoolean(result.Value) {
			var message string
			if result.Value == nil {
				message = "expression returned nil"
			} else {
				message = fmt.Sprintf("expression returned %s", result.Value.Inspect())
			}
			violations = append(violations, NewActionGuaranteeViolation(
				actionKey,
				actionName,
				g.index,
				expressionToString(g.expression),
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
) ViolationList {
	guarantees, ok := c.queryPostConditions[queryKey]
	if !ok {
		return nil // No post-conditions for this query
	}

	var violations ViolationList

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
				expressionToString(g.expression),
				instance.ID,
				fmt.Sprintf("evaluation error: %s", result.Error.Inspect()),
			))
			continue
		}

		// Check if result is TRUE
		if !isTrueBoolean(result.Value) {
			var message string
			if result.Value == nil {
				message = "expression returned nil"
			} else {
				message = fmt.Sprintf("expression returned %s", result.Value.Inspect())
			}
			violations = append(violations, NewQueryGuaranteeViolation(
				queryKey,
				queryName,
				g.index,
				expressionToString(g.expression),
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
) ViolationList {
	var violations ViolationList

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

// expressionToString converts an AST expression to a string representation.
// This is used for error messages when we don't have the original source.
func expressionToString(expr ast.Expression) string {
	if expr == nil {
		return "<nil>"
	}
	// Use the String() method if available, otherwise use the type name
	if stringer, ok := expr.(fmt.Stringer); ok {
		return stringer.String()
	}
	return fmt.Sprintf("<%T>", expr)
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

// GetModelInvariantCount returns the number of model invariants.
func (c *InvariantChecker) GetModelInvariantCount() int {
	return len(c.parsedInvariants)
}
