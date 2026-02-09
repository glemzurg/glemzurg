package invariants

import (
	"sort"
	"strings"

	"github.com/glemzurg/go-tlaplus/internal/identity"
	"github.com/glemzurg/go-tlaplus/internal/req_model"
	"github.com/glemzurg/go-tlaplus/internal/req_model/model_class"
	"github.com/glemzurg/go-tlaplus/internal/simulator/object"
	"github.com/glemzurg/go-tlaplus/internal/simulator/state"
)

// IndexDefinition describes one composite index for a class.
type IndexDefinition struct {
	IndexNum uint
	AttrNames []string                 // sorted alphabetically for deterministic tuples
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

// NewIndexUniquenessChecker creates a new index uniqueness checker from a model.
func NewIndexUniquenessChecker(model *req_model.Model) *IndexUniquenessChecker {
	checker := &IndexUniquenessChecker{
		classIndexes: make(map[identity.Key]*ClassIndexInfo),
	}

	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				// Group attributes by index number
				indexGroups := make(map[uint][]*model_class.Attribute)

				for _, attr := range class.Attributes {
					attrCopy := attr
					for _, indexNum := range attr.IndexNums {
						indexGroups[indexNum] = append(indexGroups[indexNum], &attrCopy)
					}
				}

				if len(indexGroups) == 0 {
					continue
				}

				// Build sorted index definitions
				info := &ClassIndexInfo{
					ClassKey: class.Key,
				}

				// Sort index numbers for deterministic order
				indexNums := make([]uint, 0, len(indexGroups))
				for num := range indexGroups {
					indexNums = append(indexNums, num)
				}
				sort.Slice(indexNums, func(i, j int) bool { return indexNums[i] < indexNums[j] })

				for _, indexNum := range indexNums {
					attrs := indexGroups[indexNum]
					// Sort attributes alphabetically by name
					sort.Slice(attrs, func(i, j int) bool { return attrs[i].Name < attrs[j].Name })

					names := make([]string, len(attrs))
					for i, a := range attrs {
						names[i] = a.Name
					}

					info.Indexes = append(info.Indexes, IndexDefinition{
						IndexNum:  indexNum,
						AttrNames: names,
						AttrDefs:  attrs,
					})
				}

				checker.classIndexes[class.Key] = info
			}
		}
	}

	return checker
}

// CheckState validates all instances in a simulation state for index uniqueness.
func (c *IndexUniquenessChecker) CheckState(simState *state.SimulationState) ViolationList {
	var violations ViolationList

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
	instances []*state.ClassInstance,
	indexInfo *ClassIndexInfo,
) ViolationList {
	var violations ViolationList

	for _, indexDef := range indexInfo.Indexes {
		seen := make(map[string]state.InstanceID)

		for _, instance := range instances {
			getter := func(name string) object.Object {
				return instance.GetAttribute(name)
			}
			tupleKey := BuildTupleKey(getter, indexDef.AttrNames)

			if existingID, exists := seen[tupleKey]; exists {
				// Build human-readable tuple values
				tupleValues := make([]string, len(indexDef.AttrNames))
				for i, name := range indexDef.AttrNames {
					val := instance.GetAttribute(name)
					if val == nil {
						tupleValues[i] = "<nil>"
					} else {
						tupleValues[i] = val.Inspect()
					}
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

// HasIndexes returns true if any class has indexes.
func (c *IndexUniquenessChecker) HasIndexes() bool {
	return len(c.classIndexes) > 0
}

// BuildTupleKey builds a string key from attribute values for duplicate detection.
// Uses Type() + ":" + Inspect() for each attribute, joined by "\x00".
// Nil values become "<nil>".
func BuildTupleKey(getter func(string) object.Object, attrNames []string) string {
	var parts []string
	for _, name := range attrNames {
		val := getter(name)
		if val == nil {
			parts = append(parts, "<nil>")
		} else {
			parts = append(parts, string(val.Type())+":"+val.Inspect())
		}
	}
	return strings.Join(parts, "\x00")
}
