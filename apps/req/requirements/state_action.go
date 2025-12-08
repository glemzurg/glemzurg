package requirements

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

func createKeyActionLookup(classTransitions map[string][]Transition, statStateActions map[string][]StateAction, byCategory map[string][]Action) (lookup map[string]Action) {

	// Create clean lookup for triggers.
	transitionTriggers := map[string][]Transition{}
	for _, transitions := range classTransitions {
		for _, transition := range transitions {
			if transition.ActionKey != "" {
				triggers := transitionTriggers[transition.ActionKey]
				triggers = append(triggers, transition)
				transitionTriggers[transition.ActionKey] = triggers
			}
		}
	}

	// And for state actions.
	stateActionTriggers := map[string][]StateAction{}
	for _, stateActions := range statStateActions {
		for _, stateAction := range stateActions {
			if stateAction.ActionKey != "" {
				triggers := stateActionTriggers[stateAction.ActionKey]
				triggers = append(triggers, stateAction)
				stateActionTriggers[stateAction.ActionKey] = triggers
			}
		}
	}

	lookup = map[string]Action{}
	for _, items := range byCategory {
		for _, item := range items {

			item.SetTriggers(transitionTriggers[item.Key], stateActionTriggers[item.Key])

			lookup[item.Key] = item
		}
	}
	return lookup
}
