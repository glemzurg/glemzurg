package evaluator

import "fmt"

// AssociationKey is the string form of identity.Key for an association.
// Example: "domain/DomainA/subdomain/SubB/cassociation/class/BookOrder/class/BookOrderLine/lines"
// This provides globally unique identification for associations.
type AssociationKey string

// Link represents a single association instance between two objects.
type Link struct {
	AssociationKey AssociationKey // Full identity.Key.String()
	FromID         ObjectID       // Parent object (forward direction)
	ToID           ObjectID       // Child object (forward direction)
}

// LinkTable tracks all association links in the system.
// It maintains indexes for efficient lookup in both directions.
type LinkTable struct {
	// byFrom indexes links by their from object for forward traversal
	byFrom map[ObjectID][]Link
	// byTo indexes links by their to object for reverse traversal
	byTo map[ObjectID][]Link
}

// NewLinkTable creates a new link table.
func NewLinkTable() *LinkTable {
	return &LinkTable{
		byFrom: make(map[ObjectID][]Link),
		byTo:   make(map[ObjectID][]Link),
	}
}

// AddLink creates a link between two objects for an association.
// The link is indexed in both directions for efficient lookup.
// Returns an error when the same association already links the instance pair.
func (t *LinkTable) AddLink(assocKey AssociationKey, fromID, toID ObjectID) error {
	if t.hasLink(assocKey, fromID, toID) {
		return fmt.Errorf(
			"duplicate link for association %s between instances %d and %d",
			assocKey,
			fromID,
			toID,
		)
	}

	link := Link{
		AssociationKey: assocKey,
		FromID:         fromID,
		ToID:           toID,
	}
	t.byFrom[fromID] = append(t.byFrom[fromID], link)
	t.byTo[toID] = append(t.byTo[toID], link)
	return nil
}

// RemoveLink removes a link between two objects.
// Returns true if a link was removed, false if it didn't exist.
func (t *LinkTable) RemoveLink(assocKey AssociationKey, fromID, toID ObjectID) bool {
	removed := false

	// Remove from byFrom index
	if links, ok := t.byFrom[fromID]; ok {
		for i, link := range links {
			if link.AssociationKey == assocKey && link.ToID == toID {
				t.byFrom[fromID] = append(links[:i], links[i+1:]...)
				removed = true
				break
			}
		}
		// Clean up empty slice
		if len(t.byFrom[fromID]) == 0 {
			delete(t.byFrom, fromID)
		}
	}

	// Remove from byTo index
	if links, ok := t.byTo[toID]; ok {
		for i, link := range links {
			if link.AssociationKey == assocKey && link.FromID == fromID {
				t.byTo[toID] = append(links[:i], links[i+1:]...)
				break
			}
		}
		// Clean up empty slice
		if len(t.byTo[toID]) == 0 {
			delete(t.byTo, toID)
		}
	}

	return removed
}

// GetForward returns all object IDs linked FROM the given object
// for a specific association. Used for forward traversal (.Name).
func (t *LinkTable) GetForward(fromID ObjectID, assocKey AssociationKey) []ObjectID {
	links := t.byFrom[fromID]
	var result []ObjectID
	for _, link := range links {
		if link.AssociationKey == assocKey {
			result = append(result, link.ToID)
		}
	}
	return result
}

// GetReverse returns all object IDs linked TO the given object
// for a specific association. Used for reverse traversal (._Name).
func (t *LinkTable) GetReverse(toID ObjectID, assocKey AssociationKey) []ObjectID {
	links := t.byTo[toID]
	var result []ObjectID
	for _, link := range links {
		if link.AssociationKey == assocKey {
			result = append(result, link.FromID)
		}
	}
	return result
}

// GetAllForward returns all links from a given object (any association).
func (t *LinkTable) GetAllForward(fromID ObjectID) []Link {
	return t.byFrom[fromID]
}

// GetAllReverse returns all links to a given object (any association).
func (t *LinkTable) GetAllReverse(toID ObjectID) []Link {
	return t.byTo[toID]
}

// AppendLinkWithoutValidation records a link without duplicate checking.
// Invariant tests use this to represent link tables that bypass normal insertion rules.
func (t *LinkTable) AppendLinkWithoutValidation(assocKey AssociationKey, fromID, toID ObjectID) {
	link := Link{
		AssociationKey: assocKey,
		FromID:         fromID,
		ToID:           toID,
	}
	t.byFrom[fromID] = append(t.byFrom[fromID], link)
	t.byTo[toID] = append(t.byTo[toID], link)
}

// CountPairLinks returns how many links exist for one association between a from/to pair.
func (t *LinkTable) CountPairLinks(assocKey AssociationKey, fromID, toID ObjectID) int {
	count := 0
	for _, link := range t.byFrom[fromID] {
		if link.AssociationKey == assocKey && link.ToID == toID {
			count++
		}
	}
	return count
}

// hasLink checks if a specific link already exists.
func (t *LinkTable) hasLink(assocKey AssociationKey, fromID, toID ObjectID) bool {
	for _, link := range t.byFrom[fromID] {
		if link.AssociationKey == assocKey && link.ToID == toID {
			return true
		}
	}
	return false
}

// Count returns the total number of links in the table.
func (t *LinkTable) Count() int {
	count := 0
	for _, links := range t.byFrom {
		count += len(links)
	}
	return count
}

// AllAssociationKeys returns the set of association keys that have at least one link.
func (t *LinkTable) AllAssociationKeys() map[AssociationKey]bool {
	result := make(map[AssociationKey]bool)
	for _, links := range t.byFrom {
		for _, link := range links {
			result[link.AssociationKey] = true
		}
	}
	return result
}

// Clear removes all links from the table.
func (t *LinkTable) Clear() {
	t.byFrom = make(map[ObjectID][]Link)
	t.byTo = make(map[ObjectID][]Link)
}
