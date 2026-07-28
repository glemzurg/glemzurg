package engine

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/schema"
)

// AssociationClassInfo holds native host-association metadata for one association-class role.
type AssociationClassInfo struct {
	AssociationClassKey identity.Key
	HostAssociation     model_class.Association
	FromClassKey        identity.Key
	ToClassKey          identity.Key
}

func buildAssociationClassIndexFromSchema(sch *schema.Schema) map[identity.Key]*AssociationClassInfo {
	index := make(map[identity.Key]*AssociationClassInfo)
	for _, view := range sch.ScopedAssociations() {
		assoc := view.Association
		if assoc.AssociationClassKey == nil {
			continue
		}
		acKey := *assoc.AssociationClassKey
		host, inScope, err := sch.HostAssociationForAC(acKey)
		if err != nil || !inScope || host == nil {
			continue
		}
		index[acKey] = &AssociationClassInfo{
			AssociationClassKey: host.AssociationClassKey,
			HostAssociation:     host.HostAssociation,
			FromClassKey:        host.FromClassKey,
			ToClassKey:          host.ToClassKey,
		}
	}
	return index
}

// CreationCascadeClassKey returns the class whose creation event satisfies a mandatory host association.
func CreationCascadeClassKey(ai AssociationInfo) identity.Key {
	if ai.Association.AssociationClassKey != nil {
		return *ai.Association.AssociationClassKey
	}
	return ai.ToClassKey
}
