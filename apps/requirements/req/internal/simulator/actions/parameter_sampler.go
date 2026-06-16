package actions

import (
	"errors"
	"math/rand"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// ParameterSampler generates event parameters that satisfy action requires when possible.
// Unmentioned parameters fall back to type-only random generation via ParameterBinder.
type ParameterSampler struct {
	binder         *ParameterBinder
	namedSetValues map[string]object.Object
}

// NewParameterSampler creates a sampler backed by the given binder and model named sets.
func NewParameterSampler(binder *ParameterBinder, namedSetValues map[string]object.Object) *ParameterSampler {
	return &ParameterSampler{
		binder:         binder,
		namedSetValues: namedSetValues,
	}
}

// SampleFromRequires builds parameters for paramDefs using constraints extracted from action requires.
func (s *ParameterSampler) SampleFromRequires(
	paramDefs []model_state.Parameter,
	action *model_state.Action,
	rng *rand.Rand,
) (map[string]object.Object, error) {
	if action == nil || len(action.Requires) == 0 {
		return s.binder.GenerateRandomParameters(paramDefs, rng), nil
	}

	paramNames := parameterNames(paramDefs)
	if err := ValidateRequiresSamplingSupport(action.Requires, paramNames); err != nil {
		var unsupported *UnsupportedRequiresSamplingError
		if errors.As(err, &unsupported) && action != nil {
			unsupported.ActionName = action.Name
		}
		return nil, err
	}

	constraints := extractParameterConstraints(action.Requires)
	result := s.binder.GenerateRandomParameters(paramDefs, rng)
	applyParameterConstraints(result, constraints, rng, s.namedSetValues)
	return result, nil
}

// parameterConstraints captures sampling hints extracted from require expression trees.
type parameterConstraints struct {
	nullableElseTuple *nullableElseTupleConstraint
	tupleInSet        *tupleInSetConstraint
	enumValues        map[string][]string
}

type nullableElseTupleConstraint struct {
	conditionParam string
	thenParam      string
	paramNames     []string
	setSubKey      string
}

type tupleInSetConstraint struct {
	paramNames []string
	setSubKey  string
}

func applyParameterConstraints(
	result map[string]object.Object,
	constraints parameterConstraints,
	rng *rand.Rand,
	namedSetValues map[string]object.Object,
) {
	if constraints.nullableElseTuple != nil {
		applyNullableElseTuple(result, constraints.nullableElseTuple, rng, namedSetValues)
	} else if constraints.tupleInSet != nil {
		applyTupleInSet(result, constraints.tupleInSet, rng, namedSetValues)
	}

	for paramName, values := range constraints.enumValues {
		if len(values) == 0 {
			continue
		}
		result[paramName] = object.NewString(values[rng.Intn(len(values))])
	}
}

func applyNullableElseTuple(
	result map[string]object.Object,
	constraint *nullableElseTupleConstraint,
	rng *rand.Rand,
	namedSetValues map[string]object.Object,
) {
	if rng.Intn(5) == 0 {
		for _, paramName := range constraint.paramNames {
			result[paramName] = evaluator.EMPTY_SET
		}
		return
	}

	tuple, ok := pickRandomTuple(constraint.setSubKey, namedSetValues, rng)
	if !ok {
		return
	}

	assignTupleParams(result, constraint.paramNames, tuple)
}

func applyTupleInSet(
	result map[string]object.Object,
	constraint *tupleInSetConstraint,
	rng *rand.Rand,
	namedSetValues map[string]object.Object,
) {
	tuple, ok := pickRandomTuple(constraint.setSubKey, namedSetValues, rng)
	if !ok {
		return
	}

	assignTupleParams(result, constraint.paramNames, tuple)
}

func assignTupleParams(result map[string]object.Object, paramNames []string, tuple *object.Tuple) {
	for i, paramName := range paramNames {
		result[paramName] = normalizeTupleElement(tuple.At(i + 1))
	}
}

func normalizeTupleElement(value object.Object) object.Object {
	if value == nil {
		return evaluator.EMPTY_SET
	}
	if str, ok := value.(*object.String); ok && str.Value() == "" {
		return evaluator.EMPTY_SET
	}
	return value.Clone()
}

func pickRandomTuple(
	setSubKey string,
	namedSetValues map[string]object.Object,
	rng *rand.Rand,
) (*object.Tuple, bool) {
	setObj, ok := namedSetValues[setSubKey]
	if !ok {
		return nil, false
	}

	set, ok := setObj.(*object.Set)
	if !ok || set.Size() == 0 {
		return nil, false
	}

	elements := set.Elements()
	chosen, ok := elements[rng.Intn(len(elements))].(*object.Tuple)
	if !ok {
		return nil, false
	}
	return chosen, true
}
