package actions

import (
	"math/rand"
	"regexp"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
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
) map[string]object.Object {
	if action == nil || len(action.Requires) == 0 {
		return s.binder.GenerateRandomParameters(paramDefs, rng)
	}

	constraints := extractParameterConstraints(action.Requires)
	result := s.binder.GenerateRandomParameters(paramDefs, rng)
	applyParameterConstraints(result, constraints, rng, s.namedSetValues)
	return result
}

// parameterConstraints captures sampling hints extracted from require specifications.
type parameterConstraints struct {
	nullableElseTuple *nullableElseTupleConstraint
	tupleInSet        *tupleInSetConstraint
	enumValues        map[string][]string
}

type nullableElseTupleConstraint struct {
	conditionParam string
	thenParam      string
	paramNames     [2]string
	setName        string
}

type tupleInSetConstraint struct {
	paramNames [2]string
	setName    string
}

var (
	// NULL may appear as NULL or {} in lowered specifications.
	reIfNullThenNullElseTuple = regexp.MustCompile(
		`(?i)IF\s+(\w+)\s*=\s*(?:NULL|\{\})\s+THEN\s+(\w+)\s*=\s*(?:NULL|\{\})\s+ELSE\s+(?:<<|⟨)\s*(\w+)\s*,\s*(\w+)\s*(?:>>|⟩)\s*(?:\\in|∈)\s+(_\w+)`,
	)
	reTupleInNamedSet = regexp.MustCompile(`(?:<<|⟨)\s*(\w+)\s*,\s*(\w+)\s*(?:>>|⟩)\s*(?:\\in|∈)\s+(_\w+)`)
	reParamInEnum     = regexp.MustCompile(`(\w+)\s*\\in\s*\{([^}]+)\}`)
)

func extractParameterConstraints(requires []model_logic.Logic) parameterConstraints {
	constraints := parameterConstraints{
		enumValues: make(map[string][]string),
	}

	for _, req := range requires {
		if req.Type != model_logic.LogicTypeAssessment {
			continue
		}
		spec := strings.TrimSpace(req.Spec.Specification)
		if spec == "" {
			continue
		}
		applyConstraintPatterns(spec, &constraints)
	}

	return constraints
}

func applyConstraintPatterns(spec string, constraints *parameterConstraints) {
	if constraints.nullableElseTuple == nil {
		if m := reIfNullThenNullElseTuple.FindStringSubmatch(spec); len(m) == 6 {
			constraints.nullableElseTuple = &nullableElseTupleConstraint{
				conditionParam: m[1],
				thenParam:      m[2],
				paramNames:     [2]string{m[3], m[4]},
				setName:        m[5],
			}
			return
		}
	}

	if constraints.tupleInSet == nil {
		if m := reTupleInNamedSet.FindStringSubmatch(spec); len(m) == 4 {
			constraints.tupleInSet = &tupleInSetConstraint{
				paramNames: [2]string{m[1], m[2]},
				setName:    m[3],
			}
			return
		}
	}

	if m := reParamInEnum.FindStringSubmatch(spec); len(m) == 3 {
		constraints.enumValues[m[1]] = parseEnumLiterals(m[2])
	}
}

func parseEnumLiterals(raw string) []string {
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		part = strings.Trim(part, `"`)
		if part != "" {
			values = append(values, part)
		}
	}
	return values
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
		result[constraint.paramNames[0]] = evaluator.EMPTY_SET
		result[constraint.paramNames[1]] = evaluator.EMPTY_SET
		return
	}

	tuple, ok := pickRandomTuple(constraint.setName, namedSetValues, rng)
	if !ok {
		return
	}

	result[constraint.paramNames[0]] = normalizeTupleElement(tuple.At(1))
	result[constraint.paramNames[1]] = normalizeTupleElement(tuple.At(2))
}

func applyTupleInSet(
	result map[string]object.Object,
	constraint *tupleInSetConstraint,
	rng *rand.Rand,
	namedSetValues map[string]object.Object,
) {
	tuple, ok := pickRandomTuple(constraint.setName, namedSetValues, rng)
	if !ok {
		return
	}

	result[constraint.paramNames[0]] = normalizeTupleElement(tuple.At(1))
	result[constraint.paramNames[1]] = normalizeTupleElement(tuple.At(2))
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
	setName string,
	namedSetValues map[string]object.Object,
	rng *rand.Rand,
) (*object.Tuple, bool) {
	setKey := strings.TrimPrefix(setName, "_")
	setObj, ok := namedSetValues[setKey]
	if !ok {
		setObj, ok = namedSetValues[strings.ToLower(setKey)]
	}
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
