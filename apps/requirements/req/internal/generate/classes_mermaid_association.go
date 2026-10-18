package generate

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

func attributeNamesJoined(class model_class.Class, keys []identity.Key, sep string) string {
	if len(keys) == 0 {
		return ""
	}
	parts := make([]string, len(keys))
	for i, key := range keys {
		parts[i] = attributeNameFromClass(class, key)
	}
	return strings.Join(parts, sep)
}

func attributeNameFromClass(class model_class.Class, attrKey identity.Key) string {
	for _, attr := range class.Attributes {
		if attr.Key == attrKey {
			return attr.Name
		}
	}
	return attrKey.SubKey
}

func associationUniquenessMermaidTag(
	uniqueness *model_class.AssociationUniqueness,
	fromClass, toClass model_class.Class,
) string {
	if uniqueness == nil {
		return ""
	}
	fromAttrs := attributeNamesJoined(fromClass, uniqueness.FromAttributeKeys, ", ")
	toAttrs := attributeNamesJoined(toClass, uniqueness.ToAttributeKeys, ", ")
	var tuple string
	switch {
	case fromAttrs == "" && toAttrs == "":
		return ""
	case fromAttrs == "":
		tuple = "→ " + toAttrs
	case toAttrs == "":
		tuple = fromAttrs + " →"
	default:
		tuple = fromAttrs + " → " + toAttrs
	}
	// Edge labels use Mermaid's LABEL token, which ends at the next colon; keep "unique" unquoted.
	return fmt.Sprintf("{unique %s}", tuple)
}

func associationUniquenessMermaidTags(
	uniqueness []model_class.AssociationUniqueness,
	fromClass, toClass model_class.Class,
) []string {
	var tags []string
	for i := range uniqueness {
		if tag := associationUniquenessMermaidTag(&uniqueness[i], fromClass, toClass); tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

// classesMermaidAssociationLinkLabel formats the edge label for a direct association arrow.
func classesMermaidAssociationLinkLabel(assoc model_class.Association, fromClass, toClass model_class.Class) string {
	tags := associationUniquenessMermaidTags(assoc.Uniqueness, fromClass, toClass)
	if len(tags) == 0 {
		return assoc.Name
	}
	return assoc.Name + "<br/>" + strings.Join(tags, "<br/>")
}

// classesMermaidAssociationNodeTitle formats the dashed association link node title
// when an association class decomposes the edge.
func classesMermaidAssociationNodeTitle(assoc model_class.Association, fromClass, toClass model_class.Class) string {
	return classesMermaidAssociationLinkLabel(assoc, fromClass, toClass)
}
