package parser_json

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
