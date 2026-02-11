package model_data_type

import (
	"fmt"
	"strconv"
	"strings"
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
	ConstraintType string `validate:"required,oneof=unconstrained span enumeration reference object"`
	Span           *AtomicSpan
	Reference      *string
	EnumOrdered    *bool // If defined and true, the enumeration values can be compared greater-lesser-than.
	Enums          []AtomicEnum
	ObjectClassKey *string
}

// Validate validates the Atomic struct.
func (a Atomic) Validate() error {
	// Validate struct tags (ConstraintType required + oneof).
	if err := _validate.Struct(a); err != nil {
		return err
	}

	// Reference: must be non-nil and non-empty for reference types; nil for others.
	if a.ConstraintType == _CONSTRAINT_TYPE_REFERENCE {
		if a.Reference == nil || *a.Reference == "" {
			return fmt.Errorf("Reference: Reference must not be nil or empty for reference types.")
		}
	} else {
		if a.Reference != nil {
			return fmt.Errorf("Reference: Reference must be nil for non-reference types.")
		}
	}

	// ObjectClassKey: must be non-nil and non-empty for object types; nil for others.
	if a.ConstraintType == _CONSTRAINT_TYPE_OBJECT {
		if a.ObjectClassKey == nil || *a.ObjectClassKey == "" {
			return fmt.Errorf("ObjectClassKey: ObjectClassKey must not be nil or empty for object types.")
		}
	} else {
		if a.ObjectClassKey != nil {
			return fmt.Errorf("ObjectClassKey: ObjectClassKey must be nil for non-object types.")
		}
	}

	// Enums: required when enumeration; must be empty when not enumeration; each must validate.
	if a.ConstraintType == _CONSTRAINT_TYPE_ENUMERATION {
		if len(a.Enums) == 0 {
			return fmt.Errorf("Enums: cannot be blank.")
		}
		for _, enum := range a.Enums {
			if err := enum.Validate(); err != nil {
				return fmt.Errorf("Enums: (%s).", err.Error())
			}
		}
	} else {
		if len(a.Enums) > 0 {
			return fmt.Errorf("Enums: must be blank.")
		}
	}

	// EnumOrdered: must be non-nil for enumeration; nil for others.
	if a.ConstraintType == _CONSTRAINT_TYPE_ENUMERATION {
		if a.EnumOrdered == nil {
			return fmt.Errorf("EnumOrdered: EnumOrdered must not be nil for enumeration types.")
		}
	} else {
		if a.EnumOrdered != nil {
			return fmt.Errorf("EnumOrdered: EnumOrdered must be nil for non-enumeration types.")
		}
	}

	// Span: must be non-nil for span types (and validate); nil for others.
	if a.ConstraintType == _CONSTRAINT_TYPE_SPAN {
		if a.Span == nil {
			return fmt.Errorf("Span: Span must not be nil for span types.")
		}
		if err := a.Span.Validate(); err != nil {
			return fmt.Errorf("Span: (%s).", err.Error())
		}
	} else {
		if a.Span != nil {
			return fmt.Errorf("Span: Span must be nil for non-span types.")
		}
	}

	return nil
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
