package surface

import (
	"fmt"
	"maps"
	"sort"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
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

	// ModelInvariants is a filtered copy of Model.Invariants.
	ModelInvariants []model_logic.Logic

	// Warnings collects non-fatal issues found during resolution.
	Warnings []string
}

// Resolve resolves a SurfaceSpecification against a model,
// producing the concrete set of classes, associations, and invariants
// for simulation.
func Resolve(spec *SurfaceSpecification, model *core.Model) (*ResolvedSurface, error) {
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
	collectIncludedClasses(spec, model, resolved)

	// 2. Apply excludes.
	if spec != nil {
		for _, ck := range spec.ExcludeClasses {
			delete(resolved.Classes, ck)
		}
	}

	// 3. Reject classes with no state machine — simulation cannot run them.
	var stateless []string
	for _, class := range resolved.Classes {
		if len(class.States) == 0 {
			stateless = append(stateless, class.Name)
		}
	}
	if len(stateless) > 0 {
		return nil, fmt.Errorf(
			"surface includes %d class(es) without a state machine: %s",
			len(stateless),
			strings.Join(statelessNamesSorted(stateless), ", "),
		)
	}

	// 4. Resolve associations.
	resolveAssociations(model, resolved)

	// 5. Scope invariants.
	scopeModelInvariants(model, resolved)

	// 6. Validate: at least one simulatable class must remain.
	if len(resolved.Classes) == 0 {
		return nil, fmt.Errorf("no simulatable classes remain after surface area filtering")
	}

	return resolved, nil
}

// collectIncludedClasses populates resolved.Classes based on spec includes.
func collectIncludedClasses(spec *SurfaceSpecification, model *core.Model, resolved *ResolvedSurface) {
	if spec == nil || spec.IsEmpty() {
		addAllNonRealizedClasses(model, resolved)
		return
	}

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

// resolveAssociations keeps only associations where both endpoints are in scope.
func resolveAssociations(model *core.Model, resolved *ResolvedSurface) {
	allAssocs := model.GetClassAssociations()
	for assocKey, assoc := range allAssocs {
		_, fromIn := resolved.Classes[assoc.FromClassKey]
		_, toIn := resolved.Classes[assoc.ToClassKey]
		if fromIn && toIn {
			resolved.Associations[assocKey] = assoc
		} else if fromIn || toIn {
			resolved.Warnings = append(resolved.Warnings,
				fmt.Sprintf("association %s dropped: one endpoint is outside the surface", assoc.Name))

			if fromIn && assoc.ToMultiplicity.LowerBound >= 1 {
				resolved.Warnings = append(resolved.Warnings,
					fmt.Sprintf("class %s has mandatory association to excluded class via %s — creation chain will not cascade",
						assoc.FromClassKey.String(), assoc.Name))
			}
		}
	}
}

// scopeModelInvariants filters model invariants to those relevant to the resolved surface.
func scopeModelInvariants(model *core.Model, resolved *ResolvedSurface) {
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
	included, excluded := ScopeInvariantsWithAllClasses(model.Invariants, inScopeClassNames, allClassNames)
	resolved.ModelInvariants = included
	for _, inv := range excluded {
		resolved.Warnings = append(resolved.Warnings,
			fmt.Sprintf("invariant excluded (references out-of-scope class): %s", inv.Description))
	}
}

// addAllNonRealizedClasses adds all classes from non-realized domains.
func addAllNonRealizedClasses(model *core.Model, resolved *ResolvedSurface) {
	for _, domain := range model.Domains {
		if domain.Realized {
			continue
		}
		for _, subdomain := range domain.Subdomains {
			maps.Copy(resolved.Classes, subdomain.Classes)
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

func statelessNamesSorted(names []string) []string {
	sorted := append([]string(nil), names...)
	sort.Strings(sorted)
	return sorted
}

// classSpecifier scopes a class lookup. Unqualified specifiers match any subdomain.
type classSpecifier struct {
	domainSubKey    string
	subdomainSubKey string
	classToken      string
}

// ResolveClassKeysByName resolves include-class entries to class keys.
//
// Each entry may be:
//   - class name or class subkey (case-insensitive), matching any subdomain;
//   - subdomain/class (e.g. wallet/partner);
//   - domain/subdomain/class (e.g. finance/wallet/partner).
func ResolveClassKeysByName(model *core.Model, names []string) ([]identity.Key, error) {
	seen := make(map[string]bool)
	var keys []identity.Key

	for _, name := range names {
		spec, err := parseClassSpecifier(name)
		if err != nil {
			return nil, err
		}

		matched, err := resolveClassSpecifier(model, spec)
		if err != nil {
			return nil, err
		}
		for _, key := range matched {
			keyStr := key.String()
			if seen[keyStr] {
				continue
			}
			seen[keyStr] = true
			keys = append(keys, key)
		}
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("no classes matched names: %s", strings.Join(names, ", "))
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
	return keys, nil
}

func parseClassSpecifier(raw string) (classSpecifier, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return classSpecifier{}, fmt.Errorf("class specifier is empty")
	}
	if !strings.Contains(raw, "/") {
		return classSpecifier{classToken: raw}, nil
	}

	parts := strings.Split(raw, "/")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
		if parts[i] == "" {
			return classSpecifier{}, fmt.Errorf("invalid class specifier %q: path segments must be non-empty", raw)
		}
	}

	switch len(parts) {
	case 2:
		return classSpecifier{
			subdomainSubKey: parts[0],
			classToken:      parts[1],
		}, nil
	case 3:
		return classSpecifier{
			domainSubKey:    parts[0],
			subdomainSubKey: parts[1],
			classToken:      parts[2],
		}, nil
	default:
		return classSpecifier{}, fmt.Errorf(
			"invalid class specifier %q: use class, subdomain/class, or domain/subdomain/class",
			raw,
		)
	}
}

func resolveClassSpecifier(model *core.Model, spec classSpecifier) ([]identity.Key, error) {
	if spec.subdomainSubKey == "" {
		return resolveUnqualifiedClassToken(model, spec.classToken)
	}

	var subdomainHits []struct {
		domainSubKey string
		subdomain    model_domain.Subdomain
	}

	for _, domain := range model.Domains {
		if spec.domainSubKey != "" && !strings.EqualFold(domain.Key.SubKey, spec.domainSubKey) {
			continue
		}
		for _, subdomain := range domain.Subdomains {
			if !strings.EqualFold(subdomain.Key.SubKey, spec.subdomainSubKey) {
				continue
			}
			subdomainHits = append(subdomainHits, struct {
				domainSubKey string
				subdomain    model_domain.Subdomain
			}{domainSubKey: domain.Key.SubKey, subdomain: subdomain})
		}
	}

	if len(subdomainHits) == 0 {
		if spec.domainSubKey != "" {
			return nil, fmt.Errorf(
				"no subdomain %q in domain %q for class %q",
				spec.subdomainSubKey, spec.domainSubKey, spec.classToken,
			)
		}
		return nil, fmt.Errorf("no subdomain %q for class %q", spec.subdomainSubKey, spec.classToken)
	}

	if spec.domainSubKey == "" && len(subdomainHits) > 1 {
		domains := make([]string, len(subdomainHits))
		for i, hit := range subdomainHits {
			domains[i] = hit.domainSubKey
		}
		sort.Strings(domains)
		return nil, fmt.Errorf(
			"subdomain %q is ambiguous across domains %s; use domain/subdomain/class",
			spec.subdomainSubKey,
			strings.Join(domains, ", "),
		)
	}

	var keys []identity.Key
	for _, hit := range subdomainHits {
		classKey, found := matchClassToken(hit.subdomain, spec.classToken)
		if !found {
			return nil, fmt.Errorf(
				"no class %q in subdomain %s/%s",
				spec.classToken, hit.domainSubKey, spec.subdomainSubKey,
			)
		}
		keys = append(keys, classKey)
	}
	return keys, nil
}

func resolveUnqualifiedClassToken(model *core.Model, classToken string) ([]identity.Key, error) {
	var keys []identity.Key
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			classKey, found := matchClassToken(subdomain, classToken)
			if found {
				keys = append(keys, classKey)
			}
		}
	}
	return keys, nil
}

func matchClassToken(subdomain model_domain.Subdomain, classToken string) (identity.Key, bool) {
	want := strings.ToLower(strings.TrimSpace(classToken))
	for classKey, class := range subdomain.Classes {
		if strings.EqualFold(classKey.SubKey, want) || strings.EqualFold(class.Name, want) {
			return classKey, true
		}
	}
	return identity.Key{}, false
}
