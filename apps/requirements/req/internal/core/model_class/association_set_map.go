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
// Accepts both bare association domains ({ Event : r \in Assoc }) and filtered/peer domains
// ({ Event : r \in { x \in Assoc : pred } }) so SentBy/caller metadata tracks cascade events.
func AssociationSetMapEventKey(expr me.Expression) (identity.Key, bool) {
	if expr == nil {
		return identity.Key{}, false
	}
	if _, eventCall, ok := MatchAssociationSetMapExpr(expr); ok {
		return eventCall.EventKey, true
	}
	// Filtered or peer-domain set-map: Set is not a bare AssociationRef.
	if setMap, ok := expr.(*me.SetMap); ok {
		if eventCall, ok := setMap.Transform.(*me.EventCall); ok {
			return eventCall.EventKey, true
		}
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
	// Accept both TLA+ ASCII `\in` and the Unicode membership sign used after normalize/raise.
	hasIn := strings.Contains(specification, `\in`) || strings.Contains(specification, "∈")
	return hasIn && strings.Contains(specification, ":")
}
