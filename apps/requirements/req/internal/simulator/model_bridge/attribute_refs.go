package model_bridge

import (
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// CollectAttributeRefs walks a lowered expression and returns every AttributeRef key.
func CollectAttributeRefs(expr me.Expression) map[identity.Key]bool {
	result := make(map[identity.Key]bool)
	collectAttributeRefs(expr, result)
	return result
}

// collectAttributeRefs mirrors the ContainsAnyPrimedME walk so attribute-reference
// discovery stays aligned with other expression traversals in this package.
//
//complexity:cyclo:warn=50,fail=50 Simple routing switch.
func collectAttributeRefs(expr me.Expression, out map[identity.Key]bool) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *me.AttributeRef:
		out[e.AttributeKey] = true

	case *me.IntLiteral, *me.RationalLiteral, *me.BoolLiteral, *me.StringLiteral,
		*me.SetConstant, *me.SelfRef, *me.LocalVar,
		*me.PriorFieldValue, *me.NamedSetRef, *me.ClassRef, *me.NextState:
		return

	case *me.BinaryArith:
		collectAttributeRefs(e.Left, out)
		collectAttributeRefs(e.Right, out)
	case *me.BinaryLogic:
		collectAttributeRefs(e.Left, out)
		collectAttributeRefs(e.Right, out)
	case *me.Compare:
		collectAttributeRefs(e.Left, out)
		collectAttributeRefs(e.Right, out)
	case *me.SetOp:
		collectAttributeRefs(e.Left, out)
		collectAttributeRefs(e.Right, out)
	case *me.SetCompare:
		collectAttributeRefs(e.Left, out)
		collectAttributeRefs(e.Right, out)
	case *me.BagOp:
		collectAttributeRefs(e.Left, out)
		collectAttributeRefs(e.Right, out)
	case *me.BagCompare:
		collectAttributeRefs(e.Left, out)
		collectAttributeRefs(e.Right, out)
	case *me.Membership:
		collectAttributeRefs(e.Element, out)
		collectAttributeRefs(e.Set, out)

	case *me.Negate:
		collectAttributeRefs(e.Expr, out)
	case *me.Not:
		collectAttributeRefs(e.Expr, out)

	case *me.FieldAccess:
		collectAttributeRefs(e.Base, out)
	case *me.TupleIndex:
		collectAttributeRefs(e.Tuple, out)
		collectAttributeRefs(e.Index, out)
	case *me.StringIndex:
		collectAttributeRefs(e.Str, out)
		collectAttributeRefs(e.Index, out)
	case *me.RecordUpdate:
		collectAttributeRefs(e.Base, out)
		for _, alt := range e.Alterations {
			collectAttributeRefs(alt.Value, out)
		}

	case *me.StringConcat:
		for _, operand := range e.Operands {
			collectAttributeRefs(operand, out)
		}
	case *me.TupleConcat:
		for _, operand := range e.Operands {
			collectAttributeRefs(operand, out)
		}

	case *me.IfThenElse:
		collectAttributeRefs(e.Condition, out)
		collectAttributeRefs(e.Then, out)
		collectAttributeRefs(e.Else, out)
	case *me.LetExpr:
		collectAttributeRefs(e.Value, out)
		collectAttributeRefs(e.Body, out)
	case *me.Choose:
		collectAttributeRefs(e.Set, out)
		collectAttributeRefs(e.Predicate, out)
	case *me.Case:
		for _, branch := range e.Branches {
			collectAttributeRefs(branch.Condition, out)
			collectAttributeRefs(branch.Result, out)
		}
		collectAttributeRefs(e.Otherwise, out)

	case *me.Quantifier:
		collectAttributeRefs(e.Domain, out)
		collectAttributeRefs(e.Predicate, out)
	case *me.SetFilter:
		collectAttributeRefs(e.Set, out)
		collectAttributeRefs(e.Predicate, out)
	case *me.SetMap:
		collectAttributeRefs(e.Set, out)
		collectAttributeRefs(e.Transform, out)
	case *me.SetRange:
		collectAttributeRefs(e.Start, out)
		collectAttributeRefs(e.End, out)

	case *me.SetLiteral:
		for _, element := range e.Elements {
			collectAttributeRefs(element, out)
		}
	case *me.TupleLiteral:
		for _, element := range e.Elements {
			collectAttributeRefs(element, out)
		}
	case *me.RecordLiteral:
		for _, field := range e.Fields {
			collectAttributeRefs(field.Value, out)
		}

	case *me.BuiltinCall:
		for _, arg := range e.Args {
			collectAttributeRefs(arg, out)
		}
	case *me.GlobalCall:
		for _, arg := range e.Args {
			collectAttributeRefs(arg, out)
		}
	case *me.ActionCall:
		for _, arg := range e.Args {
			collectAttributeRefs(arg, out)
		}

	default:
		return
	}
}
