package actions

import (
	"fmt"
	"math/rand"

	"github.com/glemzurg/go-tlaplus/internal/req_model/model_data_type"
	"github.com/glemzurg/go-tlaplus/internal/req_model/model_state"
	"github.com/glemzurg/go-tlaplus/internal/simulator/object"
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
		result[paramDef.Name] = value
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
		result[paramDef.Name] = generateRandomValue(paramDef.DataType, rng)
	}

	return result
}

// generateRandomValue creates a random value based on data type constraints.
func generateRandomValue(dataType *model_data_type.DataType, rng *rand.Rand) object.Object {
	if dataType == nil || dataType.Atomic == nil {
		// No type info — generate a default integer in [0, 100).
		return object.NewNatural(rng.Int63n(100))
	}

	atomic := dataType.Atomic

	switch atomic.ConstraintType {
	case model_data_type.ConstraintTypeSpan:
		return randomNumberInSpan(atomic.Span, rng)

	case model_data_type.ConstraintTypeEnumeration:
		if len(atomic.Enums) == 0 {
			return object.NewString("")
		}
		idx := rng.Intn(len(atomic.Enums))
		return object.NewString(atomic.Enums[idx].Value)

	case model_data_type.ConstraintTypeUnconstrained:
		return object.NewNatural(rng.Int63n(100))

	default:
		return object.NewNatural(rng.Int63n(100))
	}
}

// randomNumberInSpan generates a random integer within a span's bounds.
func randomNumberInSpan(span *model_data_type.AtomicSpan, rng *rand.Rand) object.Object {
	if span == nil {
		return object.NewNatural(rng.Int63n(100))
	}

	// Determine effective lower bound
	lower := int64(0)
	if span.LowerValue != nil {
		lower = int64(*span.LowerValue)
		if span.LowerType == "open" {
			lower++ // Exclude lower bound
		}
	}

	// Determine effective upper bound
	upper := lower + 100 // Default range if unconstrained
	if span.HigherValue != nil {
		upper = int64(*span.HigherValue)
		if span.HigherType == "open" {
			upper-- // Exclude upper bound
		}
	}

	if upper < lower {
		return object.NewInteger(lower)
	}

	// Generate random value in [lower, upper]
	rangeSize := upper - lower + 1
	value := lower + rng.Int63n(rangeSize)
	return object.NewInteger(value)
}
