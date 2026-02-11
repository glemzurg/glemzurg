package model_state

import (
	"sort"

	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// State is a particular set of values in a state, distinct from all other states in the state.
type State struct {
	Key        identity.Key
	Name       string `validate:"required"`
	Details    string // Markdown.
	UmlComment string
	// Children
	Actions []StateAction
}

func NewState(key identity.Key, name, details, umlComment string) (state State, err error) {

	state = State{
		Key:        key,
		Name:       name,
		Details:    details,
		UmlComment: umlComment,
	}

	if err = state.Validate(); err != nil {
		return State{}, err
	}

	return state, nil
}

// Validate validates the State struct.
func (s *State) Validate() error {
	// Validate the key.
	if err := s.Key.Validate(); err != nil {
		return err
	}
	if s.Key.KeyType() != identity.KEY_TYPE_STATE {
		return errors.Errorf("Key: invalid key type '%s' for state", s.Key.KeyType())
	}

	// Validate struct tags (Name required).
	if err := _validate.Struct(s); err != nil {
		return err
	}

	return nil
}

func (s *State) SetActions(actions []StateAction) {

	sort.Slice(actions, func(i, j int) bool {
		return lessThanStateAction(actions[i], actions[j])
	})

	s.Actions = actions
}

// ValidateWithParent validates the State, its key's parent relationship, and all children.
// The parent must be a Class.
func (s *State) ValidateWithParent(parent *identity.Key) error {
	return s.ValidateWithParentAndActions(parent, nil)
}

// ValidateWithParentAndActions validates the State with access to actions for cross-reference validation.
// The parent must be a Class.
// The actions map is used to validate that StateAction ActionKey references exist.
func (s *State) ValidateWithParentAndActions(parent *identity.Key, actions map[identity.Key]bool) error {
	// Validate the object itself.
	if err := s.Validate(); err != nil {
		return err
	}
	// Validate the key has the correct parent.
	if err := s.Key.ValidateParent(parent); err != nil {
		return err
	}
	// Validate all children.
	for i := range s.Actions {
		if err := s.Actions[i].ValidateWithParent(&s.Key); err != nil {
			return err
		}
		if err := s.Actions[i].ValidateReferences(actions); err != nil {
			return err
		}
	}
	return nil
}
