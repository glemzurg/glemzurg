package parser_json

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
)

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

// ToRequirements converts the associationInOut to model_class.Association.
func (a associationInOut) ToRequirements() (model_class.Association, error) {
	key, err := identity.ParseKey(a.Key)
	if err != nil {
		return model_class.Association{}, err
	}

	fromClassKey, err := identity.ParseKey(a.FromClassKey)
	if err != nil {
		return model_class.Association{}, err
	}

	toClassKey, err := identity.ParseKey(a.ToClassKey)
	if err != nil {
		return model_class.Association{}, err
	}

	// Handle optional pointer field - empty string means nil
	var associationClassKey *identity.Key
	if a.AssociationClassKey != "" {
		k, err := identity.ParseKey(a.AssociationClassKey)
		if err != nil {
			return model_class.Association{}, err
		}
		associationClassKey = &k
	}

	return model_class.Association{
		Key:                 key,
		Name:                a.Name,
		Details:             a.Details,
		FromClassKey:        fromClassKey,
		FromMultiplicity:    a.FromMultiplicity.ToRequirements(),
		ToClassKey:          toClassKey,
		ToMultiplicity:      a.ToMultiplicity.ToRequirements(),
		AssociationClassKey: associationClassKey,
		UmlComment:          a.UmlComment,
	}, nil
}

// FromRequirements creates a associationInOut from model_class.Association.
func FromRequirementsAssociation(a model_class.Association) associationInOut {
	// Handle optional pointer field - nil means empty string
	var associationClassKey string
	if a.AssociationClassKey != nil {
		associationClassKey = a.AssociationClassKey.String()
	}

	return associationInOut{
		Key:                 a.Key.String(),
		Name:                a.Name,
		Details:             a.Details,
		FromClassKey:        a.FromClassKey.String(),
		FromMultiplicity:    FromRequirementsMultiplicity(a.FromMultiplicity),
		ToClassKey:          a.ToClassKey.String(),
		ToMultiplicity:      FromRequirementsMultiplicity(a.ToMultiplicity),
		AssociationClassKey: associationClassKey,
		UmlComment:          a.UmlComment,
	}
}
