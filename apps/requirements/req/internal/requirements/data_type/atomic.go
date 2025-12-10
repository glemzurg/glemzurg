package data_type

import (
	"errors"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

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
	Enums          []AtomicEnum
	ObjectClassKey string
}

// Validate validates the Atomic struct.
func (a Atomic) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.ConstraintType, validation.Required, validation.In(_CONSTRAINT_TYPE_UNCONSTRAINED, _CONSTRAINT_TYPE_REFERENCE, _CONSTRAINT_TYPE_OBJECT, _CONSTRAINT_TYPE_ENUMERATION)),
		validation.Field(&a.Reference, validation.Required.When(a.ConstraintType == _CONSTRAINT_TYPE_REFERENCE)),
		validation.Field(&a.ObjectClassKey, validation.Required.When(a.ConstraintType == _CONSTRAINT_TYPE_OBJECT)),
		validation.Field(&a.Enums, validation.Required.When(a.ConstraintType == _CONSTRAINT_TYPE_ENUMERATION), validation.Empty.When(a.ConstraintType != _CONSTRAINT_TYPE_ENUMERATION), validation.Each(validation.By(func(value interface{}) error { enum := value.(AtomicEnum); return (&enum).Validate() }))),
		validation.Field(&a.EnumOrdered, validation.By(func(value interface{}) error {
			ptr, ok := value.(*bool)
			if !ok {
				return errors.New("EnumOrdered must be *bool")
			}
			if a.ConstraintType == _CONSTRAINT_TYPE_ENUMERATION {
				if ptr == nil {
					return errors.New("EnumOrdered must not be nil for enumeration types")
				}
			} else {
				if ptr != nil {
					return errors.New("EnumOrdered must be nil for non-enumeration types")
				}
			}
			return nil
		})),
	)
}

// String returns a string representation of the Atomic type.
func (a Atomic) String() string {
	switch a.ConstraintType {
	case _CONSTRAINT_TYPE_UNCONSTRAINED:
		return "unconstrained"
	case _CONSTRAINT_TYPE_REFERENCE:
		return "ref: " + a.Reference
	case _CONSTRAINT_TYPE_OBJECT:
		return "obj: " + a.ObjectClassKey
	case _CONSTRAINT_TYPE_ENUMERATION:
		var values []string
		for _, enum := range a.Enums {
			values = append(values, enum.Value)
		}
		prefix := "enum:"
		if a.EnumOrdered != nil && *a.EnumOrdered {
			prefix = "ord-enum:"
		}
		return prefix + " " + strings.Join(values, ", ")
	default:
		panic("invalid constraint type: '" + a.ConstraintType + "'")
	}
}
