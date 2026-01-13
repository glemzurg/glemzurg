package model_state

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

// Action is what happens in a transition between states.
type Action struct {
	Key        string
	Name       string
	Details    string
	Requires   []string // To enter this action.
	Guarantees []string
	// Derived values for template display.
	FromTransitions []Transition  // Where this action is called from events.
	FromStates      []StateAction // Where this action is called from a state.
}

func NewAction(key, name, details string, requires, guarantees []string) (action Action, err error) {

	action = Action{
		Key:        key,
		Name:       name,
		Details:    details,
		Requires:   requires,
		Guarantees: guarantees,
	}

	err = validation.ValidateStruct(&action,
		validation.Field(&action.Key, validation.Required),
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
