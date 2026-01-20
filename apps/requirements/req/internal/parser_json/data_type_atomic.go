package parser_json

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_data_type"

// atomicInOut represents the atomic data type (as opposed to a collection).
type atomicInOut struct {
	ConstraintType string            `json:"constraint_type"`
	Span           *atomicSpanInOut  `json:"span"`
	Reference      *string           `json:"reference"`
	EnumOrdered    *bool             `json:"enum_ordered"` // If defined and true, the enumeration values can be compared greater-lesser-than.
	Enums          []atomicEnumInOut `json:"enums"`
	ObjectClassKey *string           `json:"object_class_key"`
}

// ToRequirements converts the atomicInOut to model_data_type.Atomic.
func (a atomicInOut) ToRequirements() model_data_type.Atomic {
	atomic := model_data_type.Atomic{
		ConstraintType: a.ConstraintType,
		Span:           nil,
		Reference:      a.Reference,
		EnumOrdered:    a.EnumOrdered,
		Enums:          nil,
		ObjectClassKey: a.ObjectClassKey,
	}
	if a.Span != nil {
		s := a.Span.ToRequirements()
		atomic.Span = &s
	}
	for _, e := range a.Enums {
		atomic.Enums = append(atomic.Enums, e.ToRequirements())
	}
	return atomic
}

// FromRequirements creates a atomicInOut from model_data_type.Atomic.
func FromRequirementsAtomic(a model_data_type.Atomic) atomicInOut {
	atomic := atomicInOut{
		ConstraintType: a.ConstraintType,
		Span:           nil,
		Reference:      a.Reference,
		EnumOrdered:    a.EnumOrdered,
		Enums:          nil,
		ObjectClassKey: a.ObjectClassKey,
	}
	if a.Span != nil {
		s := FromRequirementsAtomicSpan(*a.Span)
		atomic.Span = &s
	}
	for _, e := range a.Enums {
		atomic.Enums = append(atomic.Enums, FromRequirementsAtomicEnum(e))
	}
	return atomic
}
