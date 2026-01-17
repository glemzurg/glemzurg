package model_state

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// A parameter for events.
type EventParameter struct {
	Name   string
	Source string // Where the values for this parameter are coming from.
}

func NewEventParameter(name, source string) (param EventParameter, err error) {

	param = EventParameter{
		Name:   name,
		Source: source,
	}

	if err = param.Validate(); err != nil {
		return EventParameter{}, err
	}

	return param, nil
}

// Validate validates the EventParameter struct.
func (ep *EventParameter) Validate() error {
	return validation.ValidateStruct(ep,
		validation.Field(&ep.Name, validation.Required),
		validation.Field(&ep.Source, validation.Required),
	)
}

// ValidateWithParent validates the EventParameter.
// EventParameter has no key, so parent validation is not applicable.
func (ep *EventParameter) ValidateWithParent() error {
	return ep.Validate()
}
