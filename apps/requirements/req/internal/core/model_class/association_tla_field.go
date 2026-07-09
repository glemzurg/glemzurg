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

// ClassTLAName is the TLA+ identifier for a class (display name with spaces removed).
// Used as guarantee targets for association-class reification.
func ClassTLAName(className string) string {
	return AssociationTLAFieldName(className)
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

// OutgoingAssociationClassTLANameSet returns ClassTLAName values for association classes
// on associations whose from-class is classKey. classes maps class keys to display names.
func OutgoingAssociationClassTLANameSet(
	classKey identity.Key,
	associations map[identity.Key]Association,
	classes map[identity.Key]Class,
) map[string]bool {
	if len(associations) == 0 || len(classes) == 0 {
		return nil
	}
	names := make(map[string]bool)
	for _, assoc := range associations {
		if assoc.FromClassKey != classKey || assoc.AssociationClassKey == nil {
			continue
		}
		acClass, ok := classes[*assoc.AssociationClassKey]
		if !ok {
			continue
		}
		names[ClassTLAName(acClass.Name)] = true
	}
	if len(names) == 0 {
		return nil
	}
	return names
}
