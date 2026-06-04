package generate

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// associationClassMiddleMultiplicity is the multiplicity shown on the association-class
// side of each decomposed leg in class diagrams.
const associationClassMiddleMultiplicity = "1..*"

func associationClassFromLegFromMultiplicity(assoc model_class.Association) string {
	return assoc.FromMultiplicity.String()
}

func associationClassFromLegToMultiplicity(_ model_class.Association) string {
	return associationClassMiddleMultiplicity
}

func associationClassToLegFromMultiplicity(_ model_class.Association) string {
	return associationClassMiddleMultiplicity
}

func associationClassToLegToMultiplicity(assoc model_class.Association) string {
	return associationClassToEndpointMultiplicity(assoc.ToMultiplicity)
}

// associationClassToEndpointMultiplicity is the multiplicity on the far endpoint of the
// association-class→to leg. A many-to-many endpoint association still decomposes to one
// row per association-class instance on that leg.
func associationClassToEndpointMultiplicity(to model_class.Multiplicity) string {
	if to.LowerBound == 0 && to.HigherBound == 0 {
		return "1"
	}
	value := to.String()
	if value == "*" {
		return "1"
	}
	return value
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