package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// CoerceToSet unwraps a Set or an AssociationRelation endpoint set for set operations.
func CoerceToSet(value object.Object) (*object.Set, bool) {
	switch v := value.(type) {
	case *object.Set:
		return v, true
	case *object.AssociationRelation:
		return v.Endpoints(), true
	default:
		return nil, false
	}
}
