package state

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
)

// AssociationLink materializes one host association row via its association-class instance.
type AssociationLink struct {
	HostAssocKey   identity.Key
	FromEndpointID InstanceID
	ToEndpointID   InstanceID
	LinkInstanceID InstanceID
}

// AssociationLinkTable stores reified host-association rows keyed by association class instances.
type AssociationLinkTable struct {
	byHostFrom map[evaluator.AssociationKey]map[InstanceID][]AssociationLink
	byHostTo   map[evaluator.AssociationKey]map[InstanceID][]AssociationLink
	byInstance map[InstanceID]AssociationLink
}

// NewAssociationLinkTable creates an empty association link table.
func NewAssociationLinkTable() *AssociationLinkTable {
	return &AssociationLinkTable{
		byHostFrom: make(map[evaluator.AssociationKey]map[InstanceID][]AssociationLink),
		byHostTo:   make(map[evaluator.AssociationKey]map[InstanceID][]AssociationLink),
		byInstance: make(map[InstanceID]AssociationLink),
	}
}

// AddLink records a host association materialized by one association-class instance.
// Returns an error when the host association already links the same endpoint pair.
func (t *AssociationLinkTable) AddLink(link AssociationLink) error {
	hostKey := evaluator.AssociationKey(link.HostAssocKey.String())
	if t.hasEndpointPair(hostKey, link.FromEndpointID, link.ToEndpointID) {
		return fmt.Errorf(
			"duplicate host association link between endpoints %d and %d",
			link.FromEndpointID,
			link.ToEndpointID,
		)
	}

	if t.byHostFrom[hostKey] == nil {
		t.byHostFrom[hostKey] = make(map[InstanceID][]AssociationLink)
	}
	t.byHostFrom[hostKey][link.FromEndpointID] = append(t.byHostFrom[hostKey][link.FromEndpointID], link)

	if t.byHostTo[hostKey] == nil {
		t.byHostTo[hostKey] = make(map[InstanceID][]AssociationLink)
	}
	t.byHostTo[hostKey][link.ToEndpointID] = append(t.byHostTo[hostKey][link.ToEndpointID], link)

	t.byInstance[link.LinkInstanceID] = link
	return nil
}

// AppendLinkWithoutValidation records a row without duplicate checking.
// Invariant tests use this to represent tables that bypass normal insertion rules.
func (t *AssociationLinkTable) AppendLinkWithoutValidation(link AssociationLink) {
	hostKey := evaluator.AssociationKey(link.HostAssocKey.String())

	if t.byHostFrom[hostKey] == nil {
		t.byHostFrom[hostKey] = make(map[InstanceID][]AssociationLink)
	}
	t.byHostFrom[hostKey][link.FromEndpointID] = append(t.byHostFrom[hostKey][link.FromEndpointID], link)

	if t.byHostTo[hostKey] == nil {
		t.byHostTo[hostKey] = make(map[InstanceID][]AssociationLink)
	}
	t.byHostTo[hostKey][link.ToEndpointID] = append(t.byHostTo[hostKey][link.ToEndpointID], link)

	t.byInstance[link.LinkInstanceID] = link
}

func (t *AssociationLinkTable) hasEndpointPair(
	hostKey evaluator.AssociationKey,
	fromID InstanceID,
	toID InstanceID,
) bool {
	byFrom, ok := t.byHostFrom[hostKey]
	if !ok {
		return false
	}
	for _, link := range byFrom[fromID] {
		if link.ToEndpointID == toID {
			return true
		}
	}
	return false
}

// LinksFromEndpoint returns materialized rows for a from-endpoint under the host association.
func (t *AssociationLinkTable) LinksFromEndpoint(hostAssocKey identity.Key, fromID InstanceID) []AssociationLink {
	hostKey := evaluator.AssociationKey(hostAssocKey.String())
	if byFrom, ok := t.byHostFrom[hostKey]; ok {
		return append([]AssociationLink(nil), byFrom[fromID]...)
	}
	return nil
}

// LinksToEndpoint returns materialized rows for a to-endpoint under the host association.
func (t *AssociationLinkTable) LinksToEndpoint(hostAssocKey identity.Key, toID InstanceID) []AssociationLink {
	hostKey := evaluator.AssociationKey(hostAssocKey.String())
	if byTo, ok := t.byHostTo[hostKey]; ok {
		return append([]AssociationLink(nil), byTo[toID]...)
	}
	return nil
}

// LinkByInstance returns the row materialized by the given association-class instance.
func (t *AssociationLinkTable) LinkByInstance(linkInstanceID InstanceID) (AssociationLink, bool) {
	link, ok := t.byInstance[linkInstanceID]
	return link, ok
}

// RemoveInstance drops every row touching the instance as endpoint or link instance.
func (t *AssociationLinkTable) RemoveInstance(id InstanceID) {
	var toRemove []AssociationLink
	for _, link := range t.byInstance {
		if link.LinkInstanceID == id || link.FromEndpointID == id || link.ToEndpointID == id {
			toRemove = append(toRemove, link)
		}
	}
	for _, link := range toRemove {
		t.removeLink(link)
	}
}

func (t *AssociationLinkTable) removeLink(link AssociationLink) {
	hostKey := evaluator.AssociationKey(link.HostAssocKey.String())

	if byFrom, ok := t.byHostFrom[hostKey]; ok {
		byFrom[link.FromEndpointID] = filterAssociationLinks(byFrom[link.FromEndpointID], link.LinkInstanceID)
		if len(byFrom[link.FromEndpointID]) == 0 {
			delete(byFrom, link.FromEndpointID)
		}
	}
	if byTo, ok := t.byHostTo[hostKey]; ok {
		byTo[link.ToEndpointID] = filterAssociationLinks(byTo[link.ToEndpointID], link.LinkInstanceID)
		if len(byTo[link.ToEndpointID]) == 0 {
			delete(byTo, link.ToEndpointID)
		}
	}
	delete(t.byInstance, link.LinkInstanceID)
}

func filterAssociationLinks(links []AssociationLink, linkInstanceID InstanceID) []AssociationLink {
	filtered := links[:0]
	for _, link := range links {
		if link.LinkInstanceID != linkInstanceID {
			filtered = append(filtered, link)
		}
	}
	return filtered
}

// AllHostAssociationKeys returns host association keys with at least one materialized row.
func (t *AssociationLinkTable) AllHostAssociationKeys() map[evaluator.AssociationKey]bool {
	result := make(map[evaluator.AssociationKey]bool)
	for hostKey := range t.byHostFrom {
		result[hostKey] = true
	}
	return result
}

// AllLinks returns every materialized row.
func (t *AssociationLinkTable) AllLinks() []AssociationLink {
	links := make([]AssociationLink, 0, len(t.byInstance))
	for _, link := range t.byInstance {
		links = append(links, link)
	}
	return links
}
