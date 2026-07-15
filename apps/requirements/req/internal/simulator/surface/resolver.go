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

	// 3. Warn about classes with no state machine; they stay in scope for liveness.
	for _, class := range resolved.Classes {
		if len(class.States) == 0 {
			resolved.Warnings = append(resolved.Warnings,
				fmt.Sprintf("class %s has no state machine (liveness only; not simulatable)", class.Name))
		}
	}

	// 4. Resolve associations across the full surface class set.
	// Association-class types are included only when explicitly listed — the surface
	// is the intentional subset; host associations degrade to plain endpoint links
	// when the association class is out of scope (no auto-pull).
	resolveAssociations(model, resolved)

	// 5. Scope invariants.
	scopeModelInvariants(model, resolved)

	// 6. Validate: at least one simulatable class must remain.
	if countSimulatableClasses(resolved.Classes) == 0 {
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
// When the association class is not listed on the surface, the host association is
// kept as a plain endpoint link (AssociationClassKey cleared) — never auto-included.
func resolveAssociations(model *core.Model, resolved *ResolvedSurface) {
	allAssocs := model.GetClassAssociations()
	for assocKey, assoc := range allAssocs {
		_, fromIn := resolved.Classes[assoc.FromClassKey]
		_, toIn := resolved.Classes[assoc.ToClassKey]
		if fromIn && toIn {
			resolved.Associations[assocKey] = associationForSurface(assoc, resolved)
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

// associationForSurface returns assoc for the resolved surface. Out-of-scope
// association classes are stripped so host reify becomes plain endpoint links.
func associationForSurface(assoc model_class.Association, resolved *ResolvedSurface) model_class.Association {
	if assoc.AssociationClassKey == nil {
		return assoc
	}
	if _, acIn := resolved.Classes[*assoc.AssociationClassKey]; acIn {
		return assoc
	}
	stripped := assoc
	stripped.AssociationClassKey = nil
	resolved.Warnings = append(resolved.Warnings,
		fmt.Sprintf("association %s treated as plain links: association class %s is outside the surface",
			assoc.Name, assoc.AssociationClassKey.String()))
	return stripped
}

// scopeModelInvariants filters model invariants to those relevant to the resolved surface.
func scopeModelInvariants(model *core.Model, resolved *ResolvedSurface) {
	inScopeClassNames, allClassNames := classNameSetsForScoping(model, resolved)
	included, excluded := ScopeInvariantsWithAllClasses(model.Invariants, inScopeClassNames, allClassNames)
	resolved.ModelInvariants = included
	for _, inv := range excluded {
		resolved.Warnings = append(resolved.Warnings,
			fmt.Sprintf("invariant excluded (references out-of-scope class): %s", inv.Description))
	}
}

// classNameSetsForScoping builds in-scope and full-model name sets for invariant
// filtering. Both display names and ClassTLAName forms are included so field
// navigations like AccountBalanceChange match out-of-scope association classes.
func classNameSetsForScoping(model *core.Model, resolved *ResolvedSurface) (inScope, all map[string]bool) {
	inScope = make(map[string]bool, len(resolved.Classes)*2)
	for _, class := range resolved.Classes {
		inScope[class.Name] = true
		inScope[model_class.ClassTLAName(class.Name)] = true
	}
	all = make(map[string]bool)
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				all[class.Name] = true
				all[model_class.ClassTLAName(class.Name)] = true
			}
		}
	}
	return inScope, all
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

func countSimulatableClasses(classes map[identity.Key]model_class.Class) int {
	n := 0
	for _, class := range classes {
		if len(class.States) > 0 {
			n++
		}
	}
	return n
}

// classSpecifier scopes a class lookup. Unqualified specifiers match any subdomain.
type classSpecifier struct {
	domainSubKey    string
	subdomainSubKey string
	classToken      string
}

// subdomainPathSpecifier scopes a subdomain lookup by domain/subdomain subkeys.
type subdomainPathSpecifier struct {
	domainSubKey    string
	subdomainSubKey string
}

// ResolveSubdomainKeysByPath resolves include-subdomain entries to subdomain keys.
//
// Each entry may be:
//   - subdomain subkey (e.g. wallet), matching any domain;
//   - domain/subdomain (e.g. finance/wallet).
func ResolveSubdomainKeysByPath(model *core.Model, paths []string) ([]identity.Key, error) {
	seen := make(map[string]bool)
	var keys []identity.Key

	for _, path := range paths {
		spec, err := parseSubdomainPath(path)
		if err != nil {
			return nil, err
		}

		matched, err := resolveSubdomainPath(model, spec)
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
		return nil, fmt.Errorf("no subdomains matched paths: %s", strings.Join(paths, ", "))
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
	return keys, nil
}

func parseSubdomainPath(raw string) (subdomainPathSpecifier, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return subdomainPathSpecifier{}, fmt.Errorf("subdomain path is empty")
	}
	if !strings.Contains(raw, "/") {
		return subdomainPathSpecifier{subdomainSubKey: raw}, nil
	}

	parts := strings.Split(raw, "/")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
		if parts[i] == "" {
			return subdomainPathSpecifier{}, fmt.Errorf("invalid subdomain path %q: path segments must be non-empty", raw)
		}
	}

	switch len(parts) {
	case 2:
		return subdomainPathSpecifier{
			domainSubKey:    parts[0],
			subdomainSubKey: parts[1],
		}, nil
	default:
		return subdomainPathSpecifier{}, fmt.Errorf(
			"invalid subdomain path %q: use subdomain or domain/subdomain",
			raw,
		)
	}
}

func resolveSubdomainPath(model *core.Model, spec subdomainPathSpecifier) ([]identity.Key, error) {
	var hits []identity.Key
	var domainSubKeys []string

	for _, domain := range model.Domains {
		if spec.domainSubKey != "" && !strings.EqualFold(domain.Key.SubKey, spec.domainSubKey) {
			continue
		}
		for subdomainKey, subdomain := range domain.Subdomains {
			if !strings.EqualFold(subdomain.Key.SubKey, spec.subdomainSubKey) {
				continue
			}
			hits = append(hits, subdomainKey)
			if spec.domainSubKey == "" {
				domainSubKeys = append(domainSubKeys, domain.Key.SubKey)
			}
		}
	}

	if len(hits) == 0 {
		if spec.domainSubKey != "" {
			return nil, fmt.Errorf("no subdomain %q in domain %q", spec.subdomainSubKey, spec.domainSubKey)
		}
		return nil, fmt.Errorf("no subdomain %q", spec.subdomainSubKey)
	}

	if spec.domainSubKey == "" && len(hits) > 1 {
		sort.Strings(domainSubKeys)
		return nil, fmt.Errorf(
			"subdomain %q is ambiguous across domains %s; use domain/subdomain",
			spec.subdomainSubKey,
			strings.Join(domainSubKeys, ", "),
		)
	}

	return hits, nil
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
