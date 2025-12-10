package data_type

import (
	"errors"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	_BOUND_TYPE_LIMIT_CLOSED        = "closed"        // Include the value itself.
	_BOUND_TYPE_LIMIT_OPEN          = "open"          // Do not in clude the value itself.
	_BOUND_TYPE_LIMIT_UNCONSTRAINED = "unconstrained" // Undefined what this end of the span is, at least not in requirements.
)

type AtomicSpan struct {
	// Lower bound.
	LowerType        string
	LowerValue       *int
	LowerDenominator *int // If a fraction.
	// Higher bound.
	HigherType        string
	HigherValue       *int
	HigherDenominator *int // If a fraction.
	// What are these values?
	Units string
	// What precision should we support of these values?
	Precision uint
}

func validateDenominator(value interface{}, required bool) error {
	ptr, ok := value.(*int)
	if !ok {
		return errors.New("is not int")
	}
	if ptr == nil {
		if required {
			return errors.New("cannot be blank")
		}
		return nil
	}
	if *ptr < 1 {
		return errors.New("must be no less than 1")
	}
	return nil
}

func (a *AtomicSpan) Validate() error {
	return validation.ValidateStruct(a,
		validation.Field(&a.LowerType, validation.Required, validation.In(_BOUND_TYPE_LIMIT_CLOSED, _BOUND_TYPE_LIMIT_OPEN, _BOUND_TYPE_LIMIT_UNCONSTRAINED)),
		validation.Field(&a.HigherType, validation.Required, validation.In(_BOUND_TYPE_LIMIT_CLOSED, _BOUND_TYPE_LIMIT_OPEN, _BOUND_TYPE_LIMIT_UNCONSTRAINED)),
		validation.Field(&a.Units, validation.Required),
		validation.Field(&a.LowerValue, validation.Required.When(a.LowerType != _BOUND_TYPE_LIMIT_UNCONSTRAINED)),
		validation.Field(&a.LowerDenominator, validation.By(func(value interface{}) error {
			return validateDenominator(value, a.LowerType != _BOUND_TYPE_LIMIT_UNCONSTRAINED)
		})),
		validation.Field(&a.HigherValue, validation.Required.When(a.HigherType != _BOUND_TYPE_LIMIT_UNCONSTRAINED)),
		validation.Field(&a.HigherDenominator, validation.By(func(value interface{}) error {
			return validateDenominator(value, a.HigherType != _BOUND_TYPE_LIMIT_UNCONSTRAINED)
		})),
	)
}
