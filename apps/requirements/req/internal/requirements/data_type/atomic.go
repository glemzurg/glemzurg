package data_type

import validation "github.com/go-ozzo/ozzo-validation/v4"

const (
	_CONSTRAINT_TYPE_UNCONSTRAINED = "unconstrained" // Anything.
	_CONSTRAINT_TYPE_SPAN          = "span"          // A range of allowed values.
	_CONSTRAINT_TYPE_ENUMERATION   = "enumeration"   // A set of allowed values.
	_CONSTRAINT_TYPE_REFERENCE     = "reference"     // A reference to other documentation.
	_CONSTRAINT_TYPE_OBJECT        = "object"        // An object of a class.
)

// Atomic represents the atomic data type (as opposed to a collection).
type Atomic struct {
	ConstraintType string
	Reference      string
	EnumOrdered    *bool // If defined and true, the enumeration values can be compared greater-lesser-than.
	ObjectClassKey string
}

// Validate validates the Atomic struct.
func (a Atomic) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.ConstraintType, validation.Required, validation.In(_CONSTRAINT_TYPE_UNCONSTRAINED, _CONSTRAINT_TYPE_REFERENCE)),
	)
}

// String returns a string representation of the Atomic type.
func (a Atomic) String() string {
	switch a.ConstraintType {
	case _CONSTRAINT_TYPE_UNCONSTRAINED:
		return "unconstrained"
	case _CONSTRAINT_TYPE_REFERENCE:
		return "ref: " + a.Reference
	default:
		panic("invalid constraint type: '" + a.ConstraintType + "'")
	}
}
