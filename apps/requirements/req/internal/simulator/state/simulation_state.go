// Package state provides runtime state management for TLA+ simulation.
// It tracks class instances, association links, and state machine states.
package state

import (
	"fmt"
	"maps"
	"sync"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// InstanceID uniquely identifies a class instance within a simulation.
type InstanceID uint64

// SimulationState holds the complete runtime state for a simulation.
// It tracks:
//   - All class instances with their current attribute values
//   - Association links between instances
//   - Current state machine states for each instance
type SimulationState struct {
	mu sync.RWMutex

	// instances maps instance IDs to class instances
	instances map[InstanceID]*ClassInstance

	// links tracks binary association relationships between instances.
	links *evaluator.LinkTable

	// associationLinks tracks host associations materialized by association-class instances.
	associationLinks *AssociationLinkTable

	// stateMachineStates maps instance IDs to their current state machine state
	// The value is the identity.Key of the current state
	stateMachineStates map[InstanceID]identity.Key

	// nextID is the next available instance ID
	nextID InstanceID

	// identityRegistry maps between object.Record pointers and ObjectIDs
	// This allows the evaluator to track relationships
	identityRegistry *evaluator.IdentityRegistry
}

// NewSimulationState creates a new empty simulation state.
func NewSimulationState() *SimulationState {
	return &SimulationState{
		instances:          make(map[InstanceID]*ClassInstance),
		links:              evaluator.NewLinkTable(),
		associationLinks:   NewAssociationLinkTable(),
		stateMachineStates: make(map[InstanceID]identity.Key),
		nextID:             1, // Start at 1 so 0 can indicate "no instance"
		identityRegistry:   evaluator.NewIdentityRegistry(),
	}
}

// CreateInstance creates a new class instance with the given attributes.
// Returns the newly created instance.
func (s *SimulationState) CreateInstance(classKey identity.Key, attributes *object.Record) *ClassInstance {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.nextID
	s.nextID++

	instance := &ClassInstance{
		ID:         id,
		ClassKey:   classKey,
		Attributes: attributes.Clone().(*object.Record),
	}

	s.instances[id] = instance

	// Register with identity registry for evaluator integration
	s.identityRegistry.GetOrAssign(instance.Attributes)

	return instance
}

// GetInstance retrieves an instance by ID.
// Returns nil if the instance doesn't exist.
func (s *SimulationState) GetInstance(id InstanceID) *ClassInstance {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.instances[id]
}

// UpdateInstance updates an instance's attributes.
// Returns an error if the instance doesn't exist.
func (s *SimulationState) UpdateInstance(id InstanceID, attributes *object.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	instance, ok := s.instances[id]
	if !ok {
		return fmt.Errorf("instance %d not found", id)
	}

	instance.Attributes = attributes.Clone().(*object.Record)
	return nil
}

// UpdateInstanceField updates a single field on an instance.
// Returns an error if the instance doesn't exist.
func (s *SimulationState) UpdateInstanceField(id InstanceID, fieldName string, value object.Object) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	instance, ok := s.instances[id]
	if !ok {
		return fmt.Errorf("instance %d not found", id)
	}

	instance.Attributes.Set(fieldName, value)
	return nil
}

// DeleteInstance removes an instance and all its links.
// Returns an error if the instance doesn't exist.
func (s *SimulationState) DeleteInstance(id InstanceID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	instance, ok := s.instances[id]
	if !ok {
		return fmt.Errorf("instance %d not found", id)
	}

	// Remove all links involving this instance
	s.removeAllLinks(id)
	s.associationLinks.RemoveInstance(id)

	// Remove from state machine states
	delete(s.stateMachineStates, id)

	// Remove the instance
	delete(s.instances, id)

	// Note: We don't remove from identityRegistry to avoid ID reuse issues
	_ = instance
	return nil
}

// removeAllLinks removes all links to/from an instance (must be called with lock held).
func (s *SimulationState) removeAllLinks(id InstanceID) {
	objID := evaluator.ObjectID(id)

	// Get and remove all forward links
	forwardLinks := s.links.GetAllForward(objID)
	for _, link := range forwardLinks {
		s.links.RemoveLink(link.AssociationKey, link.FromID, link.ToID)
	}

	// Get and remove all reverse links
	reverseLinks := s.links.GetAllReverse(objID)
	for _, link := range reverseLinks {
		s.links.RemoveLink(link.AssociationKey, link.FromID, link.ToID)
	}
}

// InstanceCount returns the number of instances.
func (s *SimulationState) InstanceCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.instances)
}

// AllInstances returns all instances.
func (s *SimulationState) AllInstances() []*ClassInstance {
	s.mu.RLock()
	defer s.mu.RUnlock()

	instances := make([]*ClassInstance, 0, len(s.instances))
	for _, instance := range s.instances {
		instances = append(instances, instance)
	}
	return instances
}

// InstancesByClass returns all instances of a specific class.
func (s *SimulationState) InstancesByClass(classKey identity.Key) []*ClassInstance {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var instances []*ClassInstance
	for _, instance := range s.instances {
		if instance.ClassKey == classKey {
			instances = append(instances, instance)
		}
	}
	return instances
}

// AddLink creates a link between two instances for an association.
func (s *SimulationState) AddLink(assocKey identity.Key, fromID, toID InstanceID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.links.AddLink(
		evaluator.AssociationKey(assocKey.String()),
		evaluator.ObjectID(fromID),
		evaluator.ObjectID(toID),
	)
}

// RemoveLink removes a link between two instances.
// Returns true if a link was removed.
func (s *SimulationState) RemoveLink(assocKey identity.Key, fromID, toID InstanceID) bool {
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
func (s *SimulationState) GetLinkedForward(fromID InstanceID, assocKey identity.Key) []InstanceID {
	s.mu.RLock()
	defer s.mu.RUnlock()

	objIDs := s.links.GetForward(
		evaluator.ObjectID(fromID),
		evaluator.AssociationKey(assocKey.String()),
	)

	ids := make([]InstanceID, len(objIDs))
	for i, objID := range objIDs {
		ids[i] = InstanceID(objID)
	}
	return ids
}

// GetLinkedReverse returns instance IDs linked TO the given instance
// for a specific association.
func (s *SimulationState) GetLinkedReverse(toID InstanceID, assocKey identity.Key) []InstanceID {
	s.mu.RLock()
	defer s.mu.RUnlock()

	objIDs := s.links.GetReverse(
		evaluator.ObjectID(toID),
		evaluator.AssociationKey(assocKey.String()),
	)

	ids := make([]InstanceID, len(objIDs))
	for i, objID := range objIDs {
		ids[i] = InstanceID(objID)
	}
	return ids
}

// ActiveInstanceFilter decides whether an instance counts toward association structural limits.
type ActiveInstanceFilter func(classKey identity.Key, stateName string) bool

// CountActivePairLinks counts live links for one association between a from/to instance pair.
func (s *SimulationState) CountActivePairLinks(
	assoc model_class.Association,
	fromID, toID InstanceID,
	isActive ActiveInstanceFilter,
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
			linkInst := s.instances[link.LinkInstanceID]
			if linkInst == nil {
				continue
			}
			stateName := instanceStateNameFrom(linkInst)
			if isActive != nil && !isActive(linkInst.ClassKey, stateName) {
				continue
			}
			count++
		}
		return count
	}

	return s.links.CountPairLinks(
		evaluator.AssociationKey(assoc.Key.String()),
		evaluator.ObjectID(fromID),
		evaluator.ObjectID(toID),
	)
}

func instanceStateNameFrom(instance *ClassInstance) string {
	if instance == nil {
		return ""
	}
	stateAttr := instance.GetAttribute("_state")
	if stateAttr == nil {
		return ""
	}
	if strObj, ok := stateAttr.(*object.String); ok {
		return strObj.Value()
	}
	return ""
}

// LinkCount returns the total number of links.
func (s *SimulationState) LinkCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.links.Count()
}

// Links returns the underlying link table.
// Use with caution - this exposes internal state.
func (s *SimulationState) Links() *evaluator.LinkTable {
	return s.links
}

// AddAssociationLink materializes one host association row via an association-class instance.
func (s *SimulationState) AddAssociationLink(
	hostAssocKey identity.Key,
	fromEndpointID InstanceID,
	toEndpointID InstanceID,
	linkInstanceID InstanceID,
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.associationLinks.AddLink(AssociationLink{
		HostAssocKey:   hostAssocKey,
		FromEndpointID: fromEndpointID,
		ToEndpointID:   toEndpointID,
		LinkInstanceID: linkInstanceID,
	})
}

// AssociationLinksFromEndpoint returns materialized host rows from a from-endpoint.
func (s *SimulationState) AssociationLinksFromEndpoint(hostAssocKey identity.Key, fromID InstanceID) []AssociationLink {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.associationLinks.LinksFromEndpoint(hostAssocKey, fromID)
}

// AssociationLinksToEndpoint returns materialized host rows to a to-endpoint.
func (s *SimulationState) AssociationLinksToEndpoint(hostAssocKey identity.Key, toID InstanceID) []AssociationLink {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.associationLinks.LinksToEndpoint(hostAssocKey, toID)
}

// AssociationLinkByInstance returns the host row for an association-class instance.
func (s *SimulationState) AssociationLinkByInstance(linkInstanceID InstanceID) (AssociationLink, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.associationLinks.LinkByInstance(linkInstanceID)
}

// AssociationLinks returns the underlying association link table.
func (s *SimulationState) AssociationLinks() *AssociationLinkTable {
	return s.associationLinks
}

// SetStateMachineState sets the current state machine state for an instance.
func (s *SimulationState) SetStateMachineState(id InstanceID, stateKey identity.Key) error {
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
func (s *SimulationState) GetStateMachineState(id InstanceID) (identity.Key, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, ok := s.stateMachineStates[id]
	return state, ok
}

// ClearStateMachineState removes the state machine state for an instance.
func (s *SimulationState) ClearStateMachineState(id InstanceID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.stateMachineStates, id)
}

// IdentityRegistry returns the identity registry for evaluator integration.
func (s *SimulationState) IdentityRegistry() *evaluator.IdentityRegistry {
	return s.identityRegistry
}

// Clone creates a deep copy of the simulation state.
func (s *SimulationState) Clone() *SimulationState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clone := NewSimulationState()
	clone.nextID = s.nextID

	// Clone instances
	for id, instance := range s.instances {
		clone.instances[id] = instance.Clone()
	}

	// Clone state machine states
	maps.Copy(clone.stateMachineStates, s.stateMachineStates)

	// Clone links by copying each link
	for _, instance := range s.instances {
		objID := evaluator.ObjectID(instance.ID)
		links := s.links.GetAllForward(objID)
		for _, link := range links {
			clone.links.AddLink(link.AssociationKey, link.FromID, link.ToID)
		}
	}
	for _, link := range s.associationLinks.AllLinks() {
		clone.associationLinks.AddLink(link)
	}

	return clone
}
