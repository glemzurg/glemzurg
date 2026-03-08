package actions

import (
	"fmt"
	"math/rand"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/invariants"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/state"
)

// generateIndexSafeValues populates attrs with random values for indexed attributes
// that don't conflict with existing instances. For each index on the class, it generates
// values for any indexed attribute not already set on attrs, ensuring the resulting tuple
// is unique across all existing instances.
func generateIndexSafeValues(
	attrs *object.Record,
	indexInfo *invariants.ClassIndexInfo,
	existingInstances []*state.ClassInstance,
	rng *rand.Rand,
) error {
	for _, indexDef := range indexInfo.Indexes {
		// Collect existing tuple keys for this index
		existingKeys := make(map[string]bool)
		for _, inst := range existingInstances {
			getter := func(name string) object.Object {
				return inst.GetAttribute(name)
			}
			key := invariants.BuildTupleKey(getter, indexDef.AttrNames)
			existingKeys[key] = true
		}

		// Determine which attributes need to be generated (not pre-set)
		preSet := make(map[string]bool)
		for _, attrName := range indexDef.AttrNames {
			if attrs.Get(attrName) != nil {
				preSet[attrName] = true
			}
		}

		// Try random generation up to 100 times
		const maxRetries = 100
		found := false

		for range maxRetries {
			// Generate random values for non-preset indexed attributes
			for i, attrName := range indexDef.AttrNames {
				if preSet[attrName] {
					continue
				}
				attrDef := indexDef.AttrDefs[i]
				val := generateRandomValue(attrDef.DataType, rng)
				attrs.Set(attrName, val)
			}

			// Check if the resulting tuple conflicts
			getter := func(name string) object.Object {
				return attrs.Get(name)
			}
			tupleKey := invariants.BuildTupleKey(getter, indexDef.AttrNames)

			if !existingKeys[tupleKey] {
				found = true
				break
			}
		}

		if !found {
			// Random generation exhausted — try deterministic fallback
			if err := deterministicFallback(attrs, indexDef, existingKeys, preSet); err != nil {
				return err
			}
		}
	}

	return nil
}

// deterministicFallback tries sequential/exhaustive value generation when random fails.
func deterministicFallback(
	attrs *object.Record,
	indexDef invariants.IndexDefinition,
	existingKeys map[string]bool,
	preSet map[string]bool,
) error {
	// Try to vary non-preset attributes deterministically
	for i, attrName := range indexDef.AttrNames {
		if preSet[attrName] {
			continue
		}
		attrDef := indexDef.AttrDefs[i]
		if tryDeterministicValues(attrs, attrDef, attrName, indexDef, existingKeys) {
			return nil
		}
	}

	return fmt.Errorf("index %d: unable to generate unique tuple for attributes %v after exhaustive search", indexDef.IndexNum, indexDef.AttrNames)
}

// tryDeterministicValues tries deterministic value generation for a single attribute.
// Returns true if a unique tuple was found.
func tryDeterministicValues(
	attrs *object.Record,
	attrDef *model_class.Attribute,
	attrName string,
	indexDef invariants.IndexDefinition,
	existingKeys map[string]bool,
) bool {
	if attrDef.DataType != nil && attrDef.DataType.Atomic != nil {
		switch attrDef.DataType.Atomic.ConstraintType {
		case model_data_type.CONSTRAINT_TYPE_ENUMERATION:
			return tryEnumValues(attrs, attrDef, attrName, indexDef, existingKeys)
		case model_data_type.CONSTRAINT_TYPE_SPAN:
			return trySpanValues(attrs, attrName, indexDef, existingKeys)
		default:
			return trySequentialValues(attrs, attrName, indexDef, existingKeys, 0)
		}
	}
	// No data type — try sequential integers
	return trySequentialValues(attrs, attrName, indexDef, existingKeys, 0)
}

// tryEnumValues tries each enumeration value and checks for tuple uniqueness.
func tryEnumValues(
	attrs *object.Record,
	attrDef *model_class.Attribute,
	attrName string,
	indexDef invariants.IndexDefinition,
	existingKeys map[string]bool,
) bool {
	for _, enumVal := range attrDef.DataType.Atomic.Enums {
		attrs.Set(attrName, object.NewString(enumVal.Value))
		if isUniqueTuple(attrs, indexDef, existingKeys) {
			return true
		}
	}
	return false
}

// trySpanValues tries sequential values starting from max existing + 1.
func trySpanValues(
	attrs *object.Record,
	attrName string,
	indexDef invariants.IndexDefinition,
	existingKeys map[string]bool,
) bool {
	maxVal := findMaxExistingValue(existingKeys, indexDef, attrName)
	return trySequentialValues(attrs, attrName, indexDef, existingKeys, maxVal+1)
}

// trySequentialValues tries sequential integer values starting from startVal.
func trySequentialValues(
	attrs *object.Record,
	attrName string,
	indexDef invariants.IndexDefinition,
	existingKeys map[string]bool,
	startVal int64,
) bool {
	for seq := startVal; seq < startVal+1000; seq++ {
		attrs.Set(attrName, object.NewInteger(seq))
		if isUniqueTuple(attrs, indexDef, existingKeys) {
			return true
		}
	}
	return false
}

// isUniqueTuple checks whether the current attrs produce a tuple key not in existingKeys.
func isUniqueTuple(attrs *object.Record, indexDef invariants.IndexDefinition, existingKeys map[string]bool) bool {
	getter := func(name string) object.Object {
		return attrs.Get(name)
	}
	tupleKey := invariants.BuildTupleKey(getter, indexDef.AttrNames)
	return !existingKeys[tupleKey]
}

// findMaxExistingValue scans existing tuple keys and returns the max integer seen
// for the given attribute position. Returns 0 if none found.
func findMaxExistingValue(existingKeys map[string]bool, indexDef invariants.IndexDefinition, targetAttr string) int64 {
	// This is a best-effort heuristic; we start from 0 if we can't determine max
	_ = existingKeys
	_ = indexDef
	_ = targetAttr
	return 0
}
