package model_state

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

// Guard is a constraint on an event in a state machine.
type Guard struct {
	Key     string
	Name    string // A simple unique name for a guard, for internal use.
	Details string // How the details of the guard are represented, what shows in the uml.
}

func NewGuard(key, name, details string) (guard Guard, err error) {

	guard = Guard{
		Key:     key,
		Name:    name,
		Details: details,
	}

	err = validation.ValidateStruct(&guard,
		validation.Field(&guard.Key, validation.Required),
		validation.Field(&guard.Name, validation.Required),
		validation.Field(&guard.Details, validation.Required),
	)
	if err != nil {
		return Guard{}, errors.WithStack(err)
	}

	return guard, nil
}

func CreateKeyGuardLookup(byCategory map[string][]Guard) (lookup map[string]Guard) {
	lookup = map[string]Guard{}
	for _, items := range byCategory {
		for _, item := range items {
			lookup[item.Key] = item
		}
	}
	return lookup
}
