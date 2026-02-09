package evaluator

import (
	"strings"

	"github.com/glemzurg/go-tlaplus/internal/simulator/object"
)

// Multiplicity represents the cardinality constraints of a relation end.
// Copied from model_class to avoid import cycles.
type Multiplicity struct {
	LowerBound  uint // Zero means "any" (no lower limit)
	HigherBound uint // Zero means "any" (no upper limit)
}

// RelationInfo contains metadata about an association accessible from a class.
type RelationInfo struct {
	AssociationKey AssociationKey // Full identity.Key.String() for link table lookup
	Name           string         // Display name, e.g., "Lines"
	TargetClassKey string         // The class we navigate TO (identity.Key.String())
	Multiplicity   Multiplicity   // Cardinality on the target side
	Reverse        bool           // True if this is a reverse relation (._Name)
}

// RelationContext manages association metadata and runtime link state.
// It provides the mechanism for navigating between related objects.
type RelationContext struct {
	// Association metadata indexed by class key (string form of identity.Key)
	// Maps: classKey -> fieldName -> RelationInfo
	// ForwardRelations: accessed via .Name (e.g., order.Lines)
	ForwardRelations map[string]map[string]*RelationInfo
	// ReverseRelations: accessed via ._Name (e.g., line._Lines)
	ReverseRelations map[string]map[string]*RelationInfo

	// Runtime state
	identities *IdentityRegistry
	links      *LinkTable
}

// NewRelationContext creates a new relation context.
func NewRelationContext() *RelationContext {
	return &RelationContext{
		ForwardRelations: make(map[string]map[string]*RelationInfo),
		ReverseRelations: make(map[string]map[string]*RelationInfo),
		identities:       NewIdentityRegistry(),
		links:            NewLinkTable(),
	}
}

// AddAssociation registers an association, populating both forward and reverse maps.
// - assocKey: The full identity.Key for the association (from identity.NewClassAssociationKey)
// - name: The display name (e.g., "Lines")
// - fromClassKey: The "from" class key as string
// - toClassKey: The "to" class key as string
// - fromMultiplicity: Cardinality on the "from" side
// - toMultiplicity: Cardinality on the "to" side
func (c *RelationContext) AddAssociation(
	assocKey AssociationKey,
	name string,
	fromClassKey string,
	toClassKey string,
	fromMultiplicity Multiplicity,
	toMultiplicity Multiplicity,
) {
	// Forward relation: from FromClass, access .Name to get ToClass records
	forwardInfo := &RelationInfo{
		AssociationKey: assocKey,
		Name:           name,
		TargetClassKey: toClassKey,
		Multiplicity:   toMultiplicity, // Multiplicity toward the target
		Reverse:        false,
	}

	if c.ForwardRelations[fromClassKey] == nil {
		c.ForwardRelations[fromClassKey] = make(map[string]*RelationInfo)
	}
	c.ForwardRelations[fromClassKey][name] = forwardInfo

	// Reverse relation: from ToClass, access ._Name to get FromClass records
	reverseInfo := &RelationInfo{
		AssociationKey: assocKey,
		Name:           name,
		TargetClassKey: fromClassKey,
		Multiplicity:   fromMultiplicity, // Multiplicity toward the source (now target)
		Reverse:        true,
	}

	if c.ReverseRelations[toClassKey] == nil {
		c.ReverseRelations[toClassKey] = make(map[string]*RelationInfo)
	}
	c.ReverseRelations[toClassKey][name] = reverseInfo
}

// GetForwardRelation returns relation info for a forward traversal (.Name).
// Returns nil if no such relation exists.
func (c *RelationContext) GetForwardRelation(classKey, fieldName string) *RelationInfo {
	if classMap, ok := c.ForwardRelations[classKey]; ok {
		return classMap[fieldName]
	}
	return nil
}

// GetReverseRelation returns relation info for a reverse traversal (._Name).
// The fieldName should include the underscore prefix (e.g., "_Lines").
// Returns nil if no such relation exists.
func (c *RelationContext) GetReverseRelation(classKey, fieldName string) *RelationInfo {
	// Strip the underscore prefix
	if !strings.HasPrefix(fieldName, "_") {
		return nil
	}
	name := strings.TrimPrefix(fieldName, "_")

	if classMap, ok := c.ReverseRelations[classKey]; ok {
		return classMap[name]
	}
	return nil
}

// GetRelation looks up a relation by field name, checking both forward and reverse.
// Returns the RelationInfo and whether it was found.
// For forward relations, use field name directly (e.g., "Lines").
// For reverse relations, use underscore prefix (e.g., "_Lines").
func (c *RelationContext) GetRelation(classKey, fieldName string) *RelationInfo {
	// Check if it's a reverse relation (starts with _)
	if strings.HasPrefix(fieldName, "_") {
		return c.GetReverseRelation(classKey, fieldName)
	}
	return c.GetForwardRelation(classKey, fieldName)
}

// CreateLink creates a link between two records for the given association.
// Both records will be assigned object IDs if they don't have them.
func (c *RelationContext) CreateLink(assocKey AssociationKey, from, to *object.Record) {
	fromID := c.identities.GetOrAssign(from)
	toID := c.identities.GetOrAssign(to)
	c.links.AddLink(assocKey, fromID, toID)
}

// RemoveLink removes a link between two records for the given association.
// Returns true if the link existed and was removed.
func (c *RelationContext) RemoveLink(assocKey AssociationKey, from, to *object.Record) bool {
	fromID, fromExists := c.identities.GetID(from)
	toID, toExists := c.identities.GetID(to)

	if !fromExists || !toExists {
		return false
	}

	return c.links.RemoveLink(assocKey, fromID, toID)
}

// GetRelatedRecords returns records related to the given record via an association.
// If reverse is false, returns records linked FROM this record (forward traversal).
// If reverse is true, returns records linked TO this record (reverse traversal).
func (c *RelationContext) GetRelatedRecords(record *object.Record, assocKey AssociationKey, reverse bool) []*object.Record {
	id, exists := c.identities.GetID(record)
	if !exists {
		return nil
	}

	var objectIDs []ObjectID
	if reverse {
		objectIDs = c.links.GetReverse(id, assocKey)
	} else {
		objectIDs = c.links.GetForward(id, assocKey)
	}

	records := make([]*object.Record, 0, len(objectIDs))
	for _, oid := range objectIDs {
		if rec := c.identities.GetRecord(oid); rec != nil {
			records = append(records, rec)
		}
	}

	return records
}

// RegisterRecord ensures a record has an object ID assigned.
// Returns the assigned ID.
func (c *RelationContext) RegisterRecord(record *object.Record) ObjectID {
	return c.identities.GetOrAssign(record)
}

// GetObjectID returns the object ID for a record, if assigned.
func (c *RelationContext) GetObjectID(record *object.Record) (ObjectID, bool) {
	return c.identities.GetID(record)
}

// Identities returns the underlying identity registry (for advanced use).
func (c *RelationContext) Identities() *IdentityRegistry {
	return c.identities
}

// Links returns the underlying link table (for advanced use).
func (c *RelationContext) Links() *LinkTable {
	return c.links
}

// Clear resets all runtime state (identities and links).
// Association metadata is preserved.
func (c *RelationContext) Clear() {
	c.identities.Clear()
	c.links.Clear()
}
