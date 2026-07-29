package schema

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/surface"
)

// Query methods on Schema delegate to the owned Catalog so callers can use *Schema alone.

// SetSurfaceUnavailableMembers delegates to the schema-owned Catalog.
func (s *Schema) SetSurfaceUnavailableMembers(members []surface.UnavailableMember) {
	s.Catalog().SetSurfaceUnavailableMembers(members)
}

// SurfaceUnavailableMembers delegates to the schema-owned Catalog.
func (s *Schema) SurfaceUnavailableMembers() []surface.UnavailableMember {
	return s.Catalog().SurfaceUnavailableMembers()
}

// SurfaceUnavailableDerived delegates to the schema-owned Catalog.
func (s *Schema) SurfaceUnavailableDerived(attrKey identity.Key) (surface.UnavailableMember, bool) {
	return s.Catalog().SurfaceUnavailableDerived(attrKey)
}

// SurfaceUnavailableQuery delegates to the schema-owned Catalog.
func (s *Schema) SurfaceUnavailableQuery(queryKey identity.Key) (surface.UnavailableMember, bool) {
	return s.Catalog().SurfaceUnavailableQuery(queryKey)
}

// IsSurfaceUnavailableDerived delegates to the schema-owned Catalog.
func (s *Schema) IsSurfaceUnavailableDerived(attrKey identity.Key) bool {
	return s.Catalog().IsSurfaceUnavailableDerived(attrKey)
}

// IsSurfaceUnavailableQuery delegates to the schema-owned Catalog.
func (s *Schema) IsSurfaceUnavailableQuery(queryKey identity.Key) bool {
	return s.Catalog().IsSurfaceUnavailableQuery(queryKey)
}

// LookupAssociationClass delegates to the schema-owned Catalog.
func (s *Schema) LookupAssociationClass(classKey identity.Key) *AssociationClassInfo {
	return s.Catalog().LookupAssociationClass(classKey)
}

// IsAssociationClass delegates to the schema-owned Catalog.
func (s *Schema) IsAssociationClass(classKey identity.Key) bool {
	return s.Catalog().IsAssociationClass(classKey)
}

// IsAssociationClassHost delegates to the schema-owned Catalog.
func (s *Schema) IsAssociationClassHost(assocKey identity.Key) bool {
	return s.Catalog().IsAssociationClassHost(assocKey)
}

// GetAssociationClassInfo delegates to the schema-owned Catalog.
func (s *Schema) GetAssociationClassInfo(classKey identity.Key) AssociationClassLinkInfo {
	return s.Catalog().GetAssociationClassInfo(classKey)
}

// GetClassInfo delegates to the schema-owned Catalog.
func (s *Schema) GetClassInfo(classKey identity.Key) *ClassSimInfo {
	return s.Catalog().GetClassInfo(classKey)
}

// AllScopedClasses delegates to the schema-owned Catalog.
func (s *Schema) AllScopedClasses() []*ClassSimInfo {
	return s.Catalog().AllScopedClasses()
}

// AllSimulatableClasses delegates to the schema-owned Catalog.
func (s *Schema) AllSimulatableClasses() []*ClassSimInfo {
	return s.Catalog().AllSimulatableClasses()
}

// AllEventBearingClasses delegates to the schema-owned Catalog.
func (s *Schema) AllEventBearingClasses() []*ClassSimInfo {
	return s.Catalog().AllEventBearingClasses()
}

// GetMandatoryOutboundAssociations delegates to the schema-owned Catalog.
func (s *Schema) GetMandatoryOutboundAssociations(classKey identity.Key) []AssociationInfo {
	return s.Catalog().GetMandatoryOutboundAssociations(classKey)
}

// GetActionForEvent delegates to the schema-owned Catalog.
func (s *Schema) GetActionForEvent(
	classKey identity.Key,
	eventKey identity.Key,
	instanceStateName string,
) (*model_state.Action, bool) {
	return s.Catalog().GetActionForEvent(classKey, eventKey, instanceStateName)
}

// GetCreationEvent delegates to the schema-owned Catalog.
func (s *Schema) GetCreationEvent(classKey identity.Key) (*model_state.Event, bool) {
	return s.Catalog().GetCreationEvent(classKey)
}

// AllAssociations delegates to the schema-owned Catalog.
func (s *Schema) AllAssociations() []AssociationInfo {
	return s.Catalog().AllAssociations()
}

// GetAssociationsForClass delegates to the schema-owned Catalog.
func (s *Schema) GetAssociationsForClass(classKey identity.Key) []AssociationInfo {
	return s.Catalog().GetAssociationsForClass(classKey)
}

// AssociationByKey delegates to the schema-owned Catalog.
func (s *Schema) AssociationByKey(assocKey identity.Key) (model_class.Association, bool) {
	return s.Catalog().AssociationByKey(assocKey)
}

// OutgoingAssociationByAssociationClassTLAName delegates to the schema-owned Catalog.
func (s *Schema) OutgoingAssociationByAssociationClassTLAName(
	fromClassKey identity.Key,
	classTLAName string,
) (identity.Key, model_class.Association, bool) {
	return s.Catalog().OutgoingAssociationByAssociationClassTLAName(fromClassKey, classTLAName)
}

// OutgoingAssociationByTLAField delegates to the schema-owned Catalog.
func (s *Schema) OutgoingAssociationByTLAField(
	fromClassKey identity.Key,
	tlaField string,
) (identity.Key, model_class.Association, bool) {
	return s.Catalog().OutgoingAssociationByTLAField(fromClassKey, tlaField)
}

// AssociationByNavigableTLAField delegates to the schema-owned Catalog.
func (s *Schema) AssociationByNavigableTLAField(
	classKey identity.Key,
	tlaField string,
) (identity.Key, model_class.Association, bool, bool) {
	return s.Catalog().AssociationByNavigableTLAField(classKey, tlaField)
}

// OutgoingAssociationsTo delegates to the schema-owned Catalog.
func (s *Schema) OutgoingAssociationsTo(fromClassKey, toClassKey identity.Key) []model_class.Association {
	return s.Catalog().OutgoingAssociationsTo(fromClassKey, toClassKey)
}

// PeerClass delegates to the schema-owned Catalog.
func (s *Schema) PeerClass(classKey identity.Key) (model_class.Class, bool) {
	return s.Catalog().PeerClass(classKey)
}

// PeerCreationEvent delegates to the schema-owned Catalog.
func (s *Schema) PeerCreationEvent(classKey identity.Key) (model_state.Event, bool) {
	return s.Catalog().PeerCreationEvent(classKey)
}

// PeerEvent delegates to the schema-owned Catalog.
func (s *Schema) PeerEvent(classKey identity.Key, eventKey identity.Key) (model_state.Event, bool) {
	return s.Catalog().PeerEvent(classKey, eventKey)
}

// ExternalCreationEvents delegates to the schema-owned Catalog.
func (s *Schema) ExternalCreationEvents(classKey identity.Key) []model_state.Event {
	return s.Catalog().ExternalCreationEvents(classKey)
}

// ExternalStateEvents delegates to the schema-owned Catalog.
func (s *Schema) ExternalStateEvents(classKey identity.Key, stateName string) []EventInfo {
	return s.Catalog().ExternalStateEvents(classKey, stateName)
}

// ExternalQueries delegates to the schema-owned Catalog.
func (s *Schema) ExternalQueries(classKey identity.Key) []model_state.Query {
	return s.Catalog().ExternalQueries(classKey)
}

// SurfaceDoActions delegates to the schema-owned Catalog.
func (s *Schema) SurfaceDoActions(classKey identity.Key, stateName string) []model_state.Action {
	return s.Catalog().SurfaceDoActions(classKey, stateName)
}

// ExternalDerivedAttributes delegates to the schema-owned Catalog.
func (s *Schema) ExternalDerivedAttributes(classKey identity.Key) []model_class.Attribute {
	return s.Catalog().ExternalDerivedAttributes(classKey)
}

// ClassNameMap delegates to the schema-owned Catalog.
func (s *Schema) ClassNameMap() map[identity.Key]string {
	return s.Catalog().ClassNameMap()
}
