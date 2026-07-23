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

func buildAssociationClassIndex(sch *schema.Schema, scopedClasses map[identity.Key]*ClassInfo) map[identity.Key]*AssociationClassInfo {
	index := make(map[identity.Key]*AssociationClassInfo)

	sch.ForEachAssociation(func(assoc model_class.Association) {
		if assoc.AssociationClassKey == nil {
			return
		}
		acKey := *assoc.AssociationClassKey
		if _, inScope := scopedClasses[acKey]; !inScope {
			return
		}
		if _, fromIn := scopedClasses[assoc.FromClassKey]; !fromIn {
			return
		}
		if _, toIn := scopedClasses[assoc.ToClassKey]; !toIn {
			return
		}

		index[acKey] = &AssociationClassInfo{
			AssociationClassKey: acKey,
			HostAssociation:     assoc,
			FromClassKey:        assoc.FromClassKey,
			ToClassKey:          assoc.ToClassKey,
		}
	})

	return index
}

// CreationCascadeClassKey returns the class whose creation event satisfies a mandatory host association.
func CreationCascadeClassKey(ai AssociationInfo) identity.Key {
	if ai.Association.AssociationClassKey != nil {
		return *ai.Association.AssociationClassKey
	}
	return ai.ToClassKey
}
