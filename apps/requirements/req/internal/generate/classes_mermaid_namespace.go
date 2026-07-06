package generate

import (
	"sort"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/generate/req_flat"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

type classesMermaidNamespaceGroup struct {
	Path    string
	Classes []model_class.Class
}

type classesMermaidClassLayout struct {
	LocalClasses []model_class.Class
	Namespaces   []classesMermaidNamespaceGroup
}

func groupClassesMermaidByNamespace(
	reqs *req_flat.Requirements,
	viewerSubdomainKey identity.Key,
	classes []model_class.Class,
) (classesMermaidClassLayout, error) {
	if viewerSubdomainKey.KeyType != identity.KEY_TYPE_SUBDOMAIN {
		sorted := append([]model_class.Class(nil), classes...)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Key.String() < sorted[j].Key.String()
		})
		return classesMermaidClassLayout{LocalClasses: sorted}, nil
	}

	domainLookup := reqs.ClassDomainLookup()
	subdomainLookup := reqs.ClassSubdomainLookup()

	byPath := make(map[string][]model_class.Class)
	var local []model_class.Class

	for _, class := range classes {
		targetDomain := domainLookup[class.Key.String()]
		targetSubdomain := subdomainLookup[class.Key.String()]
		segments, err := model_class.FormatClassMermaidNamespaceSegments(
			viewerSubdomainKey,
			class,
			targetDomain.Name,
			targetSubdomain.Name,
		)
		if err != nil {
			return classesMermaidClassLayout{}, err
		}
		path := mermaidNamespacePathFromSegments(segments)
		if path == "" {
			local = append(local, class)
			continue
		}
		byPath[path] = append(byPath[path], class)
	}

	sort.Slice(local, func(i, j int) bool {
		return local[i].Key.String() < local[j].Key.String()
	})

	paths := make([]string, 0, len(byPath))
	for path := range byPath {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	groups := make([]classesMermaidNamespaceGroup, 0, len(paths))
	for _, path := range paths {
		groupClasses := byPath[path]
		sort.Slice(groupClasses, func(i, j int) bool {
			return groupClasses[i].Key.String() < groupClasses[j].Key.String()
		})
		groups = append(groups, classesMermaidNamespaceGroup{
			Path:    path,
			Classes: groupClasses,
		})
	}

	return classesMermaidClassLayout{
		LocalClasses: local,
		Namespaces:   groups,
	}, nil
}

func mermaidNamespacePathFromSegments(segments []string) string {
	if len(segments) == 0 {
		return ""
	}
	parts := make([]string, len(segments))
	for i, segment := range segments {
		parts[i] = mermaidNamespaceSegment(segment)
	}
	return strings.Join(parts, ".")
}

// mermaidNamespaceSegment reduces a display name to a Mermaid namespace identifier segment.
func mermaidNamespaceSegment(displayName string) string {
	var b strings.Builder
	for _, r := range displayName {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "Scope"
	}
	return b.String()
}
