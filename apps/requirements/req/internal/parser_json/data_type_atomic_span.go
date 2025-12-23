package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/data_type"

// atomicSpanInOut represents a range of allowed values.
type atomicSpanInOut struct {
	// Lower bound.
	LowerType        string `json:"lower_type"`
	LowerValue       *int   `json:"lower_value"`
	LowerDenominator *int   `json:"lower_denominator"` // If a fraction.
	// Higher bound.
	HigherType        string `json:"higher_type"`
	HigherValue       *int   `json:"higher_value"`
	HigherDenominator *int   `json:"higher_denominator"` // If a fraction.
	// What are these values?
	Units string `json:"units"`
	// What precision should we support of these values?
	Precision float64 `json:"precision"`
}

// ToRequirements converts the atomicSpanInOut to data_type.AtomicSpan.
func (a atomicSpanInOut) ToRequirements() data_type.AtomicSpan {
	return data_type.AtomicSpan{
		LowerType:        a.LowerType,
		LowerValue:       a.LowerValue,
		LowerDenominator: a.LowerDenominator,
		HigherType:        a.HigherType,
		HigherValue:       a.HigherValue,
		HigherDenominator: a.HigherDenominator,
		Units:             a.Units,
		Precision:         a.Precision,
	}
}

// FromRequirements creates a atomicSpanInOut from data_type.AtomicSpan.
func FromRequirementsAtomicSpan(a data_type.AtomicSpan) atomicSpanInOut {
	return atomicSpanInOut{
		LowerType:        a.LowerType,
		LowerValue:       a.LowerValue,
		LowerDenominator: a.LowerDenominator,
		HigherType:        a.HigherType,
		HigherValue:       a.HigherValue,
		HigherDenominator: a.HigherDenominator,
		Units:             a.Units,
		Precision:         a.Precision,
	}
}
