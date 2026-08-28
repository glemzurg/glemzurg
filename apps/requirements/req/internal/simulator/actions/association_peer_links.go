package actions

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"
)

// linkedAssociationPeerEndpoints returns to-endpoint instance IDs linked from fromID
// for an outgoing association. AC-hosted associations resolve via materialized rows.
func linkedAssociationPeerEndpoints(
	simState *instance.State,
	fromID instance.ID,
	assoc model_class.Association,
) []instance.ID {
	if assoc.AssociationClassKey != nil {
		links := simState.AssociationLinksFromEndpoint(assoc.Key, fromID)
		if len(links) == 0 {
			return nil
		}
		peerIDs := make([]instance.ID, 0, len(links))
		for _, link := range links {
			if simState.GetInstance(link.LinkInstanceID) == nil {
				continue
			}
			peerIDs = append(peerIDs, link.ToEndpointID)
		}
		return peerIDs
	}
	return simState.GetLinkedForward(fromID, assoc.Key)
}

// associationLinkForPair returns the materialized host row for from→to when the
// association is AC-hosted.
func associationLinkForPair(
	simState *instance.State,
	assoc model_class.Association,
	fromID, toID instance.ID,
) (instance.AssociationLink, bool) {
	if assoc.AssociationClassKey == nil {
		return instance.AssociationLink{}, false
	}
	for _, link := range simState.AssociationLinksFromEndpoint(assoc.Key, fromID) {
		if link.ToEndpointID != toID {
			continue
		}
		if simState.GetInstance(link.LinkInstanceID) == nil {
			continue
		}
		return link, true
	}
	return instance.AssociationLink{}, false
}
