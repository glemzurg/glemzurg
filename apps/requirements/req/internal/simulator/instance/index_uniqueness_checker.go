package instance

import (
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/schema"
)

// IndexDefinition describes one composite index for a class.
type IndexDefinition struct {
	IndexNum  uint
	AttrNames []string                 // attribute field keys (SubKey), sorted for deterministic tuples
	AttrDefs  []*model_class.Attribute // parallel to AttrNames
}

// ClassIndexInfo holds all indexes for a single class.
type ClassIndexInfo struct {
	ClassKey identity.Key
	Indexes  []IndexDefinition
}

// IndexUniquenessChecker validates that index tuples are unique across instances of a class.
// Index definitions are loaded per class from schema at check time (no bulk dump).
type IndexUniquenessChecker struct {
	sch *schema.Schema
}

// NewIndexUniquenessChecker creates a new index uniqueness checker from schema.
func NewIndexUniquenessChecker(sch *schema.Schema) *IndexUniquenessChecker {
	return &IndexUniquenessChecker{sch: sch}
}

// CheckState validates all instances in a simulation state for index uniqueness.
func (c *IndexUniquenessChecker) CheckState(simState *State) ViolationErrors {
	var violations ViolationErrors
	if c == nil || c.sch == nil || simState == nil {
		return violations
	}

	// Group live instances by class, then ask schema for that class's indexes only.
	byClass := make(map[identity.Key][]*Instance)
	simState.ForEachInstance(func(inst *Instance) {
		byClass[inst.ClassKey] = append(byClass[inst.ClassKey], inst)
	})
	for classKey, instances := range byClass {
		if len(instances) < 2 {
			continue
		}
		indexInfo := c.GetClassIndexInfo(classKey)
		if indexInfo == nil {
			continue
		}
		violations = append(violations, c.CheckClassInstances(classKey, instances, indexInfo)...)
	}

	return violations
}

// CheckClassInstances checks index uniqueness for instances of a single class.
func (c *IndexUniquenessChecker) CheckClassInstances(
	classKey identity.Key,
	instances []*Instance,
	indexInfo *ClassIndexInfo,
) ViolationErrors {
	var violations ViolationErrors

	for _, indexDef := range indexInfo.Indexes {
		seen := make(map[string]ID)

		for _, instance := range instances {
			getter := func(name string) object.Object {
				return instance.GetAttribute(name)
			}
			tupleKey := BuildTupleKey(getter, indexDef.AttrNames)

			if existingID, exists := seen[tupleKey]; exists {
				// Build human-readable tuple values
				tupleValues := make([]string, len(indexDef.AttrNames))
				for i, name := range indexDef.AttrNames {
					tupleValues[i] = formatIndexTupleValue(instance.GetAttribute(name))
				}

				violations = append(violations, newIndexUniquenessViolation(
					existingID,
					instance.ID,
					classKey,
					indexDef.IndexNum,
					indexDef.AttrNames,
					tupleValues,
				))
			} else {
				seen[tupleKey] = instance.ID
			}
		}
	}

	return violations
}

// GetClassIndexInfo returns the index info for a class, or nil if the class has no indexes.
func (c *IndexUniquenessChecker) GetClassIndexInfo(classKey identity.Key) *ClassIndexInfo {
	if c == nil || c.sch == nil {
		return nil
	}
	defs, inScope, err := c.sch.ClassIndexes(classKey)
	if err != nil || !inScope || len(defs) == 0 {
		return nil
	}
	info := &ClassIndexInfo{ClassKey: classKey}
	for _, d := range defs {
		info.Indexes = append(info.Indexes, IndexDefinition{
			IndexNum:  d.IndexNum,
			AttrNames: d.AttrNames,
			AttrDefs:  d.AttrDefs,
		})
	}
	return info
}

// hasIndexes reports whether any in-scope class declares indexes (setup/tests).
func (c *IndexUniquenessChecker) hasIndexes() bool {
	if c == nil || c.sch == nil {
		return false
	}
	found := false
	c.sch.EachInScopeClassSim(func(sim *schema.ClassSimInfo) {
		if found || sim == nil {
			return
		}
		if info := c.GetClassIndexInfo(sim.ClassKey); info != nil {
			found = true
		}
	})
	return found
}

// BuildTupleKey builds a string key from attribute values for duplicate detection.
// NULL representations (unset attributes and simulator Null) share one canonical key
// so nullable indexes treat NULL as a single occupiable value.
func BuildTupleKey(getter func(string) object.Object, attrNames []string) string {
	parts := make([]string, len(attrNames))
	for i, name := range attrNames {
		parts[i] = indexTupleValueKey(getter(name))
	}
	return strings.Join(parts, "\x00")
}

func indexTupleValueKey(val object.Object) string {
	if object.IsNull(val) {
		return "NULL"
	}
	return string(val.Type()) + ":" + val.Inspect()
}

func formatIndexTupleValue(val object.Object) string {
	if object.IsNull(val) {
		return "NULL"
	}
	return val.Inspect()
}
