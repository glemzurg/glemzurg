package parser_json

// atomicInOut represents the atomic data type (as opposed to a collection).
type atomicInOut struct {
	ConstraintType string            `json:"constraint_type"`
	Span           *atomicSpanInOut  `json:"span"`
	Reference      *string           `json:"reference"`
	EnumOrdered    *bool             `json:"enum_ordered"` // If defined and true, the enumeration values can be compared greater-lesser-than.
	Enums          []atomicEnumInOut `json:"enums"`
	ObjectClassKey *string           `json:"object_class_key"`
}
