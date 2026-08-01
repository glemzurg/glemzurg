package instance

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// ForEachInstance calls fn for every live instance.
// fn must not call mutating State methods (may deadlock or race).
func (s *State) ForEachInstance(fn func(*Instance)) {
	if fn == nil {
		return
	}
	s.mu.RLock()
	list := make([]*Instance, 0, len(s.instances))
	for _, inst := range s.instances {
		list = append(list, inst)
	}
	s.mu.RUnlock()

	for _, inst := range list {
		fn(inst)
	}
}

// forEachInstanceOfClass calls fn for every live instance of classKey.
// fn must not call mutating State methods (may deadlock or race).
func (s *State) forEachInstanceOfClass(classKey identity.Key, fn func(*Instance)) {
	if fn == nil {
		return
	}
	s.mu.RLock()
	var list []*Instance
	for _, inst := range s.instances {
		if inst.GetClassKey() == classKey {
			list = append(list, inst)
		}
	}
	s.mu.RUnlock()

	for _, inst := range list {
		fn(inst)
	}
}

// hasInstanceOfClass reports whether any live instance has the given class key.
func (s *State) hasInstanceOfClass(classKey identity.Key) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, inst := range s.instances {
		if inst.GetClassKey() == classKey {
			return true
		}
	}
	return false
}

// countByClass returns how many live instances have the given class key.
func (s *State) countByClass(classKey identity.Key) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	n := 0
	for _, inst := range s.instances {
		if inst.GetClassKey() == classKey {
			n++
		}
	}
	return n
}

// LookupIDByRecord resolves a TLA record (extent or bare attributes) to a live instance ID.
// Prefers pointer identity on the attribute record; falls back to unique structural equality.
// Returns false when no match, or when structural equality matches more than one instance.
func (s *State) LookupIDByRecord(rec *object.Record) (ID, bool) {
	if rec == nil {
		return 0, false
	}

	if id, ok := object.ExtentID(rec); ok {
		instID := ID(id)
		s.mu.RLock()
		inst := s.instances[instID]
		s.mu.RUnlock()
		if inst != nil {
			return instID, true
		}
	}

	data := object.ExtentData(rec)

	s.mu.RLock()
	defer s.mu.RUnlock()

	var (
		found ID
		n     int
	)
	for _, inst := range s.instances {
		if inst.GetAttributes() == rec || inst.GetAttributes() == data {
			return inst.GetID(), true
		}
		if (data != nil && inst.GetAttributes().Equals(data)) || inst.GetAttributes().Equals(rec) {
			found = inst.GetID()
			n++
		}
	}
	if n == 1 {
		return found, true
	}
	return 0, false
}

// BinaryLink is one directed binary association edge between two instances.
type BinaryLink struct {
	AssocKey evaluator.AssociationKey
	FromID   ID
	ToID     ID
}

// ForEachBinaryLink calls fn for every binary association edge.
func (s *State) ForEachBinaryLink(fn func(BinaryLink)) {
	if fn == nil {
		return
	}
	s.mu.RLock()
	var edges []BinaryLink
	for _, inst := range s.instances {
		for _, link := range s.links.GetAllForward(evaluator.ObjectID(inst.GetID())) {
			edges = append(edges, BinaryLink{
				AssocKey: link.AssociationKey,
				FromID:   ID(link.FromID),
				ToID:     ID(link.ToID),
			})
		}
	}
	s.mu.RUnlock()

	for _, e := range edges {
		fn(e)
	}
}

// ForEachBinaryLinkOfAssociation calls fn for each edge of one association.
func (s *State) ForEachBinaryLinkOfAssociation(assocKey identity.Key, fn func(fromID, toID ID)) {
	if fn == nil {
		return
	}
	want := evaluator.AssociationKey(assocKey.String())
	s.ForEachBinaryLink(func(edge BinaryLink) {
		if edge.AssocKey == want {
			fn(edge.FromID, edge.ToID)
		}
	})
}

// ForEachAssociationLink calls fn for every association-class host row.
func (s *State) ForEachAssociationLink(fn func(AssociationLink)) {
	if fn == nil {
		return
	}
	s.mu.RLock()
	links := s.associationLinks.AllLinks()
	s.mu.RUnlock()

	for _, link := range links {
		fn(link)
	}
}

// ForEachAssociationLinkOfHost calls fn for host rows of one host association.
func (s *State) ForEachAssociationLinkOfHost(hostAssocKey identity.Key, fn func(AssociationLink)) {
	if fn == nil {
		return
	}
	s.ForEachAssociationLink(func(link AssociationLink) {
		if link.HostAssocKey == hostAssocKey {
			fn(link)
		}
	})
}

// ProjectToRelationContext rebuilds runtime identity, binary links, and association-class
// rows on relCtx from this state. The caller should Clear() relCtx first so association
// metadata registered on the builder is preserved while runtime graphs are refreshed.
func (s *State) ProjectToRelationContext(relCtx *evaluator.RelationContext) {
	if relCtx == nil {
		return
	}
	s.projectInstancesToRelationContext(relCtx)
	s.projectBinaryLinksToRelationContext(relCtx)
	s.projectAssociationLinksToRelationContext(relCtx)
}

func (s *State) projectInstancesToRelationContext(relCtx *evaluator.RelationContext) {
	s.ForEachInstance(func(inst *Instance) {
		id := evaluator.ObjectID(inst.GetID())
		relCtx.EnsureInstance(id, inst.GetAttributes())
		relCtx.RegisterClassKey(id, inst.GetClassKey().String())
	})
}

func (s *State) projectBinaryLinksToRelationContext(relCtx *evaluator.RelationContext) {
	s.ForEachBinaryLink(func(edge BinaryLink) {
		fromInst := s.GetInstance(edge.FromID)
		toInst := s.GetInstance(edge.ToID)
		if fromInst == nil || toInst == nil {
			return
		}
		createExtentLink(relCtx, edge.AssocKey, fromInst, toInst)
	})
}

func (s *State) projectAssociationLinksToRelationContext(relCtx *evaluator.RelationContext) {
	s.ForEachAssociationLink(func(link AssociationLink) {
		s.projectOneAssociationLink(relCtx, link)
	})
}

// projectOneAssociationLink projects a single AC host row into relation context.
// Host image is derived from AC rows only — no parallel binary host link.
func (s *State) projectOneAssociationLink(relCtx *evaluator.RelationContext, link AssociationLink) {
	fromInst := s.GetInstance(link.FromEndpointID)
	linkInst := s.GetInstance(link.LinkInstanceID)
	toInst := s.GetInstance(link.ToEndpointID)
	if fromInst == nil || linkInst == nil || toInst == nil {
		return
	}
	fromExtent := registerProjectedEndpoint(relCtx, fromInst)
	toExtent := registerProjectedEndpoint(relCtx, toInst)
	if fromExtent == nil || toExtent == nil {
		return
	}
	hostKey := evaluator.AssociationKey(link.HostAssocKey.String())
	linkExtent := registerProjectedLinkInstance(relCtx, linkInst)
	relCtx.AddAssociationClassRow(hostKey, fromExtent, toExtent, linkExtent)
}

// registerProjectedEndpoint ensures a live instance is identity-visible for AC host projection.
// Returns the TLA-visible extent record used as an AC row endpoint key.
func registerProjectedEndpoint(relCtx *evaluator.RelationContext, inst *Instance) *object.Record {
	id := evaluator.ObjectID(inst.GetID())
	relCtx.EnsureInstance(id, inst.GetAttributes())
	relCtx.RegisterClassKey(id, inst.GetClassKey().String())
	return relCtx.VisibleRecord(id)
}

func registerProjectedLinkInstance(relCtx *evaluator.RelationContext, linkInst *Instance) *object.Record {
	linkExtent := object.NewExtentElement(uint64(linkInst.GetID()), linkInst.GetAttributes())
	relCtx.EnsureInstance(evaluator.ObjectID(linkInst.GetID()), linkInst.GetAttributes())
	relCtx.RegisterClassKey(evaluator.ObjectID(linkInst.GetID()), linkInst.GetClassKey().String())
	if visible := relCtx.VisibleRecord(evaluator.ObjectID(linkInst.GetID())); visible != nil {
		return visible
	}
	return linkExtent
}

func createExtentLink(
	relCtx *evaluator.RelationContext,
	assocKey evaluator.AssociationKey,
	fromInst, toInst *Instance,
) (fromExtent, toExtent *object.Record) {
	fromExtent = object.NewExtentElement(uint64(fromInst.GetID()), fromInst.GetAttributes())
	toExtent = object.NewExtentElement(uint64(toInst.GetID()), toInst.GetAttributes())
	relCtx.CreateInstanceLink(
		assocKey,
		evaluator.InstanceEndpoint{
			ID:     evaluator.ObjectID(fromInst.GetID()),
			Extent: fromExtent,
			Data:   fromInst.GetAttributes(),
		},
		evaluator.InstanceEndpoint{
			ID:     evaluator.ObjectID(toInst.GetID()),
			Extent: toExtent,
			Data:   toInst.GetAttributes(),
		},
	)
	relCtx.RegisterClassKey(evaluator.ObjectID(fromInst.GetID()), fromInst.GetClassKey().String())
	relCtx.RegisterClassKey(evaluator.ObjectID(toInst.GetID()), toInst.GetClassKey().String())
	return fromExtent, toExtent
}
