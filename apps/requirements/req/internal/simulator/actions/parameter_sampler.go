package actions

import (
	"math/rand"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

const maxNotInNamedSetAttempts = 10

// ParameterSampler generates event parameters that satisfy action requires when possible.
// Unmentioned parameters fall back to type-only random generation via ParameterBinder.
type ParameterSampler struct {
	binder         *ParameterBinder
	namedSetValues map[string]object.Object
	// peerFieldDistinctLookup returns field values already used by class instances.
	peerFieldDistinctLookup func(classKey identity.Key, fieldSubKey string) []object.Object
	// peerFieldDistinctExcludeInstanceID skips one instance during peer lookup (the update target).
	peerFieldDistinctExcludeInstanceID state.InstanceID
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

// SetPeerFieldDistinctExcludeInstanceID configures peer lookup to skip one instance,
// typically the instance being updated so its current field value stays available.
func (s *ParameterSampler) SetPeerFieldDistinctExcludeInstanceID(id state.InstanceID) {
	s.peerFieldDistinctExcludeInstanceID = id
}

// PeerFieldDistinctExcludeInstanceID returns the instance excluded from peer lookup.
func (s *ParameterSampler) PeerFieldDistinctExcludeInstanceID() state.InstanceID {
	return s.peerFieldDistinctExcludeInstanceID
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
	nullableElseBooleanConstant   *nullableElseBooleanConstantConstraint
	tupleInSet                    *tupleInSetConstraint
	// paramInNamedSet is Param ∈ NamedSet (required membership, no null branch).
	paramInNamedSet *paramInNamedSetConstraint
	// paramInNamedSetMinusPeerField is Param ∈ (NamedSet \ { v.field : v ∈ Class }).
	paramInNamedSetMinusPeerField *paramInNamedSetMinusPeerFieldConstraint
	peerFieldDistinct             *peerFieldDistinctFromParamPattern
	enumValues                    map[string][]string
	// Partials from complementary _GZ!WhenNull / WhenNotNull pairs before coupling.
	gzNullExclusion     *gzNullBranchExclusion
	gzNullTupleFollower *gzNullBranchTupleFollower
}

// gzNullBranchExclusion is WhenNull(driver, follower ∉ namedSet) awaiting WhenNotNull equality.
type gzNullBranchExclusion struct {
	driverParam   string
	followerParam string
	setSubKey     string
}

// gzNullBranchTupleFollower is WhenNull(driver, follower = NULL) awaiting tuple membership.
type gzNullBranchTupleFollower struct {
	driverParam string
	thenParam   string
}

type nullableElseMembershipConstraint struct {
	paramName string
	setSubKey string
}

// paramInNamedSetConstraint samples a parameter from a named set.
type paramInNamedSetConstraint struct {
	paramName string
	setSubKey string
}

// paramInNamedSetMinusPeerFieldConstraint samples from a named set minus field values
// already used by class instances (set-map over the class extent).
type paramInNamedSetMinusPeerFieldConstraint struct {
	paramName   string
	setSubKey   string
	classKey    identity.Key
	className   string
	fieldSubKey string
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

// nullableElseBooleanConstantConstraint fixes the follower when the nullable driver is null:
// IF driver = NULL THEN follower = const ELSE TRUE.
type nullableElseBooleanConstantConstraint struct {
	driverParam   string
	followerParam string
	value         bool
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

// nullableMembershipWithPeerDistinct couples membership sampling with peer-field distinctness on the same parameter.
type nullableMembershipWithPeerDistinct struct {
	membership *nullableElseMembershipConstraint
	peer       *peerFieldDistinctFromParamPattern
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
	applyPrimaryNullableElseConstraints(
		result, constraints, rng, namedSetValues, nullableByName, peerFieldDistinctLookup,
	)
	applyFollowOnParameterConstraints(
		result, constraints, rng, namedSetValues, nullableByName, peerFieldDistinctLookup,
	)
}

func applyPrimaryNullableElseConstraints(
	result map[string]object.Object,
	constraints parameterConstraints,
	rng *rand.Rand,
	namedSetValues map[string]object.Object,
	nullableByName map[string]bool,
	peerFieldDistinctLookup func(classKey identity.Key, fieldSubKey string) []object.Object,
) {
	if constraints.nullableElseMembership != nil &&
		constraints.peerFieldDistinct != nil &&
		constraints.peerFieldDistinct.ParamName == constraints.nullableElseMembership.paramName {
		applyNullableElseMembershipWithPeerDistinct(
			result,
			nullableMembershipWithPeerDistinct{
				membership: constraints.nullableElseMembership,
				peer:       constraints.peerFieldDistinct,
			},
			rng,
			namedSetValues,
			nullableByName,
			peerFieldDistinctLookup,
		)
		if constraints.nullableElseBooleanConstant != nil {
			applyNullableElseBooleanConstant(result, constraints.nullableElseBooleanConstant)
		}
		return
	}

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
		return
	}

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

	if constraints.nullableElseBooleanConstant != nil {
		applyNullableElseBooleanConstant(result, constraints.nullableElseBooleanConstant)
	}
}

func applyFollowOnParameterConstraints(
	result map[string]object.Object,
	constraints parameterConstraints,
	rng *rand.Rand,
	namedSetValues map[string]object.Object,
	nullableByName map[string]bool,
	peerFieldDistinctLookup func(classKey identity.Key, fieldSubKey string) []object.Object,
) {
	integratedPeerIsoAbbr := constraints.nullableElseExclusionEquality != nil &&
		constraints.peerFieldDistinct != nil &&
		constraints.peerFieldDistinct.ParamName == constraints.nullableElseExclusionEquality.followerParam
	integratedPeerMembership := constraints.nullableElseMembership != nil &&
		constraints.peerFieldDistinct != nil &&
		constraints.peerFieldDistinct.ParamName == constraints.nullableElseMembership.paramName
	// Set-minus-peer membership already excludes used field values.
	integratedSetMinusPeer := constraints.paramInNamedSetMinusPeerField != nil
	// Plain named-set membership + peer-distinct on the same param samples set \ used.
	integratedPlainNamedSetPeer := constraints.paramInNamedSet != nil &&
		constraints.peerFieldDistinct != nil &&
		constraints.peerFieldDistinct.ParamName == constraints.paramInNamedSet.paramName

	if constraints.nullableElseEquality != nil &&
		constraints.nullableElseMirror == nil &&
		constraints.nullableElseExclusionEquality == nil &&
		!integratedPeerIsoAbbr {
		applyNullableElseEquality(result, constraints.nullableElseEquality, nullableByName)
	}

	if constraints.tupleInSet != nil {
		applyTupleInSet(result, constraints.tupleInSet, rng, namedSetValues)
	}

	switch {
	case constraints.paramInNamedSetMinusPeerField != nil:
		applyParamInNamedSetMinusPeerField(
			result, constraints.paramInNamedSetMinusPeerField, rng, namedSetValues, peerFieldDistinctLookup,
		)
	case integratedPlainNamedSetPeer:
		applyParamInNamedSetWithPeerDistinct(
			result, constraints.paramInNamedSet, constraints.peerFieldDistinct, rng, namedSetValues, peerFieldDistinctLookup,
		)
	case constraints.paramInNamedSet != nil &&
		!paramCoveredByCoupledNullableConstraint(constraints, constraints.paramInNamedSet.paramName):
		// Plain named-set membership must not overwrite ISO/Abbr coupling (or similar) already applied.
		applyParamInNamedSet(result, constraints.paramInNamedSet, rng, namedSetValues)
	}

	for paramName, values := range constraints.enumValues {
		if len(values) == 0 {
			continue
		}
		result[paramName] = object.NewString(values[rng.Intn(len(values))])
	}

	if constraints.peerFieldDistinct != nil &&
		!integratedPeerIsoAbbr &&
		!integratedPeerMembership &&
		!integratedSetMinusPeer &&
		!integratedPlainNamedSetPeer {
		applyPeerFieldDistinct(result, constraints.peerFieldDistinct, rng, peerFieldDistinctLookup)
	}
}

// paramCoveredByCoupledNullableConstraint reports whether a more specific coupling already
// owns sampling for this parameter (so plain named-set membership must not overwrite it).
func paramCoveredByCoupledNullableConstraint(constraints parameterConstraints, paramName string) bool {
	if constraints.nullableElseMembership != nil && constraints.nullableElseMembership.paramName == paramName {
		return true
	}
	if constraints.nullableElseExclusionEquality != nil &&
		(constraints.nullableElseExclusionEquality.driverParam == paramName ||
			constraints.nullableElseExclusionEquality.followerParam == paramName) {
		return true
	}
	if constraints.nullableElseMirror != nil &&
		(constraints.nullableElseMirror.driverParam == paramName ||
			constraints.nullableElseMirror.followerParam == paramName) {
		return true
	}
	if constraints.nullableElseEquality != nil &&
		(constraints.nullableElseEquality.driverParam == paramName ||
			constraints.nullableElseEquality.followerParam == paramName) {
		return true
	}
	if constraints.paramInNamedSetMinusPeerField != nil &&
		constraints.paramInNamedSetMinusPeerField.paramName == paramName {
		return true
	}
	return false
}

func applyParamInNamedSet(
	result map[string]object.Object,
	constraint *paramInNamedSetConstraint,
	rng *rand.Rand,
	namedSetValues map[string]object.Object,
) {
	value, ok := pickRandomValueFromNamedSetExcluding(constraint.setSubKey, namedSetValues, nil, rng)
	if !ok {
		return
	}
	result[constraint.paramName] = value.Clone()
}

func applyParamInNamedSetWithPeerDistinct(
	result map[string]object.Object,
	membership *paramInNamedSetConstraint,
	peer *peerFieldDistinctFromParamPattern,
	rng *rand.Rand,
	namedSetValues map[string]object.Object,
	lookup func(classKey identity.Key, fieldSubKey string) []object.Object,
) {
	used := peerUsedObjectKeys(peer, lookup)
	value, ok := pickRandomValueFromNamedSetExcluding(membership.setSubKey, namedSetValues, used, rng)
	if !ok {
		return
	}
	result[membership.paramName] = value.Clone()
}

func applyParamInNamedSetMinusPeerField(
	result map[string]object.Object,
	constraint *paramInNamedSetMinusPeerFieldConstraint,
	rng *rand.Rand,
	namedSetValues map[string]object.Object,
	lookup func(classKey identity.Key, fieldSubKey string) []object.Object,
) {
	// Peer used values come from the class field extent encoded in the set-map.
	peer := &peerFieldDistinctFromParamPattern{
		ClassKey:    constraint.classKey,
		ClassName:   constraint.className,
		FieldSubKey: constraint.fieldSubKey,
		ParamName:   constraint.paramName,
	}
	used := peerUsedObjectKeys(peer, lookup)
	value, ok := pickRandomValueFromNamedSetExcluding(constraint.setSubKey, namedSetValues, used, rng)
	if !ok {
		return
	}
	result[constraint.paramName] = value.Clone()
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
	return peerUsedObjectKeys(peer, lookup)
}

func applyPeerFieldDistinct(
	result map[string]object.Object,
	pattern *peerFieldDistinctFromParamPattern,
	rng *rand.Rand,
	lookup func(classKey identity.Key, fieldSubKey string) []object.Object,
) {
	used := peerUsedObjectKeys(pattern, lookup)
	for range maxNotInNamedSetAttempts {
		candidate := randomString(rng)
		if !used[distinctObjectKey(candidate)] {
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

func applyNullableElseBooleanConstant(
	result map[string]object.Object,
	constraint *nullableElseBooleanConstantConstraint,
) {
	if val, ok := result[constraint.driverParam]; ok && object.IsNull(val) {
		result[constraint.followerParam] = object.NewBoolean(constraint.value)
	}
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

	value, ok := pickRandomValueFromNamedSetExcluding(constraint.setSubKey, namedSetValues, nil, rng)
	if !ok {
		return
	}

	result[constraint.paramName] = value.Clone()
}

func applyNullableElseMembershipWithPeerDistinct(
	result map[string]object.Object,
	coupled nullableMembershipWithPeerDistinct,
	rng *rand.Rand,
	namedSetValues map[string]object.Object,
	nullableByName map[string]bool,
	lookup func(classKey identity.Key, fieldSubKey string) []object.Object,
) {
	constraint := coupled.membership
	used := peerUsedObjectKeys(coupled.peer, lookup)
	nullTaken := used[distinctObjectKey(nil)]

	if nullableByName[constraint.paramName] && !nullTaken && rng.Intn(5) == 0 {
		result[constraint.paramName] = evaluator.EMPTY_SET
		return
	}

	for range maxNotInNamedSetAttempts {
		value, ok := pickRandomValueFromNamedSetExcluding(constraint.setSubKey, namedSetValues, used, rng)
		if !ok {
			continue
		}
		result[constraint.paramName] = value.Clone()
		return
	}

	// Nullable index allows only one NULL; drop a generated NULL so retry can pick a concrete code.
	if nullTaken && object.IsNull(result[constraint.paramName]) {
		delete(result, constraint.paramName)
	}
}

// samplingPeerFieldDistinctConflict reports whether a sampled value collides with a peer field value.
func samplingPeerFieldDistinctConflict(
	result map[string]object.Object,
	constraints parameterConstraints,
	lookup func(classKey identity.Key, fieldSubKey string) []object.Object,
) bool {
	if lookup == nil {
		return false
	}
	if constraints.paramInNamedSetMinusPeerField != nil {
		c := constraints.paramInNamedSetMinusPeerField
		val, ok := result[c.paramName]
		if !ok {
			return false
		}
		peer := &peerFieldDistinctFromParamPattern{
			ClassKey:    c.classKey,
			ClassName:   c.className,
			FieldSubKey: c.fieldSubKey,
			ParamName:   c.paramName,
		}
		return peerUsedObjectKeys(peer, lookup)[distinctObjectKey(val)]
	}
	if constraints.peerFieldDistinct == nil {
		return false
	}
	pattern := constraints.peerFieldDistinct
	val, ok := result[pattern.ParamName]
	if !ok {
		return false
	}
	return peerUsedObjectKeys(pattern, lookup)[distinctObjectKey(val)]
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
		if excluded != nil && excluded[distinctObjectKey(candidate)] {
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
	return pickRandomValueFromNamedSetExcluding(setSubKey, namedSetValues, excluded, rng)
}

func pickRandomValueFromNamedSetExcluding(
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
		if excluded != nil && excluded[distinctObjectKey(elem)] {
			continue
		}
		available = append(available, elem)
	}
	if len(available) == 0 {
		return nil, false
	}
	return available[rng.Intn(len(available))], true
}

func distinctObjectKey(val object.Object) string {
	if object.IsNull(val) {
		return "NULL"
	}
	return string(val.Type()) + ":" + val.Inspect()
}

func peerUsedObjectKeys(
	peer *peerFieldDistinctFromParamPattern,
	lookup func(classKey identity.Key, fieldSubKey string) []object.Object,
) map[string]bool {
	used := make(map[string]bool)
	if lookup == nil {
		return used
	}
	for _, val := range lookup(peer.ClassKey, peer.FieldSubKey) {
		used[distinctObjectKey(val)] = true
	}
	return used
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
