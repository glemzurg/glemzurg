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

func NewEvent(key identity.Key, name, details string, parameters []Parameter) Event {
	return Event{
		Key:        key,
		Name:       name,
		Details:    details,
		Parameters: parameters,
	}
}

// Validate validates the Event struct.
func (e *Event) Validate(ctx *coreerr.ValidationContext) error {
	// Validate the key.
	if err := e.Key.ValidateWithContext(ctx); err != nil {
		return coreerr.New(ctx, coreerr.EventKeyInvalid, fmt.Sprintf("Key: %s", err.Error()), "Key")
	}
	if e.Key.KeyType != identity.KEY_TYPE_EVENT {
		return coreerr.NewWithValues(ctx, coreerr.EventKeyTypeInvalid, fmt.Sprintf("Key: invalid key type '%s' for event", e.Key.KeyType), "Key", e.Key.KeyType, identity.KEY_TYPE_EVENT)
	}

	if e.Name == "" {
		return coreerr.New(ctx, coreerr.EventNameRequired, "Name is required", "Name")
	}

	return nil
}

// ValidateWithParent validates the Event, its key's parent relationship, and all children.
// The parent must be a Class.
func (e *Event) ValidateWithParent(ctx *coreerr.ValidationContext, parent *identity.Key) error {
	// Validate the object itself.
	if err := e.Validate(ctx); err != nil {
		return err
	}
	// Validate the key has the correct parent.
	if err := e.Key.ValidateParentWithContext(ctx, parent); err != nil {
		return err
	}
	// Validate all children.
	for i := range e.Parameters {
		childCtx := ctx.Child("parameter", fmt.Sprintf("%d", i))
		if err := e.Parameters[i].ValidateWithParent(childCtx); err != nil {
			return err
		}
	}
	return nil
}
