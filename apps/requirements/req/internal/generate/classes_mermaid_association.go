package generate

import (
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
)

// associationUniquenessMermaidTag returns a diagram tag like "{unique}" or "{3..any}".
// An empty string means uniqueness is any and should not be shown.
func associationUniquenessMermaidTag(m model_class.Multiplicity) string {
	if m.LowerBound == 0 && m.HigherBound == 0 {
		return ""
	}
	if m.LowerBound == 1 && m.HigherBound == 1 {
		return "{unique}"
	}
	s := m.ParsedString()
	if before, found := strings.CutSuffix(s, "..*"); found {
		s = before + "..any"
	}
	return "{" + s + "}"
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
