package actions

import (
	"errors"
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
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

// ValidateRequiresSamplingSupport rejects parsed assessments that reference action parameters
// without a supported random-generation strategy.
func ValidateRequiresSamplingSupport(requires []model_logic.Logic, paramNames map[string]bool) error {
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

// ValidateActionRequiresSamplingSupport validates one action's requires against its parameters.
func ValidateActionRequiresSamplingSupport(className string, action model_state.Action) error {
	return validateOwnerRequiresSamplingSupport(className, action.Key, logicOwnerKindAction, action.Name, action.Parameters, action.Requires)
}

// ValidateQueryRequiresSamplingSupport validates one query's requires against its parameters.
func ValidateQueryRequiresSamplingSupport(className string, query model_state.Query) error {
	return validateOwnerRequiresSamplingSupport(className, query.Key, logicOwnerKindQuery, query.Name, query.Parameters, query.Requires)
}

func validateOwnerRequiresSamplingSupport(
	className string,
	ownerKey identity.Key,
	ownerKind string,
	ownerName string,
	params []model_state.Parameter,
	explicitRequires []model_logic.Logic,
) error {
	if len(params) == 0 {
		return nil
	}
	effectiveRequires, err := EffectiveRequires(ownerKey, ownerKind, params, explicitRequires)
	if err != nil {
		return err
	}
	paramNames := parameterNames(params)
	if err := ValidateRequiresSamplingSupport(effectiveRequires, paramNames); err != nil {
		var unsupported *UnsupportedRequiresSamplingError
		if errors.As(err, &unsupported) {
			unsupported.ClassName = className
			unsupported.ActionName = ownerName
			return unsupported
		}
		return err
	}
	return nil
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
	case *me.IfThenElse:
		return isNullableElseTuplePattern(node) || isNullableElseMembershipPattern(node)
	case *me.Membership:
		if node.Negated {
			return false
		}
		_, _, tupleOK := tupleMembershipInNamedSet(node)
		_, _, memberOK := paramMembershipInNamedSet(node)
		_, _, enumOK := paramInStringEnum(node)
		return tupleOK || memberOK || enumOK
	case *me.BinaryLogic:
		return expressionSupportsParamSampling(node.Left) &&
			expressionSupportsParamSampling(node.Right)
	case *me.LocalVar, *me.BoolLiteral, *me.IntLiteral, *me.RationalLiteral,
		*me.StringLiteral, *me.SetLiteral, *me.TupleLiteral, *me.RecordLiteral,
		*me.SetConstant, *me.SelfRef, *me.AttributeRef, *me.NamedSetRef, *me.PriorFieldValue:
		return true
	default:
		return false
	}
}

func isNullableElseMembershipPattern(node *me.IfThenElse) bool {
	paramName, ok := nullCompareParam(node.Condition)
	if !ok || !isTrueLiteral(node.Then) {
		return false
	}
	membership, ok := node.Else.(*me.Membership)
	if !ok || membership.Negated {
		return false
	}
	memberParam, _, ok := paramMembershipInNamedSet(membership)
	return ok && memberParam == paramName
}

func isNullableElseTuplePattern(node *me.IfThenElse) bool {
	if _, ok := nullCompareParam(node.Condition); !ok {
		return false
	}
	if _, ok := nullCompareParam(node.Then); !ok {
		return false
	}
	membership, ok := node.Else.(*me.Membership)
	if !ok || membership.Negated {
		return false
	}
	_, _, ok = tupleMembershipInNamedSet(membership)
	return ok
}
