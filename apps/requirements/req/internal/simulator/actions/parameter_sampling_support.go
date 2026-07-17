package actions

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
)

// UnsupportedRequiresSamplingError means a parsed action require references event parameters
// using expression shapes the simulator cannot turn into random parameter values.
type UnsupportedRequiresSamplingError struct {
	ClassName     string
	ActionName    string
	RequireKey    string
	Specification string
}

func (e *UnsupportedRequiresSamplingError) Error() string {
	switch {
	case e.ClassName != "" && e.ActionName != "":
		return fmt.Sprintf(
			"class %q action %q require %q: parsed requirement references parameters but the simulator cannot derive random parameter values from this expression (specification: %q)",
			e.ClassName, e.ActionName, e.RequireKey, e.Specification,
		)
	case e.ActionName != "":
		return fmt.Sprintf(
			"action %q require %q: parsed requirement references parameters but the simulator cannot derive random parameter values from this expression (specification: %q)",
			e.ActionName, e.RequireKey, e.Specification,
		)
	default:
		return fmt.Sprintf(
			"parsed requirement %q references parameters but the simulator cannot derive random parameter values from this expression (specification: %q)",
			e.RequireKey, e.Specification,
		)
	}
}

// validateRequiresSamplingSupport rejects parsed assessments that reference parameters
// without a supported random-generation strategy.
func validateRequiresSamplingSupport(requires []model_logic.Logic, paramNames map[string]bool) error {
	if len(paramNames) == 0 {
		return nil
	}

	for _, req := range requires {
		if req.Type != model_logic.LogicTypeAssessment || !req.Spec.ParseOk() {
			continue
		}
		if !expressionReferencesParams(req.Spec.Expression, paramNames) {
			continue
		}
		if expressionSupportsParamSampling(req.Spec.Expression) {
			continue
		}
		return &UnsupportedRequiresSamplingError{
			RequireKey:    req.Key.String(),
			Specification: req.Spec.Specification,
		}
	}
	return nil
}

// ValidateOwnerRequiresSamplingSupport validates one parameter owner's requires.
func ValidateOwnerRequiresSamplingSupport(className string, owner ParameterOwner) error {
	return owner.ValidateRequiresSamplingSupport(className)
}

// ValidateActionRequiresSamplingSupport validates one action's requires against its parameters.
func ValidateActionRequiresSamplingSupport(className string, action model_state.Action) error {
	return ValidateOwnerRequiresSamplingSupport(className, ParameterOwnerFromAction(action))
}

// ValidateQueryRequiresSamplingSupport validates one query's requires against its parameters.
func ValidateQueryRequiresSamplingSupport(className string, query model_state.Query) error {
	return ValidateOwnerRequiresSamplingSupport(className, ParameterOwnerFromQuery(query))
}

func parameterNames(params []model_state.Parameter) map[string]bool {
	names := make(map[string]bool, len(params))
	for _, param := range params {
		names[param.Name] = true
	}
	return names
}

func expressionReferencesParams(expr me.Expression, paramNames map[string]bool) bool {
	if expr == nil {
		return false
	}
	if localVar, ok := expr.(*me.LocalVar); ok {
		return paramNames[localVar.Name]
	}
	for _, child := range expressionChildNodes(expr) {
		if expressionReferencesParams(child, paramNames) {
			return true
		}
	}
	return false
}

func expressionChildNodes(expr me.Expression) []me.Expression {
	if children := binaryExpressionChildren(expr); len(children) > 0 {
		return children
	}
	if children := unaryExpressionChildren(expr); len(children) > 0 {
		return children
	}
	if children := collectionExpressionChildren(expr); len(children) > 0 {
		return children
	}
	if children := callExpressionChildren(expr); len(children) > 0 {
		return children
	}
	return controlFlowExpressionChildren(expr)
}

func binaryExpressionChildren(expr me.Expression) []me.Expression {
	switch node := expr.(type) {
	case *me.BinaryLogic:
		return []me.Expression{node.Left, node.Right}
	case *me.BinaryArith:
		return []me.Expression{node.Left, node.Right}
	case *me.Compare:
		return []me.Expression{node.Left, node.Right}
	case *me.SetOp:
		return []me.Expression{node.Left, node.Right}
	case *me.SetCompare:
		return []me.Expression{node.Left, node.Right}
	case *me.BagOp:
		return []me.Expression{node.Left, node.Right}
	case *me.BagCompare:
		return []me.Expression{node.Left, node.Right}
	default:
		return nil
	}
}

func unaryExpressionChildren(expr me.Expression) []me.Expression {
	switch node := expr.(type) {
	case *me.Not:
		return []me.Expression{node.Expr}
	case *me.Negate:
		return []me.Expression{node.Expr}
	case *me.NextState:
		return []me.Expression{node.Expr}
	case *me.Membership:
		return []me.Expression{node.Element}
	case *me.FieldAccess:
		return []me.Expression{node.Base}
	case *me.TupleIndex:
		return []me.Expression{node.Tuple}
	case *me.StringIndex:
		return []me.Expression{node.Str}
	default:
		return nil
	}
}

func collectionExpressionChildren(expr me.Expression) []me.Expression {
	switch node := expr.(type) {
	case *me.TupleLiteral:
		return node.Elements
	case *me.SetLiteral:
		return node.Elements
	case *me.RecordLiteral:
		children := make([]me.Expression, len(node.Fields))
		for i, field := range node.Fields {
			children[i] = field.Value
		}
		return children
	case *me.StringConcat:
		return node.Operands
	case *me.TupleConcat:
		return node.Operands
	default:
		return nil
	}
}

func callExpressionChildren(expr me.Expression) []me.Expression {
	switch node := expr.(type) {
	case *me.ActionCall:
		return node.Args
	case *me.GlobalCall:
		return node.Args
	case *me.BuiltinCall:
		return node.Args
	default:
		return nil
	}
}

func controlFlowExpressionChildren(expr me.Expression) []me.Expression {
	switch node := expr.(type) {
	case *me.IfThenElse:
		return []me.Expression{node.Condition, node.Then, node.Else}
	case *me.LetExpr:
		return []me.Expression{node.Value, node.Body}
	case *me.Choose:
		return []me.Expression{node.Set, node.Predicate}
	case *me.Quantifier:
		return []me.Expression{node.Predicate}
	case *me.SetFilter:
		return []me.Expression{node.Predicate}
	case *me.SetMap:
		return []me.Expression{node.Transform}
	case *me.Case:
		children := make([]me.Expression, 0, len(node.Branches)*2+1)
		for _, branch := range node.Branches {
			children = append(children, branch.Condition, branch.Result)
		}
		if node.Otherwise != nil {
			children = append(children, node.Otherwise)
		}
		return children
	case *me.RecordUpdate:
		children := []me.Expression{node.Base}
		for _, alt := range node.Alterations {
			children = append(children, alt.Value)
		}
		return children
	default:
		return nil
	}
}

func expressionSupportsParamSampling(expr me.Expression) bool {
	if expr == nil {
		return true
	}

	switch node := expr.(type) {
	case *me.BuiltinCall:
		return gzBuiltinSupportsParamSampling(node)
	case *me.IfThenElse:
		// Raw IF no longer drives parameter synthesis (use _GZ!When* instead).
		return false
	case *me.Membership:
		return membershipSupportsParamSampling(node)
	case *me.BinaryLogic:
		return expressionSupportsParamSampling(node.Left) &&
			expressionSupportsParamSampling(node.Right)
	case *me.Quantifier:
		return quantifierSupportsParamSampling(node)
	default:
		return expressionLeafSupportsParamSampling(expr)
	}
}

func gzBuiltinSupportsParamSampling(call *me.BuiltinCall) bool {
	if call == nil || call.Module != gzModuleName {
		// Non-_GZ builtins that reference params are not a sampling strategy.
		return false
	}
	switch call.Function {
	case gzFnWhenNotNull, gzFnWhenNull:
		if len(call.Args) != 2 {
			return false
		}
		if _, ok := gzDriverName(call.Args[0]); !ok {
			return false
		}
		return expressionSupportsParamSampling(call.Args[1]) || expressionLeafSupportsParamSampling(call.Args[1]) ||
			gzEquationSupportsSampling(call.Args[1])
	case gzFnWhenNullElse:
		if len(call.Args) != 3 {
			return false
		}
		if _, ok := gzDriverName(call.Args[0]); !ok {
			return false
		}
		return gzEquationSupportsSampling(call.Args[1]) && gzEquationSupportsSampling(call.Args[2])
	default:
		return false
	}
}

// gzEquationSupportsSampling reports whether a _GZ arm is a known synthesizable shape or TRUE.
func gzEquationSupportsSampling(expr me.Expression) bool {
	if expr == nil || isTrueLiteral(expr) {
		return true
	}
	if _, ok := nullCompareParam(expr); ok {
		return true
	}
	if _, _, ok := paramCompareBoolLiteral(expr); ok {
		return true
	}
	if _, _, ok := paramEquality(expr); ok {
		return true
	}
	if m, ok := expr.(*me.Membership); ok {
		return membershipSupportsParamSampling(m) || membershipSupportsNegatedNamedSet(m)
	}
	if and, ok := expr.(*me.BinaryLogic); ok && and.Op == me.LogicAnd {
		return gzEquationSupportsSampling(and.Left) && gzEquationSupportsSampling(and.Right)
	}
	// Compare / other leaves used in assessments.
	return expressionLeafSupportsParamSampling(expr) || expressionSupportsParamSampling(expr)
}

func membershipSupportsNegatedNamedSet(node *me.Membership) bool {
	if node == nil || !node.Negated {
		return false
	}
	_, _, ok := paramMembershipInNamedSet(node)
	return ok
}

func membershipSupportsParamSampling(node *me.Membership) bool {
	if node.Negated {
		return false
	}
	_, _, tupleOK := tupleMembershipInNamedSet(node)
	_, minusPeerOK := detectParamInNamedSetMinusPeerField(node)
	_, _, memberOK := paramMembershipInNamedSet(node)
	_, _, enumOK := paramInStringEnum(node)
	_, booleanOK := paramInBooleanSet(node)
	return tupleOK || minusPeerOK || memberOK || enumOK || booleanOK
}

func quantifierSupportsParamSampling(node *me.Quantifier) bool {
	_, ok := detectPeerFieldDistinctFromParam(node)
	return ok
}

func expressionLeafSupportsParamSampling(expr me.Expression) bool {
	switch expr.(type) {
	case *me.LocalVar, *me.BoolLiteral, *me.IntLiteral, *me.RationalLiteral,
		*me.StringLiteral, *me.SetLiteral, *me.TupleLiteral, *me.RecordLiteral,
		*me.SetConstant, *me.SelfRef, *me.AttributeRef, *me.NamedSetRef, *me.ClassRef, *me.PriorFieldValue:
		return true
	default:
		return false
	}
}
