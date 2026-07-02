package model_class

import (
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// AssociationTLAFieldName is the TLA+ identifier for an association on self:
// the association display name with spaces removed (case preserved).
func AssociationTLAFieldName(associationName string) string {
	return strings.ReplaceAll(strings.TrimSpace(associationName), " ", "")
}

// OutgoingAssociationTLAFieldSet returns TLA field names for associations whose from-class is classKey.
func OutgoingAssociationTLAFieldSet(classKey identity.Key, associations map[identity.Key]Association) map[string]bool {
	if len(associations) == 0 {
		return nil
	}
	fields := make(map[string]bool)
	for _, assoc := range associations {
		if assoc.FromClassKey == classKey {
			fields[AssociationTLAFieldName(assoc.Name)] = true
		}
	}
	if len(fields) == 0 {
		return nil
	}
	return fields
}
