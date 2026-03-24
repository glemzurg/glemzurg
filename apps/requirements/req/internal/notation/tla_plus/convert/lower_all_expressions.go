package convert

import (
	"errors"
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// LowerAllExpressions walks the model tree and re-creates all ExpressionSpecs
// using NewExpressionParseFunc with full context. This is the tolerant approach:
// parse failures leave Expression as nil rather than returning an error.
func LowerAllExpressions(model *core.Model) error {
	// Build model-level lookup maps.
	globalFunctions := BuildGlobalFunctionMap(model)
	namedSets := BuildNamedSetMap(model)
	allActions := BuildAllActionsMap(model)

	// 1. Lower model-level invariants (no class context).
	modelCtx := &LowerContext{
		GlobalFunctions: globalFunctions,
		NamedSets:       namedSets,
		AllActions:      allActions,
	}
	modelPF := NewExpressionParseFunc(modelCtx)
	for i := range model.Invariants {
		if err := relowerSpec(&model.Invariants[i].Spec, modelPF); err != nil {
			return fmt.Errorf("model invariant %d: %w", i, err)
		}
	}

	// 2. Lower global functions (each has its own parameter context).
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
		gfPF := NewExpressionParseFunc(gfCtx)
		if err := relowerSpec(&gf.Logic.Spec, gfPF); err != nil {
			return fmt.Errorf("global function %q: %w", gfKey.String(), err)
		}
		model.GlobalFunctions[gfKey] = gf
	}

	// 3. Lower named sets.
	for nsKey, ns := range model.NamedSets {
		if err := relowerSpec(&ns.Spec, modelPF); err != nil {
			return fmt.Errorf("named set %q: %w", nsKey.String(), err)
		}
		model.NamedSets[nsKey] = ns
	}

	// 4. Walk domains → subdomains → classes.
	for dKey, domain := range model.Domains {
		for sKey, subdomain := range domain.Subdomains {
			for cKey, class := range subdomain.Classes {
				if err := lowerAllClassExpressions(&class, globalFunctions, namedSets, allActions); err != nil {
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

// lowerAllClassExpressions re-creates all ExpressionSpecs in a class with full context.
func lowerAllClassExpressions(class *model_class.Class, globalFunctions, namedSets, allActions map[string]identity.Key) error {
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
	classPF := NewExpressionParseFunc(classCtx)

	// Class invariants.
	for i := range class.Invariants {
		if err := relowerSpec(&class.Invariants[i].Spec, classPF); err != nil {
			return fmt.Errorf("class invariant %d: %w", i, err)
		}
	}

	// Attributes: derivation policy and invariants.
	for aKey, attr := range class.Attributes {
		if attr.DerivationPolicy != nil {
			if err := relowerSpec(&attr.DerivationPolicy.Spec, classPF); err != nil {
				return fmt.Errorf("attribute %q derivation: %w", aKey.String(), err)
			}
		}
		for i := range attr.Invariants {
			if err := relowerSpec(&attr.Invariants[i].Spec, classPF); err != nil {
				return fmt.Errorf("attribute %q invariant %d: %w", aKey.String(), i, err)
			}
		}
		class.Attributes[aKey] = attr
	}

	// Guards.
	for gKey, guard := range class.Guards {
		if err := relowerSpec(&guard.Logic.Spec, classPF); err != nil {
			return fmt.Errorf("guard %q: %w", gKey.String(), err)
		}
		class.Guards[gKey] = guard
	}

	// Actions.
	for actKey, action := range class.Actions {
		actCtx := ContextWithParameters(classCtx, action.Parameters)
		actPF := NewExpressionParseFunc(actCtx)

		for i := range action.Requires {
			if err := relowerSpec(&action.Requires[i].Spec, actPF); err != nil {
				return fmt.Errorf("action %q require %d: %w", actKey.String(), i, err)
			}
		}
		for i := range action.Guarantees {
			if err := relowerSpec(&action.Guarantees[i].Spec, actPF); err != nil {
				return fmt.Errorf("action %q guarantee %d: %w", actKey.String(), i, err)
			}
		}
		for i := range action.SafetyRules {
			if err := relowerSpec(&action.SafetyRules[i].Spec, actPF); err != nil {
				return fmt.Errorf("action %q safety rule %d: %w", actKey.String(), i, err)
			}
		}
		class.Actions[actKey] = action
	}

	// Queries.
	for qKey, query := range class.Queries {
		qCtx := ContextWithParameters(classCtx, query.Parameters)
		qPF := NewExpressionParseFunc(qCtx)

		for i := range query.Requires {
			if err := relowerSpec(&query.Requires[i].Spec, qPF); err != nil {
				return fmt.Errorf("query %q require %d: %w", qKey.String(), i, err)
			}
		}
		for i := range query.Guarantees {
			if err := relowerSpec(&query.Guarantees[i].Spec, qPF); err != nil {
				return fmt.Errorf("query %q guarantee %d: %w", qKey.String(), i, err)
			}
		}
		class.Queries[qKey] = query
	}

	return nil
}

// relowerSpec re-creates an ExpressionSpec using the given parse function.
// If the spec has no specification text, it's a no-op.
func relowerSpec(spec *logic_spec.ExpressionSpec, pf logic_spec.ExpressionParseFunc) error {
	if spec.Specification == "" {
		return nil
	}
	newSpec, err := logic_spec.NewExpressionSpec(spec.Notation, spec.Specification, pf)
	if err != nil {
		return err
	}
	*spec = newSpec
	return nil
}

// LowerAllExpressionsStrict is like LowerAllExpressions but reports parse/lower
// errors instead of silently swallowing them. It collects ALL expression errors
// and returns them as a combined error so the caller sees every problem at once.
func LowerAllExpressionsStrict(model *core.Model) error {
	globalFunctions := BuildGlobalFunctionMap(model)
	namedSets := BuildNamedSetMap(model)
	allActions := BuildAllActionsMap(model)

	var errs []error

	// Model-level invariants.
	modelCtx := &LowerContext{
		GlobalFunctions: globalFunctions,
		NamedSets:       namedSets,
		AllActions:      allActions,
	}
	modelPF := NewExpressionParseFuncStrict(modelCtx)
	for i := range model.Invariants {
		if err := relowerSpecStrict(&model.Invariants[i].Spec, modelPF); err != nil {
			errs = append(errs, fmt.Errorf("model invariant %d: %w", i, err))
		}
	}

	// Global functions.
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
		gfPF := NewExpressionParseFuncStrict(gfCtx)
		if err := relowerSpecStrict(&gf.Logic.Spec, gfPF); err != nil {
			errs = append(errs, fmt.Errorf("global function %q: %w", gfKey.String(), err))
		}
		model.GlobalFunctions[gfKey] = gf
	}

	// Named sets.
	for nsKey, ns := range model.NamedSets {
		if err := relowerSpecStrict(&ns.Spec, modelPF); err != nil {
			errs = append(errs, fmt.Errorf("named set %q: %w", nsKey.String(), err))
		}
		model.NamedSets[nsKey] = ns
	}

	// Walk domains → subdomains → classes.
	for dKey, domain := range model.Domains {
		for sKey, subdomain := range domain.Subdomains {
			for cKey, class := range subdomain.Classes {
				if classErrs := lowerAllClassExpressionsStrict(&class, globalFunctions, namedSets, allActions); classErrs != nil {
					errs = append(errs, fmt.Errorf("class %q: %w", cKey.String(), classErrs))
				}
				subdomain.Classes[cKey] = class
			}
			domain.Subdomains[sKey] = subdomain
		}
		model.Domains[dKey] = domain
	}

	return errors.Join(errs...)
}

// lowerAllClassExpressionsStrict collects all expression errors in a class.
//
//complexity:cyclo:warn=20,fail=20 Walks all class expression sites.
func lowerAllClassExpressionsStrict(class *model_class.Class, globalFunctions, namedSets, allActions map[string]identity.Key) error {
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
	classPF := NewExpressionParseFuncStrict(classCtx)

	var errs []error

	// Class invariants.
	for i := range class.Invariants {
		if err := relowerSpecStrict(&class.Invariants[i].Spec, classPF); err != nil {
			errs = append(errs, fmt.Errorf("invariant %d: %w", i, err))
		}
	}

	// Attributes: derivation policy and invariants.
	for aKey, attr := range class.Attributes {
		if attr.DerivationPolicy != nil {
			if err := relowerSpecStrict(&attr.DerivationPolicy.Spec, classPF); err != nil {
				errs = append(errs, fmt.Errorf("attribute %q derivation: %w", aKey.String(), err))
			}
		}
		for i := range attr.Invariants {
			if err := relowerSpecStrict(&attr.Invariants[i].Spec, classPF); err != nil {
				errs = append(errs, fmt.Errorf("attribute %q invariant %d: %w", aKey.String(), i, err))
			}
		}
		class.Attributes[aKey] = attr
	}

	// Guards.
	for gKey, guard := range class.Guards {
		if err := relowerSpecStrict(&guard.Logic.Spec, classPF); err != nil {
			errs = append(errs, fmt.Errorf("guard %q: %w", gKey.String(), err))
		}
		class.Guards[gKey] = guard
	}

	// Actions.
	errs = append(errs, lowerActionExpressionsStrict(class, classCtx)...)

	// Queries.
	errs = append(errs, lowerQueryExpressionsStrict(class, classCtx)...)

	return errors.Join(errs...)
}

// lowerActionExpressionsStrict collects expression errors from all actions in a class.
func lowerActionExpressionsStrict(class *model_class.Class, classCtx *LowerContext) []error {
	var errs []error
	for actKey, action := range class.Actions {
		actCtx := ContextWithParameters(classCtx, action.Parameters)
		actPF := NewExpressionParseFuncStrict(actCtx)
		for i := range action.Requires {
			if err := relowerSpecStrict(&action.Requires[i].Spec, actPF); err != nil {
				errs = append(errs, fmt.Errorf("action %q require %d: %w", actKey.String(), i, err))
			}
		}
		for i := range action.Guarantees {
			if err := relowerSpecStrict(&action.Guarantees[i].Spec, actPF); err != nil {
				errs = append(errs, fmt.Errorf("action %q guarantee %d: %w", actKey.String(), i, err))
			}
		}
		for i := range action.SafetyRules {
			if err := relowerSpecStrict(&action.SafetyRules[i].Spec, actPF); err != nil {
				errs = append(errs, fmt.Errorf("action %q safety rule %d: %w", actKey.String(), i, err))
			}
		}
		class.Actions[actKey] = action
	}
	return errs
}

// lowerQueryExpressionsStrict collects expression errors from all queries in a class.
func lowerQueryExpressionsStrict(class *model_class.Class, classCtx *LowerContext) []error {
	var errs []error
	for qKey, query := range class.Queries {
		qCtx := ContextWithParameters(classCtx, query.Parameters)
		qPF := NewExpressionParseFuncStrict(qCtx)
		for i := range query.Requires {
			if err := relowerSpecStrict(&query.Requires[i].Spec, qPF); err != nil {
				errs = append(errs, fmt.Errorf("query %q require %d: %w", qKey.String(), i, err))
			}
		}
		for i := range query.Guarantees {
			if err := relowerSpecStrict(&query.Guarantees[i].Spec, qPF); err != nil {
				errs = append(errs, fmt.Errorf("query %q guarantee %d: %w", qKey.String(), i, err))
			}
		}
		class.Queries[qKey] = query
	}
	return errs
}

// relowerSpecStrict re-creates an ExpressionSpec using the strict parse function.
// Returns an error if parsing/lowering fails, instead of silently leaving Expression nil.
func relowerSpecStrict(spec *logic_spec.ExpressionSpec, pf StrictExpressionParseFunc) error {
	if spec.Specification == "" {
		return nil
	}
	expr, normalized, err := pf(spec.Specification)
	if err != nil {
		return fmt.Errorf("specification %q: %w", spec.Specification, err)
	}
	spec.Expression = expr
	if normalized != "" {
		spec.Specification = normalized
	}
	return nil
}
