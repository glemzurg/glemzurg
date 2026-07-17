package actions

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// NamedSetSampleDomainsAvailable reports whether every named-set sampling domain
// required by owner logics still has at least one free value.
//
// When a require is Param ∈ (NamedSet \ used peer fields) and every set element is
// already used, the domain is empty and the action must not be selected.
// Model-agnostic: driven only by extracted sampling constraints.
func (s *ParameterSampler) NamedSetSampleDomainsAvailable(
	owner ParameterOwner,
	paramDefs []model_state.Parameter,
) (bool, error) {
	if s == nil {
		return true, nil
	}
	// Domain gates come from owner requires/invariants. Prefer full owner parameters
	// when the caller has no matched subset (e.g. event name mismatch) so requires
	// like set-minus-used still exclude the action.
	if len(paramDefs) == 0 {
		paramDefs = owner.Parameters
	}
	if len(owner.Requires) == 0 && len(paramDefs) == 0 {
		return true, nil
	}
	// Always consult SamplingLogicsFor when the owner has requires, even if the
	// matched param subset looks empty of typed constraints.
	logics, err := owner.SamplingLogicsFor(paramDefs)
	if err != nil {
		return false, err
	}
	if len(logics) == 0 {
		// Fall back to explicit requires alone (invariants may be absent).
		logics = owner.Requires
	}
	if len(logics) == 0 {
		return true, nil
	}
	constraints := extractParameterConstraints(logics)
	nullableByName := parameterNullableByName(paramDefs)
	if len(nullableByName) == 0 {
		nullableByName = parameterNullableByName(owner.Parameters)
	}
	return namedSetConstraintsHaveFreeValues(
		constraints,
		nullableByName,
		s.namedSetValues,
		s.peerFieldDistinctLookup,
	), nil
}

func namedSetConstraintsHaveFreeValues(
	constraints parameterConstraints,
	nullableByName map[string]bool,
	namedSetValues map[string]object.Object,
	peerLookup func(classKey identity.Key, fieldSubKey string) []object.Object,
) bool {
	if c := constraints.paramInNamedSetMinusPeerField; c != nil {
		peer := &peerFieldDistinctFromParamPattern{
			ClassKey:    c.classKey,
			ClassName:   c.className,
			FieldSubKey: c.fieldSubKey,
			ParamName:   c.paramName,
		}
		used := peerUsedObjectKeys(peer, peerLookup)
		exists, free := namedSetFreeElementCount(c.setSubKey, namedSetValues, used)
		// Missing named-set registration: do not gate (sampler will fall back / fail later).
		if exists && free == 0 {
			return false
		}
	}

	if c := constraints.paramInNamedSet; c != nil {
		if constraints.peerFieldDistinct != nil &&
			constraints.peerFieldDistinct.ParamName == c.paramName {
			used := peerUsedObjectKeys(constraints.peerFieldDistinct, peerLookup)
			exists, free := namedSetFreeElementCount(c.setSubKey, namedSetValues, used)
			if exists && free == 0 {
				// Nullable params may still sample NULL when no peer holds NULL.
				if nullableByName[c.paramName] && !used[distinctObjectKey(nil)] {
					return true
				}
				return false
			}
		} else {
			exists, free := namedSetFreeElementCount(c.setSubKey, namedSetValues, nil)
			if exists && free == 0 {
				return false
			}
		}
	}

	if c := constraints.nullableElseMembership; c != nil {
		if constraints.peerFieldDistinct != nil &&
			constraints.peerFieldDistinct.ParamName == c.paramName {
			used := peerUsedObjectKeys(constraints.peerFieldDistinct, peerLookup)
			exists, free := namedSetFreeElementCount(c.setSubKey, namedSetValues, used)
			nullFree := nullableByName[c.paramName] && !used[distinctObjectKey(nil)]
			if exists && free == 0 && !nullFree {
				return false
			}
		} else {
			exists, free := namedSetFreeElementCount(c.setSubKey, namedSetValues, nil)
			if exists && free == 0 && !nullableByName[c.paramName] {
				return false
			}
		}
	}

	return true
}

// namedSetFreeElementCount returns whether the named set is registered and how many
// elements remain after exclusion. exists=false when the set is not in namedSetValues.
func namedSetFreeElementCount(
	setSubKey string,
	namedSetValues map[string]object.Object,
	excluded map[string]bool,
) (exists bool, free int) {
	setObj, ok := namedSetValues[setSubKey]
	if !ok {
		return false, 0
	}
	set, ok := setObj.(*object.Set)
	if !ok {
		return true, 0
	}
	if set.Size() == 0 {
		return true, 0
	}
	count := 0
	for _, elem := range set.Elements() {
		if excluded != nil && excluded[distinctObjectKey(elem)] {
			continue
		}
		count++
	}
	return true, count
}
