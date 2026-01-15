package model_state

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// Event is what triggers a transition between states.
type Event struct {
	Key        identity.Key
	Name       string
	Details    string
	Parameters []EventParameter
}

func NewEvent(key identity.Key, name, details string, parameters []EventParameter) (event Event, err error) {

	event = Event{
		Key:        key,
		Name:       name,
		Details:    details,
		Parameters: parameters,
	}

	err = validation.ValidateStruct(&event,
		validation.Field(&event.Key, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_EVENT {
				return errors.Errorf("invalid key type '%s' for event", k.KeyType())
			}
			return nil
		})),
		validation.Field(&event.Name, validation.Required),
	)
	if err != nil {
		return Event{}, errors.WithStack(err)
	}

	return event, nil
}

// ValidateWithParent validates the Event and verifies its key has the correct parent.
// The parent must be a Class.
func (e *Event) ValidateWithParent(parent *identity.Key) error {
	// Validate the key has the correct parent.
	if err := e.Key.ValidateParent(parent); err != nil {
		return err
	}
	// Event has no children with keys that need validation.
	return nil
}
