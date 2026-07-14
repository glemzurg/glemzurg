package convert

import (
	"errors"
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
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
	classNames := BuildClassNameMap(model)

	// 1. Lower model-level invariants (no class context).
	modelCtx := &LowerContext{
		GlobalFunctions: globalFunctions,
		NamedSets:       namedSets,
		ClassNames:      classNames,
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
			ClassNames:      classNames,
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

	allAssociations := model.GetClassAssociations()

	// 4. Walk domains → subdomains → classes.
	for dKey, domain := range model.Domains {
		for sKey, subdomain := range domain.Subdomains {
			for cKey, class := range subdomain.Classes {
				subdomainMaps := SubdomainClassMaps{Associations: allAssociations, Classes: subdomain.Classes}
				if err := lowerAllClassExpressions(&class, globalFunctions, namedSets, classNames, allActions, subdomainMaps); err != nil {
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
func lowerAllClassExpressions(class *model_class.Class, globalFunctions, namedSets, classNames, allActions map[string]identity.Key, subdomainMaps SubdomainClassMaps) error {
	classCtx := NewClassLowerContext(class, globalFunctions, namedSets, allActions, subdomainMaps.Associations, subdomainMaps.Classes)
	classCtx.ClassNames = classNames
	classPF := NewExpressionParseFunc(classCtx)

	// Class invariants.
	for i := range class.Invariants {
		if err := relowerSpec(&class.Invariants[i].Spec, classPF); err != nil {
			return fmt.Errorf("class invariant %d: %w", i, err)
		}
	}

	// Attributes: derivation policy and invariants.
	for i := range class.Attributes {
		attr := &class.Attributes[i]
		if attr.DerivationPolicy != nil {
			if err := relowerSpec(&attr.DerivationPolicy.Spec, classPF); err != nil {
				return fmt.Errorf("attribute %q derivation: %w", attr.Key.String(), err)
			}
		}
		for j := range attr.Invariants {
			if err := relowerSpec(&attr.Invariants[j].Spec, classPF); err != nil {
				return fmt.Errorf("attribute %q invariant %d: %w", attr.Key.String(), j, err)
			}
		}
	}

	// Guards.
	for gKey, guard := range class.Guards {
		if err := relowerSpec(&guard.Logic.Spec, classPF); err != nil {
			return fmt.Errorf("guard %q: %w", gKey.String(), err)
		}
		class.Guards[gKey] = guard
	}

	for actKey, action := range class.Actions {
		if err := relowerActionExpressions(actKey, &action, classCtx); err != nil {
			return err
		}
		class.Actions[actKey] = action
	}

	for qKey, query := range class.Queries {
		if err := relowerQueryExpressions(qKey, &query, classCtx); err != nil {
			return err
		}
		class.Queries[qKey] = query
	}

	return nil
}

func relowerActionExpressions(actKey identity.Key, action *model_state.Action, classCtx *LowerContext) error {
	actPF := NewExpressionParseFunc(ContextWithParameters(classCtx, action.Parameters))
	for i := range action.Requires {
		if err := relowerSpec(&action.Requires[i].Spec, actPF); err != nil {
			return fmt.Errorf("action %q require %d: %w", actKey.String(), i, err)
		}
	}
	for i := range action.Guarantees {
		guar := &action.Guarantees[i]
		if model_logic.IsAssociationClassReify(*guar) {
			if err := relowerSpec(&guar.EndpointSelectorSpec, actPF); err != nil {
				return fmt.Errorf("action %q guarantee %d endpoint_selector: %w", actKey.String(), i, err)
			}
			reifyPF := actPF
			if setMap, ok := guar.EndpointSelectorSpec.Expression.(*me.SetMap); ok && setMap.Variable != "" {
				reifyCtx := withLocalVar(ContextWithParameters(classCtx, action.Parameters), setMap.Variable)
				reifyPF = NewExpressionParseFunc(reifyCtx)
			}
			if err := relowerSpec(&guar.Spec, reifyPF); err != nil {
				return fmt.Errorf("action %q guarantee %d: %w", actKey.String(), i, err)
			}
			continue
		}
		if err := relowerSpec(&guar.Spec, actPF); err != nil {
			return fmt.Errorf("action %q guarantee %d: %w", actKey.String(), i, err)
		}
	}
	for i := range action.SafetyRules {
		if err := relowerSpec(&action.SafetyRules[i].Spec, actPF); err != nil {
			return fmt.Errorf("action %q safety rule %d: %w", actKey.String(), i, err)
		}
	}
	if err := relowerParameterInvariants(actKey.String(), "action", action.Parameters, actPF); err != nil {
		return err
	}
	return relowerParameterSimulation(actKey.String(), "action", action.Parameters, classCtx)
}

func relowerQueryExpressions(qKey identity.Key, query *model_state.Query, classCtx *LowerContext) error {
	qPF := NewExpressionParseFunc(ContextWithParameters(classCtx, query.Parameters))
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
	return relowerParameterInvariants(qKey.String(), "query", query.Parameters, qPF)
}

func relowerParameterInvariants(ownerKey, ownerKind string, params []model_state.Parameter, pf logic_spec.ExpressionParseFunc) error {
	for i := range params {
		for j := range params[i].Invariants {
			if err := relowerSpec(&params[i].Invariants[j].Spec, pf); err != nil {
				return fmt.Errorf("%s %q parameter %q invariant %d: %w", ownerKind, ownerKey, params[i].Name, j, err)
			}
		}
	}
	return nil
}

func relowerParameterSimulation(ownerKey, ownerKind string, params []model_state.Parameter, classCtx *LowerContext) error {
	pf := NewExpressionParseFunc(classCtx)
	for i := range params {
		if params[i].Simulation == nil {
			continue
		}
		for r := range params[i].Simulation.Rules {
			rule := &params[i].Simulation.Rules[r]
			for j := range rule.Requires {
				if err := relowerSpec(&rule.Requires[j].Spec, pf); err != nil {
					return fmt.Errorf("%s %q parameter %q simulation rule %d require %d: %w", ownerKind, ownerKey, params[i].Name, r, j, err)
				}
			}
			if rule.Specification != nil {
				if err := relowerSpec(&rule.Specification.Spec, pf); err != nil {
					return fmt.Errorf("%s %q parameter %q simulation rule %d specification: %w", ownerKind, ownerKey, params[i].Name, r, err)
				}
			}
		}
	}
	return nil
}

func relowerParameterSimulationStrict(ownerKey, ownerKind string, params []model_state.Parameter, classCtx *LowerContext) []error {
	pf := NewExpressionParseFuncStrict(classCtx)
	var errs []error
	for i := range params {
		if params[i].Simulation == nil {
			continue
		}
		for r := range params[i].Simulation.Rules {
			rule := &params[i].Simulation.Rules[r]
			for j := range rule.Requires {
				if err := relowerSpecStrict(&rule.Requires[j].Spec, pf); err != nil {
					errs = append(errs, fmt.Errorf("%s %q parameter %q simulation rule %d require %d: %w", ownerKind, ownerKey, params[i].Name, r, j, err))
				}
			}
			if rule.Specification != nil {
				if err := relowerSpecStrict(&rule.Specification.Spec, pf); err != nil {
					errs = append(errs, fmt.Errorf("%s %q parameter %q simulation rule %d specification: %w", ownerKind, ownerKey, params[i].Name, r, err))
				}
			}
		}
	}
	return errs
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
	classNames := BuildClassNameMap(model)

	var errs []error

	// Model-level invariants.
	modelCtx := &LowerContext{
		GlobalFunctions: globalFunctions,
		NamedSets:       namedSets,
		ClassNames:      classNames,
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
			ClassNames:      classNames,
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

	allAssociations := model.GetClassAssociations()

	// Walk domains → subdomains → classes.
	for dKey, domain := range model.Domains {
		for sKey, subdomain := range domain.Subdomains {
			for cKey, class := range subdomain.Classes {
				subdomainMaps := SubdomainClassMaps{Associations: allAssociations, Classes: subdomain.Classes}
				if classErrs := lowerAllClassExpressionsStrict(&class, globalFunctions, namedSets, classNames, allActions, subdomainMaps); classErrs != nil {
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
func lowerAllClassExpressionsStrict(class *model_class.Class, globalFunctions, namedSets, classNames, allActions map[string]identity.Key, subdomainMaps SubdomainClassMaps) error {
	classCtx := NewClassLowerContext(class, globalFunctions, namedSets, allActions, subdomainMaps.Associations, subdomainMaps.Classes)
	classCtx.ClassNames = classNames
	classPF := NewExpressionParseFuncStrict(classCtx)

	var errs []error

	// Class invariants.
	for i := range class.Invariants {
		if err := relowerSpecStrict(&class.Invariants[i].Spec, classPF); err != nil {
			errs = append(errs, fmt.Errorf("invariant %d: %w", i, err))
		}
	}

	// Attributes: derivation policy and invariants.
	for i := range class.Attributes {
		attr := &class.Attributes[i]
		if attr.DerivationPolicy != nil {
			if err := relowerSpecStrict(&attr.DerivationPolicy.Spec, classPF); err != nil {
				errs = append(errs, fmt.Errorf("attribute %q derivation: %w", attr.Key.String(), err))
			}
		}
		for j := range attr.Invariants {
			if err := relowerSpecStrict(&attr.Invariants[j].Spec, classPF); err != nil {
				errs = append(errs, fmt.Errorf("attribute %q invariant %d: %w", attr.Key.String(), j, err))
			}
		}
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
		errs = append(errs, lowerOneActionExpressionsStrict(actKey, &action, classCtx)...)
		class.Actions[actKey] = action
	}
	return errs
}

func lowerOneActionExpressionsStrict(actKey identity.Key, action *model_state.Action, classCtx *LowerContext) []error {
	var errs []error
	actCtx := ContextWithParameters(classCtx, action.Parameters)
	actPF := NewExpressionParseFuncStrict(actCtx)
	for i := range action.Requires {
		if err := relowerSpecStrict(&action.Requires[i].Spec, actPF); err != nil {
			errs = append(errs, fmt.Errorf("action %q require %d: %w", actKey.String(), i, err))
		}
	}
	errs = append(errs, lowerActionGuaranteesStrict(actKey, action, actCtx, actPF)...)
	for i := range action.SafetyRules {
		if err := relowerSpecStrict(&action.SafetyRules[i].Spec, actPF); err != nil {
			errs = append(errs, fmt.Errorf("action %q safety rule %d: %w", actKey.String(), i, err))
		}
	}
	for i := range action.Parameters {
		for j := range action.Parameters[i].Invariants {
			if err := relowerSpecStrict(&action.Parameters[i].Invariants[j].Spec, actPF); err != nil {
				errs = append(errs, fmt.Errorf("action %q parameter %q invariant %d: %w", actKey.String(), action.Parameters[i].Name, j, err))
			}
		}
	}
	errs = append(errs, relowerParameterSimulationStrict(actKey.String(), "action", action.Parameters, classCtx)...)
	return errs
}

func lowerActionGuaranteesStrict(
	actKey identity.Key,
	action *model_state.Action,
	actCtx *LowerContext,
	actPF StrictExpressionParseFunc,
) []error {
	var errs []error
	for i := range action.Guarantees {
		guar := &action.Guarantees[i]
		if model_logic.IsAssociationClassReify(*guar) {
			errs = append(errs, relowerAssociationClassReifyStrict(actKey, i, guar, actCtx, actPF)...)
			continue
		}
		if err := relowerSpecStrict(&guar.Spec, actPF); err != nil {
			errs = append(errs, fmt.Errorf("action %q guarantee %d: %w", actKey.String(), i, err))
		}
	}
	return errs
}

func relowerAssociationClassReifyStrict(
	actKey identity.Key,
	index int,
	guar *model_logic.Logic,
	actCtx *LowerContext,
	actPF StrictExpressionParseFunc,
) []error {
	var errs []error
	if err := relowerSpecStrict(&guar.EndpointSelectorSpec, actPF); err != nil {
		errs = append(errs, fmt.Errorf("action %q guarantee %d endpoint_selector: %w", actKey.String(), index, err))
	}
	reifyPF := actPF
	if setMap, ok := guar.EndpointSelectorSpec.Expression.(*me.SetMap); ok && setMap.Variable != "" {
		reifyCtx := withLocalVar(actCtx, setMap.Variable)
		reifyPF = NewExpressionParseFuncStrict(reifyCtx)
	}
	if err := relowerSpecStrict(&guar.Spec, reifyPF); err != nil {
		errs = append(errs, fmt.Errorf("action %q guarantee %d: %w", actKey.String(), index, err))
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
		for i := range query.Parameters {
			for j := range query.Parameters[i].Invariants {
				if err := relowerSpecStrict(&query.Parameters[i].Invariants[j].Spec, qPF); err != nil {
					errs = append(errs, fmt.Errorf("query %q parameter %q invariant %d: %w", qKey.String(), query.Parameters[i].Name, j, err))
				}
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
