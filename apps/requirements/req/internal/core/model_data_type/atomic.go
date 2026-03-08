package model_data_type

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
)

const (
	CONSTRAINT_TYPE_UNCONSTRAINED = "unconstrained" // Anything.
	CONSTRAINT_TYPE_SPAN          = "span"          // A range of allowed values.
	CONSTRAINT_TYPE_ENUMERATION   = "enumeration"   // A set of allowed values.
	CONSTRAINT_TYPE_REFERENCE     = "reference"     // A reference to other documentation.
	CONSTRAINT_TYPE_OBJECT        = "object"        // An object of a class.
)

var _validConstraintTypes = map[string]bool{
	CONSTRAINT_TYPE_UNCONSTRAINED: true,
	CONSTRAINT_TYPE_SPAN:          true,
	CONSTRAINT_TYPE_ENUMERATION:   true,
	CONSTRAINT_TYPE_REFERENCE:     true,
	CONSTRAINT_TYPE_OBJECT:        true,
}

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
	// ConstraintType: required and must be a valid value.
	if a.ConstraintType == "" {
		return &coreerr.ValidationError{
			Code:    coreerr.DtypeAtomicConstrainttypeRequired,
			Message: "ConstraintType is required",
			Field:   "ConstraintType",
			Want:    "one of: unconstrained, span, enumeration, reference, object",
		}
	}
	if !_validConstraintTypes[a.ConstraintType] {
		return &coreerr.ValidationError{
			Code:    coreerr.DtypeAtomicConstrainttypeInvalid,
			Message: "ConstraintType is not a valid value",
			Field:   "ConstraintType",
			Got:     a.ConstraintType,
			Want:    "one of: unconstrained, span, enumeration, reference, object",
		}
	}

	if err := a.validateReference(); err != nil {
		return err
	}
	if err := a.validateObjectClassKey(); err != nil {
		return err
	}
	if err := a.validateEnums(); err != nil {
		return err
	}
	if err := a.validateSpan(); err != nil {
		return err
	}
	return nil
}

func (a Atomic) validateReference() error {
	if a.ConstraintType == CONSTRAINT_TYPE_REFERENCE {
		if a.Reference == nil || *a.Reference == "" {
			return fmt.Errorf("reference: reference must not be nil or empty for reference types")
		}
	} else if a.Reference != nil {
		return fmt.Errorf("reference: reference must be nil for non-reference types")
	}
	return nil
}

func (a Atomic) validateObjectClassKey() error {
	if a.ConstraintType == CONSTRAINT_TYPE_OBJECT {
		if a.ObjectClassKey == nil || *a.ObjectClassKey == "" {
			return fmt.Errorf("objectClassKey: objectClassKey must not be nil or empty for object types")
		}
	} else if a.ObjectClassKey != nil {
		return fmt.Errorf("objectClassKey: objectClassKey must be nil for non-object types")
	}
	return nil
}

func (a Atomic) validateEnums() error {
	if a.ConstraintType == CONSTRAINT_TYPE_ENUMERATION {
		if len(a.Enums) == 0 {
			return fmt.Errorf("enums: cannot be blank")
		}
		for _, enum := range a.Enums {
			if err := enum.Validate(); err != nil {
				return fmt.Errorf("enums: (%s)", err.Error())
			}
		}
		if a.EnumOrdered == nil {
			return fmt.Errorf("enumOrdered: enumOrdered must not be nil for enumeration types")
		}
	} else {
		if len(a.Enums) > 0 {
			return fmt.Errorf("enums: must be blank")
		}
		if a.EnumOrdered != nil {
			return fmt.Errorf("enumOrdered: enumOrdered must be nil for non-enumeration types")
		}
	}
	return nil
}

func (a Atomic) validateSpan() error {
	if a.ConstraintType == CONSTRAINT_TYPE_SPAN {
		if a.Span == nil {
			return fmt.Errorf("span: span must not be nil for span types")
		}
		if err := a.Span.Validate(); err != nil {
			return fmt.Errorf("span: (%s)", err.Error())
		}
	} else if a.Span != nil {
		return fmt.Errorf("span: span must be nil for non-span types")
	}
	return nil
}

// String returns a string representation of the Atomic type.
func (a Atomic) String() string {
	switch a.ConstraintType {
	case CONSTRAINT_TYPE_UNCONSTRAINED:
		return CONSTRAINT_TYPE_UNCONSTRAINED
	case CONSTRAINT_TYPE_SPAN:
		return a.spanString()
	case CONSTRAINT_TYPE_REFERENCE:
		return "ref from " + *a.Reference
	case CONSTRAINT_TYPE_OBJECT:
		return "obj of " + *a.ObjectClassKey
	case CONSTRAINT_TYPE_ENUMERATION:
		return a.enumString()
	default:
		panic("invalid constraint type: '" + a.ConstraintType + "'")
	}
}

func (a Atomic) spanString() string {
	if a.Span == nil {
		return "span: <nil>"
	}
	lowerBracket := "("
	if a.Span.LowerType == _BOUND_TYPE_LIMIT_CLOSED {
		lowerBracket = "["
	}
	lowerStr := formatBound(a.Span.LowerValue, a.Span.LowerDenominator)
	higherStr := formatBound(a.Span.HigherValue, a.Span.HigherDenominator)
	higherBracket := ")"
	if a.Span.HigherType == _BOUND_TYPE_LIMIT_CLOSED {
		higherBracket = "]"
	}
	return lowerBracket + lowerStr + " .. " + higherStr + higherBracket + " at " + strconv.FormatFloat(a.Span.Precision, 'g', -1, 64) + " " + a.Span.Units
}

func formatBound(value *int, denominator *int) string {
	if value == nil {
		return CONSTRAINT_TYPE_UNCONSTRAINED
	}
	s := strconv.Itoa(*value)
	if denominator != nil && *denominator > 1 {
		s += "/" + strconv.Itoa(*denominator)
	}
	return s
}

func (a Atomic) enumString() string {
	var values []string
	for _, enum := range a.Enums {
		values = append(values, enum.Value)
	}
	prefix := "enum of"
	if a.EnumOrdered != nil && *a.EnumOrdered {
		prefix = "ord enum of"
	}
	return prefix + " " + strings.Join(values, ", ")
}
