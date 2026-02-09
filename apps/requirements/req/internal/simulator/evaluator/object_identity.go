package evaluator

import (
	"github.com/glemzurg/go-tlaplus/internal/simulator/object"
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
