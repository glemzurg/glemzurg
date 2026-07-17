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
		tryExtractNullableElseBooleanConstant(node, constraints)
		mergeConstraintsFromExpr(node.Then, constraints)
		mergeConstraintsFromExpr(node.Else, constraints)
	case *me.Membership:
		tryExtractMembershipConstraint(node, constraints)
	case *me.BinaryLogic:
		mergeConstraintsFromExpr(node.Left, constraints)
		mergeConstraintsFromExpr(node.Right, constraints)
	case *me.Quantifier:
		tryExtractPeerFieldDistinctFromParam(node, constraints)
	}
}

func tryExtractPeerFieldDistinctFromParam(node *me.Quantifier, constraints *parameterConstraints) {
	if constraints.peerFieldDistinct != nil {
		return
	}
	if pattern, ok := detectPeerFieldDistinctFromParam(node); ok {
		constraints.peerFieldDistinct = &pattern
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

func tryExtractNullableElseBooleanConstant(node *me.IfThenElse, constraints *parameterConstraints) {
	if constraints.nullableElseBooleanConstant != nil {
		return
	}

	driver, condOk := nullCompareParam(node.Condition)
	if !condOk || !isTrueLiteral(node.Else) {
		return
	}

	follower, value, ok := paramCompareBoolLiteral(node.Then)
	if !ok {
		return
	}

	constraints.nullableElseBooleanConstant = &nullableElseBooleanConstantConstraint{
		driverParam:   driver,
		followerParam: follower,
		value:         value,
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

	if pattern, ok := detectParamInNamedSetMinusPeerField(node); ok {
		if constraints.paramInNamedSetMinusPeerField == nil {
			constraints.paramInNamedSetMinusPeerField = &pattern
		}
		return
	}

	if paramName, setSubKey, ok := paramMembershipInNamedSet(node); ok {
		if constraints.paramInNamedSet == nil {
			constraints.paramInNamedSet = &paramInNamedSetConstraint{
				paramName: paramName,
				setSubKey: setSubKey,
			}
		}
		return
	}

	if paramName, values, ok := paramInStringEnum(node); ok {
		constraints.enumValues[paramName] = values
		return
	}

	if paramName, ok := paramInBooleanSet(node); ok {
		constraints.enumValues[paramName] = booleanTLAEnumValues
	}
}

// detectParamInNamedSetMinusPeerField matches
// Param ∈ (NamedSet \ { v.field : v ∈ Class }).
func detectParamInNamedSetMinusPeerField(node *me.Membership) (paramInNamedSetMinusPeerFieldConstraint, bool) {
	empty := paramInNamedSetMinusPeerFieldConstraint{}
	if node == nil || node.Negated {
		return empty, false
	}
	localVar, ok := node.Element.(*me.LocalVar)
	if !ok {
		return empty, false
	}
	setOp, ok := node.Set.(*me.SetOp)
	if !ok || setOp.Op != me.SetDifference {
		return empty, false
	}
	namedSet, ok := setOp.Left.(*me.NamedSetRef)
	if !ok {
		return empty, false
	}
	setMap, ok := setOp.Right.(*me.SetMap)
	if !ok {
		return empty, false
	}
	classRef, ok := setMap.Set.(*me.ClassRef)
	if !ok {
		return empty, false
	}
	fieldAccess, ok := setMap.Transform.(*me.FieldAccess)
	if !ok {
		return empty, false
	}
	baseVar, ok := fieldAccess.Base.(*me.LocalVar)
	if !ok || baseVar.Name != setMap.Variable {
		return empty, false
	}
	return paramInNamedSetMinusPeerFieldConstraint{
		paramName:   localVar.Name,
		setSubKey:   namedSet.SetKey.SubKey,
		classKey:    classRef.ClassKey,
		className:   classRef.Name,
		fieldSubKey: fieldAccess.Field,
	}, true
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

func paramCompareBoolLiteral(expr me.Expression) (paramName string, value bool, ok bool) {
	cmp, ok := expr.(*me.Compare)
	if !ok || cmp.Op != me.CompareEq {
		return "", false, false
	}

	if localVar, lok := cmp.Left.(*me.LocalVar); lok {
		if boolLit, rok := cmp.Right.(*me.BoolLiteral); rok {
			return localVar.Name, boolLit.Value, true
		}
	}
	if boolLit, lok := cmp.Left.(*me.BoolLiteral); lok {
		if localVar, rok := cmp.Right.(*me.LocalVar); rok {
			return localVar.Name, boolLit.Value, true
		}
	}
	return "", false, false
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

func paramInBooleanSet(node *me.Membership) (string, bool) {
	localVar, ok := node.Element.(*me.LocalVar)
	if !ok {
		return "", false
	}

	setConst, ok := node.Set.(*me.SetConstant)
	if !ok || setConst.Kind != me.SetConstantBoolean {
		return "", false
	}

	return localVar.Name, true
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
