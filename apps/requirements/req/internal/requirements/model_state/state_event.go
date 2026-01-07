package model_state

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

// Event is what triggers a transition between states.
type Event struct {
	Key        string
	Name       string
	Details    string
	Parameters []EventParameter
}

func NewEvent(key, name, details string, parameters []EventParameter) (event Event, err error) {

	event = Event{
		Key:        key,
		Name:       name,
		Details:    details,
		Parameters: parameters,
	}

	err = validation.ValidateStruct(&event,
		validation.Field(&event.Key, validation.Required),
		validation.Field(&event.Name, validation.Required),
	)
	if err != nil {
		return Event{}, errors.WithStack(err)
	}

	return event, nil
}

func createKeyEventLookup(byCategory map[string][]Event) (lookup map[string]Event) {
	lookup = map[string]Event{}
	for _, items := range byCategory {
		for _, item := range items {
			lookup[item.Key] = item
		}
	}
	return lookup
}
