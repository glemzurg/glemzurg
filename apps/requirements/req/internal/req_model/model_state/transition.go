package model_state

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
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
	err := validation.ValidateStruct(t,
		validation.Field(&t.Key, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_TRANSITION {
				return errors.Errorf("invalid key type '%s' for transition", k.KeyType())
			}
			return nil
		})),
		validation.Field(&t.EventKey, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_EVENT {
				return errors.Errorf("invalid key type '%s' for event", k.KeyType())
			}
			return nil
		})),
		validation.Field(&t.FromStateKey, validation.By(func(value interface{}) error {
			k := value.(*identity.Key)
			if k == nil {
				return nil
			}
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_STATE {
				return errors.Errorf("invalid key type '%s' for from state", k.KeyType())
			}
			return nil
		})),
		validation.Field(&t.ToStateKey, validation.By(func(value interface{}) error {
			k := value.(*identity.Key)
			if k == nil {
				return nil
			}
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_STATE {
				return errors.Errorf("invalid key type '%s' for to state", k.KeyType())
			}
			return nil
		})),
		validation.Field(&t.GuardKey, validation.By(func(value interface{}) error {
			k := value.(*identity.Key)
			if k == nil {
				return nil
			}
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_GUARD {
				return errors.Errorf("invalid key type '%s' for guard", k.KeyType())
			}
			return nil
		})),
		validation.Field(&t.ActionKey, validation.By(func(value interface{}) error {
			k := value.(*identity.Key)
			if k == nil {
				return nil
			}
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_ACTION {
				return errors.Errorf("invalid key type '%s' for action", k.KeyType())
			}
			return nil
		})),
	)
	if err != nil {
		return err
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
