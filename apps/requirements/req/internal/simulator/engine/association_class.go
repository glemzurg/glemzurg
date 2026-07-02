package engine

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// AssociationClassInfo holds native host-association metadata for one association-class role.
type AssociationClassInfo struct {
	AssociationClassKey identity.Key
	HostAssociation     model_class.Association
	FromClassKey        identity.Key
	ToClassKey          identity.Key
}

func buildAssociationClassIndex(model *core.Model, scopedClasses map[identity.Key]*ClassInfo) map[identity.Key]*AssociationClassInfo {
	index := make(map[identity.Key]*AssociationClassInfo)

	for _, assoc := range model.GetClassAssociations() {
		if assoc.AssociationClassKey == nil {
			continue
		}
		acKey := *assoc.AssociationClassKey
		if _, inScope := scopedClasses[acKey]; !inScope {
			continue
		}
		if _, fromIn := scopedClasses[assoc.FromClassKey]; !fromIn {
			continue
		}
		if _, toIn := scopedClasses[assoc.ToClassKey]; !toIn {
			continue
		}

		index[acKey] = &AssociationClassInfo{
			AssociationClassKey: acKey,
			HostAssociation:     assoc,
			FromClassKey:        assoc.FromClassKey,
			ToClassKey:          assoc.ToClassKey,
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
