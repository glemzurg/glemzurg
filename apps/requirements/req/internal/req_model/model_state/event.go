package model_state

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// Event is what triggers a transition between states.
type Event struct {
	Key     identity.Key
	Name    string
	Details string
	// Children
	Parameters []EventParameter
}

func NewEvent(key identity.Key, name, details string, parameters []EventParameter) (event Event, err error) {

	event = Event{
		Key:        key,
		Name:       name,
		Details:    details,
		Parameters: parameters,
	}

	if err = event.Validate(); err != nil {
		return Event{}, err
	}

	return event, nil
}

// Validate validates the Event struct.
func (e *Event) Validate() error {
	return validation.ValidateStruct(e,
		validation.Field(&e.Key, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_EVENT {
				return errors.Errorf("invalid key type '%s' for event", k.KeyType())
			}
			return nil
		})),
		validation.Field(&e.Name, validation.Required),
	)
}

// ValidateWithParent validates the Event, its key's parent relationship, and all children.
// The parent must be a Class.
func (e *Event) ValidateWithParent(parent *identity.Key) error {
	// Validate the object itself.
	if err := e.Validate(); err != nil {
		return err
	}
	// Validate the key has the correct parent.
	if err := e.Key.ValidateParent(parent); err != nil {
		return err
	}
	// Event has no children with keys that need validation.
	return nil
}
