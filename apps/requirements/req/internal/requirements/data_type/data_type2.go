package data_type


// AtomicSpan represents a span.
type AtomicSpan struct {
	LowerType         string // "closed", "open", "unconstrained"
	LowerValue        *int
	LowerDenominator  *int
	HigherType        string
	HigherValue       *int
	HigherDenominator *int
	Units             string
	Precision         int
}

// Field represents a field in a record.
type Field struct {
	Name       string
	DataType   *DataType
	Details    string
	UmlComment string
}
