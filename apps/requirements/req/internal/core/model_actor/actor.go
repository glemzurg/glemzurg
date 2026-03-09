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
func (a *Actor) Validate() error {
	// Validate the key.
	if err := a.Key.Validate(); err != nil {
		return coreerr.New(coreerr.ActorKeyInvalid, fmt.Sprintf("Key: %s", err.Error()), "Key")
	}
	if a.Key.KeyType != identity.KEY_TYPE_ACTOR {
		return coreerr.NewWithValues(coreerr.ActorKeyTypeInvalid, fmt.Sprintf("Key: invalid key type '%s' for actor", a.Key.KeyType), "Key", a.Key.KeyType, identity.KEY_TYPE_ACTOR)
	}

	if a.Name == "" {
		return coreerr.New(coreerr.ActorNameRequired, "Name is required", "Name")
	}

	if a.Type == "" {
		return coreerr.New(coreerr.ActorTypeRequired, "Type is required", "Type")
	}
	if a.Type != _USER_TYPE_PERSON && a.Type != _USER_TYPE_SYSTEM {
		return coreerr.NewWithValues(coreerr.ActorTypeInvalid, "Type must be one of: person, system", "Type", a.Type, "one of: person, system")
	}

	// Validate FK key types.
	if a.SuperclassOfKey != nil {
		if err := a.SuperclassOfKey.Validate(); err != nil {
			return coreerr.New(coreerr.ActorSuperkeyInvalid, fmt.Sprintf("SuperclassOfKey: %s", err.Error()), "SuperclassOfKey")
		}
		if a.SuperclassOfKey.KeyType != identity.KEY_TYPE_ACTOR_GENERALIZATION {
			return coreerr.NewWithValues(coreerr.ActorSuperkeyTypeInvalid, fmt.Sprintf("SuperclassOfKey: invalid key type '%s' for actor generalization", a.SuperclassOfKey.KeyType), "SuperclassOfKey", a.SuperclassOfKey.KeyType, identity.KEY_TYPE_ACTOR_GENERALIZATION)
		}
	}
	if a.SubclassOfKey != nil {
		if err := a.SubclassOfKey.Validate(); err != nil {
			return coreerr.New(coreerr.ActorSubkeyInvalid, fmt.Sprintf("SubclassOfKey: %s", err.Error()), "SubclassOfKey")
		}
		if a.SubclassOfKey.KeyType != identity.KEY_TYPE_ACTOR_GENERALIZATION {
			return coreerr.NewWithValues(coreerr.ActorSubkeyTypeInvalid, fmt.Sprintf("SubclassOfKey: invalid key type '%s' for actor generalization", a.SubclassOfKey.KeyType), "SubclassOfKey", a.SubclassOfKey.KeyType, identity.KEY_TYPE_ACTOR_GENERALIZATION)
		}
	}

	// SuperclassOfKey and SubclassOfKey cannot be the same generalization.
	if a.SuperclassOfKey != nil && a.SubclassOfKey != nil && *a.SuperclassOfKey == *a.SubclassOfKey {
		return coreerr.New(coreerr.ActorSuperSubSame, "SuperclassOfKey and SubclassOfKey cannot be the same", "SuperclassOfKey")
	}
	return nil
}

// ValidateWithParent validates the Actor, its key's parent relationship, and all children.
// The parent must be nil (actors are root-level entities).
func (a *Actor) ValidateWithParent(parent *identity.Key) error {
	// Validate the object itself.
	if err := a.Validate(); err != nil {
		return err
	}
	// Validate the key has the correct parent.
	if err := a.Key.ValidateParent(parent); err != nil {
		return err
	}
	// Actor has no children with keys that need validation.
	return nil
}

// ValidateReferences validates that the actor's reference keys point to valid entities.
// - SuperclassOfKey must exist in the generalizations map
// - SubclassOfKey must exist in the generalizations map.
func (a *Actor) ValidateReferences(generalizations map[identity.Key]bool) error {
	if a.SuperclassOfKey != nil {
		if !generalizations[*a.SuperclassOfKey] {
			return coreerr.NewWithValues(coreerr.ActorSupergenNotfound, fmt.Sprintf("actor '%s' references non-existent generalization '%s'", a.Key.String(), a.SuperclassOfKey.String()), "SuperclassOfKey", a.SuperclassOfKey.String(), "")
		}
	}
	if a.SubclassOfKey != nil {
		if !generalizations[*a.SubclassOfKey] {
			return coreerr.NewWithValues(coreerr.ActorSubgenNotfound, fmt.Sprintf("actor '%s' references non-existent generalization '%s'", a.Key.String(), a.SubclassOfKey.String()), "SubclassOfKey", a.SubclassOfKey.String(), "")
		}
	}
	return nil
}
