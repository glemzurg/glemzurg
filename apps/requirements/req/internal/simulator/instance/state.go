package instance

import (
	"fmt"
	"maps"
	"sync"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// State is the mutable world for one simulation run: instances, association
// links, state-machine positions, and identity mappings.
type State struct {
	mu sync.RWMutex

	instances map[ID]*Instance

	// links tracks binary association relationships between instances.
	links *evaluator.LinkTable

	// associationLinks tracks host associations materialized by association-class instances.
	associationLinks *AssociationLinkTable

	// stateMachineStates maps instance IDs to their current state machine state key.
	stateMachineStates map[ID]identity.Key

	nextID ID

	// identityRegistry maps between object.Record pointers and evaluator ObjectIDs.
	identityRegistry *evaluator.IdentityRegistry
}

// NewState creates a new empty simulation state.
func NewState() *State {
	return &State{
		instances:          make(map[ID]*Instance),
		links:              evaluator.NewLinkTable(),
		associationLinks:   NewAssociationLinkTable(),
		stateMachineStates: make(map[ID]identity.Key),
		nextID:             1, // Start at 1 so 0 can indicate "no instance"
		identityRegistry:   evaluator.NewIdentityRegistry(),
	}
}

// CreateInstance creates a new class instance with the given attributes.
// Returns the newly created instance. Attribute data is cloned into the store.
func (s *State) CreateInstance(classKey identity.Key, attributes *object.Record) *Instance {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.nextID
	s.nextID++

	inst := &Instance{
		ID:         id,
		ClassKey:   classKey,
		Attributes: attributes.Clone().(*object.Record),
	}

	s.instances[id] = inst

	// Register with identity registry for evaluator integration.
	s.identityRegistry.GetOrAssign(inst.Attributes)

	return inst
}

// GetInstance retrieves an instance by ID.
// Returns nil if the instance does not exist.
func (s *State) GetInstance(id ID) *Instance {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.instances[id]
}

// UpdateInstance updates an instance's attributes.
// Returns an error if the instance does not exist.
func (s *State) UpdateInstance(id ID, attributes *object.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	inst, ok := s.instances[id]
	if !ok {
		return fmt.Errorf("instance %d not found", id)
	}

	inst.Attributes = attributes.Clone().(*object.Record)
	return nil
}

// UpdateInstanceField updates a single field on an instance.
// Returns an error if the instance does not exist.
func (s *State) UpdateInstanceField(id ID, fieldName string, value object.Object) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	inst, ok := s.instances[id]
	if !ok {
		return fmt.Errorf("instance %d not found", id)
	}

	inst.Attributes.Set(fieldName, value)
	return nil
}

// DeleteInstance removes an instance and all its links and state-machine entry.
// Returns an error if the instance does not exist.
func (s *State) DeleteInstance(id ID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	inst, ok := s.instances[id]
	if !ok {
		return fmt.Errorf("instance %d not found", id)
	}

	s.removeAllLinks(id)
	s.associationLinks.RemoveInstance(id)
	delete(s.stateMachineStates, id)
	delete(s.instances, id)

	// Keep identity registry entries to avoid ID reuse issues.
	_ = inst
	return nil
}

// removeAllLinks removes all binary links to/from an instance (lock must be held).
func (s *State) removeAllLinks(id ID) {
	objID := evaluator.ObjectID(id)

	for _, link := range s.links.GetAllForward(objID) {
		s.links.RemoveLink(link.AssociationKey, link.FromID, link.ToID)
	}
	for _, link := range s.links.GetAllReverse(objID) {
		s.links.RemoveLink(link.AssociationKey, link.FromID, link.ToID)
	}
}

// InstanceCount returns the number of instances.
func (s *State) InstanceCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.instances)
}

// InstancesByClass returns all instances of a specific class.
func (s *State) InstancesByClass(classKey identity.Key) []*Instance {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var out []*Instance
	for _, inst := range s.instances {
		if inst.ClassKey == classKey {
			out = append(out, inst)
		}
	}
	return out
}

// AddLink creates a link between two instances for an association.
// Returns an error when the association already links the instance pair.
func (s *State) AddLink(assocKey identity.Key, fromID, toID ID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.links.AddLink(
		evaluator.AssociationKey(assocKey.String()),
		evaluator.ObjectID(fromID),
		evaluator.ObjectID(toID),
	)
}

// RemoveLink removes a link between two instances.
// Returns true if a link was removed.
func (s *State) RemoveLink(assocKey identity.Key, fromID, toID ID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.links.RemoveLink(
		evaluator.AssociationKey(assocKey.String()),
		evaluator.ObjectID(fromID),
		evaluator.ObjectID(toID),
	)
}

// GetLinkedForward returns instance IDs linked FROM the given instance
// for a specific association.
func (s *State) GetLinkedForward(fromID ID, assocKey identity.Key) []ID {
	s.mu.RLock()
	defer s.mu.RUnlock()

	objIDs := s.links.GetForward(
		evaluator.ObjectID(fromID),
		evaluator.AssociationKey(assocKey.String()),
	)

	ids := make([]ID, len(objIDs))
	for i, objID := range objIDs {
		ids[i] = ID(objID)
	}
	return ids
}

// GetLinkedReverse returns instance IDs linked TO the given instance
// for a specific association.
func (s *State) GetLinkedReverse(toID ID, assocKey identity.Key) []ID {
	s.mu.RLock()
	defer s.mu.RUnlock()

	objIDs := s.links.GetReverse(
		evaluator.ObjectID(toID),
		evaluator.AssociationKey(assocKey.String()),
	)

	ids := make([]ID, len(objIDs))
	for i, objID := range objIDs {
		ids[i] = ID(objID)
	}
	return ids
}

// CountActivePairLinks counts links for one association between a from/to instance pair.
// Only instances still present in simulation state count; Final transitions remove rows entirely.
func (s *State) CountActivePairLinks(
	assoc model_class.Association,
	fromID, toID ID,
) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if assoc.AssociationClassKey != nil {
		links := s.associationLinks.LinksFromEndpoint(assoc.Key, fromID)
		count := 0
		for _, link := range links {
			if link.ToEndpointID != toID {
				continue
			}
			if s.instances[link.LinkInstanceID] != nil {
				count++
			}
		}
		return count
	}

	return s.links.CountPairLinks(
		evaluator.AssociationKey(assoc.Key.String()),
		evaluator.ObjectID(fromID),
		evaluator.ObjectID(toID),
	)
}

// LinkCount returns the total number of binary links.
func (s *State) LinkCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.links.Count()
}

// Links returns the underlying binary link table.
// Migration-era escape hatch; prefer State navigation methods when possible.
func (s *State) Links() *evaluator.LinkTable {
	return s.links
}

// AddAssociationLink materializes one host association row via an association-class instance.
// Returns an error when the host association already links the endpoint pair.
func (s *State) AddAssociationLink(
	hostAssocKey identity.Key,
	fromEndpointID ID,
	toEndpointID ID,
	linkInstanceID ID,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.associationLinks.AddLink(AssociationLink{
		HostAssocKey:   hostAssocKey,
		FromEndpointID: fromEndpointID,
		ToEndpointID:   toEndpointID,
		LinkInstanceID: linkInstanceID,
	})
}

// AssociationLinksFromEndpoint returns materialized host rows from a from-endpoint.
func (s *State) AssociationLinksFromEndpoint(hostAssocKey identity.Key, fromID ID) []AssociationLink {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.associationLinks.LinksFromEndpoint(hostAssocKey, fromID)
}

// AssociationLinksToEndpoint returns materialized host rows to a to-endpoint.
func (s *State) AssociationLinksToEndpoint(hostAssocKey identity.Key, toID ID) []AssociationLink {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.associationLinks.LinksToEndpoint(hostAssocKey, toID)
}

// AssociationLinkByInstance returns the host row for an association-class instance.
func (s *State) AssociationLinkByInstance(linkInstanceID ID) (AssociationLink, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.associationLinks.LinkByInstance(linkInstanceID)
}

// AssociationLinks returns the underlying association link table.
// Migration-era escape hatch; prefer State association methods when possible.
func (s *State) AssociationLinks() *AssociationLinkTable {
	return s.associationLinks
}

// SetStateMachineState sets the current state machine state for an instance.
func (s *State) SetStateMachineState(id ID, stateKey identity.Key) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.instances[id]; !ok {
		return fmt.Errorf("instance %d not found", id)
	}

	s.stateMachineStates[id] = stateKey
	return nil
}

// GetStateMachineState returns the current state machine state for an instance.
// Returns the zero value if the instance has no state machine state set.
func (s *State) GetStateMachineState(id ID) (identity.Key, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stateKey, ok := s.stateMachineStates[id]
	return stateKey, ok
}

// ClearStateMachineState removes the state machine state for an instance.
func (s *State) ClearStateMachineState(id ID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.stateMachineStates, id)
}

// IdentityRegistry returns the identity registry for evaluator integration.
func (s *State) IdentityRegistry() *evaluator.IdentityRegistry {
	return s.identityRegistry
}

// Clone creates a deep copy of the simulation state.
func (s *State) Clone() *State {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clone := NewState()
	clone.nextID = s.nextID

	for id, inst := range s.instances {
		clone.instances[id] = inst.Clone()
	}

	maps.Copy(clone.stateMachineStates, s.stateMachineStates)

	for _, inst := range s.instances {
		objID := evaluator.ObjectID(inst.ID)
		for _, link := range s.links.GetAllForward(objID) {
			if err := clone.links.AddLink(link.AssociationKey, link.FromID, link.ToID); err != nil {
				panic(fmt.Sprintf("clone link table: %v", err))
			}
		}
	}
	for _, link := range s.associationLinks.AllLinks() {
		if err := clone.associationLinks.AddLink(link); err != nil {
			panic(fmt.Sprintf("clone association link table: %v", err))
		}
	}

	return clone
}
