package instance

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/model_bridge"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/schema"
)

type parsedAssociationInvariantItem struct {
	associationKey  identity.Key
	associationName string
	isLet           bool
	target          string
	expression      me.Expression
	originalIndex   int
	spec            string
}

// AssociationInvariantChecker evaluates association-authored invariants from the from-class anchor.
type AssociationInvariantChecker struct {
	byFromClass map[identity.Key][]parsedAssociationInvariantItem
}

// NewAssociationInvariantChecker builds association invariant metadata from schema.
func NewAssociationInvariantChecker(sch *schema.Schema) (*AssociationInvariantChecker, error) {
	checker := &AssociationInvariantChecker{
		byFromClass: make(map[identity.Key][]parsedAssociationInvariantItem),
	}

	for _, assoc := range sch.AssociationsWithInvariants() {
		items, err := parseAssociationInvariantItems(assoc)
		if err != nil {
			return nil, err
		}
		if len(items) > 0 {
			checker.byFromClass[assoc.FromClassKey] = append(checker.byFromClass[assoc.FromClassKey], items...)
		}
	}

	return checker, nil
}

func parseAssociationInvariantItems(assoc model_class.Association) ([]parsedAssociationInvariantItem, error) {
	items := make([]parsedAssociationInvariantItem, 0, len(assoc.Invariants))
	for i, inv := range assoc.Invariants {
		expr := inv.Spec.Expression
		if expr == nil {
			continue
		}
		isLet := inv.Type == model_logic.LogicTypeLet
		if !isLet && model_bridge.ContainsAnyPrimedME(expr) {
			return nil, fmt.Errorf(
				"association %q invariant %d must not contain primed variables: %s",
				assoc.Name, i, inv.Spec.Specification,
			)
		}
		items = append(items, parsedAssociationInvariantItem{
			associationKey:  assoc.Key,
			associationName: assoc.Name,
			isLet:           isLet,
			target:          inv.Target,
			expression:      expr,
			originalIndex:   i,
			spec:            inv.Spec.Specification,
		})
	}
	return items, nil
}

// CheckState validates association invariants for every from-class instance.
func (c *AssociationInvariantChecker) CheckState(
	simState *State,
	bindingsBuilder ExpressionBindings,
) ViolationErrors {
	var violations ViolationErrors
	simState.ForEachInstance(func(inst *Instance) {
		violations = append(violations, c.CheckInstance(inst, bindingsBuilder)...)
	})
	return violations
}

// CheckInstance validates association invariants for one from-class anchor instance.
func (c *AssociationInvariantChecker) CheckInstance(
	instance *Instance,
	bindingsBuilder ExpressionBindings,
) ViolationErrors {
	items, ok := c.byFromClass[instance.ClassKey]
	if !ok {
		return nil
	}

	var violations ViolationErrors
	bindings := bindingsBuilder.BuildForInstance(instance)

	for _, item := range items {
		if !item.isLet {
			continue
		}
		violations = append(violations, evalAssociationInvariantLet(instance, item, bindings)...)
	}

	for _, item := range items {
		if item.isLet {
			continue
		}
		violations = append(violations, evalAssociationInvariantAssessment(instance, item, bindings)...)
	}

	return violations
}

func evalAssociationInvariantLet(
	instance *Instance,
	item parsedAssociationInvariantItem,
	bindings *evaluator.Bindings,
) ViolationErrors {
	result := evaluator.Eval(item.expression, bindings)
	if result.IsError() {
		return ViolationErrors{newAssociationInvariantViolation(
			item.associationKey, item.associationName, instance.ID, item.originalIndex, item.spec,
			fmt.Sprintf("let evaluation error: %s", result.Error.Inspect()),
		)}
	}
	bindings.Set(item.target, result.Value, evaluator.NamespaceLocal)
	return nil
}

func evalAssociationInvariantAssessment(
	instance *Instance,
	item parsedAssociationInvariantItem,
	bindings *evaluator.Bindings,
) ViolationErrors {
	result := evaluator.Eval(item.expression, bindings)
	if result.Error != nil {
		return ViolationErrors{newAssociationInvariantViolation(
			item.associationKey, item.associationName, instance.ID, item.originalIndex, item.spec,
			fmt.Sprintf("evaluation error: %s", result.Error.Inspect()),
		)}
	}
	if isTrueBoolean(result.Value) {
		return nil
	}
	return ViolationErrors{newAssociationInvariantViolation(
		item.associationKey, item.associationName, instance.ID, item.originalIndex, item.spec,
		invariantAssessmentFailureMessage(result.Value),
	)}
}
