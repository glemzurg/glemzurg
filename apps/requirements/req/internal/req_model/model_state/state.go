package model_state

import (
	"sort"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// State is a particular set of values in a state, distinct from all other states in the state.
type State struct {
	Key        identity.Key
	Name       string
	Details    string // Markdown.
	UmlComment string
	// Part of the data in a parsed file.
	Actions []StateAction
}

func NewState(key identity.Key, name, details, umlComment string) (state State, err error) {

	state = State{
		Key:        key,
		Name:       name,
		Details:    details,
		UmlComment: umlComment,
	}

	err = validation.ValidateStruct(&state,
		validation.Field(&state.Key, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_STATE {
				return errors.Errorf("invalid key type '%s' for state", k.KeyType())
			}
			return nil
		})),
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

// ValidateWithParent validates the State and verifies its key has the correct parent.
// The parent must be a Class.
func (s *State) ValidateWithParent(parent *identity.Key) error {
	// Validate the key has the correct parent.
	if err := s.Key.ValidateParent(parent); err != nil {
		return err
	}
	// Validate all children.
	for i := range s.Actions {
		if err := s.Actions[i].ValidateWithParent(&s.Key); err != nil {
			return err
		}
	}
	return nil
}
