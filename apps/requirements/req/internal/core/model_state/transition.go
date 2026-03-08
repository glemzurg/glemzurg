package model_state

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// Transition is a move between two states.
type Transition struct {
	Key          identity.Key
	FromStateKey *identity.Key
	EventKey     identity.Key
	GuardKey     *identity.Key
	ActionKey    *identity.Key
	ToStateKey   *identity.Key
	UmlComment   string
}

func NewTransition(key identity.Key, fromStateKey *identity.Key, eventKey identity.Key, guardKey, actionKey, toStateKey *identity.Key, umlComment string) (transition Transition, err error) {
	transition = Transition{
		Key:          key,
		FromStateKey: fromStateKey,
		EventKey:     eventKey,
		GuardKey:     guardKey,
		ActionKey:    actionKey,
		ToStateKey:   toStateKey,
		UmlComment:   umlComment,
	}

	if err = transition.Validate(); err != nil {
		return Transition{}, err
	}

	return transition, nil
}

// Validate validates the Transition struct.
//
//complexity:cyclo:warn=20,fail=20 Sequential field validation.
func (t *Transition) Validate() error {
	// Validate the key.
	if err := t.Key.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.TransitionKeyInvalid,
			Message: fmt.Sprintf("Key: %s", err.Error()),
			Field:   "Key",
		}
	}
	if t.Key.KeyType != identity.KEY_TYPE_TRANSITION {
		return &coreerr.ValidationError{
			Code:    coreerr.TransitionKeyTypeInvalid,
			Message: fmt.Sprintf("Key: invalid key type '%s' for transition", t.Key.KeyType),
			Field:   "Key",
			Got:     t.Key.KeyType,
			Want:    identity.KEY_TYPE_TRANSITION,
		}
	}

	// Validate the event key (required).
	if err := t.EventKey.Validate(); err != nil {
		return &coreerr.ValidationError{
			Code:    coreerr.TransitionEventkeyInvalid,
			Message: fmt.Sprintf("EventKey: %s", err.Error()),
			Field:   "EventKey",
		}
	}
	if t.EventKey.KeyType != identity.KEY_TYPE_EVENT {
		return &coreerr.ValidationError{
			Code:    coreerr.TransitionEventkeyType,
			Message: fmt.Sprintf("EventKey: invalid key type '%s' for event", t.EventKey.KeyType),
			Field:   "EventKey",
			Got:     t.EventKey.KeyType,
			Want:    identity.KEY_TYPE_EVENT,
		}
	}

	// Validate optional key fields.
	if t.FromStateKey != nil {
		if err := t.FromStateKey.Validate(); err != nil {
			return &coreerr.ValidationError{
				Code:    coreerr.TransitionFromstatekeyInvalid,
				Message: fmt.Sprintf("FromStateKey: %s", err.Error()),
				Field:   "FromStateKey",
			}
		}
		if t.FromStateKey.KeyType != identity.KEY_TYPE_STATE {
			return &coreerr.ValidationError{
				Code:    coreerr.TransitionFromstatekeyType,
				Message: fmt.Sprintf("FromStateKey: invalid key type '%s' for from state", t.FromStateKey.KeyType),
				Field:   "FromStateKey",
				Got:     t.FromStateKey.KeyType,
				Want:    identity.KEY_TYPE_STATE,
			}
		}
	}
	if t.ToStateKey != nil {
		if err := t.ToStateKey.Validate(); err != nil {
			return &coreerr.ValidationError{
				Code:    coreerr.TransitionTostatekeyInvalid,
				Message: fmt.Sprintf("ToStateKey: %s", err.Error()),
				Field:   "ToStateKey",
			}
		}
		if t.ToStateKey.KeyType != identity.KEY_TYPE_STATE {
			return &coreerr.ValidationError{
				Code:    coreerr.TransitionTostatekeyType,
				Message: fmt.Sprintf("ToStateKey: invalid key type '%s' for to state", t.ToStateKey.KeyType),
				Field:   "ToStateKey",
				Got:     t.ToStateKey.KeyType,
				Want:    identity.KEY_TYPE_STATE,
			}
		}
	}
	if t.GuardKey != nil {
		if err := t.GuardKey.Validate(); err != nil {
			return &coreerr.ValidationError{
				Code:    coreerr.TransitionGuardkeyInvalid,
				Message: fmt.Sprintf("GuardKey: %s", err.Error()),
				Field:   "GuardKey",
			}
		}
		if t.GuardKey.KeyType != identity.KEY_TYPE_GUARD {
			return &coreerr.ValidationError{
				Code:    coreerr.TransitionGuardkeyType,
				Message: fmt.Sprintf("GuardKey: invalid key type '%s' for guard", t.GuardKey.KeyType),
				Field:   "GuardKey",
				Got:     t.GuardKey.KeyType,
				Want:    identity.KEY_TYPE_GUARD,
			}
		}
	}
	if t.ActionKey != nil {
		if err := t.ActionKey.Validate(); err != nil {
			return &coreerr.ValidationError{
				Code:    coreerr.TransitionActionkeyInvalid,
				Message: fmt.Sprintf("ActionKey: %s", err.Error()),
				Field:   "ActionKey",
			}
		}
		if t.ActionKey.KeyType != identity.KEY_TYPE_ACTION {
			return &coreerr.ValidationError{
				Code:    coreerr.TransitionActionkeyType,
				Message: fmt.Sprintf("ActionKey: invalid key type '%s' for action", t.ActionKey.KeyType),
				Field:   "ActionKey",
				Got:     t.ActionKey.KeyType,
				Want:    identity.KEY_TYPE_ACTION,
			}
		}
	}

	// We must have either from or to state or both.
	if t.FromStateKey == nil && t.ToStateKey == nil {
		return &coreerr.ValidationError{
			Code:    coreerr.TransitionNoState,
			Message: "FromStateKey, ToStateKey: cannot both be blank",
			Field:   "FromStateKey,ToStateKey",
		}
	}

	return nil
}

// ValidateWithParent validates the Transition, its key's parent relationship, and all children.
// The parent must be a Class.
func (t *Transition) ValidateWithParent(parent *identity.Key) error {
	// Validate the object itself.
	if err := t.Validate(); err != nil {
		return err
	}
	// Validate the key has the correct parent.
	if err := t.Key.ValidateParent(parent); err != nil {
		return err
	}
	// Transition has no children with keys that need validation.
	return nil
}

// ValidateReferences validates that the transition's reference keys point to valid entities in the class.
// - FromStateKey must exist in the states map (if not nil)
// - ToStateKey must exist in the states map (if not nil)
// - EventKey must exist in the events map
// - GuardKey must exist in the guards map (if not nil)
// - ActionKey must exist in the actions map (if not nil).
func (t *Transition) ValidateReferences(states, events, guards, actions map[identity.Key]bool) error {
	if t.FromStateKey != nil {
		if !states[*t.FromStateKey] {
			return &coreerr.ValidationError{
				Code:    coreerr.TransitionFromstateNotfound,
				Message: fmt.Sprintf("transition '%s' references non-existent from state '%s'", t.Key.String(), t.FromStateKey.String()),
				Field:   "FromStateKey",
				Got:     t.FromStateKey.String(),
			}
		}
	}
	if t.ToStateKey != nil {
		if !states[*t.ToStateKey] {
			return &coreerr.ValidationError{
				Code:    coreerr.TransitionTostateNotfound,
				Message: fmt.Sprintf("transition '%s' references non-existent to state '%s'", t.Key.String(), t.ToStateKey.String()),
				Field:   "ToStateKey",
				Got:     t.ToStateKey.String(),
			}
		}
	}
	if !events[t.EventKey] {
		return &coreerr.ValidationError{
			Code:    coreerr.TransitionEventNotfound,
			Message: fmt.Sprintf("transition '%s' references non-existent event '%s'", t.Key.String(), t.EventKey.String()),
			Field:   "EventKey",
			Got:     t.EventKey.String(),
		}
	}
	if t.GuardKey != nil {
		if !guards[*t.GuardKey] {
			return &coreerr.ValidationError{
				Code:    coreerr.TransitionGuardNotfound,
				Message: fmt.Sprintf("transition '%s' references non-existent guard '%s'", t.Key.String(), t.GuardKey.String()),
				Field:   "GuardKey",
				Got:     t.GuardKey.String(),
			}
		}
	}
	if t.ActionKey != nil {
		if !actions[*t.ActionKey] {
			return &coreerr.ValidationError{
				Code:    coreerr.TransitionActionNotfound,
				Message: fmt.Sprintf("transition '%s' references non-existent action '%s'", t.Key.String(), t.ActionKey.String()),
				Field:   "ActionKey",
				Got:     t.ActionKey.String(),
			}
		}
	}
	return nil
}
