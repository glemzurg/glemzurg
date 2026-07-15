package surface

import (
	"math/big"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// BuildFilteredModel creates a new Model containing only the classes,
// associations, and invariants from the resolved surface. The original
// model is not modified.
func BuildFilteredModel(original *core.Model, resolved *ResolvedSurface) (*core.Model, error) {
	filtered := core.NewModel(original.Key, core.ModelDetails{
		Name: original.Name, Details: original.Details,
	}, original.UnfinishedNotes, resolved.ModelInvariants, original.GlobalFunctions, original.NamedSets)
	filtered.Actors = original.Actors
	filtered.ActorGeneralizations = original.ActorGeneralizations

	inScopeNames, allNames := classNameSetsForScoping(original, resolved)

	// Rebuild domain/subdomain/class tree with only included classes.
	filteredDomains := make(map[identity.Key]model_domain.Domain)
	for domainKey, domain := range original.Domains {
		filteredSubdomains := make(map[identity.Key]model_domain.Subdomain)
		for subdomainKey, subdomain := range domain.Subdomains {
			filteredClasses := make(map[identity.Key]model_class.Class)
			for classKey, class := range subdomain.Classes {
				if _, inScope := resolved.Classes[classKey]; inScope {
					filteredClasses[classKey] = classWithScopedInvariants(class, inScopeNames, allNames)
				}
			}
			if len(filteredClasses) > 0 {
				filteredSub := model_domain.NewSubdomain(subdomainKey, subdomain.Name, subdomain.Details, subdomain.UnfinishedNotes, subdomain.UmlComment)
				filteredSub.Generalizations = subdomain.Generalizations
				filteredSub.UseCaseGeneralizations = subdomain.UseCaseGeneralizations
				filteredSub.Classes = filteredClasses
				filteredSub.UseCases = subdomain.UseCases
				filteredSub.ClassAssociations = filterAssociations(subdomain.ClassAssociations, resolved.Associations)
				filteredSub.UseCaseShares = subdomain.UseCaseShares
				filteredSubdomains[subdomainKey] = filteredSub
			}
		}
		if len(filteredSubdomains) > 0 {
			filteredDom := model_domain.NewDomain(domainKey, domain.Name, domain.Details, domain.UnfinishedNotes, domain.Realized, domain.UmlComment)
			filteredDom.Subdomains = filteredSubdomains
			filteredDom.ClassAssociations = filterAssociations(domain.ClassAssociations, resolved.Associations)
			filteredDomains[domainKey] = filteredDom
		}
	}
	filtered.Domains = filteredDomains

	// Filter model-level associations (resolved already strips out-of-scope AC keys).
	filtered.ClassAssociations = filterAssociations(original.ClassAssociations, resolved.Associations)

	// Preserve domain associations.
	filtered.DomainAssociations = original.DomainAssociations

	return &filtered, nil
}

// classWithScopedInvariants drops class invariants and attribute derivations that
// reference out-of-scope classes (e.g. association-class navigations when that
// class is not on the surface). Amount-bearing AC work is simply not simulated.
func classWithScopedInvariants(
	class model_class.Class,
	inScopeNames, allNames map[string]bool,
) model_class.Class {
	out := class
	changed := false

	if len(class.Invariants) > 0 {
		included, _ := ScopeInvariantsWithAllClasses(class.Invariants, inScopeNames, allNames)
		if len(included) != len(class.Invariants) {
			out.Invariants = included
			changed = true
		}
	}

	if scopedAttrs, attrChanged := attributesWithScopedDerivations(class.Attributes, inScopeNames, allNames); attrChanged {
		out.Attributes = scopedAttrs
		changed = true
	}

	if !changed {
		return class
	}
	return out
}

// attributesWithScopedDerivations replaces derivations that reference out-of-scope
// classes with a constant-zero value expression. The attribute stays derived (not a
// write-liveness target) but does not require association-class data absent from the surface.
func attributesWithScopedDerivations(
	attrs []model_class.Attribute,
	inScopeNames, allNames map[string]bool,
) ([]model_class.Attribute, bool) {
	if len(attrs) == 0 {
		return attrs, false
	}
	changed := false
	out := make([]model_class.Attribute, len(attrs))
	copy(out, attrs)
	for i := range out {
		if out[i].DerivationPolicy == nil {
			continue
		}
		included, excluded := ScopeInvariantsWithAllClasses(
			[]model_logic.Logic{*out[i].DerivationPolicy}, inScopeNames, allNames,
		)
		if len(excluded) == 0 && len(included) == 1 {
			continue
		}
		out[i].DerivationPolicy = inactiveSurfaceDerivation(*out[i].DerivationPolicy)
		changed = true
	}
	if !changed {
		return attrs, false
	}
	return out, true
}

// inactiveSurfaceDerivation keeps a value derivation that evaluates without out-of-scope
// association-class members. Constant zero is the neutral stand-in for numeric ledgers.
func inactiveSurfaceDerivation(original model_logic.Logic) *model_logic.Logic {
	stub := original
	stub.Description = "inactive on this surface (references out-of-scope class)"
	stub.Spec = logic_spec.ExpressionSpec{
		Notation:      model_logic.NotationTLAPlus,
		Specification: "0",
		Expression:    &me.IntLiteral{Value: big.NewInt(0)},
	}
	return &stub
}

// filterAssociations keeps only associations that are in the resolved set.
// Uses the resolved association value so out-of-scope association-class keys stay stripped.
func filterAssociations(
	source map[identity.Key]model_class.Association,
	resolved map[identity.Key]model_class.Association,
) map[identity.Key]model_class.Association {
	if len(source) == 0 {
		return nil
	}
	result := make(map[identity.Key]model_class.Association)
	for k := range source {
		if assoc, inScope := resolved[k]; inScope {
			result[k] = assoc
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
