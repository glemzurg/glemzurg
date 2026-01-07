package model_state

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
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

	err = validation.ValidateStruct(&param,
		validation.Field(&param.Name, validation.Required),
		validation.Field(&param.Source, validation.Required),
	)
	if err != nil {
		return EventParameter{}, errors.WithStack(err)
	}

	return param, nil
}
