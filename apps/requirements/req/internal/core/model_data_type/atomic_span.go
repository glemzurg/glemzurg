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

	// Validate lower bound fields.
	if err := a.validateLowerBound(); err != nil {
		return err
	}

	// Validate higher bound fields.
	if err := a.validateHigherBound(); err != nil {
		return err
	}

	// Precision: must be valid value.
	if err := a.validatePrecision(); err != nil {
		return err
	}

	return nil
}

// validateLowerBound validates LowerValue and LowerDenominator based on LowerType.
func (a *AtomicSpan) validateLowerBound() error {
	// LowerValue: required when LowerType != unconstrained.
	if a.LowerType != _BOUND_TYPE_LIMIT_UNCONSTRAINED {
		if a.LowerValue == nil {
			return coreerr.New(coreerr.DtypeSpanLowervalRequired, "lower value is required for constrained lower bound", "LowerValue")
		}
	}

	// LowerDenominator: conditional validation.
	if a.LowerDenominator == nil {
		if a.LowerType != _BOUND_TYPE_LIMIT_UNCONSTRAINED {
			return coreerr.New(coreerr.DtypeSpanLowerdenomRequired, "lower denominator is required for constrained lower bound", "LowerDenominator")
		}
	} else if *a.LowerDenominator < 1 {
		return coreerr.NewWithValues(coreerr.DtypeSpanLowerdenomInvalid, "lower denominator must be at least 1", "LowerDenominator", fmt.Sprintf("%d", *a.LowerDenominator), "at least 1")
	}

	return nil
}

// validateHigherBound validates HigherValue and HigherDenominator based on HigherType.
func (a *AtomicSpan) validateHigherBound() error {
	// HigherValue: required when HigherType != unconstrained.
	if a.HigherType != _BOUND_TYPE_LIMIT_UNCONSTRAINED {
		if a.HigherValue == nil {
			return coreerr.New(coreerr.DtypeSpanHighervalRequired, "higher value is required for constrained higher bound", "HigherValue")
		}
	}

	// HigherDenominator: conditional validation.
	if a.HigherDenominator == nil {
		if a.HigherType != _BOUND_TYPE_LIMIT_UNCONSTRAINED {
			return coreerr.New(coreerr.DtypeSpanHigherdenomRequired, "higher denominator is required for constrained higher bound", "HigherDenominator")
		}
	} else if *a.HigherDenominator < 1 {
		return coreerr.NewWithValues(coreerr.DtypeSpanHigherdenomInvalid, "higher denominator must be at least 1", "HigherDenominator", fmt.Sprintf("%d", *a.HigherDenominator), "at least 1")
	}

	return nil
}

// validatePrecision validates precision is a valid power of 10 between 0 and 1.
func (a *AtomicSpan) validatePrecision() error {
	if a.Precision <= 0 || a.Precision > 1 {
		return coreerr.NewWithValues(coreerr.DtypeSpanPrecisionInvalid, "precision must be greater than 0 and at most 1", "Precision", fmt.Sprintf("%g", a.Precision), "0 < precision <= 1")
	}
	log := math.Log10(a.Precision)
	if math.Floor(log) != log {
		return coreerr.NewWithValues(coreerr.DtypeSpanPrecisionNotPow10, "precision must be a power of 10 (1, 0.1, 0.01, etc.)", "Precision", fmt.Sprintf("%g", a.Precision), "power of 10")
	}
	return nil
}
