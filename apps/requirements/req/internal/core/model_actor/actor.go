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

func NewActor(key identity.Key, name, details, userType string, superclassOfKey, subclassOfKey *identity.Key, umlComment string) (actor Actor, err error) {
	actor = Actor{
		Key:             key,
		Name:            name,
		Details:         details,
		Type:            userType,
		SuperclassOfKey: superclassOfKey,
		SubclassOfKey:   subclassOfKey,
		UmlComment:      umlComment,
	}

	if err = actor.Validate(); err != nil {
		return Actor{}, err
	}

	return actor, nil
}

// Validate validates the Actor struct.
func (a *Actor) Validate() error {
	// Validate the key.
	if err := a.Key.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.ActorKeyInvalid,
			Message: fmt.Sprintf("Key: %s", err.Error()),
			Field:   "Key",
		}
	}
	if a.Key.KeyType != identity.KEY_TYPE_ACTOR {
		return &coreerr.ValidationError{
			Code:    coreerr.ActorKeyTypeInvalid,
			Message: fmt.Sprintf("Key: invalid key type '%s' for actor", a.Key.KeyType),
			Field:   "Key",
			Got:     a.Key.KeyType,
			Want:    identity.KEY_TYPE_ACTOR,
		}
	}

	if a.Name == "" {
		return &coreerr.ValidationError{
			Code:    coreerr.ActorNameRequired,
			Message: "Name is required",
			Field:   "Name",
		}
	}

	if a.Type == "" {
		return &coreerr.ValidationError{
			Code:    coreerr.ActorTypeRequired,
			Message: "Type is required",
			Field:   "Type",
		}
	}
	if a.Type != _USER_TYPE_PERSON && a.Type != _USER_TYPE_SYSTEM {
		return &coreerr.ValidationError{
			Code:    coreerr.ActorTypeInvalid,
			Message: "Type must be one of: person, system",
			Field:   "Type",
			Got:     a.Type,
			Want:    "one of: person, system",
		}
	}

	// Validate FK key types.
	if a.SuperclassOfKey != nil {
		if err := a.SuperclassOfKey.Validate(); err != nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ActorSuperkeyInvalid,
				Message: fmt.Sprintf("SuperclassOfKey: %s", err.Error()),
				Field:   "SuperclassOfKey",
			}
		}
		if a.SuperclassOfKey.KeyType != identity.KEY_TYPE_ACTOR_GENERALIZATION {
			return &coreerr.ValidationError{
				Code:    coreerr.ActorSuperkeyTypeInvalid,
				Message: fmt.Sprintf("SuperclassOfKey: invalid key type '%s' for actor generalization", a.SuperclassOfKey.KeyType),
				Field:   "SuperclassOfKey",
				Got:     a.SuperclassOfKey.KeyType,
				Want:    identity.KEY_TYPE_ACTOR_GENERALIZATION,
			}
		}
	}
	if a.SubclassOfKey != nil {
		if err := a.SubclassOfKey.Validate(); err != nil {
			return &coreerr.ValidationError{
				Code:    coreerr.ActorSubkeyInvalid,
				Message: fmt.Sprintf("SubclassOfKey: %s", err.Error()),
				Field:   "SubclassOfKey",
			}
		}
		if a.SubclassOfKey.KeyType != identity.KEY_TYPE_ACTOR_GENERALIZATION {
			return &coreerr.ValidationError{
				Code:    coreerr.ActorSubkeyTypeInvalid,
				Message: fmt.Sprintf("SubclassOfKey: invalid key type '%s' for actor generalization", a.SubclassOfKey.KeyType),
				Field:   "SubclassOfKey",
				Got:     a.SubclassOfKey.KeyType,
				Want:    identity.KEY_TYPE_ACTOR_GENERALIZATION,
			}
		}
	}

	// SuperclassOfKey and SubclassOfKey cannot be the same generalization.
	if a.SuperclassOfKey != nil && a.SubclassOfKey != nil && *a.SuperclassOfKey == *a.SubclassOfKey {
		return &coreerr.ValidationError{
			Code:    coreerr.ActorSuperSubSame,
			Message: "SuperclassOfKey and SubclassOfKey cannot be the same",
			Field:   "SuperclassOfKey",
		}
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
			return &coreerr.ValidationError{
				Code:    coreerr.ActorSupergenNotfound,
				Message: fmt.Sprintf("actor '%s' references non-existent generalization '%s'", a.Key.String(), a.SuperclassOfKey.String()),
				Field:   "SuperclassOfKey",
				Got:     a.SuperclassOfKey.String(),
			}
		}
	}
	if a.SubclassOfKey != nil {
		if !generalizations[*a.SubclassOfKey] {
			return &coreerr.ValidationError{
				Code:    coreerr.ActorSubgenNotfound,
				Message: fmt.Sprintf("actor '%s' references non-existent generalization '%s'", a.Key.String(), a.SubclassOfKey.String()),
				Field:   "SubclassOfKey",
				Got:     a.SubclassOfKey.String(),
			}
		}
	}
	return nil
}
