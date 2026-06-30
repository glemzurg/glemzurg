package generate

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

func attributeSubKeysJoined(keys []identity.Key, sep string) string {
	if len(keys) == 0 {
		return ""
	}
	parts := make([]string, len(keys))
	for i, key := range keys {
		parts[i] = key.SubKey
	}
	return strings.Join(parts, sep)
}

func associationUniquenessMermaidTag(uniqueness *model_class.AssociationUniqueness) string {
	if uniqueness == nil {
		return ""
	}
	var parts []string
	if fromAttrs := attributeSubKeysJoined(uniqueness.FromAttributeKeys, "+"); fromAttrs != "" {
		parts = append(parts, fromAttrs)
	}
	if toAttrs := attributeSubKeysJoined(uniqueness.ToAttributeKeys, "+"); toAttrs != "" {
		parts = append(parts, toAttrs)
	}
	if len(parts) == 0 {
		return ""
	}
	return fmt.Sprintf("{unique: %s}", strings.Join(parts, ", "))
}

// classesMermaidAssociationLinkLabel formats the edge label for a direct association arrow.
func classesMermaidAssociationLinkLabel(assoc model_class.Association) string {
	tag := associationUniquenessMermaidTag(assoc.Uniqueness)
	if tag == "" {
		return assoc.Name
	}
	return assoc.Name + "<br/>" + tag
}

// classesMermaidAssociationNodeTitle formats the dashed association link node title
// when an association class decomposes the edge.
func classesMermaidAssociationNodeTitle(assoc model_class.Association) string {
	return classesMermaidAssociationLinkLabel(assoc)
}
