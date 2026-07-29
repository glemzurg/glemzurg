package schema

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/surface"
)

// Query methods on Schema delegate to the private simulation index so callers can use *Schema alone.

// SetSurfaceUnavailableMembers delegates to the private simulation index.
func (s *Schema) SetSurfaceUnavailableMembers(members []surface.UnavailableMember) {
	s.mustCatalog().setSurfaceUnavailableMembers(members)
}

// SurfaceUnavailableMembers delegates to the private simulation index.
func (s *Schema) SurfaceUnavailableMembers() []surface.UnavailableMember {
	return s.mustCatalog().surfaceUnavailableMembers()
}

// SurfaceUnavailableDerived delegates to the private simulation index.
func (s *Schema) SurfaceUnavailableDerived(attrKey identity.Key) (surface.UnavailableMember, bool) {
	return s.mustCatalog().lookupUnavailableDerived(attrKey)
}

// SurfaceUnavailableQuery delegates to the private simulation index.
func (s *Schema) SurfaceUnavailableQuery(queryKey identity.Key) (surface.UnavailableMember, bool) {
	return s.mustCatalog().surfaceUnavailableQuery(queryKey)
}

// IsSurfaceUnavailableDerived delegates to the private simulation index.
func (s *Schema) IsSurfaceUnavailableDerived(attrKey identity.Key) bool {
	return s.mustCatalog().isSurfaceUnavailableDerived(attrKey)
}

// IsSurfaceUnavailableQuery delegates to the private simulation index.
func (s *Schema) IsSurfaceUnavailableQuery(queryKey identity.Key) bool {
	return s.mustCatalog().isSurfaceUnavailableQuery(queryKey)
}

// LookupAssociationClass delegates to the private simulation index.
func (s *Schema) LookupAssociationClass(classKey identity.Key) *AssociationClassInfo {
	return s.mustCatalog().lookupAssociationClass(classKey)
}

// IsAssociationClass delegates to the private simulation index.
func (s *Schema) IsAssociationClass(classKey identity.Key) bool {
	return s.mustCatalog().isAssociationClass(classKey)
}

// IsAssociationClassHost delegates to the private simulation index.
func (s *Schema) IsAssociationClassHost(assocKey identity.Key) bool {
	return s.mustCatalog().isAssociationClassHost(assocKey)
}

// GetAssociationClassInfo delegates to the private simulation index.
func (s *Schema) GetAssociationClassInfo(classKey identity.Key) AssociationClassLinkInfo {
	return s.mustCatalog().getAssociationClassInfo(classKey)
}

// GetClassInfo delegates to the private simulation index.
func (s *Schema) GetClassInfo(classKey identity.Key) *ClassSimInfo {
	return s.mustCatalog().getClassInfo(classKey)
}

// AllScopedClasses delegates to the private simulation index.
func (s *Schema) AllScopedClasses() []*ClassSimInfo {
	return s.mustCatalog().allScopedClasses()
}

// AllSimulatableClasses delegates to the private simulation index.
func (s *Schema) AllSimulatableClasses() []*ClassSimInfo {
	return s.mustCatalog().allSimulatableClasses()
}

// AllEventBearingClasses delegates to the private simulation index.
func (s *Schema) AllEventBearingClasses() []*ClassSimInfo {
	return s.mustCatalog().allEventBearingClasses()
}

// GetMandatoryOutboundAssociations delegates to the private simulation index.
func (s *Schema) GetMandatoryOutboundAssociations(classKey identity.Key) []AssociationInfo {
	return s.mustCatalog().getMandatoryOutboundAssociations(classKey)
}

// GetActionForEvent delegates to the private simulation index.
func (s *Schema) GetActionForEvent(
	classKey identity.Key,
	eventKey identity.Key,
	instanceStateName string,
) (*model_state.Action, bool) {
	return s.mustCatalog().getActionForEvent(classKey, eventKey, instanceStateName)
}

// GetCreationEvent delegates to the private simulation index.
func (s *Schema) GetCreationEvent(classKey identity.Key) (*model_state.Event, bool) {
	return s.mustCatalog().getCreationEvent(classKey)
}

// AllAssociations delegates to the private simulation index.
func (s *Schema) AllAssociations() []AssociationInfo {
	return s.mustCatalog().allAssociations()
}

// GetAssociationsForClass delegates to the private simulation index.
func (s *Schema) GetAssociationsForClass(classKey identity.Key) []AssociationInfo {
	return s.mustCatalog().getAssociationsForClass(classKey)
}

// AssociationByKey delegates to the private simulation index.
func (s *Schema) AssociationByKey(assocKey identity.Key) (model_class.Association, bool) {
	return s.mustCatalog().associationByKey(assocKey)
}

// OutgoingAssociationByAssociationClassTLAName delegates to the private simulation index.
func (s *Schema) OutgoingAssociationByAssociationClassTLAName(
	fromClassKey identity.Key,
	classTLAName string,
) (identity.Key, model_class.Association, bool) {
	return s.mustCatalog().outgoingAssociationByAssociationClassTLAName(fromClassKey, classTLAName)
}

// OutgoingAssociationByTLAField delegates to the private simulation index.
func (s *Schema) OutgoingAssociationByTLAField(
	fromClassKey identity.Key,
	tlaField string,
) (identity.Key, model_class.Association, bool) {
	return s.mustCatalog().outgoingAssociationByTLAField(fromClassKey, tlaField)
}

// AssociationByNavigableTLAField delegates to the private simulation index.
func (s *Schema) AssociationByNavigableTLAField(
	classKey identity.Key,
	tlaField string,
) (identity.Key, model_class.Association, bool, bool) {
	return s.mustCatalog().associationByNavigableTLAField(classKey, tlaField)
}

// OutgoingAssociationsTo delegates to the private simulation index.
func (s *Schema) OutgoingAssociationsTo(fromClassKey, toClassKey identity.Key) []model_class.Association {
	return s.mustCatalog().outgoingAssociationsTo(fromClassKey, toClassKey)
}

// PeerClass delegates to the private simulation index.
func (s *Schema) PeerClass(classKey identity.Key) (model_class.Class, bool) {
	return s.mustCatalog().peerClass(classKey)
}

// PeerCreationEvent delegates to the private simulation index.
func (s *Schema) PeerCreationEvent(classKey identity.Key) (model_state.Event, bool) {
	return s.mustCatalog().peerCreationEvent(classKey)
}

// PeerEvent delegates to the private simulation index.
func (s *Schema) PeerEvent(classKey identity.Key, eventKey identity.Key) (model_state.Event, bool) {
	return s.mustCatalog().peerEvent(classKey, eventKey)
}

// ExternalCreationEvents delegates to the private simulation index.
func (s *Schema) ExternalCreationEvents(classKey identity.Key) []model_state.Event {
	return s.mustCatalog().externalCreationEvents(classKey)
}

// ExternalStateEvents delegates to the private simulation index.
func (s *Schema) ExternalStateEvents(classKey identity.Key, stateName string) []EventInfo {
	return s.mustCatalog().externalStateEvents(classKey, stateName)
}

// ExternalQueries delegates to the private simulation index.
func (s *Schema) ExternalQueries(classKey identity.Key) []model_state.Query {
	return s.mustCatalog().externalQueries(classKey)
}

// SurfaceDoActions delegates to the private simulation index.
func (s *Schema) SurfaceDoActions(classKey identity.Key, stateName string) []model_state.Action {
	return s.mustCatalog().surfaceDoActions(classKey, stateName)
}

// ExternalDerivedAttributes delegates to the private simulation index.
func (s *Schema) ExternalDerivedAttributes(classKey identity.Key) []model_class.Attribute {
	return s.mustCatalog().externalDerivedAttributes(classKey)
}

// ClassNameMap delegates to the private simulation index.
func (s *Schema) ClassNameMap() map[identity.Key]string {
	return s.mustCatalog().classNameMap()
}

// CallerData exports SentBy/CalledBy metadata for surface diagnostics and tests.
func (s *Schema) CallerData() *surface.CallerData {
	return s.mustCatalog().callerData()
}

// SetEventSentBy records event senders (tests / specialized wiring).
func (s *Schema) SetEventSentBy(eventKey identity.Key, senderClassKeys []identity.Key) {
	s.mustCatalog().setEventSentBy(eventKey, senderClassKeys)
}

// SetActionCalledBy records action callers (tests / specialized wiring).
func (s *Schema) SetActionCalledBy(actionKey identity.Key, callerClassKeys []identity.Key) {
	s.mustCatalog().setActionCalledBy(actionKey, callerClassKeys)
}

// SetQueryCalledBy records query callers (tests / specialized wiring).
func (s *Schema) SetQueryCalledBy(queryKey identity.Key, callerClassKeys []identity.Key) {
	s.mustCatalog().setQueryCalledBy(queryKey, callerClassKeys)
}
