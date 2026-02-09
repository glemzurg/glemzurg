package invariants

import (
	"fmt"
	"math/big"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

// DataTypeChecker validates attribute values against their data type constraints.
// It checks:
//   - Required (non-nullable) attributes are not nil
//   - Numeric values are within span constraints
//   - Enumeration values are in the allowed set
//   - Collection sizes are within bounds
type DataTypeChecker struct {
	// classAttributes maps class key to its attribute definitions
	classAttributes map[identity.Key]map[string]*model_class.Attribute
}

// NewDataTypeChecker creates a new data type checker from a model.
// Returns an error if any class attribute has an unparsed DataType.
func NewDataTypeChecker(model *req_model.Model) (*DataTypeChecker, ViolationList) {
	checker := &DataTypeChecker{
		classAttributes: make(map[identity.Key]map[string]*model_class.Attribute),
	}

	var violations ViolationList

	// Iterate through all domains, subdomains, and classes to collect attributes
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				attrMap := make(map[string]*model_class.Attribute)

				for _, attr := range class.Attributes {
					attrCopy := attr // Make a copy to get a stable pointer
					attrMap[attr.Name] = &attrCopy

					// Check if DataType is parsed
					if attr.DataType == nil {
						violations = append(violations, NewUnparsedDataTypeViolation(
							class.Key,
							attr.Name,
							attr.DataTypeRules,
						))
					}
				}

				checker.classAttributes[class.Key] = attrMap
			}
		}
	}

	return checker, violations
}

// CheckInstance validates all attribute values on an instance against their data type constraints.
func (c *DataTypeChecker) CheckInstance(instance *state.ClassInstance) ViolationList {
	var violations ViolationList

	attrs, ok := c.classAttributes[instance.ClassKey]
	if !ok {
		// No attribute definitions for this class - skip validation
		return violations
	}

	for attrName, attrDef := range attrs {
		value := instance.GetAttribute(attrName)

		// Check required (non-nullable) constraint
		if !attrDef.Nullable && value == nil {
			violations = append(violations, NewRequiredAttributeViolation(
				instance.ID,
				instance.ClassKey,
				attrName,
			))
			continue // No point checking other constraints on nil value
		}

		// If value is nil and attribute is nullable, skip constraint checks
		if value == nil {
			continue
		}

		// Check data type constraints if DataType is parsed
		if attrDef.DataType != nil {
			typeViolations := c.checkDataTypeConstraints(
				instance.ID,
				instance.ClassKey,
				attrName,
				value,
				attrDef.DataType,
			)
			violations = append(violations, typeViolations...)
		}
	}

	return violations
}

// checkDataTypeConstraints validates a value against its data type constraints.
func (c *DataTypeChecker) checkDataTypeConstraints(
	instanceID state.InstanceID,
	classKey identity.Key,
	attrName string,
	value object.Object,
	dataType *model_data_type.DataType,
) ViolationList {
	var violations ViolationList

	// Check collection size constraints
	if sizeViolation := c.checkCollectionSize(instanceID, classKey, attrName, value, dataType); sizeViolation != nil {
		violations = append(violations, sizeViolation)
	}

	// Check atomic constraints (span, enumeration)
	if dataType.Atomic != nil {
		atomicViolations := c.checkAtomicConstraints(
			instanceID,
			classKey,
			attrName,
			value,
			dataType.Atomic,
		)
		violations = append(violations, atomicViolations...)
	}

	return violations
}

// checkCollectionSize validates collection size against min/max constraints.
func (c *DataTypeChecker) checkCollectionSize(
	instanceID state.InstanceID,
	classKey identity.Key,
	attrName string,
	value object.Object,
	dataType *model_data_type.DataType,
) *Violation {
	// Skip if no collection constraints
	if dataType.CollectionMin == nil && dataType.CollectionMax == nil {
		return nil
	}

	// Get size based on collection type
	var size int
	switch dataType.CollectionType {
	case model_data_type.CollectionTypeAtomic:
		// Atomic types don't have collection size constraints
		return nil
	case model_data_type.CollectionTypeOrdered, model_data_type.CollectionTypeQueue, model_data_type.CollectionTypeStack:
		if tuple, ok := value.(*object.Tuple); ok {
			size = tuple.Len()
		} else {
			return nil // Type mismatch handled elsewhere
		}
	case model_data_type.CollectionTypeUnordered:
		if set, ok := value.(*object.Set); ok {
			size = set.Size()
		} else {
			return nil // Type mismatch handled elsewhere
		}
	case model_data_type.CollectionTypeRecord:
		if rec, ok := value.(*object.Record); ok {
			size = len(rec.FieldNames())
		} else {
			return nil // Type mismatch handled elsewhere
		}
	default:
		return nil
	}

	// Check min constraint
	if dataType.CollectionMin != nil && size < *dataType.CollectionMin {
		return NewCollectionSizeViolation(
			instanceID,
			classKey,
			attrName,
			size,
			dataType.CollectionMin,
			dataType.CollectionMax,
		)
	}

	// Check max constraint
	if dataType.CollectionMax != nil && size > *dataType.CollectionMax {
		return NewCollectionSizeViolation(
			instanceID,
			classKey,
			attrName,
			size,
			dataType.CollectionMin,
			dataType.CollectionMax,
		)
	}

	return nil
}

// checkAtomicConstraints validates value against atomic type constraints (span, enumeration).
func (c *DataTypeChecker) checkAtomicConstraints(
	instanceID state.InstanceID,
	classKey identity.Key,
	attrName string,
	value object.Object,
	atomic *model_data_type.Atomic,
) ViolationList {
	var violations ViolationList

	switch atomic.ConstraintType {
	case model_data_type.ConstraintTypeUnconstrained:
		// No constraints to check
		return violations

	case model_data_type.ConstraintTypeSpan:
		if atomic.Span != nil {
			if violation := c.checkSpanConstraint(instanceID, classKey, attrName, value, atomic.Span); violation != nil {
				violations = append(violations, violation)
			}
		}

	case model_data_type.ConstraintTypeEnumeration:
		if len(atomic.Enums) > 0 {
			if violation := c.checkEnumConstraint(instanceID, classKey, attrName, value, atomic.Enums); violation != nil {
				violations = append(violations, violation)
			}
		}

	case model_data_type.ConstraintTypeReference, model_data_type.ConstraintTypeObject:
		// Reference and object constraints are handled by link/instance validation
		return violations
	}

	return violations
}

// checkSpanConstraint validates a numeric value against a span (range) constraint.
func (c *DataTypeChecker) checkSpanConstraint(
	instanceID state.InstanceID,
	classKey identity.Key,
	attrName string,
	value object.Object,
	span *model_data_type.AtomicSpan,
) *Violation {
	// Get numeric value
	num, ok := value.(*object.Number)
	if !ok {
		// Not a number - type mismatch handled elsewhere
		return nil
	}

	// Get the value as a big.Rat for comparison
	valueRat := num.Rat()

	// Build range string for error message
	rangeStr := formatSpanRange(span)

	// Check lower bound
	if span.LowerType != "unconstrained" && span.LowerValue != nil {
		lowerRat := spanValueToRat(span.LowerValue, span.LowerDenominator)

		cmp := valueRat.Cmp(lowerRat)
		if span.LowerType == "closed" {
			// Closed: value >= lower
			if cmp < 0 {
				return NewSpanConstraintViolation(
					instanceID,
					classKey,
					attrName,
					num.Inspect(),
					rangeStr,
				)
			}
		} else if span.LowerType == "open" {
			// Open: value > lower
			if cmp <= 0 {
				return NewSpanConstraintViolation(
					instanceID,
					classKey,
					attrName,
					num.Inspect(),
					rangeStr,
				)
			}
		}
	}

	// Check upper bound
	if span.HigherType != "unconstrained" && span.HigherValue != nil {
		higherRat := spanValueToRat(span.HigherValue, span.HigherDenominator)

		cmp := valueRat.Cmp(higherRat)
		if span.HigherType == "closed" {
			// Closed: value <= higher
			if cmp > 0 {
				return NewSpanConstraintViolation(
					instanceID,
					classKey,
					attrName,
					num.Inspect(),
					rangeStr,
				)
			}
		} else if span.HigherType == "open" {
			// Open: value < higher
			if cmp >= 0 {
				return NewSpanConstraintViolation(
					instanceID,
					classKey,
					attrName,
					num.Inspect(),
					rangeStr,
				)
			}
		}
	}

	return nil
}

// spanValueToRat converts a span value (with optional denominator) to a *big.Rat.
func spanValueToRat(value *int, denom *int) *big.Rat {
	if value == nil {
		return big.NewRat(0, 1)
	}
	denomVal := 1
	if denom != nil {
		denomVal = *denom
	}
	return big.NewRat(int64(*value), int64(denomVal))
}

// formatSpanRange creates a human-readable string for a span range.
func formatSpanRange(span *model_data_type.AtomicSpan) string {
	var lower, higher string
	var lowerBracket, higherBracket string

	// Lower bound
	if span.LowerType == "unconstrained" {
		lower = "-∞"
		lowerBracket = "("
	} else {
		lower = formatSpanValue(span.LowerValue, span.LowerDenominator)
		if span.LowerType == "closed" {
			lowerBracket = "["
		} else {
			lowerBracket = "("
		}
	}

	// Higher bound
	if span.HigherType == "unconstrained" {
		higher = "+∞"
		higherBracket = ")"
	} else {
		higher = formatSpanValue(span.HigherValue, span.HigherDenominator)
		if span.HigherType == "closed" {
			higherBracket = "]"
		} else {
			higherBracket = ")"
		}
	}

	return fmt.Sprintf("%s%s, %s%s", lowerBracket, lower, higher, higherBracket)
}

// formatSpanValue formats a span value for display.
func formatSpanValue(value *int, denom *int) string {
	if value == nil {
		return "?"
	}
	if denom == nil || *denom == 1 {
		return fmt.Sprintf("%d", *value)
	}
	return fmt.Sprintf("%d/%d", *value, *denom)
}

// checkEnumConstraint validates a value against enumeration constraint.
func (c *DataTypeChecker) checkEnumConstraint(
	instanceID state.InstanceID,
	classKey identity.Key,
	attrName string,
	value object.Object,
	enums []model_data_type.AtomicEnum,
) *Violation {
	// Get string value
	str, ok := value.(*object.String)
	if !ok {
		// Not a string - type mismatch handled elsewhere
		return nil
	}

	strVal := str.Value()

	// Check if value is in allowed enumeration
	allowedValues := make([]string, len(enums))
	for i, enum := range enums {
		allowedValues[i] = enum.Value
		if enum.Value == strVal {
			return nil // Value found in enumeration
		}
	}

	// Value not in enumeration
	return NewEnumConstraintViolation(
		instanceID,
		classKey,
		attrName,
		strVal,
		allowedValues,
	)
}

// CheckState validates all instances in a simulation state.
func (c *DataTypeChecker) CheckState(simState *state.SimulationState) ViolationList {
	var violations ViolationList

	for _, instance := range simState.AllInstances() {
		instanceViolations := c.CheckInstance(instance)
		violations = append(violations, instanceViolations...)
	}

	return violations
}

// GetAttributeDefinition returns the attribute definition for a class attribute.
func (c *DataTypeChecker) GetAttributeDefinition(classKey identity.Key, attrName string) *model_class.Attribute {
	if attrs, ok := c.classAttributes[classKey]; ok {
		return attrs[attrName]
	}
	return nil
}

// HasClass returns true if the checker has attribute definitions for the class.
func (c *DataTypeChecker) HasClass(classKey identity.Key) bool {
	_, ok := c.classAttributes[classKey]
	return ok
}
