package data_type

// DataType represents the main data type structure.
type DataType struct {
	CollectionType   string // "atomic", "ordered", "queue", "record", "stack", "unordered"
	CollectionUnique *bool
	CollectionMin    *int
	CollectionMax    *int
	Details          string
	UmlComment       string
	Atomic           *Atomic
	ElementType      *DataType // For collections
	Fields           []Field   // For records
}

// AtomicEnumValue represents an enum value.
type AtomicEnumValue struct {
	Value     string
	SortOrder int
}

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
