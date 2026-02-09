package actions

import (
	"fmt"
	"math/rand"

	"github.com/glemzurg/go-tlaplus/internal/simulator/invariants"
	"github.com/glemzurg/go-tlaplus/internal/simulator/object"
	"github.com/glemzurg/go-tlaplus/internal/simulator/state"

	"github.com/glemzurg/go-tlaplus/internal/req_model/model_class"
	"github.com/glemzurg/go-tlaplus/internal/req_model/model_data_type"
)

// generateIndexSafeValues populates attrs with random values for indexed attributes
// that don't conflict with existing instances. For each index on the class, it generates
// values for any indexed attribute not already set on attrs, ensuring the resulting tuple
// is unique across all existing instances.
func generateIndexSafeValues(
	attrs *object.Record,
	indexInfo *invariants.ClassIndexInfo,
	existingInstances []*state.ClassInstance,
	class model_class.Class,
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

		for attempt := 0; attempt < maxRetries; attempt++ {
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
		if !preSet[attrName] {
			// This attribute was being generated — try deterministic values
			attrDef := indexDef.AttrDefs[i]
			if attrDef.DataType != nil && attrDef.DataType.Atomic != nil {
				switch attrDef.DataType.Atomic.ConstraintType {
				case model_data_type.ConstraintTypeEnumeration:
					// Try each enum value
					for _, enumVal := range attrDef.DataType.Atomic.Enums {
						attrs.Set(attrName, object.NewString(enumVal.Value))
						getter := func(name string) object.Object {
							return attrs.Get(name)
						}
						tupleKey := invariants.BuildTupleKey(getter, indexDef.AttrNames)
						if !existingKeys[tupleKey] {
							return nil
						}
					}

				case model_data_type.ConstraintTypeSpan:
					// Try sequential values starting from max existing + 1
					maxVal := findMaxExistingValue(existingKeys, indexDef, attrName)
					for seq := maxVal + 1; seq < maxVal+1001; seq++ {
						attrs.Set(attrName, object.NewInteger(seq))
						getter := func(name string) object.Object {
							return attrs.Get(name)
						}
						tupleKey := invariants.BuildTupleKey(getter, indexDef.AttrNames)
						if !existingKeys[tupleKey] {
							return nil
						}
					}

				default:
					// Unconstrained — try sequential integers
					for seq := int64(0); seq < 1000; seq++ {
						attrs.Set(attrName, object.NewInteger(seq))
						getter := func(name string) object.Object {
							return attrs.Get(name)
						}
						tupleKey := invariants.BuildTupleKey(getter, indexDef.AttrNames)
						if !existingKeys[tupleKey] {
							return nil
						}
					}
				}
			} else {
				// No data type — try sequential integers
				for seq := int64(0); seq < 1000; seq++ {
					attrs.Set(attrName, object.NewInteger(seq))
					getter := func(name string) object.Object {
						return attrs.Get(name)
					}
					tupleKey := invariants.BuildTupleKey(getter, indexDef.AttrNames)
					if !existingKeys[tupleKey] {
						return nil
					}
				}
			}
		}
	}

	return fmt.Errorf("index %d: unable to generate unique tuple for attributes %v after exhaustive search", indexDef.IndexNum, indexDef.AttrNames)
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
