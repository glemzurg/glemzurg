package parser_json

// atomicInOut represents the atomic data type (as opposed to a collection).
type atomicInOut struct {
	ConstraintType string
	Span           *atomicSpanInOut
	Reference      *string
	EnumOrdered    *bool // If defined and true, the enumeration values can be compared greater-lesser-than.
	Enums          []atomicEnumInOut
	ObjectClassKey *string
}
