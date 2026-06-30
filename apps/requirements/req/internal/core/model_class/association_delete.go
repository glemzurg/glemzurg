package model_class

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// MatchAssociationDeleteGuarantee recognizes a lowered delete guarantee and its peer
// delete_event call. Specification may be either:
//   - a set-filter selection: { v \in AssocField : pred } (requires a separate state_change), or
//   - an inline association update: AssocField \ { v \in AssocField : pred }.
func MatchAssociationDeleteGuarantee(logic model_logic.Logic) (*me.AssociationRef, *me.SetFilter, *me.EventCall, bool) {
	if logic.Type != model_logic.LogicTypeDelete {
		return nil, nil, nil, false
	}
	eventCall, ok := logic.DeleteEventSpec.Expression.(*me.EventCall)
	if !ok {
		return nil, nil, nil, false
	}
	selection, assocRef, ok := matchDeleteGuaranteeSelection(logic.Spec.Expression)
	if !ok {
		return nil, nil, nil, false
	}
	return assocRef, selection, eventCall, true
}

// DeleteGuaranteeHasInlineStateChange reports whether the delete guarantee specification
// lowers to an association set-difference (state_change and delete selection in one guarantee).
func DeleteGuaranteeHasInlineStateChange(logic model_logic.Logic) bool {
	if logic.Type != model_logic.LogicTypeDelete {
		return false
	}
	setOp, ok := logic.Spec.Expression.(*me.SetOp)
	return ok && setOp.Op == me.SetDifference
}

func matchDeleteGuaranteeSelection(expr me.Expression) (*me.SetFilter, *me.AssociationRef, bool) {
	if selection, assocRef, ok := matchDeleteGuaranteeSelectionFilter(expr); ok {
		return selection, assocRef, true
	}
	return matchDeleteGuaranteeDifferenceSelection(expr)
}

func matchDeleteGuaranteeSelectionFilter(expr me.Expression) (*me.SetFilter, *me.AssociationRef, bool) {
	selection, ok := expr.(*me.SetFilter)
	if !ok {
		return nil, nil, false
	}
	assocRef, ok := selection.Set.(*me.AssociationRef)
	if !ok {
		return nil, nil, false
	}
	return selection, assocRef, true
}

func matchDeleteGuaranteeDifferenceSelection(expr me.Expression) (*me.SetFilter, *me.AssociationRef, bool) {
	setOp, ok := expr.(*me.SetOp)
	if !ok || setOp.Op != me.SetDifference {
		return nil, nil, false
	}
	assocRef, ok := setOp.Left.(*me.AssociationRef)
	if !ok {
		return nil, nil, false
	}
	selection, ok := setOp.Right.(*me.SetFilter)
	if !ok {
		return nil, nil, false
	}
	rightAssoc, ok := selection.Set.(*me.AssociationRef)
	if !ok || rightAssoc.AssociationKey != assocRef.AssociationKey {
		return nil, nil, false
	}
	return selection, assocRef, true
}

// AssociationDeleteEventKey returns the peer event key from a delete guarantee.
func AssociationDeleteEventKey(logic model_logic.Logic) (identity.Key, bool) {
	_, _, eventCall, ok := MatchAssociationDeleteGuarantee(logic)
	if !ok {
		return identity.Key{}, false
	}
	return eventCall.EventKey, true
}
