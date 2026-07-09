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
	// OutgoingAssociationByAssociationClassTLAName resolves the unique host association
	// whose association class ClassTLAName matches classTLAName (fromClassKey is the from end).
	OutgoingAssociationByAssociationClassTLAName(fromClassKey identity.Key, classTLAName string) (identity.Key, model_class.Association, bool)
	PeerClass(classKey identity.Key) (model_class.Class, bool)
	PeerCreationEvent(classKey identity.Key) (model_state.Event, bool)
	PeerEvent(classKey identity.Key, eventKey identity.Key) (model_state.Event, bool)
}
