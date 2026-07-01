// Package invariants provides invariant checking for TLA+ simulation.
// It validates both TLA+ invariants and data type constraints.
package invariants

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

// ViolationType indicates what kind of invariant was violated.
type ViolationType int

const (
	// ViolationTypeModelInvariant indicates a TLA+ model invariant violation.
	ViolationTypeModelInvariant ViolationType = iota

	// ViolationTypeClassInvariant indicates a TLA+ class-level invariant violation.
	ViolationTypeClassInvariant

	// ViolationTypeActionRequires indicates an action's TLA+ requires (precondition) was not met.
	ViolationTypeActionRequires

	// ViolationTypeActionGuarantee indicates an action's TLA+ guarantee (post-condition) violation.
	ViolationTypeActionGuarantee

	// ViolationTypeQueryGuarantee indicates a query's TLA+ guarantee (post-condition) violation.
	ViolationTypeQueryGuarantee

	// ViolationTypeAttributeInvariant indicates a failed attribute-level invariant assessment.
	ViolationTypeAttributeInvariant

	// ViolationTypeParameterInvariant indicates a failed action or query parameter invariant assessment.
	ViolationTypeParameterInvariant

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

	// ViolationTypeMissingAttributeTypeSpec indicates a written attribute has no TLA+ type_spec.
	ViolationTypeMissingAttributeTypeSpec

	// ViolationTypeMissingParameterTypeSpec indicates a simulated parameter has no TLA+ type_spec.
	ViolationTypeMissingParameterTypeSpec

	// ViolationTypeDateTimeTypeSpecMismatch indicates a datetime attribute or parameter has a type_spec other than Nat.
	ViolationTypeDateTimeTypeSpecMismatch

	// ViolationTypeDateTimeConstraint indicates a datetime value is outside the allowed Nat range.
	ViolationTypeDateTimeConstraint

	// ViolationTypeIndexUniqueness indicates two instances share the same index tuple.
	ViolationTypeIndexUniqueness

	// ViolationTypeAssociationInvariant indicates a failed association-level invariant assessment.
	ViolationTypeAssociationInvariant

	// ViolationTypeMultiplicity indicates an association multiplicity constraint is not met.
	ViolationTypeMultiplicity

	// ViolationTypeAssociationUniqueness indicates an association uniqueness tuple is duplicated.
	ViolationTypeAssociationUniqueness

	// ViolationTypeAssociationDuplicateLink indicates duplicate links between the same instance pair.
	ViolationTypeAssociationDuplicateLink

	// ViolationTypeSafetyRule indicates an action's safety rule was violated.
	ViolationTypeSafetyRule

	// ViolationTypeLivenessClassNotInstantiated indicates a class was never instantiated during simulation.
	ViolationTypeLivenessClassNotInstantiated

	// ViolationTypeLivenessAttributeNotWritten indicates a class attribute was never written during simulation.
	ViolationTypeLivenessAttributeNotWritten

	// ViolationTypeLivenessAssociationNotLinked indicates an association never had a link created during simulation.
	ViolationTypeLivenessAssociationNotLinked

	// ViolationTypeLivenessAttributeNotRead indicates an external derived attribute was never read during simulation.
	ViolationTypeLivenessAttributeNotRead

	// ViolationTypeLivenessEventNotSent indicates an event was never fired during simulation.
	ViolationTypeLivenessEventNotSent

	// ViolationTypeLivenessQueryNotRun indicates a query was never executed during simulation.
	ViolationTypeLivenessQueryNotRun

	// ViolationTypeLivenessActionNotExecuted indicates an action was never executed during simulation.
	ViolationTypeLivenessActionNotExecuted

	// ViolationTypeStateMachineIncomplete indicates a class state machine lacks the system _new event.
	ViolationTypeStateMachineIncomplete

	// ViolationTypePeerEventUnavailable indicates an association guarantee sent an event
	// the peer class cannot accept from its current state.
	ViolationTypePeerEventUnavailable
)

var violationTypeNames = map[ViolationType]string{
	ViolationTypeModelInvariant:               "model_invariant",
	ViolationTypeClassInvariant:               "class_invariant",
	ViolationTypeActionRequires:               "action_requires",
	ViolationTypeActionGuarantee:              "action_guarantee",
	ViolationTypeQueryGuarantee:               "query_guarantee",
	ViolationTypeAttributeInvariant:           "attribute_invariant",
	ViolationTypeParameterInvariant:           "parameter_invariant",
	ViolationTypeRequiredAttribute:            "required_attribute",
	ViolationTypeSpanConstraint:               "span_constraint",
	ViolationTypeEnumConstraint:               "enum_constraint",
	ViolationTypeCollectionSize:               "collection_size",
	ViolationTypeUnparsedDataType:             "unparsed_data_type",
	ViolationTypeMissingAttributeTypeSpec:     "missing_attribute_type_spec",
	ViolationTypeMissingParameterTypeSpec:     "missing_parameter_type_spec",
	ViolationTypeDateTimeTypeSpecMismatch:     "datetime_type_spec_mismatch",
	ViolationTypeDateTimeConstraint:           "datetime_constraint",
	ViolationTypeIndexUniqueness:              "index_uniqueness",
	ViolationTypeAssociationInvariant:         "association_invariant",
	ViolationTypeMultiplicity:                 "multiplicity",
	ViolationTypeAssociationUniqueness:        "association_uniqueness",
	ViolationTypeAssociationDuplicateLink:     "association_duplicate_link",
	ViolationTypeSafetyRule:                   "safety_rule",
	ViolationTypeLivenessClassNotInstantiated: "liveness_class_not_instantiated",
	ViolationTypeLivenessAttributeNotWritten:  "liveness_attribute_not_written",
	ViolationTypeLivenessAssociationNotLinked: "liveness_association_not_linked",
	ViolationTypeLivenessAttributeNotRead:     "liveness_attribute_not_read",
	ViolationTypeLivenessEventNotSent:         "liveness_event_not_sent",
	ViolationTypeLivenessQueryNotRun:          "liveness_query_not_run",
	ViolationTypeLivenessActionNotExecuted:    "liveness_action_not_executed",
	ViolationTypeStateMachineIncomplete:       "state_machine_incomplete",
	ViolationTypePeerEventUnavailable:         "peer_event_unavailable",
}

// String returns a human-readable name for the violation type.
func (v ViolationType) String() string {
	if name, ok := violationTypeNames[v]; ok {
		return name
	}
	return "unknown"
}

// ViolationError represents a detected invariant violation during simulation.
type ViolationError struct {
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
func (v *ViolationError) Error() string {
	return v.Message
}

// NewModelInvariantViolation creates a violation for a failed model invariant.
func NewModelInvariantViolation(index int, expression string, message string) *ViolationError {
	return &ViolationError{
		Type:           ViolationTypeModelInvariant,
		Message:        fmt.Sprintf("model invariant %d failed: %s - %s", index, expression, message),
		Expression:     expression,
		InvariantIndex: index,
	}
}

// NewClassInvariantViolation creates a violation for a failed class-level invariant.
func NewClassInvariantViolation(
	classKey identity.Key,
	instanceID state.InstanceID,
	index int,
	expression string,
	message string,
) *ViolationError {
	return &ViolationError{
		Type:           ViolationTypeClassInvariant,
		Message:        fmt.Sprintf("class %s invariant %d failed on instance %d: %s - %s", classKey.String(), index, instanceID, expression, message),
		InstanceID:     instanceID,
		ClassKey:       classKey,
		Expression:     expression,
		InvariantIndex: index,
	}
}

// NewAttributeInvariantViolation creates a violation for a failed attribute invariant.
func NewAttributeInvariantViolation(
	classKey identity.Key,
	instanceID state.InstanceID,
	attributeName string,
	invariantIndex int,
	expression string,
	message string,
) *ViolationError {
	return &ViolationError{
		Type:           ViolationTypeAttributeInvariant,
		Message:        fmt.Sprintf("class %s attribute %q invariant %d failed on instance %d: %s - %s", classKey.String(), attributeName, invariantIndex, instanceID, expression, message),
		InstanceID:     instanceID,
		ClassKey:       classKey,
		AttributeName:  attributeName,
		Expression:     expression,
		InvariantIndex: invariantIndex,
	}
}

// NewAssociationInvariantViolation creates a violation for a failed association invariant.
func NewAssociationInvariantViolation(
	associationKey identity.Key,
	associationName string,
	instanceID state.InstanceID,
	invariantIndex int,
	expression string,
	message string,
) *ViolationError {
	return &ViolationError{
		Type:              ViolationTypeAssociationInvariant,
		Message:           fmt.Sprintf("association %q invariant %d failed on instance %d: %s - %s", associationName, invariantIndex, instanceID, expression, message),
		InstanceID:        instanceID,
		Expression:        expression,
		InvariantIndex:    invariantIndex,
		ActionOrQueryKey:  associationKey,
		ActionOrQueryName: associationName,
	}
}

// NewParameterInvariantViolation creates a violation for a failed parameter invariant.
func NewParameterInvariantViolation(
	ownerKey identity.Key,
	ownerName string,
	invariantIndex int,
	expression string,
	instanceID state.InstanceID,
	message string,
) *ViolationError {
	return &ViolationError{
		Type:              ViolationTypeParameterInvariant,
		Message:           fmt.Sprintf("%s %s parameter invariant %d failed: %s - %s", ownerKey.KeyType, ownerName, invariantIndex, expression, message),
		InstanceID:        instanceID,
		ActionOrQueryKey:  ownerKey,
		ActionOrQueryName: ownerName,
		Expression:        expression,
		GuaranteeIndex:    invariantIndex,
	}
}

// NewActionRequiresViolation creates a violation for a failed action requires precondition.
func NewActionRequiresViolation(
	actionKey identity.Key,
	actionName string,
	requireIndex int,
	expression string,
	instanceID state.InstanceID,
	message string,
) *ViolationError {
	return &ViolationError{
		Type:              ViolationTypeActionRequires,
		Message:           fmt.Sprintf("action %s requires[%d] failed: %s - %s", actionName, requireIndex, expression, message),
		InstanceID:        instanceID,
		ActionOrQueryKey:  actionKey,
		ActionOrQueryName: actionName,
		Expression:        expression,
		GuaranteeIndex:    requireIndex,
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
) *ViolationError {
	return &ViolationError{
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
) *ViolationError {
	return &ViolationError{
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
) *ViolationError {
	return &ViolationError{
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
) *ViolationError {
	return &ViolationError{
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
) *ViolationError {
	return &ViolationError{
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
) *ViolationError {
	var rangeStr string
	switch {
	case minSize != nil && maxSize != nil:
		rangeStr = fmt.Sprintf("[%d, %d]", *minSize, *maxSize)
	case minSize != nil:
		rangeStr = fmt.Sprintf("[%d, ∞)", *minSize)
	case maxSize != nil:
		rangeStr = fmt.Sprintf("[0, %d]", *maxSize)
	default:
		rangeStr = "[0, ∞)"
	}

	return &ViolationError{
		Type:          ViolationTypeCollectionSize,
		Message:       fmt.Sprintf("attribute %s collection size %d is outside range %s on instance %d of class %s", attributeName, actualSize, rangeStr, instanceID, classKey.String()),
		InstanceID:    instanceID,
		ClassKey:      classKey,
		AttributeName: attributeName,
		ActualValue:   fmt.Sprintf("%d", actualSize),
		ExpectedValue: rangeStr,
	}
}

// NewUnparsedDataTypeViolation creates a class-level violation for an attribute without a parsed DataType.
func NewUnparsedDataTypeViolation(classKey identity.Key, attributeName string, dataTypeRules string) *ViolationError {
	return &ViolationError{
		Type:          ViolationTypeUnparsedDataType,
		Message:       fmt.Sprintf("attribute %s on class %s has unparsed data type rules: %s", attributeName, classKey.String(), dataTypeRules),
		ClassKey:      classKey,
		AttributeName: attributeName,
		ExpectedValue: dataTypeRules,
	}
}

// NewUnparsedAttributeDataTypeViolation creates an instance-level violation for an attribute
// value whose data type rules did not parse.
func NewUnparsedAttributeDataTypeViolation(
	instanceID state.InstanceID,
	classKey identity.Key,
	attributeName string,
	dataTypeRules string,
) *ViolationError {
	return &ViolationError{
		Type:          ViolationTypeUnparsedDataType,
		Message:       fmt.Sprintf("attribute %s on instance %d of class %s has unparsed data type rules: %s", attributeName, instanceID, classKey.String(), dataTypeRules),
		InstanceID:    instanceID,
		ClassKey:      classKey,
		AttributeName: attributeName,
		ExpectedValue: dataTypeRules,
	}
}

// ViolationSourceIdentity holds the key and name of the action or query that owns a parameter violation.
type ViolationSourceIdentity struct {
	Key  identity.Key
	Name string
}

// NewUnparsedParameterDataTypeViolation creates a violation when a simulated parameter's
// data type rules did not parse.
func NewUnparsedParameterDataTypeViolation(
	source ViolationSourceIdentity,
	sourceKind string,
	parameterName string,
	dataTypeRules string,
	instanceID state.InstanceID,
	classKey identity.Key,
) *ViolationError {
	return &ViolationError{
		Type:              ViolationTypeUnparsedDataType,
		Message:           fmt.Sprintf("parameter %s on %s %s has unparsed data type rules: %s", parameterName, sourceKind, source.Name, dataTypeRules),
		InstanceID:        instanceID,
		ClassKey:          classKey,
		AttributeName:     parameterName,
		ActionOrQueryKey:  source.Key,
		ActionOrQueryName: source.Name,
		ExpectedValue:     dataTypeRules,
	}
}

// NewMissingParameterTypeSpecViolation creates a violation when a simulated action or query
// parameter declares no TLA+ type_spec.
func NewMissingParameterTypeSpecViolation(
	sourceKey identity.Key,
	sourceName string,
	sourceKind string,
	parameterName string,
	instanceID state.InstanceID,
	classKey identity.Key,
) *ViolationError {
	return &ViolationError{
		Type:              ViolationTypeMissingParameterTypeSpec,
		Message:           fmt.Sprintf("parameter %s on %s %s has no TLA+ type_spec", parameterName, sourceKind, sourceName),
		InstanceID:        instanceID,
		ClassKey:          classKey,
		AttributeName:     parameterName,
		ActionOrQueryKey:  sourceKey,
		ActionOrQueryName: sourceName,
	}
}

// NewMissingAttributeTypeSpecViolation creates a violation when an instance holds a value
// for an attribute that declares no TLA+ type_spec.
func NewMissingAttributeTypeSpecViolation(
	instanceID state.InstanceID,
	classKey identity.Key,
	attributeName string,
) *ViolationError {
	return &ViolationError{
		Type:          ViolationTypeMissingAttributeTypeSpec,
		Message:       fmt.Sprintf("attribute %s on instance %d of class %s has no TLA+ type_spec but holds a value", attributeName, instanceID, classKey.String()),
		InstanceID:    instanceID,
		ClassKey:      classKey,
		AttributeName: attributeName,
	}
}

// NewDateTimeTypeSpecMismatchAttributeViolation reports a datetime attribute whose type_spec is not Nat.
func NewDateTimeTypeSpecMismatchAttributeViolation(
	instanceID state.InstanceID,
	classKey identity.Key,
	attributeName string,
	actualTypeSpec string,
) *ViolationError {
	return &ViolationError{
		Type:          ViolationTypeDateTimeTypeSpecMismatch,
		Message:       fmt.Sprintf("attribute %s on instance %d of class %s has datetime rules but type_spec %q is not Nat", attributeName, instanceID, classKey.String(), actualTypeSpec),
		InstanceID:    instanceID,
		ClassKey:      classKey,
		AttributeName: attributeName,
		ActualValue:   actualTypeSpec,
		ExpectedValue: "Nat",
	}
}

// DateTimeTypeSpecMismatchParameterParams holds parameters for a datetime parameter type_spec mismatch.
type DateTimeTypeSpecMismatchParameterParams struct {
	Source         ViolationSourceIdentity
	SourceKind     string
	ParameterName  string
	ActualTypeSpec string
	InstanceID     state.InstanceID
	ClassKey       identity.Key
}

// NewDateTimeTypeSpecMismatchParameterViolation reports a datetime parameter whose type_spec is not Nat.
func NewDateTimeTypeSpecMismatchParameterViolation(params DateTimeTypeSpecMismatchParameterParams) *ViolationError {
	return &ViolationError{
		Type:              ViolationTypeDateTimeTypeSpecMismatch,
		Message:           fmt.Sprintf("parameter %s on %s %s has datetime rules but type_spec %q is not Nat", params.ParameterName, params.SourceKind, params.Source.Name, params.ActualTypeSpec),
		InstanceID:        params.InstanceID,
		ClassKey:          params.ClassKey,
		AttributeName:     params.ParameterName,
		ActionOrQueryKey:  params.Source.Key,
		ActionOrQueryName: params.Source.Name,
		ActualValue:       params.ActualTypeSpec,
		ExpectedValue:     "Nat",
	}
}

// NewDateTimeConstraintViolation creates a violation for a datetime value outside [1, 100000000].
func NewDateTimeConstraintViolation(
	instanceID state.InstanceID,
	classKey identity.Key,
	attributeName string,
	actualValue string,
) *ViolationError {
	rangeStr := fmt.Sprintf("[%d, %d]", model_data_type.DateTimeValueMin, model_data_type.DateTimeValueMax)
	return &ViolationError{
		Type:          ViolationTypeDateTimeConstraint,
		Message:       fmt.Sprintf("attribute %s value %s is outside datetime range %s on instance %d of class %s", attributeName, actualValue, rangeStr, instanceID, classKey.String()),
		InstanceID:    instanceID,
		ClassKey:      classKey,
		AttributeName: attributeName,
		ActualValue:   actualValue,
		ExpectedValue: rangeStr,
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
) *ViolationError {
	return &ViolationError{
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
) *ViolationError {
	return &ViolationError{
		Type:              ViolationTypeSafetyRule,
		Message:           fmt.Sprintf("action %s safety rule %d failed: %s - %s", actionName, ruleIndex, expression, message),
		InstanceID:        instanceID,
		ActionOrQueryKey:  actionKey,
		ActionOrQueryName: actionName,
		Expression:        expression,
		GuaranteeIndex:    ruleIndex,
	}
}

// MultiplicityViolationParams holds the parameters for creating a multiplicity violation.
type MultiplicityViolationParams struct {
	InstanceID      state.InstanceID
	ClassKey        identity.Key
	AssociationName string
	Direction       string
	ActualCount     int
	RequiredMin     uint
	RequiredMax     uint
	Message         string
}

// NewMultiplicityViolation creates a violation for an association multiplicity constraint failure.
func NewMultiplicityViolation(params MultiplicityViolationParams) *ViolationError {
	return &ViolationError{
		Type:       ViolationTypeMultiplicity,
		Message:    fmt.Sprintf("multiplicity violation on instance %d of class %s: association %s (%s) %s", params.InstanceID, params.ClassKey.String(), params.AssociationName, params.Direction, params.Message),
		InstanceID: params.InstanceID,
		ClassKey:   params.ClassKey,
	}
}

// AssociationUniquenessViolationParams holds parameters for a per-pair association uniqueness failure.
type AssociationUniquenessViolationParams struct {
	AssociationName string
	FromInstanceID  state.InstanceID
	ToInstanceID    state.InstanceID
	ActualCount     int
	RequiredMin     uint
	RequiredMax     uint
	Message         string
}

// NewAssociationUniquenessViolation creates a violation for a duplicated association uniqueness tuple.
func NewAssociationUniquenessViolation(params AssociationUniquenessViolationParams) *ViolationError {
	return &ViolationError{
		Type: ViolationTypeAssociationUniqueness,
		Message: fmt.Sprintf(
			"association uniqueness violation: association %s between instances %d and %d %s",
			params.AssociationName,
			params.FromInstanceID,
			params.ToInstanceID,
			params.Message,
		),
		InstanceID: params.FromInstanceID,
	}
}

// AssociationDuplicateLinkViolationParams holds parameters for a duplicate instance-pair link failure.
type AssociationDuplicateLinkViolationParams struct {
	AssociationName string
	FromInstanceID  state.InstanceID
	ToInstanceID    state.InstanceID
	ActualCount     int
}

// NewAssociationDuplicateLinkViolation creates a violation for duplicate links on one instance pair.
func NewAssociationDuplicateLinkViolation(params AssociationDuplicateLinkViolationParams) *ViolationError {
	return &ViolationError{
		Type: ViolationTypeAssociationDuplicateLink,
		Message: fmt.Sprintf(
			"association duplicate link: association %q has %d links between instances %d and %d",
			params.AssociationName,
			params.ActualCount,
			params.FromInstanceID,
			params.ToInstanceID,
		),
		InstanceID: params.FromInstanceID,
	}
}

// NewLivenessClassNotInstantiatedViolation creates a violation for a class that was never instantiated.
func NewLivenessClassNotInstantiatedViolation(classKey identity.Key, className string) *ViolationError {
	return &ViolationError{
		Type:     ViolationTypeLivenessClassNotInstantiated,
		Message:  fmt.Sprintf("liveness: class %s was never instantiated during simulation", className),
		ClassKey: classKey,
	}
}

// NewLivenessAttributeNotWrittenViolation creates a violation for an attribute that was never written.
func NewLivenessAttributeNotWrittenViolation(classKey identity.Key, className, attributeName string) *ViolationError {
	return &ViolationError{
		Type:          ViolationTypeLivenessAttributeNotWritten,
		Message:       fmt.Sprintf("liveness: attribute %s on class %s was never written during simulation", attributeName, className),
		ClassKey:      classKey,
		AttributeName: attributeName,
	}
}

// NewLivenessAssociationNotLinkedViolation creates a violation for an association that was never linked.
func NewLivenessAssociationNotLinkedViolation(_ identity.Key, associationName string, fromClassKey, toClassKey identity.Key) *ViolationError {
	return &ViolationError{
		Type:    ViolationTypeLivenessAssociationNotLinked,
		Message: fmt.Sprintf("liveness: association %s (between %s and %s) never had a link created during simulation", associationName, fromClassKey.String(), toClassKey.String()),
	}
}

// NewLivenessEventNotSentViolation creates a violation for an event that was never fired.
func NewLivenessEventNotSentViolation(classKey identity.Key, className, eventName string) *ViolationError {
	return &ViolationError{
		Type:     ViolationTypeLivenessEventNotSent,
		Message:  fmt.Sprintf("liveness: event %s on class %s was never sent during simulation", eventName, className),
		ClassKey: classKey,
	}
}

// NewLivenessAttributeNotReadViolation creates a violation for an external derived attribute that was never read.
func NewLivenessAttributeNotReadViolation(classKey identity.Key, className, attributeName string) *ViolationError {
	return &ViolationError{
		Type:              ViolationTypeLivenessAttributeNotRead,
		Message:           fmt.Sprintf("liveness: derived attribute %s on class %s was never read during simulation", attributeName, className),
		ClassKey:          classKey,
		ActionOrQueryName: attributeName,
	}
}

// NewLivenessQueryNotRunViolation creates a violation for a query that was never executed.
func NewLivenessQueryNotRunViolation(classKey identity.Key, className, queryName string) *ViolationError {
	return &ViolationError{
		Type:              ViolationTypeLivenessQueryNotRun,
		Message:           fmt.Sprintf("liveness: query %s on class %s was never run during simulation", queryName, className),
		ClassKey:          classKey,
		ActionOrQueryName: queryName,
	}
}

// NewLivenessActionNotExecutedViolation creates a violation for an action that was never executed.
func NewLivenessActionNotExecutedViolation(classKey identity.Key, className, actionName string) *ViolationError {
	return &ViolationError{
		Type:              ViolationTypeLivenessActionNotExecuted,
		Message:           fmt.Sprintf("liveness: action %s on class %s was never executed during simulation", actionName, className),
		ClassKey:          classKey,
		ActionOrQueryName: actionName,
	}
}

// PeerEventUnavailableParams holds parameters for a peer event unavailable violation.
type PeerEventUnavailableParams struct {
	OwnerClassKey   identity.Key
	OwnerInstanceID state.InstanceID
	AssociationName string
	PeerClassKey    identity.Key
	PeerInstanceID  state.InstanceID
	EventKey        identity.Key
	EventName       string
	Message         string
}

// NewPeerEventUnavailableViolation creates a violation when an association guarantee
// sends an event the peer class cannot accept.
func NewPeerEventUnavailableViolation(params PeerEventUnavailableParams) *ViolationError {
	return &ViolationError{
		Type:              ViolationTypePeerEventUnavailable,
		Message:           params.Message,
		InstanceID:        params.OwnerInstanceID,
		ClassKey:          params.OwnerClassKey,
		ActionOrQueryKey:  params.EventKey,
		ActionOrQueryName: params.EventName,
		AttributeName:     params.AssociationName,
		ExpectedValue:     params.PeerClassKey.String(),
		ActualValue:       fmt.Sprintf("%d", params.PeerInstanceID),
	}
}

// NewStateMachineIncompleteViolation creates a violation for a class state machine that omits _new.
func NewStateMachineIncompleteViolation(classKey identity.Key, className string) *ViolationError {
	return &ViolationError{
		Type:     ViolationTypeStateMachineIncomplete,
		Message:  fmt.Sprintf("state machine incomplete: class %s has no «new» event for creation transitions", className),
		ClassKey: classKey,
	}
}

// ViolationErrors is a collection of violations.
type ViolationErrors []*ViolationError

// HasViolations returns true if there are any violations.
func (v ViolationErrors) HasViolations() bool {
	return len(v) > 0
}

// ByType filters violations by type.
func (v ViolationErrors) ByType(t ViolationType) ViolationErrors {
	var result ViolationErrors
	for _, violation := range v {
		if violation.Type == t {
			result = append(result, violation)
		}
	}
	return result
}

// TLAViolations returns all TLA+ related violations (model invariants and guarantees).
func (v ViolationErrors) TLAViolations() ViolationErrors {
	var result ViolationErrors
	for _, violation := range v {
		//nolint:exhaustive // Only TLA+ violation types are relevant here.
		switch violation.Type {
		case ViolationTypeModelInvariant, ViolationTypeClassInvariant, ViolationTypeActionRequires, ViolationTypeActionGuarantee, ViolationTypeQueryGuarantee, ViolationTypeAttributeInvariant, ViolationTypeParameterInvariant:
			result = append(result, violation)
		default:
			// Not a TLA+ violation; skip.
		}
	}
	return result
}

// DataTypeViolations returns all data type constraint violations.
func (v ViolationErrors) DataTypeViolations() ViolationErrors {
	var result ViolationErrors
	for _, violation := range v {
		//nolint:exhaustive // Only data type violation types are relevant here.
		switch violation.Type {
		case ViolationTypeRequiredAttribute, ViolationTypeSpanConstraint, ViolationTypeEnumConstraint, ViolationTypeCollectionSize, ViolationTypeIndexUniqueness, ViolationTypeUnparsedDataType, ViolationTypeMissingAttributeTypeSpec, ViolationTypeMissingParameterTypeSpec, ViolationTypeDateTimeTypeSpecMismatch, ViolationTypeDateTimeConstraint:
			result = append(result, violation)
		default:
			// Not a data type violation; skip.
		}
	}
	return result
}

// LivenessViolations returns all liveness check violations.
func (v ViolationErrors) LivenessViolations() ViolationErrors {
	var result ViolationErrors
	for _, violation := range v {
		//nolint:exhaustive // Only liveness violation types are relevant here.
		switch violation.Type {
		case ViolationTypeLivenessClassNotInstantiated,
			ViolationTypeLivenessAttributeNotWritten,
			ViolationTypeLivenessAssociationNotLinked,
			ViolationTypeLivenessAttributeNotRead,
			ViolationTypeLivenessEventNotSent,
			ViolationTypeLivenessQueryNotRun,
			ViolationTypeLivenessActionNotExecuted:
			result = append(result, violation)
		default:
			// Not a liveness violation; skip.
		}
	}
	return result
}

// Error returns a combined error message for all violations.
func (v ViolationErrors) Error() string {
	if len(v) == 0 {
		return ""
	}

	var messages []string
	for _, violation := range v {
		messages = append(messages, violation.Message)
	}
	return strings.Join(messages, "\n")
}
