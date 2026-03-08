package model_state

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// Event is what triggers a transition between states.
type Event struct {
	Key     identity.Key
	Name    string
	Details string
	// Children
	Parameters []Parameter
}

func NewEvent(key identity.Key, name, details string, parameters []Parameter) (event Event, err error) {
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
	// Validate the key.
	if err := e.Key.Validate(); err != nil {
		return coreerr.New(coreerr.EventKeyInvalid, fmt.Sprintf("Key: %s", err.Error()), "Key")
	}
	if e.Key.KeyType != identity.KEY_TYPE_EVENT {
		return coreerr.NewWithValues(coreerr.EventKeyTypeInvalid, fmt.Sprintf("Key: invalid key type '%s' for event", e.Key.KeyType), "Key", e.Key.KeyType, identity.KEY_TYPE_EVENT)
	}

	if e.Name == "" {
		return coreerr.New(coreerr.EventNameRequired, "Name is required", "Name")
	}

	return nil
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
	// Validate all children.
	for i := range e.Parameters {
		if err := e.Parameters[i].ValidateWithParent(); err != nil {
			return err
		}
	}
	return nil
}
