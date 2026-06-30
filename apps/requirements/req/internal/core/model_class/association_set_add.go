package model_class

import (
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
)

// MatchAssociationSetAddExpr recognizes association-field' = field \union {_new(...)} lowered form:
// SetUnion(AssociationRef, SetLiteral{EventCall}).
func MatchAssociationSetAddExpr(expr me.Expression) (*me.AssociationRef, *me.EventCall, bool) {
	setOp, ok := expr.(*me.SetOp)
	if !ok || setOp.Op != me.SetUnion {
		return nil, nil, false
	}
	assocRef, ok := setOp.Left.(*me.AssociationRef)
	if !ok {
		return nil, nil, false
	}
	lit, ok := setOp.Right.(*me.SetLiteral)
	if !ok || len(lit.Elements) != 1 {
		return nil, nil, false
	}
	eventCall, ok := lit.Elements[0].(*me.EventCall)
	if !ok {
		return nil, nil, false
	}
	return assocRef, eventCall, true
}

// IsAssociationSetAddSpecification reports the authored TLA shorthand for association set-add.
func IsAssociationSetAddSpecification(specification string) bool {
	return isAssociationSetAddSpecification(specification)
}
