package model_domain

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// ModelCrossRefs supplies model-wide lookup maps for domain and subdomain validation.
type ModelCrossRefs struct {
	Actors             map[identity.Key]bool
	Classes            map[identity.Key]bool
	AllGeneralizations map[identity.Key]bool
	AllClasses         map[identity.Key]model_class.Class
	AllAssociations    map[identity.Key]model_class.Association
}
