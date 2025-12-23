package parser_json

// DataType represents the main data type structure.
type DataType struct {
	Key              string
	CollectionType   string
	CollectionUnique *bool
	CollectionMin    *int
	CollectionMax    *int
	Atomic           *Atomic
	RecordFields     []Field
}

// Atomic represents the atomic data type (as opposed to a collection).
type Atomic struct {
	ConstraintType string
	Span           *AtomicSpan
	Reference      *string
	EnumOrdered    *bool // If defined and true, the enumeration values can be compared greater-lesser-than.
	Enums          []AtomicEnum
	ObjectClassKey *string
}

// Field represents a single field of a record datatype.
type Field struct {
	Name          string    // The name of the field.
	FieldDataType *DataType // The data type of this field.
}

// AtomicSpan represents a range of allowed values.
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

// AtomicEnum represents an allowed value in an enumeration.
type AtomicEnum struct {
	Value     string
	SortOrder int
}
