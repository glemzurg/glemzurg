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

// ForEachInstanceOfClass calls fn for every live instance of classKey.
// fn must not call mutating State methods (may deadlock or race).
func (s *State) ForEachInstanceOfClass(classKey identity.Key, fn func(*Instance)) {
	if fn == nil {
		return
	}
	s.mu.RLock()
	var list []*Instance
	for _, inst := range s.instances {
		if inst.ClassKey == classKey {
			list = append(list, inst)
		}
	}
	s.mu.RUnlock()

	for _, inst := range list {
		fn(inst)
	}
}

// HasInstanceOfClass reports whether any live instance has the given class key.
func (s *State) HasInstanceOfClass(classKey identity.Key) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, inst := range s.instances {
		if inst.ClassKey == classKey {
			return true
		}
	}
	return false
}

// CountByClass returns how many live instances have the given class key.
func (s *State) CountByClass(classKey identity.Key) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	n := 0
	for _, inst := range s.instances {
		if inst.ClassKey == classKey {
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
		if inst.Attributes == rec || inst.Attributes == data {
			return inst.ID, true
		}
		if (data != nil && inst.Attributes.Equals(data)) || inst.Attributes.Equals(rec) {
			found = inst.ID
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
		for _, link := range s.links.GetAllForward(evaluator.ObjectID(inst.ID)) {
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
		id := evaluator.ObjectID(inst.ID)
		relCtx.EnsureInstance(id, inst.Attributes)
		relCtx.RegisterClassKey(id, inst.ClassKey.String())
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
		fromInst := s.GetInstance(link.FromEndpointID)
		linkInst := s.GetInstance(link.LinkInstanceID)
		toInst := s.GetInstance(link.ToEndpointID)
		if fromInst == nil || linkInst == nil || toInst == nil {
			return
		}
		hostKey := evaluator.AssociationKey(link.HostAssocKey.String())
		fromExtent, toExtent := createExtentLink(relCtx, hostKey, fromInst, toInst)
		linkExtent := object.NewExtentElement(uint64(linkInst.ID), linkInst.Attributes)
		relCtx.EnsureInstance(evaluator.ObjectID(linkInst.ID), linkInst.Attributes)
		relCtx.AddAssociationClassRow(hostKey, fromExtent, toExtent, linkExtent)
	})
}

func createExtentLink(
	relCtx *evaluator.RelationContext,
	assocKey evaluator.AssociationKey,
	fromInst, toInst *Instance,
) (fromExtent, toExtent *object.Record) {
	fromExtent = object.NewExtentElement(uint64(fromInst.ID), fromInst.Attributes)
	toExtent = object.NewExtentElement(uint64(toInst.ID), toInst.Attributes)
	relCtx.CreateInstanceLink(
		assocKey,
		evaluator.InstanceEndpoint{
			ID:     evaluator.ObjectID(fromInst.ID),
			Extent: fromExtent,
			Data:   fromInst.Attributes,
		},
		evaluator.InstanceEndpoint{
			ID:     evaluator.ObjectID(toInst.ID),
			Extent: toExtent,
			Data:   toInst.Attributes,
		},
	)
	relCtx.RegisterClassKey(evaluator.ObjectID(fromInst.ID), fromInst.ClassKey.String())
	relCtx.RegisterClassKey(evaluator.ObjectID(toInst.ID), toInst.ClassKey.String())
	return fromExtent, toExtent
}
