package invariants

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"
)

// CheckParameterTypeSpecs reports parameters on an action or query that lack TLA+ type_spec
// when the simulator is about to use them.
func CheckParameterTypeSpecs(
	params []model_state.Parameter,
	sourceKey identity.Key,
	sourceName string,
	sourceKind string,
	instanceID instance.ID,
	classKey identity.Key,
) ViolationErrors {
	var violations ViolationErrors
	for _, param := range params {
		if param.DataType == nil {
			violations = append(violations, NewUnparsedParameterDataTypeViolation(
				ViolationSourceIdentity{Key: sourceKey, Name: sourceName},
				sourceKind, param.Name, param.DataTypeRules, instanceID, classKey,
			))
			continue
		}
		if !dataTypeHasTypeSpec(param.DataType) {
			violations = append(violations, NewMissingParameterTypeSpecViolation(
				sourceKey, sourceName, sourceKind, param.Name, instanceID, classKey,
			))
			continue
		}
		if model_data_type.IsAtomicDateTime(param.DataType) && !dateTimeTypeSpecSatisfied(param.DataType) {
			violations = append(violations, NewDateTimeTypeSpecMismatchParameterViolation(DateTimeTypeSpecMismatchParameterParams{
				Source:         ViolationSourceIdentity{Key: sourceKey, Name: sourceName},
				SourceKind:     sourceKind,
				ParameterName:  param.Name,
				ActualTypeSpec: typeSpecSpecification(param.DataType),
				InstanceID:     instanceID,
				ClassKey:       classKey,
			}))
		}
	}
	return violations
}
