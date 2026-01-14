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

	err = validation.ValidateStruct(&transition,
		validation.Field(&transition.Key, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_TRANSITION {
				return errors.Errorf("invalid key type '%s' for transition", k.KeyType())
			}
			return nil
		})),
		validation.Field(&transition.EventKey, validation.Required, validation.By(func(value interface{}) error {
			k := value.(identity.Key)
			if err := k.Validate(); err != nil {
				return err
			}
			if k.KeyType() != identity.KEY_TYPE_EVENT {
				return errors.Errorf("invalid key type '%s' for event", k.KeyType())
			}
			return nil
		})),
		validation.Field(&transition.FromStateKey, validation.By(func(value interface{}) error {
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
		validation.Field(&transition.ToStateKey, validation.By(func(value interface{}) error {
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
		validation.Field(&transition.GuardKey, validation.By(func(value interface{}) error {
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
		validation.Field(&transition.ActionKey, validation.By(func(value interface{}) error {
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
		return Transition{}, errors.WithStack(err)
	}

	// We must have either from or to state or both.
	if transition.FromStateKey == nil && transition.ToStateKey == nil {
		return Transition{}, errors.WithStack(errors.Errorf(`FromStateKey, ToStateKey: cannot both be blank`))
	}

	return transition, nil
}
