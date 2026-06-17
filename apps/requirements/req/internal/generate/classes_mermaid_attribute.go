package generate

import (
	"slices"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
)

// classesMermaidAttributeMember formats a class attribute for Mermaid member text,
// including derivation prefix and index membership suffix when present.
func classesMermaidAttributeMember(attr model_class.Attribute) string {
	var member strings.Builder
	if attr.DerivationPolicy != nil {
		member.WriteString("/")
	}
	member.WriteString(attr.Name)
	member.WriteString(classesMermaidAttributeIndexSuffix(attr.IndexNums))
	return member.String()
}

func classesMermaidAttributeIndexSuffix(indexNums []uint) string {
	if len(indexNums) == 0 {
		return ""
	}
	sorted := slices.Clone(indexNums)
	slices.Sort(sorted)

	labels := make([]string, 0, len(sorted))
	for _, indexNum := range sorted {
		labels = append(labels, attributeIndexLabel(indexNum))
	}
	return " [" + strings.Join(labels, ",") + "]"
}
