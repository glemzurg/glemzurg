package model_state

import (
	"sort"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

// State is a particular set of values in a state, distinct from all other states in the state.
type State struct {
	Key        string
	Name       string
	Details    string // Markdown.
	UmlComment string
	// Part of the data in a parsed file.
	Actions []StateAction
}

func NewState(key, name, details, umlComment string) (state State, err error) {

	state = State{
		Key:        key,
		Name:       name,
		Details:    details,
		UmlComment: umlComment,
	}

	err = validation.ValidateStruct(&state,
		validation.Field(&state.Key, validation.Required),
		validation.Field(&state.Name, validation.Required),
	)
	if err != nil {
		return State{}, errors.WithStack(err)
	}

	return state, nil
}

func (s *State) SetActions(actions []StateAction) {

	sort.Slice(actions, func(i, j int) bool {
		return lessThanStateAction(actions[i], actions[j])
	})

	s.Actions = actions
}
