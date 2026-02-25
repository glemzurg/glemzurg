package model_actor

import (
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

var _validate = validator.New()

const (
	_USER_TYPE_PERSON = "person"
	_USER_TYPE_SYSTEM = "system"
)

// An actor is a external user of this sytem, either a person or another system.
type Actor struct {
	Key             identity.Key
	Name            string        `validate:"required"`
	Details         string        // Markdown.
	Type            string        `validate:"required,oneof=person system"` // "person" or "system"
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
		return err
	}
	if a.Key.KeyType != identity.KEY_TYPE_ACTOR {
		return errors.Errorf("Key: invalid key type '%s' for actor.", a.Key.KeyType)
	}

	// Validate struct tags (Name required, Type required+oneof).
	if err := _validate.Struct(a); err != nil {
		return err
	}

	// Validate FK key types.
	if a.SuperclassOfKey != nil {
		if err := a.SuperclassOfKey.Validate(); err != nil {
			return errors.Wrap(err, "SuperclassOfKey")
		}
		if a.SuperclassOfKey.KeyType != identity.KEY_TYPE_ACTOR_GENERALIZATION {
			return errors.Errorf("SuperclassOfKey: invalid key type '%s' for actor generalization", a.SuperclassOfKey.KeyType)
		}
	}
	if a.SubclassOfKey != nil {
		if err := a.SubclassOfKey.Validate(); err != nil {
			return errors.Wrap(err, "SubclassOfKey")
		}
		if a.SubclassOfKey.KeyType != identity.KEY_TYPE_ACTOR_GENERALIZATION {
			return errors.Errorf("SubclassOfKey: invalid key type '%s' for actor generalization", a.SubclassOfKey.KeyType)
		}
	}

	// SuperclassOfKey and SubclassOfKey cannot be the same generalization.
	if a.SuperclassOfKey != nil && a.SubclassOfKey != nil && *a.SuperclassOfKey == *a.SubclassOfKey {
		return errors.New("SuperclassOfKey and SubclassOfKey cannot be the same")
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
// - SubclassOfKey must exist in the generalizations map
func (a *Actor) ValidateReferences(generalizations map[identity.Key]bool) error {
	if a.SuperclassOfKey != nil {
		if !generalizations[*a.SuperclassOfKey] {
			return errors.Errorf("actor '%s' references non-existent generalization '%s'", a.Key.String(), a.SuperclassOfKey.String())
		}
	}
	if a.SubclassOfKey != nil {
		if !generalizations[*a.SubclassOfKey] {
			return errors.Errorf("actor '%s' references non-existent generalization '%s'", a.Key.String(), a.SubclassOfKey.String())
		}
	}
	return nil
}
