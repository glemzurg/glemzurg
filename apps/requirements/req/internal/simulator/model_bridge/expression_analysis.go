package model_bridge

import (
	"slices"

	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
)

// ContainsGlobalCall reports whether an expression invokes a model global function.
func ContainsGlobalCall(expr me.Expression) bool {
	if expr == nil {
		return false
	}
	if _, ok := expr.(*me.GlobalCall); ok {
		return true
	}
	if containsGlobalCallLeaf(expr) {
		return false
	}
	if containsGlobalCallOperator(expr) {
		return true
	}
	return containsGlobalCallCompound(expr)
}

func containsGlobalCallLeaf(expr me.Expression) bool {
	switch expr.(type) {
	case *me.IntLiteral, *me.RationalLiteral, *me.BoolLiteral, *me.StringLiteral,
		*me.SetConstant, *me.SelfRef, *me.AttributeRef, *me.LocalVar,
		*me.PriorFieldValue, *me.NamedSetRef, *me.ClassRef, *me.NextState:
		return true
	default:
		return false
	}
}

func containsGlobalCallOperator(expr me.Expression) bool {
	switch expr.(type) {
	case *me.BinaryArith, *me.BinaryLogic, *me.Compare, *me.SetOp, *me.SetCompare,
		*me.BagOp, *me.BagCompare, *me.Membership:
		left, right := binaryChildren(expr)
		return ContainsGlobalCall(left) || ContainsGlobalCall(right)
	case *me.Negate, *me.Not, *me.NextState:
		return ContainsGlobalCall(unaryChild(expr))
	default:
		return false
	}
}

func containsGlobalCallCompound(expr me.Expression) bool {
	if containsGlobalCallAccess(expr) {
		return true
	}
	if containsGlobalCallCollection(expr) {
		return true
	}
	return containsGlobalCallControlFlow(expr)
}

func containsGlobalCallAccess(expr me.Expression) bool {
	switch e := expr.(type) {
	case *me.FieldAccess:
		return ContainsGlobalCall(e.Base)
	case *me.TupleIndex:
		return ContainsGlobalCall(e.Tuple) || ContainsGlobalCall(e.Index)
	case *me.StringIndex:
		return ContainsGlobalCall(e.Str) || ContainsGlobalCall(e.Index)
	case *me.RecordUpdate:
		if ContainsGlobalCall(e.Base) {
			return true
		}
		for _, alt := range e.Alterations {
			if ContainsGlobalCall(alt.Value) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func containsGlobalCallCollection(expr me.Expression) bool {
	switch e := expr.(type) {
	case *me.StringConcat:
		return slices.ContainsFunc(e.Operands, ContainsGlobalCall)
	case *me.TupleConcat:
		return slices.ContainsFunc(e.Operands, ContainsGlobalCall)
	case *me.SetLiteral:
		return slices.ContainsFunc(e.Elements, ContainsGlobalCall)
	case *me.TupleLiteral:
		return slices.ContainsFunc(e.Elements, ContainsGlobalCall)
	case *me.RecordLiteral:
		return slices.ContainsFunc(e.Fields, func(f me.RecordField) bool {
			return ContainsGlobalCall(f.Value)
		})
	default:
		return false
	}
}

func containsGlobalCallControlFlow(expr me.Expression) bool {
	switch e := expr.(type) {
	case *me.IfThenElse:
		return ContainsGlobalCall(e.Condition) || ContainsGlobalCall(e.Then) || ContainsGlobalCall(e.Else)
	case *me.LetExpr:
		return ContainsGlobalCall(e.Value) || ContainsGlobalCall(e.Body)
	case *me.Choose:
		return ContainsGlobalCall(e.Set) || ContainsGlobalCall(e.Predicate)
	case *me.Case:
		for _, branch := range e.Branches {
			if ContainsGlobalCall(branch.Condition) || ContainsGlobalCall(branch.Result) {
				return true
			}
		}
		return ContainsGlobalCall(e.Otherwise)
	case *me.Quantifier:
		return ContainsGlobalCall(e.Domain) || ContainsGlobalCall(e.Predicate)
	case *me.SetFilter:
		return ContainsGlobalCall(e.Set) || ContainsGlobalCall(e.Predicate)
	case *me.SetMap:
		return ContainsGlobalCall(e.Set) || ContainsGlobalCall(e.Transform)
	case *me.SetRange:
		return ContainsGlobalCall(e.Start) || ContainsGlobalCall(e.End)
	case *me.BuiltinCall:
		return slices.ContainsFunc(e.Args, ContainsGlobalCall)
	case *me.ActionCall:
		return slices.ContainsFunc(e.Args, ContainsGlobalCall)
	case *me.GlobalCall:
		return slices.ContainsFunc(e.Args, ContainsGlobalCall)
	default:
		return false
	}
}

func binaryChildren(expr me.Expression) (me.Expression, me.Expression) {
	switch e := expr.(type) {
	case *me.BinaryArith:
		return e.Left, e.Right
	case *me.BinaryLogic:
		return e.Left, e.Right
	case *me.Compare:
		return e.Left, e.Right
	case *me.SetOp:
		return e.Left, e.Right
	case *me.SetCompare:
		return e.Left, e.Right
	case *me.BagOp:
		return e.Left, e.Right
	case *me.BagCompare:
		return e.Left, e.Right
	case *me.Membership:
		return e.Element, e.Set
	default:
		return nil, nil
	}
}

func unaryChild(expr me.Expression) me.Expression {
	switch e := expr.(type) {
	case *me.Negate:
		return e.Expr
	case *me.Not:
		return e.Expr
	case *me.NextState:
		return e.Expr
	default:
		return nil
	}
}
