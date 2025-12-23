package parser_json

// atomic represents the atomic data type (as opposed to a collection).
type atomic struct {
	ConstraintType string
	Span           *atomicSpan
	Reference      *string
	EnumOrdered    *bool // If defined and true, the enumeration values can be compared greater-lesser-than.
	Enums          []atomicEnum
	ObjectClassKey *string
}
