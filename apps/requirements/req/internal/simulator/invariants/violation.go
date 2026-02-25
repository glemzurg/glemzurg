// Package invariants provides invariant checking for TLA+ simulation.
// It validates both TLA+ invariants and data type constraints.
package invariants

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

// ViolationType indicates what kind of invariant was violated.
type ViolationType int

const (
	// ViolationTypeModelInvariant indicates a TLA+ model invariant violation.
	ViolationTypeModelInvariant ViolationType = iota

	// ViolationTypeActionGuarantee indicates an action's TLA+ guarantee (post-condition) violation.
	ViolationTypeActionGuarantee

	// ViolationTypeQueryGuarantee indicates a query's TLA+ guarantee (post-condition) violation.
	ViolationTypeQueryGuarantee

	// ViolationTypeRequiredAttribute indicates a required (non-nullable) attribute is nil.
	ViolationTypeRequiredAttribute

	// ViolationTypeSpanConstraint indicates a numeric value is outside the allowed range.
	ViolationTypeSpanConstraint

	// ViolationTypeEnumConstraint indicates a value is not in the allowed enumeration.
	ViolationTypeEnumConstraint

	// ViolationTypeCollectionSize indicates a collection's size is outside allowed bounds.
	ViolationTypeCollectionSize

	// ViolationTypeUnparsedDataType indicates an attribute has no parsed DataType.
	ViolationTypeUnparsedDataType

	// ViolationTypeIndexUniqueness indicates two instances share the same index tuple.
	ViolationTypeIndexUniqueness

	// ViolationTypeMultiplicity indicates an association multiplicity constraint is not met.
	ViolationTypeMultiplicity

	// ViolationTypeSafetyRule indicates an action's safety rule was violated.
	ViolationTypeSafetyRule

	// ViolationTypeLivenessClassNotInstantiated indicates a class was never instantiated during simulation.
	ViolationTypeLivenessClassNotInstantiated

	// ViolationTypeLivenessAttributeNotWritten indicates a class attribute was never written during simulation.
	ViolationTypeLivenessAttributeNotWritten

	// ViolationTypeLivenessAssociationNotLinked indicates an association never had a link created during simulation.
	ViolationTypeLivenessAssociationNotLinked

	// ViolationTypeLivenessAttributeNotRead indicates a class attribute was never read during simulation (reserved for future use).
	ViolationTypeLivenessAttributeNotRead
)

// String returns a human-readable name for the violation type.
func (v ViolationType) String() string {
	switch v {
	case ViolationTypeModelInvariant:
		return "model_invariant"
	case ViolationTypeActionGuarantee:
		return "action_guarantee"
	case ViolationTypeQueryGuarantee:
		return "query_guarantee"
	case ViolationTypeRequiredAttribute:
		return "required_attribute"
	case ViolationTypeSpanConstraint:
		return "span_constraint"
	case ViolationTypeEnumConstraint:
		return "enum_constraint"
	case ViolationTypeCollectionSize:
		return "collection_size"
	case ViolationTypeUnparsedDataType:
		return "unparsed_data_type"
	case ViolationTypeIndexUniqueness:
		return "index_uniqueness"
	case ViolationTypeMultiplicity:
		return "multiplicity"
	case ViolationTypeSafetyRule:
		return "safety_rule"
	case ViolationTypeLivenessClassNotInstantiated:
		return "liveness_class_not_instantiated"
	case ViolationTypeLivenessAttributeNotWritten:
		return "liveness_attribute_not_written"
	case ViolationTypeLivenessAssociationNotLinked:
		return "liveness_association_not_linked"
	case ViolationTypeLivenessAttributeNotRead:
		return "liveness_attribute_not_read"
	default:
		return "unknown"
	}
}

// Violation represents a detected invariant violation during simulation.
type Violation struct {
	// Type indicates what kind of invariant was violated.
	Type ViolationType

	// Message is a human-readable description of the violation.
	Message string

	// InstanceID is the ID of the instance where the violation occurred.
	// Zero for model-level violations.
	InstanceID state.InstanceID

	// ClassKey identifies the class of the instance (if applicable).
	ClassKey identity.Key

	// AttributeName is the name of the attribute involved (for data type violations).
	AttributeName string

	// Expression is the TLA+ expression that was evaluated (for TLA+ violations).
	Expression string

	// ActionOrQueryKey identifies the action or query (for guarantee violations).
	ActionOrQueryKey identity.Key

	// ActionOrQueryName is the name of the action or query (for guarantee violations).
	ActionOrQueryName string

	// ExpectedValue is what the value should have been (for constraint violations).
	ExpectedValue string

	// ActualValue is what the value actually was (for constraint violations).
	ActualValue string

	// InvariantIndex is the index in Model.Invariants (for model invariants).
	InvariantIndex int

	// GuaranteeIndex is the index in the guarantee array (for guarantee violations).
	GuaranteeIndex int
}

// Error implements the error interface.
func (v *Violation) Error() string {
	return v.Message
}

// NewModelInvariantViolation creates a violation for a failed model invariant.
func NewModelInvariantViolation(index int, expression string, message string) *Violation {
	return &Violation{
		Type:           ViolationTypeModelInvariant,
		Message:        fmt.Sprintf("model invariant %d failed: %s - %s", index, expression, message),
		Expression:     expression,
		InvariantIndex: index,
	}
}

// NewActionGuaranteeViolation creates a violation for a failed action guarantee.
func NewActionGuaranteeViolation(
	actionKey identity.Key,
	actionName string,
	guaranteeIndex int,
	expression string,
	instanceID state.InstanceID,
	message string,
) *Violation {
	return &Violation{
		Type:              ViolationTypeActionGuarantee,
		Message:           fmt.Sprintf("action %s guarantee %d failed: %s - %s", actionName, guaranteeIndex, expression, message),
		InstanceID:        instanceID,
		ActionOrQueryKey:  actionKey,
		ActionOrQueryName: actionName,
		Expression:        expression,
		GuaranteeIndex:    guaranteeIndex,
	}
}

// NewQueryGuaranteeViolation creates a violation for a failed query guarantee.
func NewQueryGuaranteeViolation(
	queryKey identity.Key,
	queryName string,
	guaranteeIndex int,
	expression string,
	instanceID state.InstanceID,
	message string,
) *Violation {
	return &Violation{
		Type:              ViolationTypeQueryGuarantee,
		Message:           fmt.Sprintf("query %s guarantee %d failed: %s - %s", queryName, guaranteeIndex, expression, message),
		InstanceID:        instanceID,
		ActionOrQueryKey:  queryKey,
		ActionOrQueryName: queryName,
		Expression:        expression,
		GuaranteeIndex:    guaranteeIndex,
	}
}

// NewRequiredAttributeViolation creates a violation for a missing required attribute.
func NewRequiredAttributeViolation(
	instanceID state.InstanceID,
	classKey identity.Key,
	attributeName string,
) *Violation {
	return &Violation{
		Type:          ViolationTypeRequiredAttribute,
		Message:       fmt.Sprintf("required attribute %s is nil on instance %d of class %s", attributeName, instanceID, classKey.String()),
		InstanceID:    instanceID,
		ClassKey:      classKey,
		AttributeName: attributeName,
	}
}

// NewSpanConstraintViolation creates a violation for a value outside its allowed range.
func NewSpanConstraintViolation(
	instanceID state.InstanceID,
	classKey identity.Key,
	attributeName string,
	actualValue string,
	expectedRange string,
) *Violation {
	return &Violation{
		Type:          ViolationTypeSpanConstraint,
		Message:       fmt.Sprintf("attribute %s value %s is outside range %s on instance %d of class %s", attributeName, actualValue, expectedRange, instanceID, classKey.String()),
		InstanceID:    instanceID,
		ClassKey:      classKey,
		AttributeName: attributeName,
		ActualValue:   actualValue,
		ExpectedValue: expectedRange,
	}
}

// NewEnumConstraintViolation creates a violation for a value not in the allowed enumeration.
func NewEnumConstraintViolation(
	instanceID state.InstanceID,
	classKey identity.Key,
	attributeName string,
	actualValue string,
	allowedValues []string,
) *Violation {
	return &Violation{
		Type:          ViolationTypeEnumConstraint,
		Message:       fmt.Sprintf("attribute %s value %s is not in allowed values [%s] on instance %d of class %s", attributeName, actualValue, strings.Join(allowedValues, ", "), instanceID, classKey.String()),
		InstanceID:    instanceID,
		ClassKey:      classKey,
		AttributeName: attributeName,
		ActualValue:   actualValue,
		ExpectedValue: strings.Join(allowedValues, ", "),
	}
}

// NewCollectionSizeViolation creates a violation for a collection with invalid size.
func NewCollectionSizeViolation(
	instanceID state.InstanceID,
	classKey identity.Key,
	attributeName string,
	actualSize int,
	minSize *int,
	maxSize *int,
) *Violation {
	var rangeStr string
	if minSize != nil && maxSize != nil {
		rangeStr = fmt.Sprintf("[%d, %d]", *minSize, *maxSize)
	} else if minSize != nil {
		rangeStr = fmt.Sprintf("[%d, ∞)", *minSize)
	} else if maxSize != nil {
		rangeStr = fmt.Sprintf("[0, %d]", *maxSize)
	} else {
		rangeStr = "[0, ∞)"
	}

	return &Violation{
		Type:          ViolationTypeCollectionSize,
		Message:       fmt.Sprintf("attribute %s collection size %d is outside range %s on instance %d of class %s", attributeName, actualSize, rangeStr, instanceID, classKey.String()),
		InstanceID:    instanceID,
		ClassKey:      classKey,
		AttributeName: attributeName,
		ActualValue:   fmt.Sprintf("%d", actualSize),
		ExpectedValue: rangeStr,
	}
}

// NewUnparsedDataTypeViolation creates a violation for an attribute without a parsed DataType.
func NewUnparsedDataTypeViolation(classKey identity.Key, attributeName string, dataTypeRules string) *Violation {
	return &Violation{
		Type:          ViolationTypeUnparsedDataType,
		Message:       fmt.Sprintf("attribute %s on class %s has unparsed data type: %s", attributeName, classKey.String(), dataTypeRules),
		ClassKey:      classKey,
		AttributeName: attributeName,
		ExpectedValue: dataTypeRules,
	}
}

// NewIndexUniquenessViolation creates a violation for duplicate index tuples.
func NewIndexUniquenessViolation(
	instanceID state.InstanceID,
	conflictingInstanceID state.InstanceID,
	classKey identity.Key,
	indexNum uint,
	attrNames []string,
	tupleValues []string,
) *Violation {
	return &Violation{
		Type:       ViolationTypeIndexUniqueness,
		Message:    fmt.Sprintf("index %d uniqueness violated: attributes [%s] = [%s] duplicated on instances %d and %d of class %s", indexNum, strings.Join(attrNames, ", "), strings.Join(tupleValues, ", "), instanceID, conflictingInstanceID, classKey.String()),
		InstanceID: instanceID,
		ClassKey:   classKey,
	}
}

// NewSafetyRuleViolation creates a violation for a failed action safety rule.
func NewSafetyRuleViolation(
	actionKey identity.Key,
	actionName string,
	ruleIndex int,
	expression string,
	instanceID state.InstanceID,
	message string,
) *Violation {
	return &Violation{
		Type:              ViolationTypeSafetyRule,
		Message:           fmt.Sprintf("action %s safety rule %d failed: %s - %s", actionName, ruleIndex, expression, message),
		InstanceID:        instanceID,
		ActionOrQueryKey:  actionKey,
		ActionOrQueryName: actionName,
		Expression:        expression,
		GuaranteeIndex:    ruleIndex,
	}
}

// NewMultiplicityViolation creates a violation for an association multiplicity constraint failure.
func NewMultiplicityViolation(
	instanceID state.InstanceID,
	classKey identity.Key,
	associationName string,
	direction string,
	actualCount int,
	requiredMin uint,
	requiredMax uint,
	message string,
) *Violation {
	return &Violation{
		Type:       ViolationTypeMultiplicity,
		Message:    fmt.Sprintf("multiplicity violation on instance %d of class %s: association %s (%s) %s", instanceID, classKey.String(), associationName, direction, message),
		InstanceID: instanceID,
		ClassKey:   classKey,
	}
}

// NewLivenessClassNotInstantiatedViolation creates a violation for a class that was never instantiated.
func NewLivenessClassNotInstantiatedViolation(classKey identity.Key, className string) *Violation {
	return &Violation{
		Type:     ViolationTypeLivenessClassNotInstantiated,
		Message:  fmt.Sprintf("liveness: class %s was never instantiated during simulation", className),
		ClassKey: classKey,
	}
}

// NewLivenessAttributeNotWrittenViolation creates a violation for an attribute that was never written.
func NewLivenessAttributeNotWrittenViolation(classKey identity.Key, className, attributeName string) *Violation {
	return &Violation{
		Type:          ViolationTypeLivenessAttributeNotWritten,
		Message:       fmt.Sprintf("liveness: attribute %s on class %s was never written during simulation", attributeName, className),
		ClassKey:      classKey,
		AttributeName: attributeName,
	}
}

// NewLivenessAssociationNotLinkedViolation creates a violation for an association that was never linked.
func NewLivenessAssociationNotLinkedViolation(associationKey identity.Key, associationName string, fromClassKey, toClassKey identity.Key) *Violation {
	return &Violation{
		Type:    ViolationTypeLivenessAssociationNotLinked,
		Message: fmt.Sprintf("liveness: association %s (between %s and %s) never had a link created during simulation", associationName, fromClassKey.String(), toClassKey.String()),
	}
}

// ViolationList is a collection of violations.
type ViolationList []*Violation

// HasViolations returns true if there are any violations.
func (v ViolationList) HasViolations() bool {
	return len(v) > 0
}

// ByType filters violations by type.
func (v ViolationList) ByType(t ViolationType) ViolationList {
	var result ViolationList
	for _, violation := range v {
		if violation.Type == t {
			result = append(result, violation)
		}
	}
	return result
}

// TLAViolations returns all TLA+ related violations (model invariants and guarantees).
func (v ViolationList) TLAViolations() ViolationList {
	var result ViolationList
	for _, violation := range v {
		switch violation.Type {
		case ViolationTypeModelInvariant, ViolationTypeActionGuarantee, ViolationTypeQueryGuarantee:
			result = append(result, violation)
		}
	}
	return result
}

// DataTypeViolations returns all data type constraint violations.
func (v ViolationList) DataTypeViolations() ViolationList {
	var result ViolationList
	for _, violation := range v {
		switch violation.Type {
		case ViolationTypeRequiredAttribute, ViolationTypeSpanConstraint, ViolationTypeEnumConstraint, ViolationTypeCollectionSize, ViolationTypeIndexUniqueness:
			result = append(result, violation)
		}
	}
	return result
}

// LivenessViolations returns all liveness check violations.
func (v ViolationList) LivenessViolations() ViolationList {
	var result ViolationList
	for _, violation := range v {
		switch violation.Type {
		case ViolationTypeLivenessClassNotInstantiated,
			ViolationTypeLivenessAttributeNotWritten,
			ViolationTypeLivenessAssociationNotLinked,
			ViolationTypeLivenessAttributeNotRead:
			result = append(result, violation)
		}
	}
	return result
}

// Error returns a combined error message for all violations.
func (v ViolationList) Error() string {
	if len(v) == 0 {
		return ""
	}

	var messages []string
	for _, violation := range v {
		messages = append(messages, violation.Message)
	}
	return strings.Join(messages, "\n")
}
