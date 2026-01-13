package model_class

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

// Association is how two classes relate to each other.
type Association struct {
	Key                 identity.Key
	Name                string
	Details             string       // Markdown.
	FromClassKey        identity.Key // The class on one end of the association.
	FromMultiplicity    Multiplicity // The multiplicity from one end of the association.
	ToClassKey          identity.Key // The class on the other end of the association.
	ToMultiplicity      Multiplicity // The multiplicity on the other end of the association.
	AssociationClassKey identity.Key // Any class that points to this association.
	UmlComment          string
}

func NewAssociation(key, name, details, fromClassKey string, fromMultiplicity Multiplicity, toClassKey string, toMultiplicity Multiplicity, associationClassKey, umlComment string) (association Association, err error) {

	association = Association{
		Key:                 key,
		Name:                name,
		Details:             details,
		FromClassKey:        fromClassKey,
		FromMultiplicity:    fromMultiplicity,
		ToClassKey:          toClassKey,
		ToMultiplicity:      toMultiplicity,
		AssociationClassKey: associationClassKey,
		UmlComment:          umlComment,
	}

	err = validation.ValidateStruct(&association,
		validation.Field(&association.Key, validation.Required),
		validation.Field(&association.Name, validation.Required),
		validation.Field(&association.FromClassKey, validation.Required),
		validation.Field(&association.ToClassKey, validation.Required),
	)
	if err != nil {
		return Association{}, errors.WithStack(err)
	}

	return association, nil
}

func (a *Association) Includes(classKey string) (included bool) {
	return a.FromClassKey == classKey || a.ToClassKey == classKey || a.AssociationClassKey == classKey
}

func (a *Association) Other(classKey string) (otherKey string, err error) {
	if !a.Includes(classKey) {
		return "", errors.WithStack(errors.Errorf(`association does not include class: '%s'`, classKey))
	}
	if a.FromClassKey != classKey {
		return a.FromClassKey, nil
	}
	return a.ToClassKey, nil
}

func CreateKeyAssociationLookup(items []Association) (lookup map[string]Association) {
	lookup = map[string]Association{}
	for _, item := range items {
		lookup[item.Key] = item
	}
	return lookup
}
