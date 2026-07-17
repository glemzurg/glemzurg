package surface

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// BuildFilteredModel creates a new Model containing only the classes,
// associations, and invariants from the resolved surface. The original
// model is not modified.
//
// Class and attribute invariants that reference out-of-scope classes (by name
// or association navigation) are dropped — they only run when every involved
// class is on the surface. Derived attributes and queries that depend on
// out-of-scope classes stay intact for traceability; the catalog marks them
// surface-unavailable so evaluation produces a violation when something calls them.
func BuildFilteredModel(original *core.Model, resolved *ResolvedSurface) (*core.Model, error) {
	filtered := core.NewModel(original.Key, core.ModelDetails{
		Name: original.Name, Details: original.Details,
	}, original.UnfinishedNotes, resolved.ModelInvariants, original.GlobalFunctions, original.NamedSets)
	filtered.Actors = original.Actors
	filtered.ActorGeneralizations = original.ActorGeneralizations

	scope := newSurfaceClassScope(original, resolved)

	// Rebuild domain/subdomain/class tree with only included classes.
	filteredDomains := make(map[identity.Key]model_domain.Domain)
	for domainKey, domain := range original.Domains {
		filteredSubdomains := make(map[identity.Key]model_domain.Subdomain)
		for subdomainKey, subdomain := range domain.Subdomains {
			filteredClasses := make(map[identity.Key]model_class.Class)
			for classKey, class := range subdomain.Classes {
				if _, inScope := resolved.Classes[classKey]; inScope {
					filteredClasses[classKey] = classWithScopedInvariants(class, scope, resolved)
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

// surfaceClassScope holds lookup tables for association-aware surface dependency checks.
type surfaceClassScope struct {
	inScope     map[identity.Key]model_class.Class
	classByName map[string]identity.Key
	classNames  map[identity.Key]string
	nav         associationNavDeps
}

func newSurfaceClassScope(original *core.Model, resolved *ResolvedSurface) surfaceClassScope {
	return surfaceClassScope{
		inScope:     resolved.Classes,
		classByName: classLookupByNameAndTLA(original),
		classNames:  classDisplayNames(original),
		nav:         buildAssociationNavigationDeps(original),
	}
}

// classWithScopedInvariants drops class and attribute invariants that involve any
// class outside the surface (including association navigations). Common pattern:
// cross-class invariants only run when every participant is simulated.
func classWithScopedInvariants(
	class model_class.Class,
	scope surfaceClassScope,
	resolved *ResolvedSurface,
) model_class.Class {
	out := class
	changed := false

	if len(class.Invariants) > 0 {
		included, excluded := filterLogicsForSurface(class.Invariants, class.Key, scope)
		if len(excluded) > 0 {
			out.Invariants = included
			changed = true
			for _, inv := range excluded {
				resolved.Warnings = append(resolved.Warnings,
					fmt.Sprintf("class %s invariant excluded (references out-of-scope class): %s",
						class.Name, invDescription(inv)))
			}
		}
	}

	if scopedAttrs, attrChanged := attributesWithScopedInvariants(class.Attributes, class.Key, scope, resolved, class.Name); attrChanged {
		out.Attributes = scopedAttrs
		changed = true
	}

	if !changed {
		return class
	}
	return out
}

func attributesWithScopedInvariants(
	attrs []model_class.Attribute,
	ownerClassKey identity.Key,
	scope surfaceClassScope,
	resolved *ResolvedSurface,
	className string,
) ([]model_class.Attribute, bool) {
	if len(attrs) == 0 {
		return attrs, false
	}
	changed := false
	out := make([]model_class.Attribute, len(attrs))
	copy(out, attrs)
	for i := range out {
		if len(out[i].Invariants) == 0 {
			continue
		}
		included, excluded := filterLogicsForSurface(out[i].Invariants, ownerClassKey, scope)
		if len(excluded) == 0 {
			continue
		}
		out[i].Invariants = included
		changed = true
		for _, inv := range excluded {
			resolved.Warnings = append(resolved.Warnings,
				fmt.Sprintf("class %s attribute %s invariant excluded (references out-of-scope class): %s",
					className, out[i].Name, invDescription(inv)))
		}
	}
	if !changed {
		return attrs, false
	}
	return out, true
}

// filterLogicsForSurface keeps only logics whose expressions reference no out-of-scope class.
func filterLogicsForSurface(
	logics []model_logic.Logic,
	ownerClassKey identity.Key,
	scope surfaceClassScope,
) (included, excluded []model_logic.Logic) {
	for _, logic := range logics {
		missing := missingClassesForLogics(
			[]model_logic.Logic{logic}, ownerClassKey, scope.inScope, scope.classByName, scope.classNames, scope.nav,
		)
		if len(missing) == 0 {
			included = append(included, logic)
			continue
		}
		excluded = append(excluded, logic)
	}
	return included, excluded
}

func invDescription(inv model_logic.Logic) string {
	if inv.Description != "" {
		return inv.Description
	}
	if inv.Spec.Specification != "" {
		return inv.Spec.Specification
	}
	return inv.Key.String()
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
