package actions

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// PeerCreationCatalog resolves associations and peer-class events for association
// set-add and set-map guarantees.
type PeerCreationCatalog interface {
	AssociationClassIndex
	AssociationByKey(assocKey identity.Key) (model_class.Association, bool)
	OutgoingAssociationByTLAField(fromClassKey identity.Key, tlaField string) (identity.Key, model_class.Association, bool)
	// AssociationByNavigableTLAField resolves a forward (AssocName) or reverse (_AssocName)
	// navigation field on classKey. reverse is true when the class is the association to-end.
	AssociationByNavigableTLAField(classKey identity.Key, tlaField string) (assocKey identity.Key, assoc model_class.Association, reverse bool, found bool)
	// OutgoingAssociationByAssociationClassTLAName resolves the unique host association
	// whose association class ClassTLAName matches classTLAName (fromClassKey is the from end).
	OutgoingAssociationByAssociationClassTLAName(fromClassKey identity.Key, classTLAName string) (identity.Key, model_class.Association, bool)
	// OutgoingAssociationsTo lists outgoing associations from fromClassKey whose to-class is toClassKey.
	OutgoingAssociationsTo(fromClassKey, toClassKey identity.Key) []model_class.Association
	PeerClass(classKey identity.Key) (model_class.Class, bool)
	PeerCreationEvent(classKey identity.Key) (model_state.Event, bool)
	PeerEvent(classKey identity.Key, eventKey identity.Key) (model_state.Event, bool)
}
