package generate

import (
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// renderAssociationClassMermaid reports whether the association should be decomposed into
// a dashed title-only link node between endpoints, with the association class linked
// alongside via a dotted line.
func renderAssociationClassMermaid(assoc model_class.Association) bool {
	return assoc.AssociationClassKey != nil
}

// renderAssociationLinkNodeMermaid reports whether the association gets a dashed link node.
// Association classes always use one; direct associations use one when they carry a uml_comment
// so Mermaid can attach a note to the node.
func renderAssociationLinkNodeMermaid(assoc model_class.Association) bool {
	return renderAssociationClassMermaid(assoc) || associationHasUmlComment(assoc)
}

func associationHasUmlComment(assoc model_class.Association) bool {
	return strings.TrimSpace(assoc.UmlComment) != ""
}

// associationClassKeyNode is a thin wrapper so templates can pass the association class key.
func associationClassKeyNode(assoc model_class.Association) identity.Key {
	return *assoc.AssociationClassKey
}
