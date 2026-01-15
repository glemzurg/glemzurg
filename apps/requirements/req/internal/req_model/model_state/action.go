package model_state

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// Action is what happens in a transition between states.
type Action struct {
	Key        identity.Key
	Name       string
	Details    string
	Requires   []string // To enter this action.
	Guarantees []string
	// Derived values for template display.
	FromTransitions []Transition  // Where this action is called from events.
	FromStates      []StateAction // Where this action is called from a state.
}

func NewAction(key identity.Key, name, details string, requires, guarantees []string) (action Action, err error) {

	action = Action{
		Key:        key,
		Name:       name,
		Details:    details,
		Requires:   requires,
		Guarantees: guarantees,
	}

	err = validation.ValidateStruct(&action,
		validation.Field(&action.Key, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_ACTION {
				return errors.Errorf("invalid key type '%s' for action", k.KeyType())
			}
			return nil
		})),
		validation.Field(&action.Name, validation.Required),
	)
	if err != nil {
		return Action{}, errors.WithStack(err)
	}

	return action, nil
}

func (a *Action) SetTriggers(transitions []Transition, stateActions []StateAction) {
	a.FromTransitions = transitions
	a.FromStates = stateActions
}

// ValidateWithParent validates the Action and verifies its key has the correct parent.
// The parent must be a Class.
func (a *Action) ValidateWithParent(parent *identity.Key) error {
	// Validate the key has the correct parent.
	if err := a.Key.ValidateParent(parent); err != nil {
		return err
	}
	// Action has no children with keys that need validation.
	return nil
}
