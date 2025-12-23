package parser_json

// atomicInOut represents the atomic data type (as opposed to a collection).
type atomicInOut struct {
	ConstraintType string            `json:"constraint_type,omitempty"`
	Span           *atomicSpanInOut  `json:"span,omitempty"`
	Reference      *string           `json:"reference,omitempty"`
	EnumOrdered    *bool             `json:"enum_ordered,omitempty"` // If defined and true, the enumeration values can be compared greater-lesser-than.
	Enums          []atomicEnumInOut `json:"enums,omitempty"`
	ObjectClassKey *string           `json:"object_class_key,omitempty"`
}
