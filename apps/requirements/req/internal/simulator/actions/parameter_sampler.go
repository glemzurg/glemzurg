package actions

import (
	"math/rand"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

const maxNotInNamedSetAttempts = 10

// ParameterSampler generates event parameters that satisfy action requires when possible.
// Unmentioned parameters fall back to type-only random generation via ParameterBinder.
type ParameterSampler struct {
	binder         *ParameterBinder
	namedSetValues map[string]object.Object
	// peerFieldDistinctLookup returns field values already used by class instances.
	peerFieldDistinctLookup func(classKey identity.Key, fieldSubKey string) []object.Object
	// generateOverride is set only by tests to force deterministic parameter draws.
	generateOverride func(paramDefs []model_state.Parameter, rng *rand.Rand) map[string]object.Object
}

// NewParameterSampler creates a sampler backed by the given binder and model named sets.
func NewParameterSampler(binder *ParameterBinder, namedSetValues map[string]object.Object) *ParameterSampler {
	return &ParameterSampler{
		binder:         binder,
		namedSetValues: namedSetValues,
	}
}

// SetPeerFieldDistinctLookup configures lookup of field values already used by class instances.
func (s *ParameterSampler) SetPeerFieldDistinctLookup(
	lookup func(classKey identity.Key, fieldSubKey string) []object.Object,
) {
	s.peerFieldDistinctLookup = lookup
}

// SampleFromRequires builds parameters for paramDefs using the action's effective requires.
func (s *ParameterSampler) SampleFromRequires(
	paramDefs []model_state.Parameter,
	action *model_state.Action,
	rng *rand.Rand,
) (map[string]object.Object, error) {
	if action == nil {
		return s.binder.GenerateRandomParameters(paramDefs, rng), nil
	}
	return s.SampleParameters(ParameterOwnerFromAction(*action), paramDefs, rng)
}

// SampleQueryFromRequires builds parameters for a query using its effective requires.
func (s *ParameterSampler) SampleQueryFromRequires(
	query model_state.Query,
	rng *rand.Rand,
) (map[string]object.Object, error) {
	return s.SampleParameters(ParameterOwnerFromQuery(query), query.Parameters, rng)
}

// parameterConstraints captures sampling hints extracted from require expression trees.
type parameterConstraints struct {
	nullableElseTuple             *nullableElseTupleConstraint
	nullableElseMirror            *nullableElseMirrorConstraint
	nullableElseExclusionEquality *nullableElseExclusionEqualityConstraint
	nullableElseMembership        *nullableElseMembershipConstraint
	nullableElseEquality          *nullableElseEqualityConstraint
	tupleInSet                    *tupleInSetConstraint
	peerFieldDistinct             *peerFieldDistinctFromParamPattern
	enumValues                    map[string][]string
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

// nullableElseMirrorConstraint couples a nullable driver to a follower when the driver is set:
// IF driver = NULL THEN TRUE ELSE (driver ∈ namedSet ∧ driver = follower).
type nullableElseMirrorConstraint struct {
	driverParam   string
	followerParam string
	setSubKey     string
}

// nullableElseEqualityConstraint copies the driver onto the follower when the driver is set:
// IF driver = NULL THEN TRUE ELSE driver = follower.
type nullableElseEqualityConstraint struct {
	driverParam   string
	followerParam string
}

// nullableElseExclusionEqualityConstraint couples a nullable driver to a follower across branches:
// IF driver = NULL THEN follower ∉ namedSet ELSE driver = follower (driver sampled from namedSet).
type nullableElseExclusionEqualityConstraint struct {
	driverParam   string
	followerParam string
	setSubKey     string
}

// nullableExclusionWithPeerDistinct couples exclusion-equality sampling with peer-field distinctness on the follower.
type nullableExclusionWithPeerDistinct struct {
	exclusion *nullableElseExclusionEqualityConstraint
	peer      *peerFieldDistinctFromParamPattern
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
	peerFieldDistinctLookup func(classKey identity.Key, fieldSubKey string) []object.Object,
) {
	if constraints.nullableElseExclusionEquality != nil &&
		constraints.peerFieldDistinct != nil &&
		constraints.peerFieldDistinct.ParamName == constraints.nullableElseExclusionEquality.followerParam {
		applyNullableElseExclusionWithPeerDistinct(
			result,
			nullableExclusionWithPeerDistinct{
				exclusion: constraints.nullableElseExclusionEquality,
				peer:      constraints.peerFieldDistinct,
			},
			rng,
			namedSetValues,
			nullableByName,
			peerFieldDistinctLookup,
		)
	} else {
		switch {
		case constraints.nullableElseExclusionEquality != nil:
			applyNullableElseExclusionEquality(result, constraints.nullableElseExclusionEquality, rng, namedSetValues, nullableByName)
		case constraints.nullableElseTuple != nil:
			applyNullableElseTuple(result, constraints.nullableElseTuple, rng, namedSetValues, nullableByName)
		case constraints.nullableElseMirror != nil:
			applyNullableElseMirror(result, constraints.nullableElseMirror, rng, namedSetValues, nullableByName)
		case constraints.nullableElseMembership != nil:
			applyNullableElseMembership(result, constraints.nullableElseMembership, rng, namedSetValues, nullableByName)
		}
	}

	integratedPeerIsoAbbr := constraints.nullableElseExclusionEquality != nil &&
		constraints.peerFieldDistinct != nil &&
		constraints.peerFieldDistinct.ParamName == constraints.nullableElseExclusionEquality.followerParam

	if constraints.nullableElseEquality != nil &&
		constraints.nullableElseMirror == nil &&
		constraints.nullableElseExclusionEquality == nil &&
		!integratedPeerIsoAbbr {
		applyNullableElseEquality(result, constraints.nullableElseEquality, nullableByName)
	}

	if constraints.tupleInSet != nil {
		applyTupleInSet(result, constraints.tupleInSet, rng, namedSetValues)
	}

	for paramName, values := range constraints.enumValues {
		if len(values) == 0 {
			continue
		}
		result[paramName] = object.NewString(values[rng.Intn(len(values))])
	}

	if constraints.peerFieldDistinct != nil &&
		(constraints.nullableElseExclusionEquality == nil ||
			constraints.peerFieldDistinct.ParamName != constraints.nullableElseExclusionEquality.followerParam) {
		applyPeerFieldDistinct(result, constraints.peerFieldDistinct, rng, peerFieldDistinctLookup)
	}
}

// applyNullableElseExclusionWithPeerDistinct samples ISO/Abbr coupling and peer-uniqueness together
// so Abbr is never overwritten after ISO is set.
func applyNullableElseExclusionWithPeerDistinct(
	result map[string]object.Object,
	coupled nullableExclusionWithPeerDistinct,
	rng *rand.Rand,
	namedSetValues map[string]object.Object,
	nullableByName map[string]bool,
	lookup func(classKey identity.Key, fieldSubKey string) []object.Object,
) {
	exclusion := coupled.exclusion
	usedAbbrs := peerUsedStringSet(coupled.peer, lookup)

	if nullableByName[exclusion.driverParam] && rng.Intn(5) == 0 {
		result[exclusion.driverParam] = evaluator.EMPTY_SET
		if value, ok := pickRandomStringNotInNamedSetExcluding(
			exclusion.setSubKey, namedSetValues, usedAbbrs, rng,
		); ok {
			result[exclusion.followerParam] = value
		}
		return
	}

	value, ok := pickRandomStringFromNamedSetExcluding(exclusion.setSubKey, namedSetValues, usedAbbrs, rng)
	if !ok {
		return
	}
	result[exclusion.driverParam] = value
	result[exclusion.followerParam] = value.Clone()
}

func peerUsedStringSet(
	peer *peerFieldDistinctFromParamPattern,
	lookup func(classKey identity.Key, fieldSubKey string) []object.Object,
) map[string]bool {
	used := make(map[string]bool)
	if lookup == nil {
		return used
	}
	for _, val := range lookup(peer.ClassKey, peer.FieldSubKey) {
		if str, ok := val.(*object.String); ok {
			used[str.Value()] = true
		}
	}
	return used
}

func applyPeerFieldDistinct(
	result map[string]object.Object,
	pattern *peerFieldDistinctFromParamPattern,
	rng *rand.Rand,
	lookup func(classKey identity.Key, fieldSubKey string) []object.Object,
) {
	if lookup == nil {
		return
	}
	used := lookup(pattern.ClassKey, pattern.FieldSubKey)
	if len(used) == 0 {
		return
	}
	usedSet := make(map[string]bool, len(used))
	for _, val := range used {
		if str, ok := val.(*object.String); ok {
			usedSet[str.Value()] = true
		}
	}
	for range maxNotInNamedSetAttempts {
		candidate := randomString(rng)
		if str, ok := candidate.(*object.String); ok && !usedSet[str.Value()] {
			result[pattern.ParamName] = candidate
			return
		}
	}
}

func applyNullableElseExclusionEquality(
	result map[string]object.Object,
	constraint *nullableElseExclusionEqualityConstraint,
	rng *rand.Rand,
	namedSetValues map[string]object.Object,
	nullableByName map[string]bool,
) {
	if nullableByName[constraint.driverParam] && rng.Intn(5) == 0 {
		result[constraint.driverParam] = evaluator.EMPTY_SET
		value, ok := pickRandomStringNotInNamedSet(constraint.setSubKey, namedSetValues, rng)
		if ok {
			result[constraint.followerParam] = value
		}
		return
	}

	value, ok := pickRandomStringFromNamedSet(constraint.setSubKey, namedSetValues, rng)
	if !ok {
		return
	}

	result[constraint.driverParam] = value
	result[constraint.followerParam] = value.Clone()
}

func applyNullableElseMirror(
	result map[string]object.Object,
	constraint *nullableElseMirrorConstraint,
	rng *rand.Rand,
	namedSetValues map[string]object.Object,
	nullableByName map[string]bool,
) {
	if nullableByName[constraint.driverParam] && rng.Intn(5) == 0 {
		result[constraint.driverParam] = evaluator.EMPTY_SET
		return
	}

	value, ok := pickRandomStringFromNamedSet(constraint.setSubKey, namedSetValues, rng)
	if !ok {
		return
	}

	result[constraint.driverParam] = value
	result[constraint.followerParam] = value.Clone()
}

func applyNullableElseEquality(
	result map[string]object.Object,
	constraint *nullableElseEqualityConstraint,
	nullableByName map[string]bool,
) {
	if nullableByName[constraint.driverParam] && object.IsNull(result[constraint.driverParam]) {
		return
	}
	if val, ok := result[constraint.driverParam]; ok && !object.IsNull(val) {
		result[constraint.followerParam] = val.Clone()
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

func pickRandomStringNotInNamedSet(
	setSubKey string,
	namedSetValues map[string]object.Object,
	rng *rand.Rand,
) (object.Object, bool) {
	return pickRandomStringNotInNamedSetExcluding(setSubKey, namedSetValues, nil, rng)
}

func pickRandomStringNotInNamedSetExcluding(
	setSubKey string,
	namedSetValues map[string]object.Object,
	excluded map[string]bool,
	rng *rand.Rand,
) (object.Object, bool) {
	setObj, ok := namedSetValues[setSubKey]
	if !ok {
		return nil, false
	}

	set, ok := setObj.(*object.Set)
	if !ok {
		return nil, false
	}

	for range maxNotInNamedSetAttempts {
		candidate := randomString(rng)
		if str, ok := candidate.(*object.String); ok && excluded != nil && excluded[str.Value()] {
			continue
		}
		if !set.Contains(candidate) {
			return candidate, true
		}
	}
	return nil, false
}

func pickRandomStringFromNamedSet(
	setSubKey string,
	namedSetValues map[string]object.Object,
	rng *rand.Rand,
) (object.Object, bool) {
	return pickRandomStringFromNamedSetExcluding(setSubKey, namedSetValues, nil, rng)
}

func pickRandomStringFromNamedSetExcluding(
	setSubKey string,
	namedSetValues map[string]object.Object,
	excluded map[string]bool,
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

	available := make([]object.Object, 0, set.Size())
	for _, elem := range set.Elements() {
		if str, ok := elem.(*object.String); ok && excluded != nil && excluded[str.Value()] {
			continue
		}
		available = append(available, elem)
	}
	if len(available) == 0 {
		return nil, false
	}
	return available[rng.Intn(len(available))].Clone(), true
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
