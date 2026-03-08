package model_state

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
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
	// Validate the key.
	if err := sa.Key.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.StateactionKeyInvalid,
			Message: fmt.Sprintf("Key: %s", err.Error()),
			Field:   "Key",
		}
	}
	if sa.Key.KeyType != identity.KEY_TYPE_STATE_ACTION {
		return &coreerr.ValidationError{
			Code:    coreerr.StateactionKeyTypeInvalid,
			Message: fmt.Sprintf("Key: invalid key type '%s' for state action", sa.Key.KeyType),
			Field:   "Key",
			Got:     sa.Key.KeyType,
			Want:    identity.KEY_TYPE_STATE_ACTION,
		}
	}

	// Validate the action key.
	if err := sa.ActionKey.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.StateactionActionkeyInvalid,
			Message: fmt.Sprintf("ActionKey: %s", err.Error()),
			Field:   "ActionKey",
		}
	}
	if sa.ActionKey.KeyType != identity.KEY_TYPE_ACTION {
		return &coreerr.ValidationError{
			Code:    coreerr.StateactionActionkeyType,
			Message: fmt.Sprintf("ActionKey: invalid key type '%s' for action", sa.ActionKey.KeyType),
			Field:   "ActionKey",
			Got:     sa.ActionKey.KeyType,
			Want:    identity.KEY_TYPE_ACTION,
		}
	}

	// Validate When field.
	if sa.When == "" {
		return &coreerr.ValidationError{
			Code:    coreerr.StateactionWhenRequired,
			Message: "When is required",
			Field:   "When",
		}
	}
	if sa.When != _WHEN_ENTRY && sa.When != _WHEN_EXIT && sa.When != _WHEN_DO {
		return &coreerr.ValidationError{
			Code:    coreerr.StateactionWhenInvalid,
			Message: fmt.Sprintf("When '%s' is not valid", sa.When),
			Field:   "When",
			Got:     sa.When,
			Want:    "one of: entry, exit, do",
		}
	}

	return nil
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

// ValidateReferences validates that the state action's ActionKey references a real action in the class.
func (sa *StateAction) ValidateReferences(actions map[identity.Key]bool) error {
	if !actions[sa.ActionKey] {
		return &coreerr.ValidationError{
			Code:    coreerr.StateactionActionNotfound,
			Message: fmt.Sprintf("state action '%s' references non-existent action '%s'", sa.Key.String(), sa.ActionKey.String()),
			Field:   "ActionKey",
			Got:     sa.ActionKey.String(),
		}
	}
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
