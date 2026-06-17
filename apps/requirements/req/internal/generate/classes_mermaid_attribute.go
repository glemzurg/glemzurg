package generate

import (
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
	member.WriteString(attributeIndexBracketSuffix(attr.IndexNums))
	return member.String()
}
