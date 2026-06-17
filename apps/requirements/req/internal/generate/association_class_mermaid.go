package generate

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// renderAssociationClassMermaid reports whether the association should be decomposed into
// a dashed title-only link node between endpoints, with the association class linked
// alongside via a dotted line.
func renderAssociationClassMermaid(assoc model_class.Association) bool {
	return assoc.AssociationClassKey != nil
}

// associationClassKeyNode is a thin wrapper so templates can pass the association class key.
func associationClassKeyNode(assoc model_class.Association) identity.Key {
	return *assoc.AssociationClassKey
}
