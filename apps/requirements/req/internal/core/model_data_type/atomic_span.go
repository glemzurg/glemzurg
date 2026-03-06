package model_data_type

import (
	"fmt"
	"math"
)

const (
	_BOUND_TYPE_LIMIT_CLOSED        = "closed"        // Include the value itself.
	_BOUND_TYPE_LIMIT_OPEN          = "open"          // Do not include the value itself.
	_BOUND_TYPE_LIMIT_UNCONSTRAINED = "unconstrained" // Undefined what this end of the span is, at least not in requirements.
)

type AtomicSpan struct {
	// Lower bound.
	LowerType        string `validate:"required,oneof=closed open unconstrained"`
	LowerValue       *int
	LowerDenominator *int // If a fraction.
	// Higher bound.
	HigherType        string `validate:"required,oneof=closed open unconstrained"`
	HigherValue       *int
	HigherDenominator *int // If a fraction.
	// What are these values?
	Units string `validate:"required"`
	// What precision should we support of these values?
	Precision float64 `validate:"required"`
}

func validateDenominator(ptr *int, required bool) error {
	if ptr == nil {
		if required {
			return fmt.Errorf("cannot be blank.")
		}
		return nil
	}
	if *ptr < 1 {
		return fmt.Errorf("must be no less than 1.")
	}
	return nil
}

func precisionValidator(v float64) error {
	if v <= 0 || v > 1 {
		return fmt.Errorf("must be greater than 0 and less than or equal to 1.")
	}

	log := math.Log10(v)
	if math.Floor(log) != log {
		return fmt.Errorf("must be exactly 1.0, 0.1, 0.01, etc.")
	}

	return nil
}

func (a *AtomicSpan) Validate() error {
	// Validate struct tags (LowerType, HigherType, Units, Precision required + oneof).
	if err := _validate.Struct(a); err != nil {
		return err
	}

	// LowerValue: required when LowerType != unconstrained.
	if a.LowerType != _BOUND_TYPE_LIMIT_UNCONSTRAINED {
		if a.LowerValue == nil {
			return fmt.Errorf("LowerValue: cannot be blank.")
		}
	}

	// LowerDenominator: conditional validation.
	if err := validateDenominator(a.LowerDenominator, a.LowerType != _BOUND_TYPE_LIMIT_UNCONSTRAINED); err != nil {
		return fmt.Errorf("LowerDenominator: %s", err.Error())
	}

	// HigherValue: required when HigherType != unconstrained.
	if a.HigherType != _BOUND_TYPE_LIMIT_UNCONSTRAINED {
		if a.HigherValue == nil {
			return fmt.Errorf("HigherValue: cannot be blank.")
		}
	}

	// HigherDenominator: conditional validation.
	if err := validateDenominator(a.HigherDenominator, a.HigherType != _BOUND_TYPE_LIMIT_UNCONSTRAINED); err != nil {
		return fmt.Errorf("HigherDenominator: %s", err.Error())
	}

	// Precision: must be valid value.
	if err := precisionValidator(a.Precision); err != nil {
		return fmt.Errorf("Precision: %s", err.Error())
	}

	return nil
}
