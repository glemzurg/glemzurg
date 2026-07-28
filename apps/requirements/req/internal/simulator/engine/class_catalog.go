package engine

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/schema"
)

// ClassCatalog is the simulation static catalog; implementation lives in schema.
type ClassCatalog = schema.Catalog

// ClassInfo is simulation metadata for one class (schema.ClassSimInfo).
type ClassInfo = schema.ClassSimInfo

// EventInfo pairs an event with transitions (schema.EventInfo).
type EventInfo = schema.EventInfo

// AssociationInfo holds association metadata for simulation queries.
type AssociationInfo = schema.AssociationInfo

// AssociationClassInfo holds host-association metadata for an association-class role.
type AssociationClassInfo = schema.AssociationClassInfo

// NewClassCatalog returns the schema-owned catalog (built at schema.New).
func NewClassCatalog(sch *schema.Schema) *ClassCatalog {
	if sch == nil {
		return nil
	}
	return sch.Catalog()
}

// CreationCascadeClassKey returns the class whose creation event satisfies a mandatory host association.
func CreationCascadeClassKey(ai AssociationInfo) identity.Key {
	return schema.CreationCascadeClassKey(ai)
}

// Silence unused import if only used for documentation of model types.
var _ = model_class.Class{}
