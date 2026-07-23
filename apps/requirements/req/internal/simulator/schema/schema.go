package schema

import (
	"slices"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// Schema is the sole home of model facts for one simulation run.
//
// Construction is the intake gate: pass a (typically surface-filtered) *core.Model
// into [NewFromModel]. After that, the running simulator must obtain model data
// only through Schema (or components built from Schema). Do not keep a second
// *core.Model pointer for the same run.
//
// Schema is immutable for the run: do not mutate the underlying model after
// NewFromModel, or indexes will disagree with the model.
//
// Construct via [NewEmpty] or [NewFromModel].
type Schema struct {
	// model is the authoritative static model for this run (may be nil for empty schema).
	model *core.Model

	// Indexed views derived once from model for hot lookups.
	classes      map[identity.Key]Class
	associations map[identity.Key]Association
}

// Class is a run-facing view of one in-scope class (not the full model_class.Class).
type Class struct {
	Key        identity.Key
	Name       string
	Attributes []model_class.Attribute
}

// Association is a run-facing view of one class association.
type Association struct {
	Key                 identity.Key
	Name                string
	FromClassKey        identity.Key
	ToClassKey          identity.Key
	AssociationClassKey *identity.Key
	FromMultiplicity    model_class.Multiplicity
	ToMultiplicity      model_class.Multiplicity
}

// NewEmpty returns a schema with no model (tests / bootstrap).
func NewEmpty() *Schema {
	return &Schema{
		classes:      make(map[identity.Key]Class),
		associations: make(map[identity.Key]Association),
	}
}

// NewFromModel takes ownership of model as the sole static model for a run.
// Typically pass the surface-filtered active model. The caller must not mutate
// model afterward and must not retain a separate model pointer for simulator use.
func NewFromModel(model *core.Model) *Schema {
	sch := NewEmpty()
	if model == nil {
		return sch
	}
	sch.model = model
	sch.reindex()
	return sch
}

// reindex rebuilds lookup maps from the owned model.
func (s *Schema) reindex() {
	s.classes = make(map[identity.Key]Class)
	s.associations = make(map[identity.Key]Association)
	if s.model == nil {
		return
	}

	for _, domain := range s.model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				s.classes[class.Key] = Class{
					Key:        class.Key,
					Name:       class.Name,
					Attributes: slices.Clone(class.Attributes),
				}
			}
		}
	}

	for _, assoc := range s.model.GetClassAssociations() {
		var acKey *identity.Key
		if assoc.AssociationClassKey != nil {
			k := *assoc.AssociationClassKey
			acKey = &k
		}
		s.associations[assoc.Key] = Association{
			Key:                 assoc.Key,
			Name:                assoc.Name,
			FromClassKey:        assoc.FromClassKey,
			ToClassKey:          assoc.ToClassKey,
			AssociationClassKey: acKey,
			FromMultiplicity:    assoc.FromMultiplicity,
			ToMultiplicity:      assoc.ToMultiplicity,
		}
	}
}

// CoreModel returns the owned model for this run.
//
// This is the only legitimate *core.Model for simulator components during a run.
// Prefer Schema methods (Class, Association, …) when they cover the need.
// Callers must not mutate the returned model.
//
// Migration note: catalog, checkers, and expression setup still consume CoreModel
// until they are rewritten against Schema-only APIs. New code should not store the
// pointer beyond the construction of those components.
func (s *Schema) CoreModel() *core.Model {
	if s == nil {
		return nil
	}
	return s.model
}

// IsClassInScope reports whether classKey is registered on this schema.
func (s *Schema) IsClassInScope(classKey identity.Key) bool {
	if s == nil {
		return false
	}
	_, ok := s.classes[classKey]
	return ok
}

// Class returns a run-facing view of a class.
func (s *Schema) Class(classKey identity.Key) (Class, bool) {
	if s == nil {
		return Class{}, false
	}
	c, ok := s.classes[classKey]
	return c, ok
}

// ModelClass returns the full model class for classKey, if present on the owned model.
func (s *Schema) ModelClass(classKey identity.Key) (model_class.Class, bool) {
	if s == nil || s.model == nil {
		return model_class.Class{}, false
	}
	for _, domain := range s.model.Domains {
		for _, subdomain := range domain.Subdomains {
			if class, ok := subdomain.Classes[classKey]; ok {
				return class, true
			}
		}
	}
	return model_class.Class{}, false
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

// Association returns a run-facing view of an association.
func (s *Schema) Association(assocKey identity.Key) (Association, bool) {
	if s == nil {
		return Association{}, false
	}
	a, ok := s.associations[assocKey]
	return a, ok
}

// ModelAssociation returns the full model association if present.
func (s *Schema) ModelAssociation(assocKey identity.Key) (model_class.Association, bool) {
	if s == nil || s.model == nil {
		return model_class.Association{}, false
	}
	assocs := s.model.GetClassAssociations()
	a, ok := assocs[assocKey]
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
