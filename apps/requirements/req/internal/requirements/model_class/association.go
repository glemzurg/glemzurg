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

func NewAssociation(key identity.Key, name, details string, fromClassKey identity.Key, fromMultiplicity Multiplicity, toClassKey identity.Key, toMultiplicity Multiplicity, associationClassKey identity.Key, umlComment string) (association Association, err error) {

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
		validation.Field(&association.Key, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if k.KeyType() != identity.KEY_TYPE_CLASS_ASSOCIATION {
				return errors.Errorf("key must be of type '%s', not '%s'", identity.KEY_TYPE_CLASS_ASSOCIATION, k.KeyType())
			}
			return nil
		})),
		validation.Field(&association.Name, validation.Required),
		validation.Field(&association.FromClassKey, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if k.KeyType() != identity.KEY_TYPE_CLASS {
				return errors.Errorf("key must be of type '%s', not '%s'", identity.KEY_TYPE_CLASS, k.KeyType())
			}
			return nil
		})),
		validation.Field(&association.ToClassKey, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if k.KeyType() != identity.KEY_TYPE_CLASS {
				return errors.Errorf("key must be of type '%s', not '%s'", identity.KEY_TYPE_CLASS, k.KeyType())
			}
			return nil
		})),
	)
	if err != nil {
		return Association{}, errors.WithStack(err)
	}

	return association, nil
}

func (a *Association) Includes(classKey identity.Key) (included bool) {
	return a.FromClassKey == classKey || a.ToClassKey == classKey || a.AssociationClassKey == classKey
}

func (a *Association) Other(classKey identity.Key) (otherKey identity.Key, err error) {
	if !a.Includes(classKey) {
		return identity.Key{}, errors.WithStack(errors.Errorf(`association does not include class: '%s'`, classKey.String()))
	}
	if a.FromClassKey != classKey {
		return a.FromClassKey, nil
	}
	return a.ToClassKey, nil
}
