package model_class

import "github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

// IsReverseInvariantOnlyAssociation reports whether assoc is a from-to mirror of another
// association that owns the association class. Such reverse declarations exist only to author
// invariants from the opposite anchor and must not register separate link metadata.
func IsReverseInvariantOnlyAssociation(associations map[identity.Key]Association, assoc Association) bool {
	if assoc.AssociationClassKey != nil {
		return false
	}
	for _, other := range associations {
		if other.Key == assoc.Key || other.Name != assoc.Name {
			continue
		}
		if other.FromClassKey != assoc.ToClassKey || other.ToClassKey != assoc.FromClassKey {
			continue
		}
		if other.AssociationClassKey != nil {
			return true
		}
	}
	return false
}
