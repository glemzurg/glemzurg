package surface

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
)

// ResolvedSurface is the concrete output of resolving a SurfaceSpecification
// against a model. It contains the exact set of classes, associations, and
// invariants that will participate in the simulation.
type ResolvedSurface struct {
	// Classes is the set of class keys included in the simulation.
	Classes map[identity.Key]model_class.Class

	// Associations contains only associations where BOTH endpoints are
	// in the Classes set. Cross-boundary associations are excluded.
	Associations map[identity.Key]model_class.Association

	// ModelInvariants is a filtered copy of Model.TlaInvariants.
	ModelInvariants []string

	// Warnings collects non-fatal issues found during resolution.
	Warnings []string
}

// Resolve resolves a SurfaceSpecification against a model,
// producing the concrete set of classes, associations, and invariants
// for simulation.
func Resolve(spec *SurfaceSpecification, model *req_model.Model) (*ResolvedSurface, error) {
	// Validate spec keys exist in model.
	if spec != nil {
		if err := spec.Validate(model); err != nil {
			return nil, fmt.Errorf("surface specification validation: %w", err)
		}
	}

	resolved := &ResolvedSurface{
		Classes:      make(map[identity.Key]model_class.Class),
		Associations: make(map[identity.Key]model_class.Association),
	}

	// 1. Collect candidate classes by walking includes.
	if spec == nil || spec.IsEmpty() {
		// Empty spec — include ALL classes from ALL non-realized domains.
		addAllNonRealizedClasses(model, resolved)
	} else {
		// Walk includes.
		includeDomainSet := toKeySet(spec.IncludeDomains)
		includeSubdomainSet := toKeySet(spec.IncludeSubdomains)
		includeClassSet := toKeySet(spec.IncludeClasses)

		for domainKey, domain := range model.Domains {
			if domain.Realized {
				if includeDomainSet[domainKey] {
					resolved.Warnings = append(resolved.Warnings,
						fmt.Sprintf("domain %s is realized (external system) and was excluded", domain.Name))
				}
				continue
			}

			includeDomain := includeDomainSet[domainKey]

			for subdomainKey, subdomain := range domain.Subdomains {
				includeSubdomain := includeSubdomainSet[subdomainKey]

				for classKey, class := range subdomain.Classes {
					if includeDomain || includeSubdomain || includeClassSet[classKey] {
						resolved.Classes[classKey] = class
					}
				}
			}
		}
	}

	// 2. Apply excludes.
	if spec != nil {
		for _, ck := range spec.ExcludeClasses {
			delete(resolved.Classes, ck)
		}
	}

	// 3. Filter to simulatable: remove classes with no states.
	for classKey, class := range resolved.Classes {
		if len(class.States) == 0 {
			delete(resolved.Classes, classKey)
		}
	}

	// 4. Resolve associations: keep only those where both endpoints are in scope.
	allAssocs := model.GetClassAssociations()
	for assocKey, assoc := range allAssocs {
		_, fromIn := resolved.Classes[assoc.FromClassKey]
		_, toIn := resolved.Classes[assoc.ToClassKey]
		if fromIn && toIn {
			resolved.Associations[assocKey] = assoc
		} else if fromIn || toIn {
			resolved.Warnings = append(resolved.Warnings,
				fmt.Sprintf("association %s dropped: one endpoint is outside the surface", assoc.Name))

			// Check for broken mandatory chains.
			if fromIn && assoc.ToMultiplicity.LowerBound >= 1 {
				resolved.Warnings = append(resolved.Warnings,
					fmt.Sprintf("class %s has mandatory association to excluded class via %s — creation chain will not cascade",
						assoc.FromClassKey.String(), assoc.Name))
			}
		}
	}

	// 5. Scope invariants.
	inScopeClassNames := make(map[string]bool, len(resolved.Classes))
	for _, class := range resolved.Classes {
		inScopeClassNames[class.Name] = true
	}
	allClassNames := make(map[string]bool)
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				allClassNames[class.Name] = true
			}
		}
	}
	included, excluded := ScopeInvariantsWithAllClasses(model.TlaInvariants, inScopeClassNames, allClassNames)
	resolved.ModelInvariants = included
	for _, inv := range excluded {
		resolved.Warnings = append(resolved.Warnings,
			fmt.Sprintf("invariant excluded (references out-of-scope class): %s", inv))
	}

	// 6. Validate: at least one simulatable class must remain.
	if len(resolved.Classes) == 0 {
		return nil, fmt.Errorf("no simulatable classes remain after surface area filtering")
	}

	return resolved, nil
}

// addAllNonRealizedClasses adds all classes from non-realized domains.
func addAllNonRealizedClasses(model *req_model.Model, resolved *ResolvedSurface) {
	for _, domain := range model.Domains {
		if domain.Realized {
			continue
		}
		for _, subdomain := range domain.Subdomains {
			for classKey, class := range subdomain.Classes {
				resolved.Classes[classKey] = class
			}
		}
	}
}

// toKeySet converts a slice of keys to a set for O(1) lookup.
func toKeySet(keys []identity.Key) map[identity.Key]bool {
	set := make(map[identity.Key]bool, len(keys))
	for _, k := range keys {
		set[k] = true
	}
	return set
}
