package actions

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
)

func extractParameterConstraints(requires []model_logic.Logic) parameterConstraints {
	constraints := parameterConstraints{
		enumValues: make(map[string][]string),
	}

	for _, req := range requires {
		if req.Type != model_logic.LogicTypeAssessment || !req.Spec.ParseOk() {
			continue
		}
		mergeConstraintsFromExpr(req.Spec.Expression, &constraints)
	}

	return constraints
}

func mergeConstraintsFromExpr(expr me.Expression, constraints *parameterConstraints) {
	if expr == nil {
		return
	}

	switch node := expr.(type) {
	case *me.IfThenElse:
		tryExtractNullableElseTuple(node, constraints)
		tryExtractNullableElseMembership(node, constraints)
	case *me.Membership:
		tryExtractMembershipConstraint(node, constraints)
	case *me.BinaryLogic:
		mergeConstraintsFromExpr(node.Left, constraints)
		mergeConstraintsFromExpr(node.Right, constraints)
	}
}

func tryExtractNullableElseMembership(node *me.IfThenElse, constraints *parameterConstraints) {
	if constraints.nullableElseMembership != nil {
		return
	}

	paramName, condOk := nullCompareParam(node.Condition)
	if !condOk || !isTrueLiteral(node.Then) {
		return
	}

	membership, ok := node.Else.(*me.Membership)
	if !ok || membership.Negated {
		return
	}

	memberParam, setSubKey, ok := paramMembershipInNamedSet(membership)
	if !ok || memberParam != paramName {
		return
	}

	constraints.nullableElseMembership = &nullableElseMembershipConstraint{
		paramName: paramName,
		setSubKey: setSubKey,
	}
}

func tryExtractNullableElseTuple(node *me.IfThenElse, constraints *parameterConstraints) {
	if constraints.nullableElseTuple != nil {
		return
	}

	conditionParam, condOk := nullCompareParam(node.Condition)
	thenParam, thenOk := nullCompareParam(node.Then)
	if !condOk || !thenOk {
		return
	}

	membership, ok := node.Else.(*me.Membership)
	if !ok || membership.Negated {
		return
	}

	paramNames, setSubKey, ok := tupleMembershipInNamedSet(membership)
	if !ok {
		return
	}

	constraints.nullableElseTuple = &nullableElseTupleConstraint{
		conditionParam: conditionParam,
		thenParam:      thenParam,
		paramNames:     paramNames,
		setSubKey:      setSubKey,
	}
}

func tryExtractMembershipConstraint(node *me.Membership, constraints *parameterConstraints) {
	if node.Negated {
		return
	}

	if paramNames, setSubKey, ok := tupleMembershipInNamedSet(node); ok {
		if constraints.tupleInSet == nil {
			constraints.tupleInSet = &tupleInSetConstraint{
				paramNames: paramNames,
				setSubKey:  setSubKey,
			}
		}
		return
	}

	if paramName, values, ok := paramInStringEnum(node); ok {
		constraints.enumValues[paramName] = values
	}
}

func nullCompareParam(expr me.Expression) (string, bool) {
	cmp, ok := expr.(*me.Compare)
	if !ok || cmp.Op != me.CompareEq {
		return "", false
	}

	localVar, ok := cmp.Left.(*me.LocalVar)
	if !ok || !isEmptySetLiteral(cmp.Right) {
		return "", false
	}

	return localVar.Name, true
}

func isEmptySetLiteral(expr me.Expression) bool {
	literal, ok := expr.(*me.SetLiteral)
	return ok && len(literal.Elements) == 0
}

func isTrueLiteral(expr me.Expression) bool {
	literal, ok := expr.(*me.BoolLiteral)
	return ok && literal.Value
}

func paramMembershipInNamedSet(node *me.Membership) (string, string, bool) {
	localVar, ok := node.Element.(*me.LocalVar)
	if !ok {
		return "", "", false
	}

	ref, ok := node.Set.(*me.NamedSetRef)
	if !ok {
		return "", "", false
	}

	return localVar.Name, ref.SetKey.SubKey, true
}

func tupleMembershipInNamedSet(node *me.Membership) ([]string, string, bool) {
	tuple, ok := node.Element.(*me.TupleLiteral)
	if !ok || len(tuple.Elements) == 0 {
		return nil, "", false
	}

	paramNames := make([]string, len(tuple.Elements))
	for i, element := range tuple.Elements {
		localVar, ok := element.(*me.LocalVar)
		if !ok {
			return nil, "", false
		}
		paramNames[i] = localVar.Name
	}

	ref, ok := node.Set.(*me.NamedSetRef)
	if !ok {
		return nil, "", false
	}

	return paramNames, ref.SetKey.SubKey, true
}

func paramInStringEnum(node *me.Membership) (string, []string, bool) {
	localVar, ok := node.Element.(*me.LocalVar)
	if !ok {
		return "", nil, false
	}

	setLiteral, ok := node.Set.(*me.SetLiteral)
	if !ok || len(setLiteral.Elements) == 0 {
		return "", nil, false
	}

	values := make([]string, 0, len(setLiteral.Elements))
	for _, element := range setLiteral.Elements {
		stringLiteral, ok := element.(*me.StringLiteral)
		if !ok {
			return "", nil, false
		}
		values = append(values, stringLiteral.Value)
	}

	return localVar.Name, values, true
}
