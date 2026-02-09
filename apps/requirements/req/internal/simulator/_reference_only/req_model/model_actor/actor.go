package model_actor

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

	"github.com/glemzurg/go-tlaplus/internal/identity"
)

const (
	_USER_TYPE_PERSON = "person"
	_USER_TYPE_SYSTEM = "system"
)

// An actor is a external user of this sytem, either a person or another system.
type Actor struct {
	Key        identity.Key
	Name       string
	Details    string // Markdown.
	Type       string // "person" or "system"
	UmlComment string
}

func NewActor(key identity.Key, name, details, userType, umlComment string) (actor Actor, err error) {

	actor = Actor{
		Key:        key,
		Name:       name,
		Details:    details,
		Type:       userType,
		UmlComment: umlComment,
	}

	if err = actor.Validate(); err != nil {
		return Actor{}, err
	}

	return actor, nil
}

// Validate validates the Actor struct.
func (a *Actor) Validate() error {
	return validation.ValidateStruct(a,
		validation.Field(&a.Key, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_ACTOR {
				return errors.Errorf("invalid key type '%s' for actor", k.KeyType())
			}
			return nil
		})),
		validation.Field(&a.Name, validation.Required),
		validation.Field(&a.Type, validation.Required, validation.In(_USER_TYPE_PERSON, _USER_TYPE_SYSTEM)),
	)
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
