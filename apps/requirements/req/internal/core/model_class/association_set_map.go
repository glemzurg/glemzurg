package model_class

import (
	"strings"

	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// MatchAssociationSetMapExpr recognizes {Transform(r) : r \in associationField} lowered form:
// SetMap{Set: AssociationRef, Transform: EventCall}.
func MatchAssociationSetMapExpr(expr me.Expression) (*me.AssociationRef, *me.EventCall, bool) {
	setMap, ok := expr.(*me.SetMap)
	if !ok {
		return nil, nil, false
	}
	assocRef, ok := setMap.Set.(*me.AssociationRef)
	if !ok {
		return nil, nil, false
	}
	eventCall, ok := setMap.Transform.(*me.EventCall)
	if !ok {
		return nil, nil, false
	}
	return assocRef, eventCall, true
}

// MatchAssociationAddOrUpdateExpr recognizes IF empty THEN set-add ELSE set-map on one association.
func MatchAssociationAddOrUpdateExpr(expr me.Expression) (*me.AssociationRef, *me.EventCall, *me.EventCall, bool) {
	ifte, ok := expr.(*me.IfThenElse)
	if !ok {
		return nil, nil, nil, false
	}
	assocThen, createCall, ok := MatchAssociationSetAddExpr(ifte.Then)
	if !ok {
		return nil, nil, nil, false
	}
	assocElse, updateCall, ok := MatchAssociationSetMapExpr(ifte.Else)
	if !ok {
		return nil, nil, nil, false
	}
	if assocThen.AssociationKey != assocElse.AssociationKey {
		return nil, nil, nil, false
	}
	return assocThen, createCall, updateCall, true
}

// IsAssociationSetMapSpecification reports the authored TLA shorthand for association set-map.
func IsAssociationSetMapSpecification(specification string) bool {
	return isAssociationSetMapSpecification(specification)
}

// IsAssociationAddOrUpdateSpecification reports add-or-update IF/union/set-map guarantee text.
func IsAssociationAddOrUpdateSpecification(specification string) bool {
	if specification == "" {
		return false
	}
	lower := strings.ToLower(specification)
	hasUnion := strings.Contains(lower, `\union`) || strings.Contains(specification, "∪")
	return strings.Contains(lower, "if ") &&
		strings.Contains(lower, "then ") &&
		strings.Contains(lower, "else ") &&
		hasUnion &&
		strings.Contains(lower, `\in`)
}

// AssociationSetMapEventKey returns the peer event key referenced by a set-map guarantee expression.
func AssociationSetMapEventKey(expr me.Expression) (identity.Key, bool) {
	if expr == nil {
		return identity.Key{}, false
	}
	if _, eventCall, ok := MatchAssociationSetMapExpr(expr); ok {
		return eventCall.EventKey, true
	}
	if _, _, updateCall, ok := MatchAssociationAddOrUpdateExpr(expr); ok {
		return updateCall.EventKey, true
	}
	return identity.Key{}, false
}

func isAssociationSetMapSpecification(specification string) bool {
	if specification == "" {
		return false
	}
	if isAssociationSetAddSpecification(specification) {
		return false
	}
	lower := strings.ToLower(specification)
	return strings.Contains(lower, `\in`) && strings.Contains(specification, ":")
}
