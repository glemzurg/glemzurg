package schema

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
)

// AssociationBinder receives association registrations for relation/eval bindings.
// Implemented by state.BindingsBuilder.
type AssociationBinder interface {
	AddAssociationClassHost(
		assocKey identity.Key,
		assocName string,
		endpoints evaluator.AssociationHostEndpoints,
		linkClassName string,
		mults evaluator.AssociationHostMultiplicities,
	)
	AddAssociation(
		assocKey identity.Key,
		assocName string,
		fromClassKey, toClassKey identity.Key,
		fromMult, toMult evaluator.Multiplicity,
	)
}

// RegisterAssociationBindings wires every catalog association (scoped + boundary)
// into the binder once at setup. Callers do not dump associations themselves.
func (s *Schema) RegisterAssociationBindings(b AssociationBinder) {
	if s == nil || b == nil {
		return
	}
	for _, ai := range s.mustCatalog().allAssociations() {
		assoc := ai.Association
		fromMult := evaluator.Multiplicity{
			LowerBound:  assoc.FromMultiplicity.LowerBound,
			HigherBound: assoc.FromMultiplicity.HigherBound,
		}
		toMult := evaluator.Multiplicity{
			LowerBound:  assoc.ToMultiplicity.LowerBound,
			HigherBound: assoc.ToMultiplicity.HigherBound,
		}
		if assoc.AssociationClassKey != nil {
			if linkInfo := s.GetClassInfo(*assoc.AssociationClassKey); linkInfo != nil {
				b.AddAssociationClassHost(
					assoc.Key,
					assoc.Name,
					evaluator.AssociationHostEndpoints{
						FromClassKey: assoc.FromClassKey.String(),
						ToClassKey:   assoc.ToClassKey.String(),
					},
					linkInfo.Class.Name,
					evaluator.AssociationHostMultiplicities{From: fromMult, To: toMult},
				)
				continue
			}
		}
		b.AddAssociation(
			assoc.Key,
			assoc.Name,
			assoc.FromClassKey,
			assoc.ToClassKey,
			fromMult,
			toMult,
		)
	}
}

// ResolveObjectClassRef maps an object-of class reference to a class key.
// Prefers in-scope classes; falls back to full extent names (out-of-scope).
func (s *Schema) ResolveObjectClassRef(objectClassRef string) (key identity.Key, inScope bool, ok bool) {
	if s == nil || objectClassRef == "" {
		return identity.Key{}, false, false
	}
	want := identity.NormalizeSubKey(objectClassRef)
	var found identity.Key
	var foundInScope bool
	var matched bool
	s.EachInScopeClassSim(func(info *ClassSimInfo) {
		if matched || info == nil {
			return
		}
		if objectClassRefMatches(want, objectClassRef, info) {
			found = info.ClassKey
			foundInScope = true
			matched = true
		}
	})
	if matched {
		return found, foundInScope, true
	}
	// Known only as out-of-scope extent.
	for classKey, tlaName := range s.ClassNameMap() {
		if classKey.SubKey == objectClassRef || classKey.String() == objectClassRef {
			return classKey, false, true
		}
		if identity.NormalizeSubKey(tlaName) == want || tlaName == objectClassRef {
			return classKey, false, true
		}
	}
	return identity.Key{}, false, false
}

func objectClassRefMatches(wantNorm, objectClassRef string, info *ClassSimInfo) bool {
	if info == nil {
		return false
	}
	if info.ClassKey.SubKey == objectClassRef || info.ClassKey.String() == objectClassRef {
		return true
	}
	if identity.NormalizeSubKey(info.Class.Name) == wantNorm {
		return true
	}
	if model_class.ClassTLAName(info.Class.Name) == objectClassRef {
		return true
	}
	return identity.NormalizeSubKey(model_class.ClassTLAName(info.Class.Name)) == wantNorm
}
