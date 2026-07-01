package actions

import (
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// nullableNullSampleDenom is the denominator for sampling NULL on nullable parameters (1/denom chance).
const nullableNullSampleDenom = 10

const (
	spanBoundUnconstrained = "unconstrained"
	spanBoundOpen          = "open"
	// spanDefaultHalfWidth scales unconstrained span ends: ±(halfWidth × precision).
	spanDefaultHalfWidth = 100
)

// ParameterBinder validates and generates parameter values for actions and queries.
type ParameterBinder struct{}

// NewParameterBinder creates a new parameter binder.
func NewParameterBinder() *ParameterBinder {
	return &ParameterBinder{}
}

// BindParameters validates that all required parameters are provided and returns
// a map suitable for use as evaluator bindings.
func (b *ParameterBinder) BindParameters(
	paramDefs []model_state.Parameter,
	paramValues map[string]object.Object,
) (map[string]object.Object, error) {
	result := make(map[string]object.Object)

	for _, paramDef := range paramDefs {
		value, exists := paramValues[paramDef.Name]
		if !exists {
			return nil, fmt.Errorf("missing required parameter: %s", paramDef.Name)
		}
		result[paramDef.Name] = CoerceValueForDataType(paramDef.DataType, value)
	}

	return result, nil
}

// GenerateRandomParameters creates random parameter values that satisfy data type constraints.
// Used by the simulation engine to drive random exploration.
func (b *ParameterBinder) GenerateRandomParameters(
	paramDefs []model_state.Parameter,
	rng *rand.Rand,
) map[string]object.Object {
	result := make(map[string]object.Object)

	for _, paramDef := range paramDefs {
		result[paramDef.Name] = sampleParameterValue(paramDef, rng)
	}

	coerceSampledParameters(paramDefs, result)
	return result
}

// sampleParameterValue generates a random value for one action/query parameter.
// Nullable parameters may be NULL; non-nullable parameters never are.
func sampleParameterValue(param model_state.Parameter, rng *rand.Rand) object.Object {
	if param.Nullable && rng.Intn(nullableNullSampleDenom) == 0 {
		return evaluator.EMPTY_SET
	}
	return generateRandomValue(param.DataType, rng)
}

// generateRandomValue creates a random non-null value based on data type constraints.
func generateRandomValue(dataType *model_data_type.DataType, rng *rand.Rand) object.Object {
	if values := model_data_type.EnumerationValues(dataType); len(values) > 0 {
		return randomEnumerationValue(dataType, values, rng)
	}

	if dataType != nil && dataType.TypeSpec != nil {
		switch strings.ToUpper(strings.TrimSpace(dataType.TypeSpec.Specification)) {
		case "STRING":
			return randomString(rng)
		case "BOOLEAN":
			if rng.Intn(2) == 0 {
				return object.NewBoolean(false)
			}
			return object.NewBoolean(true)
		}
	}

	if dataType == nil || dataType.Atomic == nil {
		// No type info — generate a default integer in [0, 99].
		return randomDefaultNumber(rng)
	}

	atomic := dataType.Atomic

	switch atomic.ConstraintType {
	case model_data_type.CONSTRAINT_TYPE_SPAN:
		return randomNumberInSpan(atomic.Span, rng)

	case model_data_type.CONSTRAINT_TYPE_ENUMERATION:
		if len(atomic.Enums) == 0 {
			return evaluator.EMPTY_SET
		}
		values := make([]string, len(atomic.Enums))
		for i, enum := range atomic.Enums {
			values[i] = enum.Value
		}
		return randomEnumerationValue(dataType, values, rng)

	case model_data_type.CONSTRAINT_TYPE_UNCONSTRAINED:
		return randomString(rng)

	case model_data_type.CONSTRAINT_TYPE_REFERENCE, model_data_type.CONSTRAINT_TYPE_OBJECT:
		return randomString(rng)

	default:
		return randomDefaultNumber(rng)
	}
}

func randomDefaultNumber(rng *rand.Rand) object.Object {
	return object.NewNatural(rng.Int63n(100))
}

func randomString(rng *rand.Rand) object.Object {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	length := 1 + rng.Intn(8)
	var b strings.Builder
	for range length {
		b.WriteByte(letters[rng.Intn(len(letters))])
	}
	return object.NewString(b.String())
}

// randomNumberInSpan generates a random Real on a span's precision lattice.
func randomNumberInSpan(span *model_data_type.AtomicSpan, rng *rand.Rand) object.Object {
	step := spanPrecision(span)
	if span == nil {
		extent := spanDefaultHalfWidth * step
		return randomRealOnPrecisionGrid(-extent, extent, step, rng)
	}

	lower, upper := spanSamplingInterval(span)
	if upper <= lower {
		return object.NewFloat(lower)
	}

	return randomRealOnPrecisionGrid(lower, upper, step, rng)
}

func spanPrecision(span *model_data_type.AtomicSpan) float64 {
	if span == nil || span.Precision <= 0 {
		return 1
	}
	return span.Precision
}

func randomRealOnPrecisionGrid(lower, upper, step float64, rng *rand.Rand) object.Object {
	if step <= 0 {
		return object.NewFloat(lower)
	}

	lowerSteps := int(math.Round(lower / step))
	upperSteps := int(math.Round(upper / step))
	if upperSteps < lowerSteps {
		return object.NewFloat(lower)
	}

	stepIndex := lowerSteps + rng.Intn(upperSteps-lowerSteps+1)
	return object.NewFloat(float64(stepIndex) * step)
}

func spanSamplingInterval(span *model_data_type.AtomicSpan) (lower, upper float64) {
	step := spanPrecision(span)
	defaultExtent := spanDefaultHalfWidth * step

	lower = -defaultExtent
	if span.LowerType != spanBoundUnconstrained && span.LowerValue != nil {
		lower, _ = spanValueToRat(span.LowerValue, span.LowerDenominator).Float64()
		if span.LowerType == spanBoundOpen {
			lower += step
		}
	}

	upper = defaultExtent
	if span.HigherType != spanBoundUnconstrained && span.HigherValue != nil {
		upper, _ = spanValueToRat(span.HigherValue, span.HigherDenominator).Float64()
		if span.HigherType == spanBoundOpen {
			upper -= step
		}
	}

	return lower, upper
}

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
