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
	case *me.BuiltinCall:
		// Null-branch synthesis is driven only by _GZ sugar; raw IF no longer synthesizes.
		tryExtractGZConstraints(node, constraints)
	case *me.IfThenElse:
		// Assessment still runs IF; sampling constraints are not inferred from IF shapes.
		return
	case *me.Membership:
		tryExtractMembershipConstraint(node, constraints)
	case *me.BinaryLogic:
		mergeConstraintsFromExpr(node.Left, constraints)
		mergeConstraintsFromExpr(node.Right, constraints)
		tryCoupleComplementaryGZConstraints(constraints)
	case *me.Quantifier:
		tryExtractPeerFieldDistinctFromParam(node, constraints)
	}
}

// tryExtractGZConstraints maps _GZ!WhenNotNull / WhenNull / WhenNullElse onto sampling constraints.
func tryExtractGZConstraints(call *me.BuiltinCall, constraints *parameterConstraints) {
	if call == nil || call.Module != gzModuleName {
		return
	}
	switch call.Function {
	case gzFnWhenNotNull:
		if len(call.Args) != 2 {
			return
		}
		driver, ok := gzDriverName(call.Args[0])
		if !ok {
			return
		}
		extractWhenNotNullEquation(driver, call.Args[1], constraints)
	case gzFnWhenNull:
		if len(call.Args) != 2 {
			return
		}
		driver, ok := gzDriverName(call.Args[0])
		if !ok {
			return
		}
		extractWhenNullEquation(driver, call.Args[1], constraints)
	case gzFnWhenNullElse:
		if len(call.Args) != 3 {
			return
		}
		driver, ok := gzDriverName(call.Args[0])
		if !ok {
			return
		}
		extractWhenNullElseEquations(driver, call.Args[1], call.Args[2], constraints)
	}
}

func extractWhenNotNullEquation(driver string, eq me.Expression, constraints *parameterConstraints) {
	if tryExtractMirrorFromEquation(driver, eq, constraints) {
		return
	}
	if m, ok := eq.(*me.Membership); ok && !m.Negated {
		if paramName, setSubKey, ok := paramMembershipInNamedSet(m); ok && paramName == driver {
			if constraints.nullableElseMembership == nil {
				constraints.nullableElseMembership = &nullableElseMembershipConstraint{
					paramName: paramName,
					setSubKey: setSubKey,
				}
			}
			return
		}
		// Non-nullable-style membership still useful under WhenNotNull (nullable driver).
		tryExtractMembershipConstraint(m, constraints)
		return
	}
	if eqDriver, follower, ok := paramEquality(eq); ok && eqDriver == driver {
		if constraints.nullableElseEquality == nil {
			constraints.nullableElseEquality = &nullableElseEqualityConstraint{
				driverParam:   driver,
				followerParam: follower,
			}
		}
		return
	}
	// Auto-wrapped parameter invariants may nest peer-distinct quantifiers under WhenNotNull.
	if q, ok := eq.(*me.Quantifier); ok {
		tryExtractPeerFieldDistinctFromParam(q, constraints)
		return
	}
	// AND of synthesizable arms (e.g. membership ∧ peer-distinct) inside WhenNotNull.
	if and, ok := eq.(*me.BinaryLogic); ok && and.Op == me.LogicAnd {
		extractWhenNotNullEquation(driver, and.Left, constraints)
		extractWhenNotNullEquation(driver, and.Right, constraints)
	}
}

func extractWhenNullEquation(driver string, eq me.Expression, constraints *parameterConstraints) {
	if follower, value, ok := paramCompareBoolLiteral(eq); ok {
		if constraints.nullableElseBooleanConstant == nil {
			constraints.nullableElseBooleanConstant = &nullableElseBooleanConstantConstraint{
				driverParam:   driver,
				followerParam: follower,
				value:         value,
			}
		}
		return
	}
	// Partial for complementary pairing: follower ∉ named set when driver is null.
	if follower, setSubKey, ok := paramNotMembershipInNamedSet(eq); ok {
		if constraints.gzNullExclusion == nil {
			constraints.gzNullExclusion = &gzNullBranchExclusion{
				driverParam:   driver,
				followerParam: follower,
				setSubKey:     setSubKey,
			}
		}
		return
	}
	// Partial for tuple: follower = NULL when driver is null.
	if thenParam, ok := nullCompareParam(eq); ok {
		if constraints.gzNullTupleFollower == nil {
			constraints.gzNullTupleFollower = &gzNullBranchTupleFollower{
				driverParam: driver,
				thenParam:   thenParam,
			}
		}
	}
}

func extractWhenNullElseEquations(driver string, nullEq, setEq me.Expression, constraints *parameterConstraints) {
	// Exclusion equality: IF D=NULL THEN F∉S ELSE D=F
	if follower, setSubKey, thenOk := paramNotMembershipInNamedSet(nullEq); thenOk {
		eqDriver, eqFollower, elseOk := paramEquality(setEq)
		if elseOk && eqDriver == driver && eqFollower == follower {
			if constraints.nullableElseExclusionEquality == nil {
				constraints.nullableElseExclusionEquality = &nullableElseExclusionEqualityConstraint{
					driverParam:   driver,
					followerParam: follower,
					setSubKey:     setSubKey,
				}
			}
			return
		}
	}

	// Tuple: IF C=NULL THEN T=NULL ELSE <<…>> ∈ S
	if thenParam, thenOk := nullCompareParam(nullEq); thenOk {
		if membership, ok := setEq.(*me.Membership); ok && !membership.Negated {
			paramNames, setSubKey, ok := tupleMembershipInNamedSet(membership)
			if ok {
				if constraints.nullableElseTuple == nil {
					constraints.nullableElseTuple = &nullableElseTupleConstraint{
						conditionParam: driver,
						thenParam:      thenParam,
						paramNames:     paramNames,
						setSubKey:      setSubKey,
					}
				}
				return
			}
		}
	}

	// Boolean when null, true when set: IF D=NULL THEN F=c ELSE TRUE
	if isTrueLiteral(setEq) {
		extractWhenNullEquation(driver, nullEq, constraints)
		return
	}

	// Vacuous null arm: IF D=NULL THEN TRUE ELSE eq  ≡ WhenNotNull
	if isTrueLiteral(nullEq) {
		extractWhenNotNullEquation(driver, setEq, constraints)
		return
	}

	// Fall back: extract each arm independently then couple.
	extractWhenNullEquation(driver, nullEq, constraints)
	extractWhenNotNullEquation(driver, setEq, constraints)
	tryCoupleComplementaryGZConstraints(constraints)
}

func tryExtractMirrorFromEquation(driver string, eq me.Expression, constraints *parameterConstraints) bool {
	membership, equality, ok := mirrorElseMembershipAndEquality(eq)
	if !ok {
		return false
	}
	memberParam, setSubKey, ok := paramMembershipInNamedSet(membership)
	if !ok || memberParam != driver {
		return false
	}
	eqDriver, follower, ok := paramEquality(equality)
	if !ok || eqDriver != driver {
		return false
	}
	if constraints.nullableElseMirror == nil {
		constraints.nullableElseMirror = &nullableElseMirrorConstraint{
			driverParam:   driver,
			followerParam: follower,
			setSubKey:     setSubKey,
		}
	}
	return true
}

// tryCoupleComplementaryGZConstraints rebuilds exclusion/tuple constraints from
// complementary WhenNull + WhenNotNull partials (or WhenNullElse fall-back pieces).
func tryCoupleComplementaryGZConstraints(constraints *parameterConstraints) {
	if constraints.nullableElseExclusionEquality == nil &&
		constraints.gzNullExclusion != nil &&
		constraints.nullableElseEquality != nil &&
		constraints.gzNullExclusion.driverParam == constraints.nullableElseEquality.driverParam &&
		constraints.gzNullExclusion.followerParam == constraints.nullableElseEquality.followerParam {
		constraints.nullableElseExclusionEquality = &nullableElseExclusionEqualityConstraint{
			driverParam:   constraints.gzNullExclusion.driverParam,
			followerParam: constraints.gzNullExclusion.followerParam,
			setSubKey:     constraints.gzNullExclusion.setSubKey,
		}
		constraints.gzNullExclusion = nil
		constraints.nullableElseEquality = nil
	}

	if constraints.nullableElseTuple == nil &&
		constraints.gzNullTupleFollower != nil &&
		constraints.tupleInSet != nil &&
		len(constraints.tupleInSet.paramNames) > 0 {
		// Driver should be first tuple element for jurisdiction-style tuples.
		driver := constraints.gzNullTupleFollower.driverParam
		if constraints.tupleInSet.paramNames[0] == driver {
			constraints.nullableElseTuple = &nullableElseTupleConstraint{
				conditionParam: driver,
				thenParam:      constraints.gzNullTupleFollower.thenParam,
				paramNames:     constraints.tupleInSet.paramNames,
				setSubKey:      constraints.tupleInSet.setSubKey,
			}
			constraints.gzNullTupleFollower = nil
			constraints.tupleInSet = nil
		}
	}
}

const (
	gzModuleName     = "_GZ"
	gzFnWhenNotNull  = "WhenNotNull"
	gzFnWhenNull     = "WhenNull"
	gzFnWhenNullElse = "WhenNullElse"
)

// gzDriverName returns the synthesis key for a _GZ driver argument.
// Accepts bare params (LocalVar) and field paths (self.field / FieldAccess).
func gzDriverName(expr me.Expression) (string, bool) {
	if expr == nil {
		return "", false
	}
	if localVar, ok := expr.(*me.LocalVar); ok {
		return localVar.Name, true
	}
	if attr, ok := expr.(*me.AttributeRef); ok {
		return attr.AttributeKey.SubKey, true
	}
	if fa, ok := expr.(*me.FieldAccess); ok {
		return fa.Field, true
	}
	return "", false
}

func tryExtractPeerFieldDistinctFromParam(node *me.Quantifier, constraints *parameterConstraints) {
	if constraints.peerFieldDistinct != nil {
		return
	}
	if pattern, ok := detectPeerFieldDistinctFromParam(node); ok {
		constraints.peerFieldDistinct = &pattern
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
