package parser_json

// atomicSpanInOut represents a range of allowed values.
type atomicSpanInOut struct {
	// Lower bound.
	LowerType        string `json:"lower_type,omitempty"`
	LowerValue       *int   `json:"lower_value,omitempty"`
	LowerDenominator *int   `json:"lower_denominator,omitempty"` // If a fraction.
	// Higher bound.
	HigherType        string `json:"higher_type,omitempty"`
	HigherValue       *int   `json:"higher_value,omitempty"`
	HigherDenominator *int   `json:"higher_denominator,omitempty"` // If a fraction.
	// What are these values?
	Units string `json:"units,omitempty"`
	// What precision should we support of these values?
	Precision float64 `json:"precision,omitempty"`
}
