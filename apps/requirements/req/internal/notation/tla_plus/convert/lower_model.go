package convert

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/parser"
)

// LowerModel walks the entire model tree, parsing and lowering every ExpressionSpec
// that has a TLA+ specification string into a model_expression.Expression.
// It returns the first error encountered, leaving the model partially populated.
func LowerModel(model *core.Model) error {
	// Build model-level lookup maps for global functions and named sets.
	globalFunctions := BuildGlobalFunctionMap(model)
	namedSets := BuildNamedSetMap(model)

	// Build cross-class action lookup (AllActions) across the entire model.
	allActions := BuildAllActionsMap(model)

	// 1. Lower model-level invariants (no class context).
	modelCtx := &LowerContext{
		GlobalFunctions: globalFunctions,
		NamedSets:       namedSets,
		AllActions:      allActions,
	}
	for i := range model.Invariants {
		if err := lowerLogicSpec(&model.Invariants[i].Spec, modelCtx); err != nil {
			return fmt.Errorf("model invariant %d: %w", i, err)
		}
	}

	// 2. Lower global functions.
	for gfKey, gf := range model.GlobalFunctions {
		params := make(map[string]bool, len(gf.Parameters))
		for _, p := range gf.Parameters {
			params[p] = true
		}
		gfCtx := &LowerContext{
			GlobalFunctions: globalFunctions,
			NamedSets:       namedSets,
			AllActions:      allActions,
			Parameters:      params,
		}
		if err := lowerLogicSpec(&gf.Logic.Spec, gfCtx); err != nil {
			return fmt.Errorf("global function %q: %w", gfKey.String(), err)
		}
		model.GlobalFunctions[gfKey] = gf
	}

	// 3. Lower named sets.
	for nsKey, ns := range model.NamedSets {
		if err := lowerLogicSpec(&ns.Spec, modelCtx); err != nil {
			return fmt.Errorf("named set %q: %w", nsKey.String(), err)
		}
		model.NamedSets[nsKey] = ns
	}

	// 4. Walk domains → subdomains → classes.
	for dKey, domain := range model.Domains {
		for sKey, subdomain := range domain.Subdomains {
			for cKey, class := range subdomain.Classes {
				if err := lowerClass(&class, globalFunctions, namedSets, allActions); err != nil {
					return fmt.Errorf("class %q: %w", cKey.String(), err)
				}
				subdomain.Classes[cKey] = class
			}
			domain.Subdomains[sKey] = subdomain
		}
		model.Domains[dKey] = domain
	}

	return nil
}

// lowerClass populates all ExpressionSpec.Expression fields within a class.
func lowerClass(
	class *model_class.Class,
	globalFunctions map[string]identity.Key,
	namedSets map[string]identity.Key,
	allActions map[string]identity.Key,
) error {
	// Build class-level context maps.
	attrNames := BuildAttributeNameMap(class)
	actionNames := BuildActionNameMap(class)
	queryNames := BuildQueryNameMap(class)

	classCtx := &LowerContext{
		ClassKey:        class.Key,
		AttributeNames:  attrNames,
		ActionNames:     actionNames,
		QueryNames:      queryNames,
		GlobalFunctions: globalFunctions,
		NamedSets:       namedSets,
		AllActions:      allActions,
	}

	// Class invariants.
	for i := range class.Invariants {
		if err := lowerLogicSpec(&class.Invariants[i].Spec, classCtx); err != nil {
			return fmt.Errorf("class invariant %d: %w", i, err)
		}
	}

	// Attributes: derivation policy and invariants.
	for aKey, attr := range class.Attributes {
		if err := lowerAttribute(&attr, classCtx); err != nil {
			return fmt.Errorf("attribute %q: %w", aKey.String(), err)
		}
		class.Attributes[aKey] = attr
	}

	// Guards.
	for gKey, guard := range class.Guards {
		if err := lowerLogicSpec(&guard.Logic.Spec, classCtx); err != nil {
			return fmt.Errorf("guard %q: %w", gKey.String(), err)
		}
		class.Guards[gKey] = guard
	}

	// Actions.
	for actKey, action := range class.Actions {
		if err := lowerAction(&action, classCtx); err != nil {
			return fmt.Errorf("action %q: %w", actKey.String(), err)
		}
		class.Actions[actKey] = action
	}

	// Queries.
	for qKey, query := range class.Queries {
		if err := lowerQuery(&query, classCtx); err != nil {
			return fmt.Errorf("query %q: %w", qKey.String(), err)
		}
		class.Queries[qKey] = query
	}

	return nil
}

func lowerAttribute(attr *model_class.Attribute, baseCtx *LowerContext) error {
	if attr.DerivationPolicy != nil {
		if err := lowerLogicSpec(&attr.DerivationPolicy.Spec, baseCtx); err != nil {
			return fmt.Errorf("derivation policy: %w", err)
		}
	}
	for i := range attr.Invariants {
		if err := lowerLogicSpec(&attr.Invariants[i].Spec, baseCtx); err != nil {
			return fmt.Errorf("invariant %d: %w", i, err)
		}
	}
	return nil
}

func lowerAction(action *model_state.Action, baseCtx *LowerContext) error {
	// Actions have parameters that become local vars.
	ctx := ContextWithParameters(baseCtx, action.Parameters)

	for i := range action.Requires {
		if err := lowerLogicSpec(&action.Requires[i].Spec, ctx); err != nil {
			return fmt.Errorf("require %d: %w", i, err)
		}
	}
	for i := range action.Guarantees {
		if err := lowerLogicSpec(&action.Guarantees[i].Spec, ctx); err != nil {
			return fmt.Errorf("guarantee %d: %w", i, err)
		}
	}
	for i := range action.SafetyRules {
		if err := lowerLogicSpec(&action.SafetyRules[i].Spec, ctx); err != nil {
			return fmt.Errorf("safety rule %d: %w", i, err)
		}
	}
	return nil
}

func lowerQuery(query *model_state.Query, baseCtx *LowerContext) error {
	ctx := ContextWithParameters(baseCtx, query.Parameters)

	for i := range query.Requires {
		if err := lowerLogicSpec(&query.Requires[i].Spec, ctx); err != nil {
			return fmt.Errorf("require %d: %w", i, err)
		}
	}
	for i := range query.Guarantees {
		if err := lowerLogicSpec(&query.Guarantees[i].Spec, ctx); err != nil {
			return fmt.Errorf("guarantee %d: %w", i, err)
		}
	}
	return nil
}

// contextWithParameters creates a child context with action/query parameters added.
func ContextWithParameters(base *LowerContext, params []model_state.Parameter) *LowerContext {
	if len(params) == 0 {
		return base
	}
	child := *base
	child.Parameters = make(map[string]bool)
	if base.Parameters != nil {
		for k, v := range base.Parameters {
			child.Parameters[k] = v
		}
	}
	for _, p := range params {
		child.Parameters[p.Name] = true
	}
	return &child
}

// lowerLogicSpec parses and lowers a single ExpressionSpec if it has a TLA+ specification
// and hasn't been lowered yet.
func lowerLogicSpec(spec *model_spec.ExpressionSpec, ctx *LowerContext) error {
	// Skip if not TLA+, no specification text, or already lowered.
	if spec.Notation != "tla_plus" || spec.Specification == "" || spec.Expression != nil {
		return nil
	}

	// Parse the TLA+ specification string to AST.
	astExpr, err := parser.ParseExpression(spec.Specification)
	if err != nil {
		return fmt.Errorf("parse %q: %w", spec.Specification, err)
	}

	// Lower AST to model_expression.
	meExpr, err := Lower(astExpr, ctx)
	if err != nil {
		return fmt.Errorf("lower %q: %w", spec.Specification, err)
	}

	spec.Expression = meExpr
	return nil
}

// --- Map builders ---

func BuildGlobalFunctionMap(model *core.Model) map[string]identity.Key {
	m := make(map[string]identity.Key, len(model.GlobalFunctions))
	for _, gf := range model.GlobalFunctions {
		m[gf.Name] = gf.Key
	}
	return m
}

func BuildNamedSetMap(model *core.Model) map[string]identity.Key {
	m := make(map[string]identity.Key, len(model.NamedSets))
	for _, ns := range model.NamedSets {
		m[ns.Name] = ns.Key
	}
	return m
}

func BuildAllActionsMap(model *core.Model) map[string]identity.Key {
	m := make(map[string]identity.Key)
	for _, domain := range model.Domains {
		dName := domain.Name
		for _, subdomain := range domain.Subdomains {
			sName := subdomain.Name
			for _, class := range subdomain.Classes {
				cName := class.Name
				for _, action := range class.Actions {
					// Register with multiple scoping levels for lookup flexibility.
					// Full: Domain!Subdomain!Class!Action
					m[dName+"!"+sName+"!"+cName+"!"+action.Name] = action.Key
					// Partial: Subdomain!Class!Action
					m[sName+"!"+cName+"!"+action.Name] = action.Key
					// Partial: Class!Action
					m[cName+"!"+action.Name] = action.Key
				}
				for _, query := range class.Queries {
					m[dName+"!"+sName+"!"+cName+"!"+query.Name] = query.Key
					m[sName+"!"+cName+"!"+query.Name] = query.Key
					m[cName+"!"+query.Name] = query.Key
				}
			}
		}
	}
	return m
}

func BuildAttributeNameMap(class *model_class.Class) map[string]identity.Key {
	m := make(map[string]identity.Key, len(class.Attributes))
	for _, attr := range class.Attributes {
		m[attr.Name] = attr.Key
	}
	return m
}

func BuildActionNameMap(class *model_class.Class) map[string]identity.Key {
	m := make(map[string]identity.Key, len(class.Actions))
	for _, action := range class.Actions {
		m[action.Name] = action.Key
	}
	return m
}

func BuildQueryNameMap(class *model_class.Class) map[string]identity.Key {
	m := make(map[string]identity.Key, len(class.Queries))
	for _, query := range class.Queries {
		m[query.Name] = query.Key
	}
	return m
}
