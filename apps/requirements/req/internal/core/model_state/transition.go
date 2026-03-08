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
		return coreerr.New(coreerr.TransitionKeyInvalid, fmt.Sprintf("Key: %s", err.Error()), "Key")
	}
	if t.Key.KeyType != identity.KEY_TYPE_TRANSITION {
		return coreerr.NewWithValues(coreerr.TransitionKeyTypeInvalid, fmt.Sprintf("Key: invalid key type '%s' for transition", t.Key.KeyType), "Key", t.Key.KeyType, identity.KEY_TYPE_TRANSITION)
	}

	// Validate the event key (required).
	if err := t.EventKey.Validate(); err != nil {
		return coreerr.New(coreerr.TransitionEventkeyInvalid, fmt.Sprintf("EventKey: %s", err.Error()), "EventKey")
	}
	if t.EventKey.KeyType != identity.KEY_TYPE_EVENT {
		return coreerr.NewWithValues(coreerr.TransitionEventkeyType, fmt.Sprintf("EventKey: invalid key type '%s' for event", t.EventKey.KeyType), "EventKey", t.EventKey.KeyType, identity.KEY_TYPE_EVENT)
	}

	// Validate optional key fields.
	if t.FromStateKey != nil {
		if err := t.FromStateKey.Validate(); err != nil {
			return coreerr.New(coreerr.TransitionFromstatekeyInvalid, fmt.Sprintf("FromStateKey: %s", err.Error()), "FromStateKey")
		}
		if t.FromStateKey.KeyType != identity.KEY_TYPE_STATE {
			return coreerr.NewWithValues(coreerr.TransitionFromstatekeyType, fmt.Sprintf("FromStateKey: invalid key type '%s' for from state", t.FromStateKey.KeyType), "FromStateKey", t.FromStateKey.KeyType, identity.KEY_TYPE_STATE)
		}
	}
	if t.ToStateKey != nil {
		if err := t.ToStateKey.Validate(); err != nil {
			return coreerr.New(coreerr.TransitionTostatekeyInvalid, fmt.Sprintf("ToStateKey: %s", err.Error()), "ToStateKey")
		}
		if t.ToStateKey.KeyType != identity.KEY_TYPE_STATE {
			return coreerr.NewWithValues(coreerr.TransitionTostatekeyType, fmt.Sprintf("ToStateKey: invalid key type '%s' for to state", t.ToStateKey.KeyType), "ToStateKey", t.ToStateKey.KeyType, identity.KEY_TYPE_STATE)
		}
	}
	if t.GuardKey != nil {
		if err := t.GuardKey.Validate(); err != nil {
			return coreerr.New(coreerr.TransitionGuardkeyInvalid, fmt.Sprintf("GuardKey: %s", err.Error()), "GuardKey")
		}
		if t.GuardKey.KeyType != identity.KEY_TYPE_GUARD {
			return coreerr.NewWithValues(coreerr.TransitionGuardkeyType, fmt.Sprintf("GuardKey: invalid key type '%s' for guard", t.GuardKey.KeyType), "GuardKey", t.GuardKey.KeyType, identity.KEY_TYPE_GUARD)
		}
	}
	if t.ActionKey != nil {
		if err := t.ActionKey.Validate(); err != nil {
			return coreerr.New(coreerr.TransitionActionkeyInvalid, fmt.Sprintf("ActionKey: %s", err.Error()), "ActionKey")
		}
		if t.ActionKey.KeyType != identity.KEY_TYPE_ACTION {
			return coreerr.NewWithValues(coreerr.TransitionActionkeyType, fmt.Sprintf("ActionKey: invalid key type '%s' for action", t.ActionKey.KeyType), "ActionKey", t.ActionKey.KeyType, identity.KEY_TYPE_ACTION)
		}
	}

	// We must have either from or to state or both.
	if t.FromStateKey == nil && t.ToStateKey == nil {
		return coreerr.New(coreerr.TransitionNoState, "FromStateKey, ToStateKey: cannot both be blank", "FromStateKey,ToStateKey")
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
			return coreerr.NewWithValues(coreerr.TransitionFromstateNotfound, fmt.Sprintf("transition '%s' references non-existent from state '%s'", t.Key.String(), t.FromStateKey.String()), "FromStateKey", t.FromStateKey.String(), "")
		}
	}
	if t.ToStateKey != nil {
		if !states[*t.ToStateKey] {
			return coreerr.NewWithValues(coreerr.TransitionTostateNotfound, fmt.Sprintf("transition '%s' references non-existent to state '%s'", t.Key.String(), t.ToStateKey.String()), "ToStateKey", t.ToStateKey.String(), "")
		}
	}
	if !events[t.EventKey] {
		return coreerr.NewWithValues(coreerr.TransitionEventNotfound, fmt.Sprintf("transition '%s' references non-existent event '%s'", t.Key.String(), t.EventKey.String()), "EventKey", t.EventKey.String(), "")
	}
	if t.GuardKey != nil {
		if !guards[*t.GuardKey] {
			return coreerr.NewWithValues(coreerr.TransitionGuardNotfound, fmt.Sprintf("transition '%s' references non-existent guard '%s'", t.Key.String(), t.GuardKey.String()), "GuardKey", t.GuardKey.String(), "")
		}
	}
	if t.ActionKey != nil {
		if !actions[*t.ActionKey] {
			return coreerr.NewWithValues(coreerr.TransitionActionNotfound, fmt.Sprintf("transition '%s' references non-existent action '%s'", t.Key.String(), t.ActionKey.String()), "ActionKey", t.ActionKey.String(), "")
		}
	}
	return nil
}
