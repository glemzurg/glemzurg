package parser_human

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_spec"
)

// lowerAllExpressions walks the model tree and re-creates all ExpressionSpecs
// using convert.NewExpressionParseFunc with full context. This is Phase 2 of
// the two-phase parsing approach: Phase 1 parses structure with nil parseFunc,
// Phase 2 re-creates specs with full context so expressions get fully parsed.
func lowerAllExpressions(model *req_model.Model) error {
	// Build model-level lookup maps.
	globalFunctions := convert.BuildGlobalFunctionMap(model)
	namedSets := convert.BuildNamedSetMap(model)
	allActions := convert.BuildAllActionsMap(model)

	// 1. Lower model-level invariants (no class context).
	modelCtx := &convert.LowerContext{
		GlobalFunctions: globalFunctions,
		NamedSets:       namedSets,
		AllActions:      allActions,
	}
	modelPF := convert.NewExpressionParseFunc(modelCtx)
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
		gfCtx := &convert.LowerContext{
			GlobalFunctions: globalFunctions,
			NamedSets:       namedSets,
			AllActions:      allActions,
			Parameters:      params,
		}
		gfPF := convert.NewExpressionParseFunc(gfCtx)
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
				if err := lowerClassExpressions(&class, globalFunctions, namedSets, allActions); err != nil {
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

// lowerClassExpressions re-creates all ExpressionSpecs in a class with full context.
func lowerClassExpressions(class *model_class.Class, globalFunctions, namedSets, allActions map[string]identity.Key) error {
	// Build class-level context maps.
	attrNames := convert.BuildAttributeNameMap(class)
	actionNames := convert.BuildActionNameMap(class)
	queryNames := convert.BuildQueryNameMap(class)

	classCtx := &convert.LowerContext{
		ClassKey:        class.Key,
		AttributeNames:  attrNames,
		ActionNames:     actionNames,
		QueryNames:      queryNames,
		GlobalFunctions: globalFunctions,
		NamedSets:       namedSets,
		AllActions:      allActions,
	}
	classPF := convert.NewExpressionParseFunc(classCtx)

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
		actCtx := convert.ContextWithParameters(classCtx, action.Parameters)
		actPF := convert.NewExpressionParseFunc(actCtx)

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
		qCtx := convert.ContextWithParameters(classCtx, query.Parameters)
		qPF := convert.NewExpressionParseFunc(qCtx)

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
func relowerSpec(spec *model_spec.ExpressionSpec, pf model_spec.ExpressionParseFunc) error {
	if spec.Specification == "" {
		return nil
	}
	newSpec, err := model_spec.NewExpressionSpec(spec.Notation, spec.Specification, pf)
	if err != nil {
		return err
	}
	*spec = newSpec
	return nil
}
