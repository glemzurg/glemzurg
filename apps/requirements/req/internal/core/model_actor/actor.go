package model_actor

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

const (
	_USER_TYPE_PERSON = "person"
	_USER_TYPE_SYSTEM = "system"
)

// An actor is an external user of this system, either a person or another system.
type Actor struct {
	Key             identity.Key
	Name            string
	Details         string        // Markdown.
	Type            string        // "person" or "system"
	SuperclassOfKey *identity.Key // If this actor is part of a generalization as the superclass.
	SubclassOfKey   *identity.Key // If this actor is part of a generalization as a subclass.
	UmlComment      string
}

func NewActor(key identity.Key, name, details, userType string, superclassOfKey, subclassOfKey *identity.Key, umlComment string) Actor {
	return Actor{
		Key:             key,
		Name:            name,
		Details:         details,
		Type:            userType,
		SuperclassOfKey: superclassOfKey,
		SubclassOfKey:   subclassOfKey,
		UmlComment:      umlComment,
	}
}

// Validate validates the Actor struct.
func (a *Actor) Validate(ctx *coreerr.ValidationContext) error {
	// Validate the key.
	if err := a.Key.ValidateWithContext(ctx); err != nil {
		return coreerr.New(ctx, coreerr.ActorKeyInvalid, fmt.Sprintf("Key: %s", err.Error()), "Key")
	}
	if a.Key.KeyType != identity.KEY_TYPE_ACTOR {
		return coreerr.NewWithValues(ctx, coreerr.ActorKeyTypeInvalid, fmt.Sprintf("Key: invalid key type '%s' for actor", a.Key.KeyType), "Key", a.Key.KeyType, identity.KEY_TYPE_ACTOR)
	}

	if a.Name == "" {
		return coreerr.New(ctx, coreerr.ActorNameRequired, "Name is required", "Name")
	}

	if a.Type == "" {
		return coreerr.New(ctx, coreerr.ActorTypeRequired, "Type is required", "Type")
	}
	if a.Type != _USER_TYPE_PERSON && a.Type != _USER_TYPE_SYSTEM {
		return coreerr.NewWithValues(ctx, coreerr.ActorTypeInvalid, "Type must be one of: person, system", "Type", a.Type, "one of: person, system")
	}

	// Validate FK key types.
	if a.SuperclassOfKey != nil {
		if err := a.SuperclassOfKey.ValidateWithContext(ctx); err != nil {
			return coreerr.New(ctx, coreerr.ActorSuperkeyInvalid, fmt.Sprintf("SuperclassOfKey: %s", err.Error()), "SuperclassOfKey")
		}
		if a.SuperclassOfKey.KeyType != identity.KEY_TYPE_ACTOR_GENERALIZATION {
			return coreerr.NewWithValues(ctx, coreerr.ActorSuperkeyTypeInvalid, fmt.Sprintf("SuperclassOfKey: invalid key type '%s' for actor generalization", a.SuperclassOfKey.KeyType), "SuperclassOfKey", a.SuperclassOfKey.KeyType, identity.KEY_TYPE_ACTOR_GENERALIZATION)
		}
	}
	if a.SubclassOfKey != nil {
		if err := a.SubclassOfKey.ValidateWithContext(ctx); err != nil {
			return coreerr.New(ctx, coreerr.ActorSubkeyInvalid, fmt.Sprintf("SubclassOfKey: %s", err.Error()), "SubclassOfKey")
		}
		if a.SubclassOfKey.KeyType != identity.KEY_TYPE_ACTOR_GENERALIZATION {
			return coreerr.NewWithValues(ctx, coreerr.ActorSubkeyTypeInvalid, fmt.Sprintf("SubclassOfKey: invalid key type '%s' for actor generalization", a.SubclassOfKey.KeyType), "SubclassOfKey", a.SubclassOfKey.KeyType, identity.KEY_TYPE_ACTOR_GENERALIZATION)
		}
	}

	// SuperclassOfKey and SubclassOfKey cannot be the same generalization.
	if a.SuperclassOfKey != nil && a.SubclassOfKey != nil && *a.SuperclassOfKey == *a.SubclassOfKey {
		return coreerr.New(ctx, coreerr.ActorSuperSubSame, "SuperclassOfKey and SubclassOfKey cannot be the same", "SuperclassOfKey")
	}
	return nil
}

// ValidateWithParent validates the Actor, its key's parent relationship, and all children.
// The parent must be nil (actors are root-level entities).
func (a *Actor) ValidateWithParent(ctx *coreerr.ValidationContext, parent *identity.Key) error {
	// Validate the object itself.
	if err := a.Validate(ctx); err != nil {
		return err
	}
	// Validate the key has the correct parent.
	if err := a.Key.ValidateParentWithContext(ctx, parent); err != nil {
		return err
	}
	// Actor has no children with keys that need validation.
	return nil
}

// ValidateReferences validates that the actor's reference keys point to valid entities.
// - SuperclassOfKey must exist in the generalizations map
// - SubclassOfKey must exist in the generalizations map.
func (a *Actor) ValidateReferences(ctx *coreerr.ValidationContext, generalizations map[identity.Key]bool) error {
	if a.SuperclassOfKey != nil {
		if !generalizations[*a.SuperclassOfKey] {
			return coreerr.NewWithValues(ctx, coreerr.ActorSupergenNotfound, fmt.Sprintf("actor '%s' references non-existent generalization '%s'", a.Key.String(), a.SuperclassOfKey.String()), "SuperclassOfKey", a.SuperclassOfKey.String(), "")
		}
	}
	if a.SubclassOfKey != nil {
		if !generalizations[*a.SubclassOfKey] {
			return coreerr.NewWithValues(ctx, coreerr.ActorSubgenNotfound, fmt.Sprintf("actor '%s' references non-existent generalization '%s'", a.Key.String(), a.SubclassOfKey.String()), "SubclassOfKey", a.SubclassOfKey.String(), "")
		}
	}
	return nil
}
