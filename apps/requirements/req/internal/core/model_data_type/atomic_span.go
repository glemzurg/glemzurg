package model_data_type

import (
	"fmt"
	"math"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
)

const (
	_BOUND_TYPE_LIMIT_CLOSED        = "closed"        // Include the value itself.
	_BOUND_TYPE_LIMIT_OPEN          = "open"          // Do not include the value itself.
	_BOUND_TYPE_LIMIT_UNCONSTRAINED = "unconstrained" // Undefined what this end of the span is, at least not in requirements.
)

var _validBoundTypes = map[string]bool{
	_BOUND_TYPE_LIMIT_CLOSED:        true,
	_BOUND_TYPE_LIMIT_OPEN:          true,
	_BOUND_TYPE_LIMIT_UNCONSTRAINED: true,
}

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
	Precision float64
}

func validateDenominator(ptr *int, required bool) error {
	if ptr == nil {
		if required {
			return fmt.Errorf("cannot be blank")
		}
		return nil
	}
	if *ptr < 1 {
		return fmt.Errorf("must be no less than 1")
	}
	return nil
}

func precisionValidator(v float64) error {
	if v <= 0 || v > 1 {
		return fmt.Errorf("must be greater than 0 and less than or equal to 1")
	}

	log := math.Log10(v)
	if math.Floor(log) != log {
		return fmt.Errorf("must be exactly 1.0, 0.1, 0.01, etc")
	}

	return nil
}

func (a *AtomicSpan) Validate() error {
	// LowerType: required and must be one of closed, open, unconstrained.
	if a.LowerType == "" {
		return coreerr.NewWithValues(coreerr.DtypeSpanLowertypeRequired, "LowerType is required", "LowerType", "", "one of: closed, open, unconstrained")
	}
	if !_validBoundTypes[a.LowerType] {
		return coreerr.NewWithValues(coreerr.DtypeSpanLowertypeInvalid, "LowerType is not a valid value", "LowerType", a.LowerType, "one of: closed, open, unconstrained")
	}

	// HigherType: required and must be one of closed, open, unconstrained.
	if a.HigherType == "" {
		return coreerr.NewWithValues(coreerr.DtypeSpanHighertypeRequired, "HigherType is required", "HigherType", "", "one of: closed, open, unconstrained")
	}
	if !_validBoundTypes[a.HigherType] {
		return coreerr.NewWithValues(coreerr.DtypeSpanHighertypeInvalid, "HigherType is not a valid value", "HigherType", a.HigherType, "one of: closed, open, unconstrained")
	}

	// Units: required.
	if a.Units == "" {
		return coreerr.New(coreerr.DtypeSpanUnitsRequired, "Units is required", "Units")
	}

	// Precision: required (non-zero).
	if a.Precision == 0 {
		return coreerr.New(coreerr.DtypeSpanPrecisionRequired, "Precision is required", "Precision")
	}

	// LowerValue: required when LowerType != unconstrained.
	if a.LowerType != _BOUND_TYPE_LIMIT_UNCONSTRAINED {
		if a.LowerValue == nil {
			return fmt.Errorf("LowerValue: cannot be blank")
		}
	}

	// LowerDenominator: conditional validation.
	if err := validateDenominator(a.LowerDenominator, a.LowerType != _BOUND_TYPE_LIMIT_UNCONSTRAINED); err != nil {
		return fmt.Errorf("LowerDenominator: %s", err.Error())
	}

	// HigherValue: required when HigherType != unconstrained.
	if a.HigherType != _BOUND_TYPE_LIMIT_UNCONSTRAINED {
		if a.HigherValue == nil {
			return fmt.Errorf("HigherValue: cannot be blank")
		}
	}

	// HigherDenominator: conditional validation.
	if err := validateDenominator(a.HigherDenominator, a.HigherType != _BOUND_TYPE_LIMIT_UNCONSTRAINED); err != nil {
		return fmt.Errorf("HigherDenominator: %s", err.Error())
	}

	// Precision: must be valid value.
	if err := precisionValidator(a.Precision); err != nil {
		return fmt.Errorf("precision: %s", err.Error())
	}

	return nil
}
