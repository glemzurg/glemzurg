package model_data_type

import (
	"errors"
	"strconv"
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
	Span           *AtomicSpan
	Reference      *string
	EnumOrdered    *bool // If defined and true, the enumeration values can be compared greater-lesser-than.
	Enums          []AtomicEnum
	ObjectClassKey *string
}

// Validate validates the Atomic struct.
func (a Atomic) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.ConstraintType, validation.Required, validation.In(_CONSTRAINT_TYPE_UNCONSTRAINED, _CONSTRAINT_TYPE_SPAN, _CONSTRAINT_TYPE_REFERENCE, _CONSTRAINT_TYPE_OBJECT, _CONSTRAINT_TYPE_ENUMERATION)),
		validation.Field(&a.Reference, validation.By(func(value interface{}) error {
			ptr, ok := value.(*string)
			if !ok {
				return errors.New("Reference must be *string")
			}
			if a.ConstraintType == _CONSTRAINT_TYPE_REFERENCE {
				if ptr == nil || *ptr == "" {
					return errors.New("Reference must not be nil or empty for reference types")
				}
			} else {
				if ptr != nil {
					return errors.New("Reference must be nil for non-reference types")
				}
			}
			return nil
		})),
		validation.Field(&a.ObjectClassKey, validation.By(func(value interface{}) error {
			ptr, ok := value.(*string)
			if !ok {
				return errors.New("ObjectClassKey must be *string")
			}
			if a.ConstraintType == _CONSTRAINT_TYPE_OBJECT {
				if ptr == nil || *ptr == "" {
					return errors.New("ObjectClassKey must not be nil or empty for object types")
				}
			} else {
				if ptr != nil {
					return errors.New("ObjectClassKey must be nil for non-object types")
				}
			}
			return nil
		})),
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
		validation.Field(&a.Span, validation.By(func(value interface{}) error {
			ptr, ok := value.(*AtomicSpan)
			if !ok {
				return errors.New("Span must be *AtomicSpan")
			}
			if a.ConstraintType == _CONSTRAINT_TYPE_SPAN {
				if ptr == nil {
					return errors.New("Span must not be nil for span types")
				}
				return ptr.Validate()
			} else {
				if ptr != nil {
					return errors.New("Span must be nil for non-span types")
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

	case _CONSTRAINT_TYPE_SPAN:
		if a.Span == nil {
			return "span: <nil>"

		}

		lowerBracket := "("
		if a.Span.LowerType == "closed" {
			lowerBracket = "["
		}

		lowerStr := "unconstrained"
		if a.Span.LowerValue != nil {
			lowerStr = strconv.Itoa(*a.Span.LowerValue)
			if a.Span.LowerDenominator != nil && *a.Span.LowerDenominator > 1 {
				lowerStr += "/" + strconv.Itoa(*a.Span.LowerDenominator)
			}
		}

		higherStr := "unconstrained"
		if a.Span.HigherValue != nil {
			higherStr = strconv.Itoa(*a.Span.HigherValue)
			if a.Span.HigherDenominator != nil && *a.Span.HigherDenominator > 1 {
				higherStr += "/" + strconv.Itoa(*a.Span.HigherDenominator)
			}
		}

		higherBracket := ")"
		if a.Span.HigherType == "closed" {
			higherBracket = "]"
		}

		return lowerBracket + lowerStr + " .. " + higherStr + higherBracket + " at " + strconv.FormatFloat(a.Span.Precision, 'g', -1, 64) + " " + a.Span.Units

	case _CONSTRAINT_TYPE_REFERENCE:
		return "ref from " + *a.Reference

	case _CONSTRAINT_TYPE_OBJECT:
		return "obj of " + *a.ObjectClassKey

	case _CONSTRAINT_TYPE_ENUMERATION:
		var values []string
		for _, enum := range a.Enums {
			values = append(values, enum.Value)
		}
		prefix := "enum of"
		if a.EnumOrdered != nil && *a.EnumOrdered {
			prefix = "ord enum of"
		}
		return prefix + " " + strings.Join(values, ", ")
	default:
		panic("invalid constraint type: '" + a.ConstraintType + "'")
	}
}
