package object

import "fmt"

// AssociationRelation is the runtime value of an association-class host traversal.
// It behaves as the endpoint set in quantifiers and membership, and exposes
// exactly one association-class row per endpoint via a named member.
type AssociationRelation struct {
	endpoints       *Set
	linkClassMember string
	linkByEndpoint  map[*Record]*Record
}

// NewAssociationRelation builds an association traversal value.
// linkByEndpoint maps each far-endpoint record to its association-class row.
func NewAssociationRelation(
	endpoints *Set,
	linkClassMember string,
	linkByEndpoint map[*Record]*Record,
) *AssociationRelation {
	if endpoints == nil {
		endpoints = NewSet()
	}
	if linkByEndpoint == nil {
		linkByEndpoint = make(map[*Record]*Record)
	}
	return &AssociationRelation{
		endpoints:       endpoints,
		linkClassMember: linkClassMember,
		linkByEndpoint:  linkByEndpoint,
	}
}

func (a *AssociationRelation) Type() ObjectType { return TypeAssociationRelation }

func (a *AssociationRelation) Inspect() string {
	return a.endpoints.Inspect()
}

func (a *AssociationRelation) SetValue(source Object) error {
	src, ok := source.(*AssociationRelation)
	if !ok {
		return fmt.Errorf("cannot assign %T to AssociationRelation", source)
	}
	if err := a.endpoints.SetValue(src.endpoints); err != nil {
		return err
	}
	a.linkClassMember = src.linkClassMember
	a.linkByEndpoint = cloneLinkByEndpoint(src.linkByEndpoint)
	return nil
}

func (a *AssociationRelation) Clone() Object {
	return NewAssociationRelation(
		a.endpoints.Clone().(*Set),
		a.linkClassMember,
		cloneLinkByEndpoint(a.linkByEndpoint),
	)
}

// Endpoints returns the related endpoint instances (the far side of the association).
func (a *AssociationRelation) Endpoints() *Set {
	return a.endpoints
}

// LinkClassMember returns the TLA+ member name for association-class rows.
func (a *AssociationRelation) LinkClassMember() string {
	return a.linkClassMember
}

// LinkByEndpoint returns the association-class row for a far-endpoint record.
func (a *AssociationRelation) LinkByEndpoint() map[*Record]*Record {
	return a.linkByEndpoint
}

// LinkForEndpoint returns the association-class row materializing one host row.
func (a *AssociationRelation) LinkForEndpoint(endpoint *Record) (*Record, bool) {
	if endpoint == nil {
		return nil, false
	}
	if link, ok := a.linkByEndpoint[endpoint]; ok {
		return link, true
	}
	for ep, link := range a.linkByEndpoint {
		if ep.Equals(endpoint) {
			return link, true
		}
	}
	return nil, false
}

// Equals reports whether two association traversals expose the same endpoints and link rows.
func (a *AssociationRelation) Equals(other *AssociationRelation) bool {
	if other == nil {
		return false
	}
	if a.linkClassMember != other.linkClassMember {
		return false
	}
	if !a.endpoints.Equals(other.endpoints) {
		return false
	}
	if len(a.linkByEndpoint) != len(other.linkByEndpoint) {
		return false
	}
	for ep, link := range a.linkByEndpoint {
		otherLink, ok := other.LinkForEndpoint(ep)
		if !ok || !link.Equals(otherLink) {
			return false
		}
	}
	return true
}

func cloneLinkByEndpoint(src map[*Record]*Record) map[*Record]*Record {
	if len(src) == 0 {
		return make(map[*Record]*Record)
	}
	dst := make(map[*Record]*Record, len(src))
	for ep, link := range src {
		dst[ep] = link.Clone().(*Record)
	}
	return dst
}
