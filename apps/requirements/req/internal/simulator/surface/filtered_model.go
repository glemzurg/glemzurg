package surface

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// BuildFilteredModel creates a new Model containing only the classes,
// associations, and invariants from the resolved surface. The original
// model is not modified.
//
// Derived attributes and queries that depend on out-of-scope classes are kept
// intact on the filtered model for traceability; the catalog marks them
// surface-unavailable so they are not selected for external simulation steps
// and evaluation produces a violation when something calls them.
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

	// Record derived/query members that depend on out-of-scope association data.
	resolved.UnavailableMembers = CollectUnavailableMembers(original, resolved)

	return &filtered, nil
}

// classWithScopedInvariants drops class invariants that reference out-of-scope classes.
// Derived attributes and queries are not rewritten here — unavailability is tracked
// separately so callers get a surface-out-of-scope violation instead of a silent stub.
func classWithScopedInvariants(
	class model_class.Class,
	inScopeNames, allNames map[string]bool,
) model_class.Class {
	if len(class.Invariants) == 0 {
		return class
	}
	included, _ := ScopeInvariantsWithAllClasses(class.Invariants, inScopeNames, allNames)
	if len(included) == len(class.Invariants) {
		return class
	}
	out := class
	out.Invariants = included
	return out
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
