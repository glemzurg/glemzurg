package schema

import (
	"maps"

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
// Class and association values are [model_class.Class] and
// [model_class.Association] from the model tree — not parallel schema DTOs.
//
// Construct via [NewFromModel] only.
type Schema struct {
	// model is the authoritative static model for this run (always non-nil after NewFromModel).
	model *core.Model

	// Indexed views of model types for hot lookups (same values as in model).
	classes      map[identity.Key]model_class.Class
	associations map[identity.Key]model_class.Association
}

// NewFromModel takes ownership of model as the sole static model for a run.
// model must be non-nil. Typically pass the surface-filtered active model.
// The caller must not mutate model afterward and must not retain a separate
// model pointer for simulator use.
func NewFromModel(model *core.Model) *Schema {
	if model == nil {
		panic("schema.NewFromModel: model is required")
	}
	sch := &Schema{
		model:        model,
		classes:      make(map[identity.Key]model_class.Class),
		associations: make(map[identity.Key]model_class.Association),
	}
	sch.reindex()
	return sch
}

// reindex rebuilds lookup maps from the owned model.
func (s *Schema) reindex() {
	s.classes = make(map[identity.Key]model_class.Class)
	s.associations = make(map[identity.Key]model_class.Association)
	for _, domain := range s.model.Domains {
		for _, subdomain := range domain.Subdomains {
			maps.Copy(s.classes, subdomain.Classes)
		}
	}
	maps.Copy(s.associations, s.model.GetClassAssociations())
}

// EmptyModel returns a new empty *core.Model (no domains/classes) for building a
// Schema when tests need instance.State without surface content.
func EmptyModel() *core.Model {
	m := core.NewModel("empty", core.ModelDetails{Name: "empty", Details: ""}, "", nil, nil, nil)
	return &m
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

// Class returns the model class for classKey, if present on the owned model.
func (s *Schema) Class(classKey identity.Key) (model_class.Class, bool) {
	if s == nil {
		return model_class.Class{}, false
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

// Association returns the model association for assocKey, if present.
func (s *Schema) Association(assocKey identity.Key) (model_class.Association, bool) {
	if s == nil {
		return model_class.Association{}, false
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
