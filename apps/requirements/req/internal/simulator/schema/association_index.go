package schema

import (
	"fmt"
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// AssociationView is a both-ends-in-scope association with multiplicity summary.
// Built once at schema construction; callers must not mutate Association.
type AssociationView struct {
	Association   model_class.Association
	FromClassKey  identity.Key
	ToClassKey    identity.Key
	MandatoryTo   bool
	MandatoryFrom bool
	MinTo         uint
	MinFrom       uint
}

// HostAssociationInfo links an in-scope association-class to its host association.
type HostAssociationInfo struct {
	AssociationClassKey identity.Key
	HostAssociation     model_class.Association
	FromClassKey        identity.Key
	ToClassKey          identity.Key
}

// UniquenessBinding is a scoped association that declares uniqueness.
type UniquenessBinding struct {
	Association model_class.Association
	Uniqueness  model_class.AssociationUniqueness
}

// reindexAssociationGraph builds both-ends-in-scope projections after class scope is known.
func (s *Schema) reindexAssociationGraph() {
	s.scopedAssociations = nil
	s.assocsByClass = make(map[identity.Key][]AssociationView)
	s.hostByACKey = make(map[identity.Key]*HostAssociationInfo)
	s.uniquenessBindings = nil
	s.assocsWithInvariants = nil

	var scoped []AssociationView
	for _, assoc := range s.associations {
		if !s.inScope[assoc.FromClassKey] || !s.inScope[assoc.ToClassKey] {
			continue
		}
		view := associationViewFrom(*assoc)
		scoped = append(scoped, view)

		if assoc.Uniqueness != nil {
			s.uniquenessBindings = append(s.uniquenessBindings, UniquenessBinding{
				Association: *assoc,
				Uniqueness:  *assoc.Uniqueness,
			})
		}
		if len(assoc.Invariants) > 0 {
			s.assocsWithInvariants = append(s.assocsWithInvariants, *assoc)
		}
		if assoc.AssociationClassKey != nil {
			acKey := *assoc.AssociationClassKey
			if s.inScope[acKey] {
				info := &HostAssociationInfo{
					AssociationClassKey: acKey,
					HostAssociation:     *assoc,
					FromClassKey:        assoc.FromClassKey,
					ToClassKey:          assoc.ToClassKey,
				}
				s.hostByACKey[acKey] = info
			}
		}
	}

	sort.Slice(scoped, func(i, j int) bool {
		return scoped[i].Association.Key.String() < scoped[j].Association.Key.String()
	})
	s.scopedAssociations = scoped

	for _, view := range scoped {
		s.assocsByClass[view.FromClassKey] = append(s.assocsByClass[view.FromClassKey], view)
		if view.FromClassKey != view.ToClassKey {
			s.assocsByClass[view.ToClassKey] = append(s.assocsByClass[view.ToClassKey], view)
		}
	}
	for classKey, views := range s.assocsByClass {
		sort.Slice(views, func(i, j int) bool {
			return views[i].Association.Key.String() < views[j].Association.Key.String()
		})
		s.assocsByClass[classKey] = views
	}

	sort.Slice(s.uniquenessBindings, func(i, j int) bool {
		return s.uniquenessBindings[i].Association.Key.String() < s.uniquenessBindings[j].Association.Key.String()
	})
	sort.Slice(s.assocsWithInvariants, func(i, j int) bool {
		return s.assocsWithInvariants[i].Key.String() < s.assocsWithInvariants[j].Key.String()
	})
}

func associationViewFrom(assoc model_class.Association) AssociationView {
	return AssociationView{
		Association:   assoc,
		FromClassKey:  assoc.FromClassKey,
		ToClassKey:    assoc.ToClassKey,
		MandatoryTo:   assoc.ToMultiplicity.LowerBound >= 1,
		MandatoryFrom: assoc.FromMultiplicity.LowerBound >= 1,
		MinTo:         assoc.ToMultiplicity.LowerBound,
		MinFrom:       assoc.FromMultiplicity.LowerBound,
	}
}

// ScopedAssociations returns both-ends-in-scope associations (sorted by key).
func (s *Schema) ScopedAssociations() []AssociationView {
	if s == nil || len(s.scopedAssociations) == 0 {
		return nil
	}
	out := make([]AssociationView, len(s.scopedAssociations))
	copy(out, s.scopedAssociations)
	return out
}

// AssociationsForClass returns scoped association views involving classKey.
// Out of scope class: (nil, false, nil). Unknown class: error.
func (s *Schema) AssociationsForClass(classKey identity.Key) ([]AssociationView, bool, error) {
	if s == nil {
		return nil, false, fmt.Errorf("schema.AssociationsForClass: nil schema")
	}
	if _, ok := s.classes[classKey]; !ok {
		return nil, false, fmt.Errorf("unknown class: %s", classKey.String())
	}
	if !s.inScope[classKey] {
		return nil, false, nil
	}
	views := s.assocsByClass[classKey]
	if len(views) == 0 {
		return nil, true, nil
	}
	out := make([]AssociationView, len(views))
	copy(out, views)
	return out, true, nil
}

// HostAssociationForAC returns host association metadata for an association-class key.
// Out of scope / not an AC: (nil, false, nil). Unknown key: error if class unknown.
func (s *Schema) HostAssociationForAC(acKey identity.Key) (*HostAssociationInfo, bool, error) {
	if s == nil {
		return nil, false, fmt.Errorf("schema.HostAssociationForAC: nil schema")
	}
	if _, ok := s.classes[acKey]; !ok {
		return nil, false, fmt.Errorf("unknown class: %s", acKey.String())
	}
	info, ok := s.hostByACKey[acKey]
	if !ok {
		// Known class but not an in-scope AC host role.
		return nil, false, nil
	}
	return info, true, nil
}

// AssociationsWithUniqueness returns scoped associations that declare uniqueness.
func (s *Schema) AssociationsWithUniqueness() []UniquenessBinding {
	if s == nil || len(s.uniquenessBindings) == 0 {
		return nil
	}
	out := make([]UniquenessBinding, len(s.uniquenessBindings))
	copy(out, s.uniquenessBindings)
	return out
}

// AssociationsWithInvariants returns scoped associations that author instance.
func (s *Schema) AssociationsWithInvariants() []model_class.Association {
	if s == nil || len(s.assocsWithInvariants) == 0 {
		return nil
	}
	out := make([]model_class.Association, len(s.assocsWithInvariants))
	copy(out, s.assocsWithInvariants)
	return out
}

// allAssociationsMap returns every association on the full model by key (including OOS).
// Used only inside schema reindex / caller-graph install.
func (s *Schema) allAssociationsMap() map[identity.Key]model_class.Association {
	if s == nil || len(s.associations) == 0 {
		return nil
	}
	out := make(map[identity.Key]model_class.Association, len(s.associations))
	for k, a := range s.associations {
		out[k] = *a
	}
	return out
}
