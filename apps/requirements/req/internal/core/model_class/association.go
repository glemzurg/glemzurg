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

func NewAssociation(key identity.Key, name, details string, from, to AssociationEnd, associationClassKey *identity.Key, umlComment string) Association {
	return Association{
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
}

// Validate validates the Association struct.
func (a *Association) Validate(ctx *coreerr.ValidationContext) error {
	// Validate the key.
	if err := a.Key.ValidateWithContext(ctx); err != nil {
		return coreerr.New(ctx, coreerr.AssocKeyInvalid, fmt.Sprintf("Key: %s", err.Error()), "Key")
	}
	if a.Key.KeyType != identity.KEY_TYPE_CLASS_ASSOCIATION {
		return coreerr.NewWithValues(ctx, coreerr.AssocKeyTypeInvalid, fmt.Sprintf("key: invalid key type '%s' for association", a.Key.KeyType), "Key", a.Key.KeyType, identity.KEY_TYPE_CLASS_ASSOCIATION)
	}

	// Name is required.
	if a.Name == "" {
		return coreerr.New(ctx, coreerr.AssocNameRequired, "Name is required", "Name")
	}

	// Validate the FromClassKey.
	if err := a.FromClassKey.ValidateWithContext(ctx); err != nil {
		return coreerr.New(ctx, coreerr.AssocFromkeyInvalid, fmt.Sprintf("FromClassKey: %s", err.Error()), "FromClassKey")
	}
	if a.FromClassKey.KeyType != identity.KEY_TYPE_CLASS {
		return coreerr.NewWithValues(ctx, coreerr.AssocFromkeyTypeInvalid, fmt.Sprintf("fromClassKey: invalid key type '%s' for from class", a.FromClassKey.KeyType), "FromClassKey", a.FromClassKey.KeyType, identity.KEY_TYPE_CLASS)
	}

	// Validate the ToClassKey.
	if err := a.ToClassKey.ValidateWithContext(ctx); err != nil {
		return coreerr.New(ctx, coreerr.AssocTokeyInvalid, fmt.Sprintf("ToClassKey: %s", err.Error()), "ToClassKey")
	}
	if a.ToClassKey.KeyType != identity.KEY_TYPE_CLASS {
		return coreerr.NewWithValues(ctx, coreerr.AssocTokeyTypeInvalid, fmt.Sprintf("toClassKey: invalid key type '%s' for to class", a.ToClassKey.KeyType), "ToClassKey", a.ToClassKey.KeyType, identity.KEY_TYPE_CLASS)
	}

	// Validate multiplicities as properties.
	if err := a.FromMultiplicity.Validate(ctx); err != nil {
		return coreerr.New(ctx, coreerr.AssocFromMultInvalid, fmt.Sprintf("FromMultiplicity: %s", err.Error()), "FromMultiplicity")
	}
	if err := a.ToMultiplicity.Validate(ctx); err != nil {
		return coreerr.New(ctx, coreerr.AssocToMultInvalid, fmt.Sprintf("ToMultiplicity: %s", err.Error()), "ToMultiplicity")
	}
	// Validate AssociationClassKey FK key type and constraints.
	if a.AssociationClassKey != nil {
		if err := a.validateAssociationClassKey(ctx); err != nil {
			return err
		}
	}
	return nil
}

// validateAssociationClassKey validates the AssociationClassKey field.
func (a *Association) validateAssociationClassKey(ctx *coreerr.ValidationContext) error {
	if err := a.AssociationClassKey.ValidateWithContext(ctx); err != nil {
		return coreerr.New(ctx, coreerr.AssocAssocclassInvalid, fmt.Sprintf("AssociationClassKey: %s", err.Error()), "AssociationClassKey")
	}
	if a.AssociationClassKey.KeyType != identity.KEY_TYPE_CLASS {
		return coreerr.NewWithValues(ctx, coreerr.AssocAssocclassType, fmt.Sprintf("AssociationClassKey: invalid key type '%s' for class", a.AssociationClassKey.KeyType), "AssociationClassKey", a.AssociationClassKey.KeyType, identity.KEY_TYPE_CLASS)
	}
	if *a.AssociationClassKey == a.FromClassKey {
		return coreerr.NewWithValues(ctx, coreerr.AssocAssocclassSameFrom, "AssociationClassKey cannot be the same as FromClassKey", "AssociationClassKey", a.AssociationClassKey.String(), "")
	}
	if *a.AssociationClassKey == a.ToClassKey {
		return coreerr.NewWithValues(ctx, coreerr.AssocAssocclassSameTo, "AssociationClassKey cannot be the same as ToClassKey", "AssociationClassKey", a.AssociationClassKey.String(), "")
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
func (a *Association) ValidateWithParent(ctx *coreerr.ValidationContext, parent *identity.Key) error {
	// Validate the object itself.
	if err := a.Validate(ctx); err != nil {
		return err
	}
	// Validate the key has the correct parent.
	if err := a.Key.ValidateParentWithContext(ctx, parent); err != nil {
		return err
	}
	// Association has no children with keys that need validation.
	return nil
}

// ValidateReferences validates that the association's class keys reference real classes.
// - FromClassKey must exist in the classes map
// - ToClassKey must exist in the classes map
// - AssociationClassKey (if set) must exist in the classes map.
func (a *Association) ValidateReferences(ctx *coreerr.ValidationContext, classes map[identity.Key]bool) error {
	if !classes[a.FromClassKey] {
		return coreerr.NewWithValues(ctx, coreerr.AssocFromNotfound, fmt.Sprintf("association '%s' references non-existent from class '%s'", a.Key.String(), a.FromClassKey.String()), "FromClassKey", a.FromClassKey.String(), "")
	}
	if !classes[a.ToClassKey] {
		return coreerr.NewWithValues(ctx, coreerr.AssocToNotfound, fmt.Sprintf("association '%s' references non-existent to class '%s'", a.Key.String(), a.ToClassKey.String()), "ToClassKey", a.ToClassKey.String(), "")
	}
	if a.AssociationClassKey != nil {
		if !classes[*a.AssociationClassKey] {
			return coreerr.NewWithValues(ctx, coreerr.AssocAssocclassNotfound, fmt.Sprintf("association '%s' references non-existent association class '%s'", a.Key.String(), a.AssociationClassKey.String()), "AssociationClassKey", a.AssociationClassKey.String(), "")
		}
	}
	return nil
}
