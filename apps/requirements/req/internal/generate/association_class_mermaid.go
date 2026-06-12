package generate

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// associationClassFromLegFromMultiplicity is the multiplicity on the from-class end of the
// from→association-class leg. Each association-class row belongs to exactly one from instance.
func associationClassFromLegFromMultiplicity(_ model_class.Association) string {
	return "1"
}

// associationClassFromLegToMultiplicity is the multiplicity on the association-class end
// of the from→association-class leg: how many association-class rows per from instance,
// which matches the parent association's to end.
func associationClassFromLegToMultiplicity(assoc model_class.Association) string {
	return assoc.ToMultiplicity.String()
}

// associationClassToLegFromMultiplicity is the multiplicity on the association-class end
// of the association-class→to leg: how many association-class rows per to instance,
// which matches the parent association's from end.
func associationClassToLegFromMultiplicity(assoc model_class.Association) string {
	return assoc.FromMultiplicity.String()
}

func associationClassToLegToMultiplicity(assoc model_class.Association) string {
	return associationClassToEndpointMultiplicity(assoc.ToMultiplicity)
}

// associationClassToEndpointMultiplicity is the multiplicity on the far endpoint of the
// association-class→to leg. Each association-class instance links to exactly one to-class
// instance regardless of how many to endpoints the parent association allows per from.
func associationClassToEndpointMultiplicity(_ model_class.Multiplicity) string {
	return "1"
}

// renderAssociationClassMermaid reports whether the association should be drawn as
// decomposed solid legs through its association class rather than a direct endpoint link.
func renderAssociationClassMermaid(assoc model_class.Association) bool {
	return assoc.AssociationClassKey != nil
}

// associationClassKeyNode is a thin wrapper so templates can pass the association class key.
func associationClassKeyNode(assoc model_class.Association) identity.Key {
	return *assoc.AssociationClassKey
}