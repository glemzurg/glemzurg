package model_class

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// MatchAssociationDeleteGuarantee recognizes a lowered delete guarantee:
// selection is a set-filter over an association field; delete_event is a peer EventCall.
func MatchAssociationDeleteGuarantee(logic model_logic.Logic) (*me.AssociationRef, *me.SetFilter, *me.EventCall, bool) {
	if logic.Type != model_logic.LogicTypeDelete {
		return nil, nil, nil, false
	}
	selection, ok := logic.Spec.Expression.(*me.SetFilter)
	if !ok {
		return nil, nil, nil, false
	}
	assocRef, ok := selection.Set.(*me.AssociationRef)
	if !ok {
		return nil, nil, nil, false
	}
	eventCall, ok := logic.DeleteEventSpec.Expression.(*me.EventCall)
	if !ok {
		return nil, nil, nil, false
	}
	return assocRef, selection, eventCall, true
}

// AssociationDeleteEventKey returns the peer event key from a delete guarantee.
func AssociationDeleteEventKey(logic model_logic.Logic) (identity.Key, bool) {
	_, _, eventCall, ok := MatchAssociationDeleteGuarantee(logic)
	if !ok {
		return identity.Key{}, false
	}
	return eventCall.EventKey, true
}
