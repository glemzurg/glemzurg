package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

// associationInOut is how two classes relate to each other.
type associationInOut struct {
	Key                 string            `json:"key"`
	Name                string            `json:"name"`
	Details             string            `json:"details"` // Markdown.
	FromClassKey        string            `json:"from_class_key"`
	FromMultiplicity    multiplicityInOut `json:"from_multiplicity"`
	ToClassKey          string            `json:"to_class_key"`
	ToMultiplicity      multiplicityInOut `json:"to_multiplicity"`
	AssociationClassKey string            `json:"association_class_key"`
	UmlComment          string            `json:"uml_comment"`
}

// ToRequirements converts the associationInOut to requirements.Association.
func (a associationInOut) ToRequirements() requirements.Association {
	return requirements.Association{
		Key:                 a.Key,
		Name:                a.Name,
		Details:             a.Details,
		FromClassKey:        a.FromClassKey,
		FromMultiplicity:    a.FromMultiplicity.ToRequirements(),
		ToClassKey:          a.ToClassKey,
		ToMultiplicity:      a.ToMultiplicity.ToRequirements(),
		AssociationClassKey: a.AssociationClassKey,
		UmlComment:          a.UmlComment,
	}
}

// FromRequirements creates a associationInOut from requirements.Association.
func FromRequirementsAssociation(a requirements.Association) associationInOut {
	return associationInOut{
		Key:                 a.Key,
		Name:                a.Name,
		Details:             a.Details,
		FromClassKey:        a.FromClassKey,
		FromMultiplicity:    FromRequirementsMultiplicity(a.FromMultiplicity),
		ToClassKey:          a.ToClassKey,
		ToMultiplicity:      FromRequirementsMultiplicity(a.ToMultiplicity),
		AssociationClassKey: a.AssociationClassKey,
		UmlComment:          a.UmlComment,
	}
}
