package model_class

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// MatchAssociationDestroyGuarantee recognizes a lowered destroy guarantee and its peer
// destroy_event call. Specification may be either:
//   - a set-filter selection: { v \in AssocField : pred } (requires a separate state_change), or
//   - an inline association update: AssocField \ { v \in AssocField : pred }.
func MatchAssociationDestroyGuarantee(logic model_logic.Logic) (*me.AssociationRef, *me.SetFilter, *me.EventCall, bool) {
	if logic.Type != model_logic.LogicTypeDestroy {
		return nil, nil, nil, false
	}
	eventCall, ok := logic.DestroyEventSpec.Expression.(*me.EventCall)
	if !ok {
		return nil, nil, nil, false
	}
	selection, assocRef, ok := matchDestroyGuaranteeSelection(logic.Spec.Expression)
	if !ok {
		return nil, nil, nil, false
	}
	return assocRef, selection, eventCall, true
}

// DestroyGuaranteeHasInlineStateChange reports whether the destroy guarantee specification
// lowers to an association set-difference (state_change and destroy selection in one guarantee).
func DestroyGuaranteeHasInlineStateChange(logic model_logic.Logic) bool {
	if logic.Type != model_logic.LogicTypeDestroy {
		return false
	}
	setOp, ok := logic.Spec.Expression.(*me.SetOp)
	return ok && setOp.Op == me.SetDifference
}

func matchDestroyGuaranteeSelection(expr me.Expression) (*me.SetFilter, *me.AssociationRef, bool) {
	if selection, assocRef, ok := matchDestroyGuaranteeSelectionFilter(expr); ok {
		return selection, assocRef, true
	}
	return matchDestroyGuaranteeDifferenceSelection(expr)
}

func matchDestroyGuaranteeSelectionFilter(expr me.Expression) (*me.SetFilter, *me.AssociationRef, bool) {
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

func matchDestroyGuaranteeDifferenceSelection(expr me.Expression) (*me.SetFilter, *me.AssociationRef, bool) {
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

// AssociationDestroyEventKey returns the peer event key from a destroy guarantee.
func AssociationDestroyEventKey(logic model_logic.Logic) (identity.Key, bool) {
	_, _, eventCall, ok := MatchAssociationDestroyGuarantee(logic)
	if !ok {
		return identity.Key{}, false
	}
	return eventCall.EventKey, true
}
