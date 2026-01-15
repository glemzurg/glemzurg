package model_actor

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
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
	// Helpful data.
	ClassKeys []identity.Key // Classes that implement this actor.
}

func NewActor(key identity.Key, name, details, userType, umlComment string) (actor Actor, err error) {

	actor = Actor{
		Key:        key,
		Name:       name,
		Details:    details,
		Type:       userType,
		UmlComment: umlComment,
	}

	err = validation.ValidateStruct(&actor,
		validation.Field(&actor.Key, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_ACTOR {
				return errors.Errorf("invalid key type '%s' for actor", k.KeyType())
			}
			return nil
		})),
		validation.Field(&actor.Name, validation.Required),
		validation.Field(&actor.Type, validation.Required, validation.In(_USER_TYPE_PERSON, _USER_TYPE_SYSTEM)),
	)
	if err != nil {
		return Actor{}, errors.WithStack(err)
	}

	return actor, nil
}
