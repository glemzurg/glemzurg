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
	s.mustCatalog().SetSurfaceUnavailableMembers(members)
}

// SurfaceUnavailableMembers delegates to the schema-owned Catalog.
func (s *Schema) SurfaceUnavailableMembers() []surface.UnavailableMember {
	return s.mustCatalog().SurfaceUnavailableMembers()
}

// SurfaceUnavailableDerived delegates to the schema-owned Catalog.
func (s *Schema) SurfaceUnavailableDerived(attrKey identity.Key) (surface.UnavailableMember, bool) {
	return s.mustCatalog().SurfaceUnavailableDerived(attrKey)
}

// SurfaceUnavailableQuery delegates to the schema-owned Catalog.
func (s *Schema) SurfaceUnavailableQuery(queryKey identity.Key) (surface.UnavailableMember, bool) {
	return s.mustCatalog().SurfaceUnavailableQuery(queryKey)
}

// IsSurfaceUnavailableDerived delegates to the schema-owned Catalog.
func (s *Schema) IsSurfaceUnavailableDerived(attrKey identity.Key) bool {
	return s.mustCatalog().IsSurfaceUnavailableDerived(attrKey)
}

// IsSurfaceUnavailableQuery delegates to the schema-owned Catalog.
func (s *Schema) IsSurfaceUnavailableQuery(queryKey identity.Key) bool {
	return s.mustCatalog().IsSurfaceUnavailableQuery(queryKey)
}

// LookupAssociationClass delegates to the schema-owned Catalog.
func (s *Schema) LookupAssociationClass(classKey identity.Key) *AssociationClassInfo {
	return s.mustCatalog().LookupAssociationClass(classKey)
}

// IsAssociationClass delegates to the schema-owned Catalog.
func (s *Schema) IsAssociationClass(classKey identity.Key) bool {
	return s.mustCatalog().IsAssociationClass(classKey)
}

// IsAssociationClassHost delegates to the schema-owned Catalog.
func (s *Schema) IsAssociationClassHost(assocKey identity.Key) bool {
	return s.mustCatalog().IsAssociationClassHost(assocKey)
}

// GetAssociationClassInfo delegates to the schema-owned Catalog.
func (s *Schema) GetAssociationClassInfo(classKey identity.Key) AssociationClassLinkInfo {
	return s.mustCatalog().GetAssociationClassInfo(classKey)
}

// GetClassInfo delegates to the schema-owned Catalog.
func (s *Schema) GetClassInfo(classKey identity.Key) *ClassSimInfo {
	return s.mustCatalog().GetClassInfo(classKey)
}

// AllScopedClasses delegates to the schema-owned Catalog.
func (s *Schema) AllScopedClasses() []*ClassSimInfo {
	return s.mustCatalog().AllScopedClasses()
}

// AllSimulatableClasses delegates to the schema-owned Catalog.
func (s *Schema) AllSimulatableClasses() []*ClassSimInfo {
	return s.mustCatalog().AllSimulatableClasses()
}

// AllEventBearingClasses delegates to the schema-owned Catalog.
func (s *Schema) AllEventBearingClasses() []*ClassSimInfo {
	return s.mustCatalog().AllEventBearingClasses()
}

// GetMandatoryOutboundAssociations delegates to the schema-owned Catalog.
func (s *Schema) GetMandatoryOutboundAssociations(classKey identity.Key) []AssociationInfo {
	return s.mustCatalog().GetMandatoryOutboundAssociations(classKey)
}

// GetActionForEvent delegates to the schema-owned Catalog.
func (s *Schema) GetActionForEvent(
	classKey identity.Key,
	eventKey identity.Key,
	instanceStateName string,
) (*model_state.Action, bool) {
	return s.mustCatalog().GetActionForEvent(classKey, eventKey, instanceStateName)
}

// GetCreationEvent delegates to the schema-owned Catalog.
func (s *Schema) GetCreationEvent(classKey identity.Key) (*model_state.Event, bool) {
	return s.mustCatalog().GetCreationEvent(classKey)
}

// AllAssociations delegates to the schema-owned Catalog.
func (s *Schema) AllAssociations() []AssociationInfo {
	return s.mustCatalog().AllAssociations()
}

// GetAssociationsForClass delegates to the schema-owned Catalog.
func (s *Schema) GetAssociationsForClass(classKey identity.Key) []AssociationInfo {
	return s.mustCatalog().GetAssociationsForClass(classKey)
}

// AssociationByKey delegates to the schema-owned Catalog.
func (s *Schema) AssociationByKey(assocKey identity.Key) (model_class.Association, bool) {
	return s.mustCatalog().AssociationByKey(assocKey)
}

// OutgoingAssociationByAssociationClassTLAName delegates to the schema-owned Catalog.
func (s *Schema) OutgoingAssociationByAssociationClassTLAName(
	fromClassKey identity.Key,
	classTLAName string,
) (identity.Key, model_class.Association, bool) {
	return s.mustCatalog().OutgoingAssociationByAssociationClassTLAName(fromClassKey, classTLAName)
}

// OutgoingAssociationByTLAField delegates to the schema-owned Catalog.
func (s *Schema) OutgoingAssociationByTLAField(
	fromClassKey identity.Key,
	tlaField string,
) (identity.Key, model_class.Association, bool) {
	return s.mustCatalog().OutgoingAssociationByTLAField(fromClassKey, tlaField)
}

// AssociationByNavigableTLAField delegates to the schema-owned Catalog.
func (s *Schema) AssociationByNavigableTLAField(
	classKey identity.Key,
	tlaField string,
) (identity.Key, model_class.Association, bool, bool) {
	return s.mustCatalog().AssociationByNavigableTLAField(classKey, tlaField)
}

// OutgoingAssociationsTo delegates to the schema-owned Catalog.
func (s *Schema) OutgoingAssociationsTo(fromClassKey, toClassKey identity.Key) []model_class.Association {
	return s.mustCatalog().OutgoingAssociationsTo(fromClassKey, toClassKey)
}

// PeerClass delegates to the schema-owned Catalog.
func (s *Schema) PeerClass(classKey identity.Key) (model_class.Class, bool) {
	return s.mustCatalog().PeerClass(classKey)
}

// PeerCreationEvent delegates to the schema-owned Catalog.
func (s *Schema) PeerCreationEvent(classKey identity.Key) (model_state.Event, bool) {
	return s.mustCatalog().PeerCreationEvent(classKey)
}

// PeerEvent delegates to the schema-owned Catalog.
func (s *Schema) PeerEvent(classKey identity.Key, eventKey identity.Key) (model_state.Event, bool) {
	return s.mustCatalog().PeerEvent(classKey, eventKey)
}

// ExternalCreationEvents delegates to the schema-owned Catalog.
func (s *Schema) ExternalCreationEvents(classKey identity.Key) []model_state.Event {
	return s.mustCatalog().ExternalCreationEvents(classKey)
}

// ExternalStateEvents delegates to the schema-owned Catalog.
func (s *Schema) ExternalStateEvents(classKey identity.Key, stateName string) []EventInfo {
	return s.mustCatalog().ExternalStateEvents(classKey, stateName)
}

// ExternalQueries delegates to the schema-owned Catalog.
func (s *Schema) ExternalQueries(classKey identity.Key) []model_state.Query {
	return s.mustCatalog().ExternalQueries(classKey)
}

// SurfaceDoActions delegates to the schema-owned Catalog.
func (s *Schema) SurfaceDoActions(classKey identity.Key, stateName string) []model_state.Action {
	return s.mustCatalog().SurfaceDoActions(classKey, stateName)
}

// ExternalDerivedAttributes delegates to the schema-owned Catalog.
func (s *Schema) ExternalDerivedAttributes(classKey identity.Key) []model_class.Attribute {
	return s.mustCatalog().ExternalDerivedAttributes(classKey)
}

// ClassNameMap delegates to the schema-owned Catalog.
func (s *Schema) ClassNameMap() map[identity.Key]string {
	return s.mustCatalog().ClassNameMap()
}

// CallerData exports SentBy/CalledBy metadata for surface diagnostics and tests.
func (s *Schema) CallerData() *surface.CallerData {
	return s.mustCatalog().CallerData()
}

// SetEventSentBy records event senders (tests / specialized wiring).
func (s *Schema) SetEventSentBy(eventKey identity.Key, senderClassKeys []identity.Key) {
	s.mustCatalog().SetEventSentBy(eventKey, senderClassKeys)
}

// SetActionCalledBy records action callers (tests / specialized wiring).
func (s *Schema) SetActionCalledBy(actionKey identity.Key, callerClassKeys []identity.Key) {
	s.mustCatalog().SetActionCalledBy(actionKey, callerClassKeys)
}

// SetQueryCalledBy records query callers (tests / specialized wiring).
func (s *Schema) SetQueryCalledBy(queryKey identity.Key, callerClassKeys []identity.Key) {
	s.mustCatalog().SetQueryCalledBy(queryKey, callerClassKeys)
}
