package actions

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
)

// SamplingConstraintsForTest exposes constraint extraction for integration tests.
type SamplingConstraintsForTest struct {
	NullableElseMembership        *nullableElseMembershipConstraint
	NullableElseEquality          *nullableElseEqualityConstraint
	NullableElseExclusionEquality *nullableElseExclusionEqualityConstraint
	NullableElseMirror            *nullableElseMirrorConstraint
}

// ExtractSamplingConstraintsForTest returns constraints extracted from sampling logics.
func ExtractSamplingConstraintsForTest(logics []model_logic.Logic) SamplingConstraintsForTest {
	c := extractParameterConstraints(logics)
	return SamplingConstraintsForTest{
		NullableElseMembership:        c.nullableElseMembership,
		NullableElseEquality:          c.nullableElseEquality,
		NullableElseExclusionEquality: c.nullableElseExclusionEquality,
		NullableElseMirror:            c.nullableElseMirror,
	}
}

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
		tryExtractNullableElseExclusionEquality(node, constraints)
		tryExtractNullableElseMirror(node, constraints)
		tryExtractNullableElseMembership(node, constraints)
		tryExtractNullableElseEquality(node, constraints)
	case *me.Membership:
		tryExtractMembershipConstraint(node, constraints)
	case *me.BinaryLogic:
		mergeConstraintsFromExpr(node.Left, constraints)
		mergeConstraintsFromExpr(node.Right, constraints)
	}
}

func tryExtractNullableElseMirror(node *me.IfThenElse, constraints *parameterConstraints) {
	if constraints.nullableElseMirror != nil {
		return
	}

	driver, condOk := nullCompareParam(node.Condition)
	if !condOk || !isTrueLiteral(node.Then) {
		return
	}

	membership, equality, ok := mirrorElseMembershipAndEquality(node.Else)
	if !ok {
		return
	}

	memberParam, setSubKey, ok := paramMembershipInNamedSet(membership)
	if !ok || memberParam != driver {
		return
	}
	eqDriver, follower, ok := paramEquality(equality)
	if !ok || eqDriver != driver {
		return
	}

	constraints.nullableElseMirror = &nullableElseMirrorConstraint{
		driverParam:   driver,
		followerParam: follower,
		setSubKey:     setSubKey,
	}
}

func tryExtractNullableElseExclusionEquality(node *me.IfThenElse, constraints *parameterConstraints) {
	if constraints.nullableElseExclusionEquality != nil {
		return
	}

	driver, condOk := nullCompareParam(node.Condition)
	if !condOk {
		return
	}

	follower, setSubKey, thenOk := paramNotMembershipInNamedSet(node.Then)
	if !thenOk {
		return
	}

	eqDriver, eqFollower, elseOk := paramEquality(node.Else)
	if !elseOk || eqDriver != driver || eqFollower != follower {
		return
	}

	constraints.nullableElseExclusionEquality = &nullableElseExclusionEqualityConstraint{
		driverParam:   driver,
		followerParam: follower,
		setSubKey:     setSubKey,
	}
}

func tryExtractNullableElseEquality(node *me.IfThenElse, constraints *parameterConstraints) {
	if constraints.nullableElseEquality != nil {
		return
	}

	driver, condOk := nullCompareParam(node.Condition)
	if !condOk || !isTrueLiteral(node.Then) {
		return
	}

	eqDriver, follower, ok := paramEquality(node.Else)
	if !ok || eqDriver != driver {
		return
	}

	constraints.nullableElseEquality = &nullableElseEqualityConstraint{
		driverParam:   driver,
		followerParam: follower,
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

func paramNotMembershipInNamedSet(expr me.Expression) (string, string, bool) {
	membership, ok := expr.(*me.Membership)
	if !ok || !membership.Negated {
		return "", "", false
	}
	return paramMembershipInNamedSet(membership)
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

func mirrorElseMembershipAndEquality(
	elseExpr me.Expression,
) (membership *me.Membership, equality *me.Compare, ok bool) {
	and, ok := elseExpr.(*me.BinaryLogic)
	if !ok || and.Op != me.LogicAnd {
		return nil, nil, false
	}

	leftMembership, leftEq := membershipAndEqualityBranches(and.Left, and.Right)
	if leftMembership != nil {
		return leftMembership, leftEq, leftEq != nil
	}

	rightMembership, rightEq := membershipAndEqualityBranches(and.Right, and.Left)
	if rightMembership != nil {
		return rightMembership, rightEq, rightEq != nil
	}
	return nil, nil, false
}

func membershipAndEqualityBranches(
	first, second me.Expression,
) (membership *me.Membership, equality *me.Compare) {
	if m, ok := first.(*me.Membership); ok && !m.Negated {
		if eq, ok := second.(*me.Compare); ok && eq.Op == me.CompareEq {
			return m, eq
		}
	}
	return nil, nil
}

func paramEquality(expr me.Expression) (driver, follower string, ok bool) {
	cmp, ok := expr.(*me.Compare)
	if !ok || cmp.Op != me.CompareEq {
		return "", "", false
	}
	left, lok := cmp.Left.(*me.LocalVar)
	right, rok := cmp.Right.(*me.LocalVar)
	if !lok || !rok {
		return "", "", false
	}
	return left.Name, right.Name, true
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
