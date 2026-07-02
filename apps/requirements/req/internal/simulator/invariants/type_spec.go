package invariants

import (
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

func dataTypeHasTypeSpec(dataType *model_data_type.DataType) bool {
	if dataType == nil || dataType.TypeSpec == nil {
		return false
	}
	return strings.TrimSpace(dataType.TypeSpec.Specification) != ""
}

func attributeHasTypeSpec(attr *model_class.Attribute) bool {
	if attr == nil {
		return false
	}
	return dataTypeHasTypeSpec(attr.DataType)
}

func typeSpecSpecification(dataType *model_data_type.DataType) string {
	if dataType == nil || dataType.TypeSpec == nil {
		return ""
	}
	return strings.TrimSpace(dataType.TypeSpec.Specification)
}

func dateTimeTypeSpecSatisfied(dataType *model_data_type.DataType) bool {
	return model_data_type.HasNatTypeSpec(dataType)
}

func attributeDefinitionViolations(
	instanceID state.InstanceID,
	classKey identity.Key,
	attrDef *model_class.Attribute,
) ViolationErrors {
	if attrDef.DataType == nil {
		return ViolationErrors{NewUnparsedAttributeDataTypeViolation(
			instanceID, classKey, attrDef.Name, attrDef.DataTypeRules,
		)}
	}
	if !attributeHasTypeSpec(attrDef) {
		return ViolationErrors{NewMissingAttributeTypeSpecViolation(
			instanceID, classKey, attrDef.Name,
		)}
	}
	if model_data_type.IsAtomicDateTime(attrDef.DataType) && !dateTimeTypeSpecSatisfied(attrDef.DataType) {
		return ViolationErrors{NewDateTimeTypeSpecMismatchAttributeViolation(
			instanceID, classKey, attrDef.Name, typeSpecSpecification(attrDef.DataType),
		)}
	}
	return nil
}
