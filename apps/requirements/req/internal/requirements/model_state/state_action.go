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

	err = validation.ValidateStruct(&stateAction,
		validation.Field(&stateAction.Key, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_STATE_ACTION {
				return errors.Errorf("invalid key type '%s' for state action", k.KeyType())
			}
			return nil
		})),
		validation.Field(&stateAction.ActionKey, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_ACTION {
				return errors.Errorf("invalid key type '%s' for action", k.KeyType())
			}
			return nil
		})),
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
	return a.Key.String() < b.Key.String()
}
