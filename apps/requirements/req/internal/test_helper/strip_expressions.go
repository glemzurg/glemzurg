package test_helper

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_spec"
)

// StripExpressions returns a copy of the model with all parsed Expression and
// ExpressionType trees set to nil. This is used by database round-trip tests
// because the database layer stores only the specification text, not the
// parsed intermediate representation.
func StripExpressions(model req_model.Model) req_model.Model {

	// Model invariants.
	model.Invariants = stripLogicSlice(model.Invariants)

	// Global functions.
	if model.GlobalFunctions != nil {
		gfs := make(map[identity.Key]model_logic.GlobalFunction, len(model.GlobalFunctions))
		for k, gf := range model.GlobalFunctions {
			gf.Logic = stripLogic(gf.Logic)
			gfs[k] = gf
		}
		model.GlobalFunctions = gfs
	}

	// Named sets.
	if model.NamedSets != nil {
		for k, ns := range model.NamedSets {
			ns.Spec = stripExpressionSpec(ns.Spec)
			if ns.TypeSpec != nil {
				ts := stripTypeSpec(*ns.TypeSpec)
				ns.TypeSpec = &ts
			}
			model.NamedSets[k] = ns
		}
	}

	// Walk domains > subdomains > classes.
	for domainKey, domain := range model.Domains {
		for subdomainKey, subdomain := range domain.Subdomains {
			for classKey, class := range subdomain.Classes {
				// Class invariants.
				class.Invariants = stripLogicSlice(class.Invariants)

				// Attributes: derivation policy + attribute invariants.
				for attrKey, attr := range class.Attributes {
					if attr.DerivationPolicy != nil {
						l := stripLogic(*attr.DerivationPolicy)
						attr.DerivationPolicy = &l
					}
					attr.Invariants = stripLogicSlice(attr.Invariants)
					class.Attributes[attrKey] = attr
				}

				// Guards.
				for guardKey, guard := range class.Guards {
					guard.Logic = stripLogic(guard.Logic)
					class.Guards[guardKey] = guard
				}

				// Actions: requires, guarantees, safety rules.
				for actionKey, action := range class.Actions {
					action.Requires = stripLogicSlice(action.Requires)
					action.Guarantees = stripLogicSlice(action.Guarantees)
					action.SafetyRules = stripLogicSlice(action.SafetyRules)
					class.Actions[actionKey] = action
				}

				// Queries: requires, guarantees.
				for queryKey, query := range class.Queries {
					query.Requires = stripLogicSlice(query.Requires)
					query.Guarantees = stripLogicSlice(query.Guarantees)
					class.Queries[queryKey] = query
				}

				subdomain.Classes[classKey] = class
			}
			domain.Subdomains[subdomainKey] = subdomain
		}
		model.Domains[domainKey] = domain
	}

	return model
}

func stripLogicSlice(logics []model_logic.Logic) []model_logic.Logic {
	if logics == nil {
		return nil
	}
	result := make([]model_logic.Logic, len(logics))
	for i, l := range logics {
		result[i] = stripLogic(l)
	}
	return result
}

func stripLogic(l model_logic.Logic) model_logic.Logic {
	l.Spec = stripExpressionSpec(l.Spec)
	if l.TargetTypeSpec != nil {
		ts := stripTypeSpec(*l.TargetTypeSpec)
		l.TargetTypeSpec = &ts
	}
	return l
}

func stripExpressionSpec(spec model_spec.ExpressionSpec) model_spec.ExpressionSpec {
	spec.Expression = nil
	return spec
}

func stripTypeSpec(spec model_spec.TypeSpec) model_spec.TypeSpec {
	spec.ExpressionType = nil
	return spec
}
