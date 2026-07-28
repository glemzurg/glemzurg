package invariants

import (
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"
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
type IndexUniquenessChecker struct {
	classIndexes map[identity.Key]*ClassIndexInfo
}

// NewIndexUniquenessChecker creates a new index uniqueness checker from schema.
func NewIndexUniquenessChecker(sch *schema.Schema) *IndexUniquenessChecker {
	checker := &IndexUniquenessChecker{
		classIndexes: make(map[identity.Key]*ClassIndexInfo),
	}
	for classKey, defs := range sch.AllClassIndexes() {
		info := &ClassIndexInfo{ClassKey: classKey}
		for _, d := range defs {
			info.Indexes = append(info.Indexes, IndexDefinition{
				IndexNum:  d.IndexNum,
				AttrNames: d.AttrNames,
				AttrDefs:  d.AttrDefs,
			})
		}
		checker.classIndexes[classKey] = info
	}
	return checker
}

// CheckState validates all instances in a simulation state for index uniqueness.
func (c *IndexUniquenessChecker) CheckState(simState *instance.State) ViolationErrors {
	var violations ViolationErrors

	for classKey, indexInfo := range c.classIndexes {
		instances := simState.InstancesByClass(classKey)
		if len(instances) < 2 {
			continue
		}
		violations = append(violations, c.CheckClassInstances(classKey, instances, indexInfo)...)
	}

	return violations
}

// CheckClassInstances checks index uniqueness for instances of a single class.
func (c *IndexUniquenessChecker) CheckClassInstances(
	classKey identity.Key,
	instances []*instance.Instance,
	indexInfo *ClassIndexInfo,
) ViolationErrors {
	var violations ViolationErrors

	for _, indexDef := range indexInfo.Indexes {
		seen := make(map[string]instance.ID)

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

				violations = append(violations, NewIndexUniquenessViolation(
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
	return c.classIndexes[classKey]
}

// hasIndexes returns true if any class has indexes.
func (c *IndexUniquenessChecker) hasIndexes() bool {
	return len(c.classIndexes) > 0
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
