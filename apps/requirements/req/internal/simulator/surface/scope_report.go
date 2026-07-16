package surface

import (
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// ScopeKind is how a scope entry is summarized for testers.
type ScopeKind string

const (
	// ScopeSubdomain means every class in that subdomain is in the run.
	ScopeSubdomain ScopeKind = "subdomain"
	// ScopeClass means only this class (not the whole subdomain) is in the run.
	ScopeClass ScopeKind = "class"
)

// ScopeEntry is one line of simulation scope: a full subdomain path or a class path.
// Paths use domain/subdomain or domain/subdomain/class subkeys (same style as include flags).
type ScopeEntry struct {
	Kind ScopeKind `json:"kind"`
	Path string    `json:"path"`
}

// BuildScopeEntries summarizes which model classes participate in a run.
// When every class of a subdomain is in scope, emit one subdomain path; otherwise
// list each in-scope class path. Realized (external) domains are ignored.
func BuildScopeEntries(model *core.Model, inScope map[identity.Key]model_class.Class) []ScopeEntry {
	if model == nil || len(inScope) == 0 {
		return nil
	}

	var entries []ScopeEntry
	for _, domain := range model.Domains {
		if domain.Realized {
			continue
		}
		domainSub := domain.Key.SubKey
		for _, subdomain := range domain.Subdomains {
			subPath := domainSub + "/" + subdomain.Key.SubKey
			total := len(subdomain.Classes)
			if total == 0 {
				continue
			}
			var scoped []model_class.Class
			for classKey, class := range subdomain.Classes {
				if _, ok := inScope[classKey]; ok {
					scoped = append(scoped, class)
				}
			}
			if len(scoped) == 0 {
				continue
			}
			if len(scoped) == total {
				entries = append(entries, ScopeEntry{Kind: ScopeSubdomain, Path: subPath})
				continue
			}
			sort.Slice(scoped, func(i, j int) bool {
				return scoped[i].Key.SubKey < scoped[j].Key.SubKey
			})
			for _, class := range scoped {
				entries = append(entries, ScopeEntry{
					Kind: ScopeClass,
					Path: subPath + "/" + class.Key.SubKey,
				})
			}
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Path != entries[j].Path {
			return entries[i].Path < entries[j].Path
		}
		return entries[i].Kind < entries[j].Kind
	})
	return entries
}

// AllNonRealizedClasses maps every class key in non-realized domains (full-model scope).
func AllNonRealizedClasses(model *core.Model) map[identity.Key]model_class.Class {
	if model == nil {
		return nil
	}
	out := make(map[identity.Key]model_class.Class)
	for _, domain := range model.Domains {
		if domain.Realized {
			continue
		}
		for _, subdomain := range domain.Subdomains {
			for classKey, class := range subdomain.Classes {
				out[classKey] = class
			}
		}
	}
	return out
}
