package model_state

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

const (
	_WHEN_ENTRY = "entry" // An action triggered on entering a state.
	_WHEN_EXIT  = "exit"  // An action triggered on exiting a state.
	_WHEN_DO    = "do"    // An action that runs persistenly in a state.
)

var _whenSortValue = map[string]int{
	_WHEN_ENTRY: 1, // Sorts first.
	_WHEN_DO:    2,
	_WHEN_EXIT:  3,
}

// StateAction is a action that triggers when a state is entered or exited or happens perpetually.
type StateAction struct {
	Key       string
	ActionKey string
	When      string
	// Derived data for templates.
	StateKey string
}

func NewStateAction(key, actionKey string, when string) (stateAction StateAction, err error) {

	stateAction = StateAction{
		Key:       key,
		ActionKey: actionKey,
		When:      when,
	}

	err = validation.ValidateStruct(&stateAction,
		validation.Field(&stateAction.Key, validation.Required),
		validation.Field(&stateAction.ActionKey, validation.Required),
		validation.Field(&stateAction.When, validation.Required, validation.In(_WHEN_ENTRY, _WHEN_EXIT, _WHEN_DO)),
	)
	if err != nil {
		return StateAction{}, errors.WithStack(err)
	}

	return stateAction, nil
}

func lessThanStateAction(a, b StateAction) (less bool) {

	// Sort by when first.
	if _whenSortValue[a.When] != _whenSortValue[b.When] {
		return _whenSortValue[a.When] < _whenSortValue[b.When]
	}

	// Sort by key next.
	return a.Key < b.Key
}

func CreateKeyStateActionLookup(byCategory map[string][]StateAction) (lookup map[string]StateAction) {
	lookup = map[string]StateAction{}
	for _, items := range byCategory {
		for _, item := range items {
			lookup[item.Key] = item
		}
	}
	return lookup
}
