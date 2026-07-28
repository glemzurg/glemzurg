package schema

import (
	"fmt"
	"slices"
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/model_bridge"
)

// IndexDefinition describes one composite index for a class.
type IndexDefinition struct {
	IndexNum  uint
	AttrNames []string
	AttrDefs  []*model_class.Attribute
}

// DerivedAttrDef is a pre-indexed derivation policy for one attribute.
type DerivedAttrDef struct {
	AttrKey    identity.Key
	AttrSubKey string
	AttrName   string
	Expression me.Expression
}

func (s *Schema) reindexAttributeProjections() {
	s.classIndexes = make(map[identity.Key][]IndexDefinition)
	s.attrsBySubKey = make(map[identity.Key]map[string]*model_class.Attribute)
	s.derivedByClass = make(map[identity.Key][]DerivedAttrDef)
	s.derivedByKey = make(map[identity.Key]DerivedAttrDef)

	s.forEachClassInScope(func(c *model_class.Class) {
		s.indexClassAttributes(c)
	})
}

func (s *Schema) indexClassAttributes(c *model_class.Class) {
	attrMap := make(map[string]*model_class.Attribute, len(c.Attributes))
	indexGroups := make(map[uint][]*model_class.Attribute)

	for i := range c.Attributes {
		attr := c.Attributes[i]
		attrCopy := attr
		attrMap[attr.Key.SubKey] = &attrCopy
		for _, indexNum := range attr.IndexNums {
			indexGroups[indexNum] = append(indexGroups[indexNum], &attrCopy)
		}
		if attr.DerivationPolicy != nil {
			expr := attr.DerivationPolicy.Spec.Expression
			if expr == nil {
				if attr.DerivationPolicy.Spec.Specification == "" {
					continue
				}
				// Leave unindexed; DerivedAttributes load will surface via evaluator error path
				// only when consumers call ValidateDerivedAttributes.
				continue
			}
			if model_bridge.ContainsAnyPrimedME(expr) {
				continue // ValidateDerivedAttributes reports these
			}
			def := DerivedAttrDef{
				AttrKey:    attr.Key,
				AttrSubKey: attr.Key.SubKey,
				AttrName:   attr.Name,
				Expression: expr,
			}
			s.derivedByClass[c.Key] = append(s.derivedByClass[c.Key], def)
			s.derivedByKey[attr.Key] = def
		}
	}
	s.attrsBySubKey[c.Key] = attrMap

	if len(indexGroups) == 0 {
		return
	}
	indexNums := make([]uint, 0, len(indexGroups))
	for num := range indexGroups {
		indexNums = append(indexNums, num)
	}
	slices.Sort(indexNums)
	var defs []IndexDefinition
	for _, indexNum := range indexNums {
		attrs := indexGroups[indexNum]
		sort.Slice(attrs, func(i, j int) bool { return attrs[i].Key.SubKey < attrs[j].Key.SubKey })
		names := make([]string, len(attrs))
		for i, a := range attrs {
			names[i] = a.Key.SubKey
		}
		defs = append(defs, IndexDefinition{
			IndexNum:  indexNum,
			AttrNames: names,
			AttrDefs:  attrs,
		})
	}
	s.classIndexes[c.Key] = defs
}

// ClassIndexes returns composite index definitions for an in-scope class.
func (s *Schema) ClassIndexes(classKey identity.Key) ([]IndexDefinition, bool, error) {
	if s == nil {
		return nil, false, fmt.Errorf("schema.ClassIndexes: nil schema")
	}
	if _, ok := s.classes[classKey]; !ok {
		return nil, false, fmt.Errorf("unknown class: %s", classKey.String())
	}
	if !s.inScope[classKey] {
		return nil, false, nil
	}
	defs := s.classIndexes[classKey]
	if len(defs) == 0 {
		return nil, true, nil
	}
	out := make([]IndexDefinition, len(defs))
	copy(out, defs)
	return out, true, nil
}

// AllClassIndexes returns index definitions for every in-scope class that has indexes.
func (s *Schema) AllClassIndexes() map[identity.Key][]IndexDefinition {
	if s == nil || len(s.classIndexes) == 0 {
		return nil
	}
	out := make(map[identity.Key][]IndexDefinition, len(s.classIndexes))
	for k, defs := range s.classIndexes {
		cp := make([]IndexDefinition, len(defs))
		copy(cp, defs)
		out[k] = cp
	}
	return out
}

// AttributesBySubKey returns attribute defs keyed by SubKey for an in-scope class.
func (s *Schema) AttributesBySubKey(classKey identity.Key) (map[string]*model_class.Attribute, bool, error) {
	if s == nil {
		return nil, false, fmt.Errorf("schema.AttributesBySubKey: nil schema")
	}
	if _, ok := s.classes[classKey]; !ok {
		return nil, false, fmt.Errorf("unknown class: %s", classKey.String())
	}
	if !s.inScope[classKey] {
		return nil, false, nil
	}
	return s.attrsBySubKey[classKey], true, nil
}

// AllAttributesBySubKey returns attr maps for every in-scope class.
func (s *Schema) AllAttributesBySubKey() map[identity.Key]map[string]*model_class.Attribute {
	if s == nil || len(s.attrsBySubKey) == 0 {
		return nil
	}
	return s.attrsBySubKey
}

// DerivedAttributes returns derived attribute defs for an in-scope class.
func (s *Schema) DerivedAttributes(classKey identity.Key) ([]DerivedAttrDef, bool, error) {
	if s == nil {
		return nil, false, fmt.Errorf("schema.DerivedAttributes: nil schema")
	}
	if _, ok := s.classes[classKey]; !ok {
		return nil, false, fmt.Errorf("unknown class: %s", classKey.String())
	}
	if !s.inScope[classKey] {
		return nil, false, nil
	}
	defs := s.derivedByClass[classKey]
	if len(defs) == 0 {
		return nil, true, nil
	}
	out := make([]DerivedAttrDef, len(defs))
	copy(out, defs)
	return out, true, nil
}

// DerivedAttribute returns a single derived attribute def by attribute key.
func (s *Schema) DerivedAttribute(attrKey identity.Key) (DerivedAttrDef, bool, error) {
	if s == nil {
		return DerivedAttrDef{}, false, fmt.Errorf("schema.DerivedAttribute: nil schema")
	}
	def, ok := s.derivedByKey[attrKey]
	if !ok {
		// Not derived or not indexed — treat as not in derived set (not an unknown-class error).
		return DerivedAttrDef{}, false, nil
	}
	return def, true, nil
}

// DerivedAttributeKeys returns the set of derived attribute keys in scope.
func (s *Schema) DerivedAttributeKeys() map[identity.Key]bool {
	if s == nil || len(s.derivedByKey) == 0 {
		return nil
	}
	out := make(map[identity.Key]bool, len(s.derivedByKey))
	for k := range s.derivedByKey {
		out[k] = true
	}
	return out
}

// EachDerivedAttributeClass calls fn for each in-scope class that has derived attributes.
func (s *Schema) EachDerivedAttributeClass(fn func(classKey identity.Key, defs []DerivedAttrDef)) {
	if s == nil || fn == nil || len(s.derivedByClass) == 0 {
		return
	}
	keys := make([]identity.Key, 0, len(s.derivedByClass))
	for k := range s.derivedByClass {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].String() < keys[j].String()
	})
	for _, k := range keys {
		defs := s.derivedByClass[k]
		cp := make([]DerivedAttrDef, len(defs))
		copy(cp, defs)
		fn(k, cp)
	}
}

// ValidateDerivedAttributes reports derivation policies that failed to index (unparsed / primed).
func (s *Schema) ValidateDerivedAttributes() error {
	if s == nil {
		return nil
	}
	var first error
	s.forEachClassInScope(func(c *model_class.Class) {
		if first != nil {
			return
		}
		for _, attr := range c.Attributes {
			if attr.DerivationPolicy == nil {
				continue
			}
			expr := attr.DerivationPolicy.Spec.Expression
			if expr == nil {
				if attr.DerivationPolicy.Spec.Specification == "" {
					continue
				}
				first = fmt.Errorf(
					"class %s attribute %s DerivationPolicy: expression not lowered",
					c.Name, attr.Name,
				)
				return
			}
			if model_bridge.ContainsAnyPrimedME(expr) {
				first = fmt.Errorf(
					"class %s attribute %s DerivationPolicy must not contain primed variables",
					c.Name, attr.Name,
				)
				return
			}
		}
	})
	return first
}
