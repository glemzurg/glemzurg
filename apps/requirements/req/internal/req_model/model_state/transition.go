package model_state

import (
	"github.com/pkg/errors"

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
func (t *Transition) Validate() error {
	// Validate the key.
	if err := t.Key.Validate(); err != nil {
		return err
	}
	if t.Key.KeyType() != identity.KEY_TYPE_TRANSITION {
		return errors.Errorf("Key: invalid key type '%s' for transition", t.Key.KeyType())
	}

	// Validate the event key (required).
	if err := t.EventKey.Validate(); err != nil {
		return errors.Wrap(err, "EventKey")
	}
	if t.EventKey.KeyType() != identity.KEY_TYPE_EVENT {
		return errors.Errorf("EventKey: invalid key type '%s' for event", t.EventKey.KeyType())
	}

	// Validate optional key fields.
	if t.FromStateKey != nil {
		if err := t.FromStateKey.Validate(); err != nil {
			return errors.Wrap(err, "FromStateKey")
		}
		if t.FromStateKey.KeyType() != identity.KEY_TYPE_STATE {
			return errors.Errorf("FromStateKey: invalid key type '%s' for from state", t.FromStateKey.KeyType())
		}
	}
	if t.ToStateKey != nil {
		if err := t.ToStateKey.Validate(); err != nil {
			return errors.Wrap(err, "ToStateKey")
		}
		if t.ToStateKey.KeyType() != identity.KEY_TYPE_STATE {
			return errors.Errorf("ToStateKey: invalid key type '%s' for to state", t.ToStateKey.KeyType())
		}
	}
	if t.GuardKey != nil {
		if err := t.GuardKey.Validate(); err != nil {
			return errors.Wrap(err, "GuardKey")
		}
		if t.GuardKey.KeyType() != identity.KEY_TYPE_GUARD {
			return errors.Errorf("GuardKey: invalid key type '%s' for guard", t.GuardKey.KeyType())
		}
	}
	if t.ActionKey != nil {
		if err := t.ActionKey.Validate(); err != nil {
			return errors.Wrap(err, "ActionKey")
		}
		if t.ActionKey.KeyType() != identity.KEY_TYPE_ACTION {
			return errors.Errorf("ActionKey: invalid key type '%s' for action", t.ActionKey.KeyType())
		}
	}

	// We must have either from or to state or both.
	if t.FromStateKey == nil && t.ToStateKey == nil {
		return errors.Errorf(`FromStateKey, ToStateKey: cannot both be blank`)
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
// - ActionKey must exist in the actions map (if not nil)
func (t *Transition) ValidateReferences(states, events, guards, actions map[identity.Key]bool) error {
	if t.FromStateKey != nil {
		if !states[*t.FromStateKey] {
			return errors.Errorf("transition '%s' references non-existent from state '%s'", t.Key.String(), t.FromStateKey.String())
		}
	}
	if t.ToStateKey != nil {
		if !states[*t.ToStateKey] {
			return errors.Errorf("transition '%s' references non-existent to state '%s'", t.Key.String(), t.ToStateKey.String())
		}
	}
	if !events[t.EventKey] {
		return errors.Errorf("transition '%s' references non-existent event '%s'", t.Key.String(), t.EventKey.String())
	}
	if t.GuardKey != nil {
		if !guards[*t.GuardKey] {
			return errors.Errorf("transition '%s' references non-existent guard '%s'", t.Key.String(), t.GuardKey.String())
		}
	}
	if t.ActionKey != nil {
		if !actions[*t.ActionKey] {
			return errors.Errorf("transition '%s' references non-existent action '%s'", t.Key.String(), t.ActionKey.String())
		}
	}
	return nil
}
