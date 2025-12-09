package data_type

// Atomic represents the atomic data type.
type Atomic struct {
	ConstraintType string // "unconstrained", "enumeration", "object", "reference", "span"
	Reference      string
	EnumOrdered    *bool
	ObjectClassKey string
	Details        string
	UmlComment     string
	EnumValues     []AtomicEnumValue
	Spans          []AtomicSpan
}
