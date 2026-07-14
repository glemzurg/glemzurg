package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// ObjectID is a hidden identity for tracking object relations.
// Not visible to TLA+ expressions - purely an implementation detail
// for the evaluator to track object relationships.
type ObjectID uint64

// IdentityRegistry assigns and tracks object identities.
// It maintains bidirectional mappings between records and their IDs.
type IdentityRegistry struct {
	nextID     ObjectID
	recordToID map[*object.Record]ObjectID
	idToRecord map[ObjectID]*object.Record
}

// NewIdentityRegistry creates a new identity registry.
func NewIdentityRegistry() *IdentityRegistry {
	return &IdentityRegistry{
		nextID:     1, // Start at 1 so 0 can indicate "no ID"
		recordToID: make(map[*object.Record]ObjectID),
		idToRecord: make(map[ObjectID]*object.Record),
	}
}

// GetOrAssign returns the ID for a record, assigning a new one if needed.
// This is the primary method for getting an object's identity.
func (r *IdentityRegistry) GetOrAssign(record *object.Record) ObjectID {
	if record == nil {
		return 0
	}

	if id, exists := r.recordToID[record]; exists {
		return id
	}

	// Assign new ID
	id := r.nextID
	r.nextID++

	r.recordToID[record] = id
	r.idToRecord[id] = record

	return id
}

// RegisterVisible binds a fixed ObjectID to a TLA-visible record and optional aliases.
// GetRecord returns visible; GetID succeeds for visible and every alias (e.g. bare
// self data plus the [id, data] extent element for the same instance).
func (r *IdentityRegistry) RegisterVisible(id ObjectID, visible *object.Record, aliases ...*object.Record) {
	if id == 0 || visible == nil {
		return
	}
	r.recordToID[visible] = id
	r.idToRecord[id] = visible
	for _, alias := range aliases {
		if alias != nil {
			r.recordToID[alias] = id
		}
	}
	if id >= r.nextID {
		r.nextID = id + 1
	}
}

// RegisterAlias maps an additional record pointer to an existing ObjectID without
// changing GetRecord. Used when self is a derived-attribute clone of instance data.
func (r *IdentityRegistry) RegisterAlias(id ObjectID, alias *object.Record) {
	if id == 0 || alias == nil {
		return
	}
	if _, ok := r.idToRecord[id]; !ok {
		return
	}
	r.recordToID[alias] = id
}

// GetID returns the ID for a record if it has been assigned.
// Returns (0, false) if the record has no assigned ID.
func (r *IdentityRegistry) GetID(record *object.Record) (ObjectID, bool) {
	if record == nil {
		return 0, false
	}
	id, exists := r.recordToID[record]
	return id, exists
}

// GetRecord returns the record for a given ID.
// Returns nil if the ID is not found.
func (r *IdentityRegistry) GetRecord(id ObjectID) *object.Record {
	return r.idToRecord[id]
}

// Count returns the number of tracked records.
func (r *IdentityRegistry) Count() int {
	return len(r.recordToID)
}

// Clear removes all tracked identities.
// Use with caution - this invalidates all existing links.
func (r *IdentityRegistry) Clear() {
	r.recordToID = make(map[*object.Record]ObjectID)
	r.idToRecord = make(map[ObjectID]*object.Record)
	// Don't reset nextID to avoid ID reuse after clear
}
