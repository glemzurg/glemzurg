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
	fromAttrs := attributeNamesJoined(fromClass, uniqueness.FromAttributeKeys, "+")
	toAttrs := attributeNamesJoined(toClass, uniqueness.ToAttributeKeys, "+")
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
	return fmt.Sprintf("{unique: %s}", tuple)
}

// classesMermaidAssociationLinkLabel formats the edge label for a direct association arrow.
func classesMermaidAssociationLinkLabel(assoc model_class.Association, fromClass, toClass model_class.Class) string {
	tag := associationUniquenessMermaidTag(assoc.Uniqueness, fromClass, toClass)
	if tag == "" {
		return assoc.Name
	}
	return assoc.Name + "<br/>" + tag
}

// classesMermaidAssociationNodeTitle formats the dashed association link node title
// when an association class decomposes the edge.
func classesMermaidAssociationNodeTitle(assoc model_class.Association, fromClass, toClass model_class.Class) string {
	return classesMermaidAssociationLinkLabel(assoc, fromClass, toClass)
}
