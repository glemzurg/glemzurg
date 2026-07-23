package schema

import (
	"slices"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// Schema is the immutable metadata for one simulation run.
// Construct via [NewEmpty] or [NewFromModel]; do not mutate after construction.
type Schema struct {
	classes      map[identity.Key]Class
	associations map[identity.Key]Association
}

// Class is static metadata for one class on the simulation surface.
type Class struct {
	Key        identity.Key
	Name       string
	Attributes []model_class.Attribute
}

// Association is static metadata for one class association.
type Association struct {
	Key                 identity.Key
	Name                string
	FromClassKey        identity.Key
	ToClassKey          identity.Key
	AssociationClassKey *identity.Key
	FromMultiplicity    model_class.Multiplicity
	ToMultiplicity      model_class.Multiplicity
}

// NewEmpty returns a schema with no classes or associations (tests / bootstrap).
func NewEmpty() *Schema {
	return &Schema{
		classes:      make(map[identity.Key]Class),
		associations: make(map[identity.Key]Association),
	}
}

// NewFromModel builds a schema from a model. Typically the surface-filtered model
// so only in-scope classes appear as [Class] entries.
func NewFromModel(model *core.Model) *Schema {
	sch := NewEmpty()
	if model == nil {
		return sch
	}

	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				sch.classes[class.Key] = Class{
					Key:        class.Key,
					Name:       class.Name,
					Attributes: slices.Clone(class.Attributes),
				}
			}
		}
	}

	for _, assoc := range model.GetClassAssociations() {
		var acKey *identity.Key
		if assoc.AssociationClassKey != nil {
			k := *assoc.AssociationClassKey
			acKey = &k
		}
		sch.associations[assoc.Key] = Association{
			Key:                 assoc.Key,
			Name:                assoc.Name,
			FromClassKey:        assoc.FromClassKey,
			ToClassKey:          assoc.ToClassKey,
			AssociationClassKey: acKey,
			FromMultiplicity:    assoc.FromMultiplicity,
			ToMultiplicity:      assoc.ToMultiplicity,
		}
	}

	return sch
}

// IsClassInScope reports whether classKey is registered on this schema
// (i.e. was present in the model used to build it).
func (s *Schema) IsClassInScope(classKey identity.Key) bool {
	if s == nil {
		return false
	}
	_, ok := s.classes[classKey]
	return ok
}

// Class returns static metadata for a class.
func (s *Schema) Class(classKey identity.Key) (Class, bool) {
	if s == nil {
		return Class{}, false
	}
	c, ok := s.classes[classKey]
	return c, ok
}

// Attributes returns the attribute definitions for a class, or nil if unknown.
func (s *Schema) Attributes(classKey identity.Key) []model_class.Attribute {
	c, ok := s.Class(classKey)
	if !ok {
		return nil
	}
	return c.Attributes
}

// ClassKeys returns every in-scope class key (order is not significant).
func (s *Schema) ClassKeys() []identity.Key {
	if s == nil || len(s.classes) == 0 {
		return nil
	}
	keys := make([]identity.Key, 0, len(s.classes))
	for k := range s.classes {
		keys = append(keys, k)
	}
	return keys
}

// Association returns static metadata for an association.
func (s *Schema) Association(assocKey identity.Key) (Association, bool) {
	if s == nil {
		return Association{}, false
	}
	a, ok := s.associations[assocKey]
	return a, ok
}

// IsAssociationClass reports whether classKey is an association-class for some association.
func (s *Schema) IsAssociationClass(classKey identity.Key) bool {
	if s == nil {
		return false
	}
	for _, a := range s.associations {
		if a.AssociationClassKey != nil && *a.AssociationClassKey == classKey {
			return true
		}
	}
	return false
}

// AssociationKeys returns every association key (order is not significant).
func (s *Schema) AssociationKeys() []identity.Key {
	if s == nil || len(s.associations) == 0 {
		return nil
	}
	keys := make([]identity.Key, 0, len(s.associations))
	for k := range s.associations {
		keys = append(keys, k)
	}
	return keys
}
