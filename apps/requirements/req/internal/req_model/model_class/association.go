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
	Details             string        // Markdown.
	FromClassKey        identity.Key  // The class on one end of the association.
	FromMultiplicity    Multiplicity  // The multiplicity from one end of the association.
	ToClassKey          identity.Key  // The class on the other end of the association.
	ToMultiplicity      Multiplicity  // The multiplicity on the other end of the association.
	AssociationClassKey *identity.Key // Any class that points to this association.
	UmlComment          string
}

func NewAssociation(key identity.Key, name, details string, fromClassKey identity.Key, fromMultiplicity Multiplicity, toClassKey identity.Key, toMultiplicity Multiplicity, associationClassKey *identity.Key, umlComment string) (association Association, err error) {

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

	if err = association.Validate(); err != nil {
		return Association{}, err
	}

	return association, nil
}

// Validate validates the Association struct.
func (a *Association) Validate() error {
	return validation.ValidateStruct(a,
		validation.Field(&a.Key, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_CLASS_ASSOCIATION {
				return errors.Errorf("invalid key type '%s' for association", k.KeyType())
			}
			return nil
		})),
		validation.Field(&a.Name, validation.Required),
		validation.Field(&a.FromClassKey, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_CLASS {
				return errors.Errorf("invalid key type '%s' for from class", k.KeyType())
			}
			return nil
		})),
		validation.Field(&a.ToClassKey, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_CLASS {
				return errors.Errorf("invalid key type '%s' for to class", k.KeyType())
			}
			return nil
		})),
	)
}

func (a *Association) Includes(classKey identity.Key) (included bool) {
	if a.FromClassKey == classKey || a.ToClassKey == classKey {
		return true
	}
	if a.AssociationClassKey != nil && *a.AssociationClassKey == classKey {
		return true
	}
	return false
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

// ValidateWithParent validates the Association, its key's parent relationship, and all children.
// The parent may be a Subdomain, Domain, or nil (for model-level associations).
func (a *Association) ValidateWithParent(parent *identity.Key) error {
	// Validate the object itself.
	if err := a.Validate(); err != nil {
		return err
	}
	// Validate the key has the correct parent.
	if err := a.Key.ValidateParent(parent); err != nil {
		return err
	}
	// Association has no children with keys that need validation.
	return nil
}
