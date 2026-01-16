package model_state

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
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
	Key       identity.Key
	ActionKey identity.Key
	When      string
	// Derived data for templates.
	StateKey identity.Key
}

func NewStateAction(key, actionKey identity.Key, when string) (stateAction StateAction, err error) {

	stateAction = StateAction{
		Key:       key,
		ActionKey: actionKey,
		When:      when,
	}

	if err = stateAction.Validate(); err != nil {
		return StateAction{}, err
	}

	return stateAction, nil
}

// Validate validates the StateAction struct.
func (sa *StateAction) Validate() error {
	return validation.ValidateStruct(sa,
		validation.Field(&sa.Key, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_STATE_ACTION {
				return errors.Errorf("invalid key type '%s' for state action", k.KeyType())
			}
			return nil
		})),
		validation.Field(&sa.ActionKey, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_ACTION {
				return errors.Errorf("invalid key type '%s' for action", k.KeyType())
			}
			return nil
		})),
		validation.Field(&sa.When, validation.Required, validation.In(_WHEN_ENTRY, _WHEN_EXIT, _WHEN_DO)),
	)
}

// ValidateWithParent validates the StateAction, its key's parent relationship, and all children.
// The parent must be a State.
func (sa *StateAction) ValidateWithParent(parent *identity.Key) error {
	// Validate the object itself.
	if err := sa.Validate(); err != nil {
		return err
	}
	// Validate the key has the correct parent.
	if err := sa.Key.ValidateParent(parent); err != nil {
		return err
	}
	// StateAction has no children with keys that need validation.
	return nil
}

func lessThanStateAction(a, b StateAction) (less bool) {

	// Sort by when first.
	if _whenSortValue[a.When] != _whenSortValue[b.When] {
		return _whenSortValue[a.When] < _whenSortValue[b.When]
	}

	// Sort by key next.
	return a.Key.String() < b.Key.String()
}
