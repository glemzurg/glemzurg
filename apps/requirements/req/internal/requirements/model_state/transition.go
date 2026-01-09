package model_state

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

// Transition is a move between two states.
type Transition struct {
	Key          string
	FromStateKey string
	EventKey     string
	GuardKey     string
	ActionKey    string
	ToStateKey   string
	UmlComment   string
}

func NewTransition(key, fromStateKey, eventKey, guardKey, actionKey, toStateKey, umlComment string) (transition Transition, err error) {

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
		validation.Field(&transition.Key, validation.Required),
		validation.Field(&transition.EventKey, validation.Required),
	)
	if err != nil {
		return Transition{}, errors.WithStack(err)
	}

	// We must have either from or two state or both.
	if transition.FromStateKey == "" && transition.ToStateKey == "" {
		return Transition{}, errors.WithStack(errors.Errorf(`FromStateKey, ToStateKey: cannot both be blank`))
	}

	return transition, nil
}

func CreateKeyTransitionLookup(byCategory map[string][]Transition) (lookup map[string]Transition) {
	lookup = map[string]Transition{}
	for _, items := range byCategory {
		for _, item := range items {
			lookup[item.Key] = item
		}
	}
	return lookup
}
