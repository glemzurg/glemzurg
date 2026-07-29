package schema

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// RunScope is the set of classes included in a simulation run.
// It is intake-only: pass it to [New], then discard it — schema answers scope
// questions via lookup triples and scoped bulk APIs.
//
// A zero value (nil classKeys) means every class in the model is in scope.
// A non-nil map is the exact in-scope set (keys absent from the model are ignored).
type RunScope struct {
	classKeys map[identity.Key]struct{}
}

// RunScopeAll returns a scope that includes every class in the model.
func RunScopeAll() RunScope {
	return RunScope{}
}

// NewRunScope builds a scope from an explicit class key list.
// An empty list yields RunScopeAll (same as zero value) so callers that
// resolved "simulate everything" need not special-case.
func NewRunScope(classKeys []identity.Key) RunScope {
	if len(classKeys) == 0 {
		return RunScopeAll()
	}
	m := make(map[identity.Key]struct{}, len(classKeys))
	for _, k := range classKeys {
		m[k] = struct{}{}
	}
	return RunScope{classKeys: m}
}

// Schema is the sole home of model facts and run scope for one simulation run.
//
// Construction is the intake gate: pass a full *core.Model and [RunScope] into
// [New]. After that, the running simulator obtains model data and scope only
// through Schema (or components built from Schema). Do not keep a second
// *core.Model or include-list as authority for the same run.
//
// Schema is immutable for the run: do not mutate the underlying model after New.
//
// Construct via [New] only.
type Schema struct {
	// model is the authoritative static model for this run (always non-nil after New).
	model *core.Model

	// Full-universe indexes (stable pointers; values never mutated after New).
	classes      map[identity.Key]*model_class.Class
	associations map[identity.Key]*model_class.Association

	// inScope marks classes that participate in this run (may hold instances).
	inScope map[identity.Key]bool

	// extentNames maps every known class key to its TLA extent name (in and out of scope).
	extentNames map[identity.Key]string

	// boundaryAssociations have exactly one endpoint in scope (link no-ops / empty peers).
	boundaryAssociations []*model_class.Association

	// Association graph projections (both ends in scope unless noted).
	scopedAssociations   []AssociationView
	assocsByClass        map[identity.Key][]AssociationView
	hostByACKey          map[identity.Key]*HostAssociationInfo
	uniquenessBindings   []UniquenessBinding
	assocsWithInvariants []model_class.Association

	// Class simulation metadata (in-scope only).
	classSim map[identity.Key]*ClassSimInfo

	// Attribute / derived projections (in-scope classes).
	classIndexes   map[identity.Key][]IndexDefinition
	attrsBySubKey  map[identity.Key]map[string]*model_class.Attribute
	derivedByClass map[identity.Key][]DerivedAttrDef
	derivedByKey   map[identity.Key]DerivedAttrDef

	// modelInvariants is the run-scoped model-level invariant list (may differ from model.Invariants).
	modelInvariants []model_logic.Logic

	// catalog holds association navigation, caller graphs, and surface-unavailable indexes.
	catalog *catalog
}

// New takes ownership of model as the sole static model for a run and internalizes scope.
// model must be non-nil. The caller must not mutate model afterward and must not retain
// a separate model pointer or include-list as simulator authority.
//
// scope zero value (RunScopeAll) marks every class in the model as in-scope.
func New(model *core.Model, scope RunScope) *Schema {
	if model == nil {
		panic("schema.New: model is required")
	}
	sch := &Schema{
		model:        model,
		classes:      make(map[identity.Key]*model_class.Class),
		associations: make(map[identity.Key]*model_class.Association),
		inScope:      make(map[identity.Key]bool),
		extentNames:  make(map[identity.Key]string),
	}
	sch.reindex(scope)
	// Default model invariants from the owned model; surface wiring may replace via SetModelInvariants.
	sch.modelInvariants = model.Invariants
	return sch
}

// SetModelInvariants replaces the run-scoped model-level invariant list (surface filtering).
func (s *Schema) SetModelInvariants(invs []model_logic.Logic) {
	if s == nil {
		return
	}
	s.modelInvariants = invs
}

// ReplaceInScopeClass replaces the stored body of an in-scope class (e.g. surface-scoped invariants).
// Returns an error if the class is unknown or out of scope.
func (s *Schema) ReplaceInScopeClass(class model_class.Class) error {
	if s == nil {
		return fmt.Errorf("schema.ReplaceInScopeClass: nil schema")
	}
	if _, ok := s.classes[class.Key]; !ok {
		return fmt.Errorf("unknown class: %s", class.Key.String())
	}
	if !s.inScope[class.Key] {
		return fmt.Errorf("class not in run scope: %s", class.Key.String())
	}
	c := class
	s.classes[class.Key] = &c
	s.extentNames[class.Key] = model_class.ClassTLAName(class.Name)
	// Class body changed (e.g. surface-scoped invariants); rebuild projections that depend on it.
	s.classSim[class.Key] = buildClassSimInfo(c)
	s.indexClassAttributes(&c)
	// Caller graphs and association nav depend on class bodies; rebuild catalog.
	s.catalog = newCatalog(s)
	return nil
}

// reindex rebuilds lookup maps and scope partitions from the owned model.
func (s *Schema) reindex(scope RunScope) {
	s.classes = make(map[identity.Key]*model_class.Class)
	s.associations = make(map[identity.Key]*model_class.Association)
	s.inScope = make(map[identity.Key]bool)
	s.extentNames = make(map[identity.Key]string)
	s.boundaryAssociations = nil

	for _, domain := range s.model.Domains {
		for _, subdomain := range domain.Subdomains {
			for key, class := range subdomain.Classes {
				c := class // stable copy for pointer
				s.classes[key] = &c
				s.extentNames[key] = model_class.ClassTLAName(class.Name)
			}
		}
	}

	for key, assoc := range s.model.GetClassAssociations() {
		a := assoc
		s.associations[key] = &a
	}

	// Apply scope: nil classKeys ⇒ all known classes in scope.
	if scope.classKeys == nil {
		for key := range s.classes {
			s.inScope[key] = true
		}
	} else {
		for key := range scope.classKeys {
			if _, ok := s.classes[key]; ok {
				s.inScope[key] = true
			}
		}
	}

	s.reindexBoundaryAssociations()
	s.reindexAssociationGraph()
	s.reindexClassSim()
	s.reindexAttributeProjections()
	s.catalog = newCatalog(s)
}

func (s *Schema) reindexBoundaryAssociations() {
	var boundary []*model_class.Association
	for _, assoc := range s.associations {
		fromIn := s.inScope[assoc.FromClassKey]
		toIn := s.inScope[assoc.ToClassKey]
		if fromIn == toIn {
			continue // both in or both out
		}
		// Boundary: keep AC key only when the association class is in scope.
		surfaceAssoc := *assoc
		if surfaceAssoc.AssociationClassKey != nil {
			if !s.inScope[*surfaceAssoc.AssociationClassKey] {
				surfaceAssoc.AssociationClassKey = nil
			}
		}
		a := surfaceAssoc
		boundary = append(boundary, &a)
	}
	s.boundaryAssociations = boundary
}

// Class returns the class when it is in run scope.
// Out of scope but known: (nil, false, nil). Unknown key: error.
func (s *Schema) Class(classKey identity.Key) (*model_class.Class, bool, error) {
	if s == nil {
		return nil, false, fmt.Errorf("schema.Class: nil schema")
	}
	c, ok := s.classes[classKey]
	if !ok {
		return nil, false, fmt.Errorf("unknown class: %s", classKey.String())
	}
	if !s.inScope[classKey] {
		return nil, false, nil
	}
	return c, true, nil
}

// Association returns the association when both endpoints are in run scope.
// Known association with an out-of-scope endpoint: (nil, false, nil).
// Unknown key: error.
//
// Boundary associations (exactly one end in scope) are not returned here;
// use [BoundaryAssociations].
func (s *Schema) Association(assocKey identity.Key) (*model_class.Association, bool, error) {
	if s == nil {
		return nil, false, fmt.Errorf("schema.Association: nil schema")
	}
	a, ok := s.associations[assocKey]
	if !ok {
		return nil, false, fmt.Errorf("unknown association: %s", assocKey.String())
	}
	if !s.inScope[a.FromClassKey] || !s.inScope[a.ToClassKey] {
		return nil, false, nil
	}
	return a, true, nil
}

// ExtentName returns the TLA extent name for a class key.
// Out-of-scope classes still return a name (for empty-set bindings) with inScope false.
// Unknown key: error.
func (s *Schema) ExtentName(classKey identity.Key) (name string, inScope bool, err error) {
	if s == nil {
		return "", false, fmt.Errorf("schema.ExtentName: nil schema")
	}
	name, ok := s.extentNames[classKey]
	if !ok {
		return "", false, fmt.Errorf("unknown class: %s", classKey.String())
	}
	return name, s.inScope[classKey], nil
}

// BoundaryAssociations returns associations with exactly one endpoint in scope.
// Callers must not mutate the returned associations.
func (s *Schema) BoundaryAssociations() []*model_class.Association {
	if s == nil || len(s.boundaryAssociations) == 0 {
		return nil
	}
	out := make([]*model_class.Association, len(s.boundaryAssociations))
	copy(out, s.boundaryAssociations)
	return out
}

// ForEachExtent calls fn for every class known to the schema (in and out of scope).
func (s *Schema) ForEachExtent(fn func(classKey identity.Key, name string, inScope bool)) {
	if s == nil || fn == nil {
		return
	}
	for key, name := range s.extentNames {
		fn(key, name, s.inScope[key])
	}
}

// IsClassInScope reports whether classKey participates in this run.
// Prefer Class(key) for access; this remains for transitional call sites.
func (s *Schema) IsClassInScope(classKey identity.Key) bool {
	if s == nil {
		return false
	}
	return s.inScope[classKey]
}

// class returns the model class for classKey if in scope (legacy unexported helper).
func (s *Schema) class(classKey identity.Key) (model_class.Class, bool) {
	c, inScope, err := s.Class(classKey)
	if err != nil || !inScope || c == nil {
		return model_class.Class{}, false
	}
	return *c, true
}

// attributes returns the attribute definitions for an in-scope class, or nil.
func (s *Schema) attributes(classKey identity.Key) []model_class.Attribute {
	c, ok := s.class(classKey)
	if !ok {
		return nil
	}
	return c.Attributes
}

// classKeys returns every in-scope class key (order is not significant).
func (s *Schema) classKeys() []identity.Key {
	if s == nil || len(s.inScope) == 0 {
		return nil
	}
	keys := make([]identity.Key, 0, len(s.inScope))
	for k, ok := range s.inScope {
		if ok {
			keys = append(keys, k)
		}
	}
	return keys
}

// association returns the association when both ends are in scope (legacy helper).
func (s *Schema) association(assocKey identity.Key) (model_class.Association, bool) {
	a, inScope, err := s.Association(assocKey)
	if err != nil || !inScope || a == nil {
		return model_class.Association{}, false
	}
	return *a, true
}

// isAssociationClass reports whether classKey is an association-class for some association.
func (s *Schema) isAssociationClass(classKey identity.Key) bool {
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

// associationKeys returns every association key on the full model (order not significant).
func (s *Schema) associationKeys() []identity.Key {
	if s == nil || len(s.associations) == 0 {
		return nil
	}
	keys := make([]identity.Key, 0, len(s.associations))
	for k := range s.associations {
		keys = append(keys, k)
	}
	return keys
}

// forEachClassInScope calls fn for every in-scope class (internal bulk walk).
func (s *Schema) forEachClassInScope(fn func(*model_class.Class)) {
	if s == nil || fn == nil {
		return
	}
	for key, c := range s.classes {
		if s.inScope[key] {
			fn(c)
		}
	}
}
