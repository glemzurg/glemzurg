package model_class

import (
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// Association is how two classes relate to each other.
type Association struct {
	Key                 identity.Key
	Name                string `validate:"required"`
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
	// Validate the key.
	if err := a.Key.Validate(); err != nil {
		return err
	}
	if a.Key.KeyType != identity.KEY_TYPE_CLASS_ASSOCIATION {
		return errors.Errorf("Key: invalid key type '%s' for association.", a.Key.KeyType)
	}

	// Validate struct tags (Name required).
	if err := _validate.Struct(a); err != nil {
		return err
	}

	// Validate the FromClassKey.
	if err := a.FromClassKey.Validate(); err != nil {
		return err
	}
	if a.FromClassKey.KeyType != identity.KEY_TYPE_CLASS {
		return errors.Errorf("FromClassKey: invalid key type '%s' for from class.", a.FromClassKey.KeyType)
	}

	// Validate the ToClassKey.
	if err := a.ToClassKey.Validate(); err != nil {
		return err
	}
	if a.ToClassKey.KeyType != identity.KEY_TYPE_CLASS {
		return errors.Errorf("ToClassKey: invalid key type '%s' for to class.", a.ToClassKey.KeyType)
	}

	// Validate multiplicities as properties.
	if err := a.FromMultiplicity.Validate(); err != nil {
		return err
	}
	if err := a.ToMultiplicity.Validate(); err != nil {
		return err
	}
	// AssociationClassKey cannot match FromClassKey or ToClassKey.
	if a.AssociationClassKey != nil {
		if *a.AssociationClassKey == a.FromClassKey {
			return errors.New("AssociationClassKey cannot be the same as FromClassKey")
		}
		if *a.AssociationClassKey == a.ToClassKey {
			return errors.New("AssociationClassKey cannot be the same as ToClassKey")
		}
	}
	return nil
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

// ValidateReferences validates that the association's class keys reference real classes.
// - FromClassKey must exist in the classes map
// - ToClassKey must exist in the classes map
// - AssociationClassKey (if set) must exist in the classes map
func (a *Association) ValidateReferences(classes map[identity.Key]bool) error {
	if !classes[a.FromClassKey] {
		return errors.Errorf("association '%s' references non-existent from class '%s'", a.Key.String(), a.FromClassKey.String())
	}
	if !classes[a.ToClassKey] {
		return errors.Errorf("association '%s' references non-existent to class '%s'", a.Key.String(), a.ToClassKey.String())
	}
	if a.AssociationClassKey != nil {
		if !classes[*a.AssociationClassKey] {
			return errors.Errorf("association '%s' references non-existent association class '%s'", a.Key.String(), a.AssociationClassKey.String())
		}
	}
	return nil
}
