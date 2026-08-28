package evaluator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// CoerceToSet unwraps a Set or an AssociationRelation endpoint set for set operations.
// A sole association-class row (Record) becomes a singleton set so filters like
// { x \in self.IsSubdividedInto.CurrencyWalletDefinition : … } work when only one link exists.
func CoerceToSet(value object.Object) (*object.Set, bool) {
	switch v := value.(type) {
	case *object.Set:
		return v, true
	case *object.AssociationRelation:
		return v.Endpoints(), true
	case *object.Record:
		s := object.NewSet()
		s.Add(v)
		return s, true
	default:
		return nil, false
	}
}
