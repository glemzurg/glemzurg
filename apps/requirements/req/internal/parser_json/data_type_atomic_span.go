package parser_json

// atomicSpan represents a range of allowed values.
type atomicSpan struct {
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
