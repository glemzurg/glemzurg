package model_state

import (
	"sort"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

	"github.com/glemzurg/go-tlaplus/internal/identity"
)

// State is a particular set of values in a state, distinct from all other states in the state.
type State struct {
	Key        identity.Key
	Name       string
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
	return validation.ValidateStruct(s,
		validation.Field(&s.Key, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_STATE {
				return errors.Errorf("invalid key type '%s' for state", k.KeyType())
			}
			return nil
		})),
		validation.Field(&s.Name, validation.Required),
	)
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
