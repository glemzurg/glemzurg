package model_class

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
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

// AssociationEnd describes one end of an association: which class and its multiplicity.
type AssociationEnd struct {
	ClassKey     identity.Key
	Multiplicity Multiplicity
}

func NewAssociation(key identity.Key, name, details string, from, to AssociationEnd, associationClassKey *identity.Key, umlComment string) (association Association, err error) {
	association = Association{
		Key:                 key,
		Name:                name,
		Details:             details,
		FromClassKey:        from.ClassKey,
		FromMultiplicity:    from.Multiplicity,
		ToClassKey:          to.ClassKey,
		ToMultiplicity:      to.Multiplicity,
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
		return &coreerr.ValidationError{
			Code:    coreerr.AssocKeyInvalid,
			Message: fmt.Sprintf("Key: %s", err.Error()),
			Field:   "Key",
		}
	}
	if a.Key.KeyType != identity.KEY_TYPE_CLASS_ASSOCIATION {
		return &coreerr.ValidationError{
			Code:    coreerr.AssocKeyTypeInvalid,
			Message: fmt.Sprintf("key: invalid key type '%s' for association", a.Key.KeyType),
			Field:   "Key",
			Got:     a.Key.KeyType,
			Want:    identity.KEY_TYPE_CLASS_ASSOCIATION,
		}
	}

	// Name is required.
	if a.Name == "" {
		return &coreerr.ValidationError{
			Code:    coreerr.AssocNameRequired,
			Message: "Name is required",
			Field:   "Name",
		}
	}

	// Validate the FromClassKey.
	if err := a.FromClassKey.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.AssocFromkeyInvalid,
			Message: fmt.Sprintf("FromClassKey: %s", err.Error()),
			Field:   "FromClassKey",
		}
	}
	if a.FromClassKey.KeyType != identity.KEY_TYPE_CLASS {
		return &coreerr.ValidationError{
			Code:    coreerr.AssocFromkeyTypeInvalid,
			Message: fmt.Sprintf("fromClassKey: invalid key type '%s' for from class", a.FromClassKey.KeyType),
			Field:   "FromClassKey",
			Got:     a.FromClassKey.KeyType,
			Want:    identity.KEY_TYPE_CLASS,
		}
	}

	// Validate the ToClassKey.
	if err := a.ToClassKey.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.AssocTokeyInvalid,
			Message: fmt.Sprintf("ToClassKey: %s", err.Error()),
			Field:   "ToClassKey",
		}
	}
	if a.ToClassKey.KeyType != identity.KEY_TYPE_CLASS {
		return &coreerr.ValidationError{
			Code:    coreerr.AssocTokeyTypeInvalid,
			Message: fmt.Sprintf("toClassKey: invalid key type '%s' for to class", a.ToClassKey.KeyType),
			Field:   "ToClassKey",
			Got:     a.ToClassKey.KeyType,
			Want:    identity.KEY_TYPE_CLASS,
		}
	}

	// Validate multiplicities as properties.
	if err := a.FromMultiplicity.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.AssocFromMultInvalid,
			Message: fmt.Sprintf("FromMultiplicity: %s", err.Error()),
			Field:   "FromMultiplicity",
		}
	}
	if err := a.ToMultiplicity.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.AssocToMultInvalid,
			Message: fmt.Sprintf("ToMultiplicity: %s", err.Error()),
			Field:   "ToMultiplicity",
		}
	}
	// Validate AssociationClassKey FK key type and constraints.
	if a.AssociationClassKey != nil {
		if err := a.validateAssociationClassKey(); err != nil {
			return err
		}
	}
	return nil
}

// validateAssociationClassKey validates the AssociationClassKey field.
func (a *Association) validateAssociationClassKey() error {
	if err := a.AssociationClassKey.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.AssocAssocclassInvalid,
			Message: fmt.Sprintf("AssociationClassKey: %s", err.Error()),
			Field:   "AssociationClassKey",
		}
	}
	if a.AssociationClassKey.KeyType != identity.KEY_TYPE_CLASS {
		return &coreerr.ValidationError{
			Code:    coreerr.AssocAssocclassType,
			Message: fmt.Sprintf("AssociationClassKey: invalid key type '%s' for class", a.AssociationClassKey.KeyType),
			Field:   "AssociationClassKey",
			Got:     a.AssociationClassKey.KeyType,
			Want:    identity.KEY_TYPE_CLASS,
		}
	}
	if *a.AssociationClassKey == a.FromClassKey {
		return &coreerr.ValidationError{
			Code:    coreerr.AssocAssocclassSameFrom,
			Message: "AssociationClassKey cannot be the same as FromClassKey",
			Field:   "AssociationClassKey",
			Got:     a.AssociationClassKey.String(),
		}
	}
	if *a.AssociationClassKey == a.ToClassKey {
		return &coreerr.ValidationError{
			Code:    coreerr.AssocAssocclassSameTo,
			Message: "AssociationClassKey cannot be the same as ToClassKey",
			Field:   "AssociationClassKey",
			Got:     a.AssociationClassKey.String(),
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
// - AssociationClassKey (if set) must exist in the classes map.
func (a *Association) ValidateReferences(classes map[identity.Key]bool) error {
	if !classes[a.FromClassKey] {
		return &coreerr.ValidationError{
			Code:    coreerr.AssocFromNotfound,
			Message: fmt.Sprintf("association '%s' references non-existent from class '%s'", a.Key.String(), a.FromClassKey.String()),
			Field:   "FromClassKey",
			Got:     a.FromClassKey.String(),
		}
	}
	if !classes[a.ToClassKey] {
		return &coreerr.ValidationError{
			Code:    coreerr.AssocToNotfound,
			Message: fmt.Sprintf("association '%s' references non-existent to class '%s'", a.Key.String(), a.ToClassKey.String()),
			Field:   "ToClassKey",
			Got:     a.ToClassKey.String(),
		}
	}
	if a.AssociationClassKey != nil {
		if !classes[*a.AssociationClassKey] {
			return &coreerr.ValidationError{
				Code:    coreerr.AssocAssocclassNotfound,
				Message: fmt.Sprintf("association '%s' references non-existent association class '%s'", a.Key.String(), a.AssociationClassKey.String()),
				Field:   "AssociationClassKey",
				Got:     a.AssociationClassKey.String(),
			}
		}
	}
	return nil
}
