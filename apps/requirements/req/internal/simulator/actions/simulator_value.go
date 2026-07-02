package actions

import (
	"math/rand"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

func valueForEnumerationLiteral(dataType *model_data_type.DataType, literal string) object.Object {
	if model_data_type.HasBooleanTypeSpec(dataType) {
		if value, ok := model_data_type.BooleanFromEnumerationLiteral(literal); ok {
			return object.NewBoolean(value)
		}
	}
	return object.NewString(literal)
}

func randomEnumerationValue(dataType *model_data_type.DataType, values []string, rng *rand.Rand) object.Object {
	if len(values) == 0 {
		return object.NewString("")
	}
	return valueForEnumerationLiteral(dataType, values[rng.Intn(len(values))])
}

// CoerceValueForDataType normalizes sampled or assigned values to match type_spec storage.
func CoerceValueForDataType(dataType *model_data_type.DataType, value object.Object) object.Object {
	if !model_data_type.HasBooleanTypeSpec(dataType) {
		return value
	}
	if _, ok := value.(*object.Boolean); ok {
		return value
	}
	if str, ok := value.(*object.String); ok {
		if boolValue, ok := model_data_type.BooleanFromEnumerationLiteral(str.Value()); ok {
			return object.NewBoolean(boolValue)
		}
	}
	return value
}

func coerceSampledParameters(
	paramDefs []model_state.Parameter,
	values map[string]object.Object,
) {
	for _, param := range paramDefs {
		value, ok := values[param.Name]
		if !ok {
			continue
		}
		values[param.Name] = CoerceValueForDataType(param.DataType, value)
	}
}
