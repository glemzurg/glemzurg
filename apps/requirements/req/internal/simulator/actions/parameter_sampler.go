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
	nullableByName := parameterNullableByName(paramDefs)
	applyParameterConstraints(result, constraints, rng, s.namedSetValues, nullableByName)
	enforceParameterNullability(result, paramDefs, rng)
	return result, nil
}

// parameterConstraints captures sampling hints extracted from require expression trees.
type parameterConstraints struct {
	nullableElseTuple      *nullableElseTupleConstraint
	nullableElseMembership *nullableElseMembershipConstraint
	tupleInSet             *tupleInSetConstraint
	enumValues             map[string][]string
}

type nullableElseMembershipConstraint struct {
	paramName string
	setSubKey string
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

func parameterNullableByName(paramDefs []model_state.Parameter) map[string]bool {
	nullable := make(map[string]bool, len(paramDefs))
	for _, param := range paramDefs {
		nullable[param.Name] = param.Nullable
	}
	return nullable
}

// enforceParameterNullability replaces NULL on non-nullable parameters after constraint application.
func enforceParameterNullability(
	result map[string]object.Object,
	paramDefs []model_state.Parameter,
	rng *rand.Rand,
) {
	for _, param := range paramDefs {
		if param.Nullable {
			continue
		}
		if object.IsNull(result[param.Name]) {
			result[param.Name] = generateRandomValue(param.DataType, rng)
		}
	}
}

func applyParameterConstraints(
	result map[string]object.Object,
	constraints parameterConstraints,
	rng *rand.Rand,
	namedSetValues map[string]object.Object,
	nullableByName map[string]bool,
) {
	switch {
	case constraints.nullableElseTuple != nil:
		applyNullableElseTuple(result, constraints.nullableElseTuple, rng, namedSetValues, nullableByName)
	case constraints.nullableElseMembership != nil:
		applyNullableElseMembership(result, constraints.nullableElseMembership, rng, namedSetValues, nullableByName)
	case constraints.tupleInSet != nil:
		applyTupleInSet(result, constraints.tupleInSet, rng, namedSetValues)
	}

	for paramName, values := range constraints.enumValues {
		if len(values) == 0 {
			continue
		}
		result[paramName] = object.NewString(values[rng.Intn(len(values))])
	}
}

func applyNullableElseMembership(
	result map[string]object.Object,
	constraint *nullableElseMembershipConstraint,
	rng *rand.Rand,
	namedSetValues map[string]object.Object,
	nullableByName map[string]bool,
) {
	if nullableByName[constraint.paramName] && rng.Intn(5) == 0 {
		result[constraint.paramName] = evaluator.EMPTY_SET
		return
	}

	value, ok := pickRandomStringFromNamedSet(constraint.setSubKey, namedSetValues, rng)
	if !ok {
		return
	}

	result[constraint.paramName] = value
}

func applyNullableElseTuple(
	result map[string]object.Object,
	constraint *nullableElseTupleConstraint,
	rng *rand.Rand,
	namedSetValues map[string]object.Object,
	nullableByName map[string]bool,
) {
	if nullableElseTupleMayBeNull(constraint.paramNames, nullableByName) && rng.Intn(5) == 0 {
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

func nullableElseTupleMayBeNull(paramNames []string, nullableByName map[string]bool) bool {
	for _, paramName := range paramNames {
		if !nullableByName[paramName] {
			return false
		}
	}
	return true
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

func pickRandomStringFromNamedSet(
	setSubKey string,
	namedSetValues map[string]object.Object,
	rng *rand.Rand,
) (object.Object, bool) {
	setObj, ok := namedSetValues[setSubKey]
	if !ok {
		return nil, false
	}

	set, ok := setObj.(*object.Set)
	if !ok || set.Size() == 0 {
		return nil, false
	}

	elements := set.Elements()
	chosen := elements[rng.Intn(len(elements))]
	return chosen.Clone(), true
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
