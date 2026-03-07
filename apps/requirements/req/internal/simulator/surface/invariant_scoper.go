package surface

import (
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
)

// ScopeInvariants filters model invariants to those relevant to the given
// surface. An invariant is excluded if it references a class name that exists
// in the model but is NOT in the in-scope set.
//
// classNames is the set of in-scope class names.
// allClassNames is the set of ALL class names in the model.
func ScopeInvariants(invariants []model_logic.Logic, _ map[string]bool) (included []model_logic.Logic, excluded []model_logic.Logic) {
	// Without knowing all class names, we can't detect out-of-scope references.
	// Include everything by default. Use ScopeInvariantsWithAllClasses for full filtering.
	included = append(included, invariants...)
	return included, excluded
}

// ScopeInvariantsWithAllClasses filters invariants using both in-scope and
// all class names. An invariant is excluded if it references a class name
// that exists in the model but is NOT in the in-scope set.
//
//complexity:cyclo:warn=60,fail=60 Simple routing switch.
func ScopeInvariantsWithAllClasses(
	invariants []model_logic.Logic,
	inScopeClassNames map[string]bool,
	allClassNames map[string]bool,
) (included []model_logic.Logic, excluded []model_logic.Logic) {
	for _, inv := range invariants {
		expr := inv.Spec.Expression
		if expr == nil {
			// If IR is not available, include it — better safe than sorry.
			included = append(included, inv)
			continue
		}

		identifiers := collectIdentifiersFromIR(expr)

		// Check if any identifier matches a class name NOT in scope.
		shouldExclude := false
		for ident := range identifiers {
			if allClassNames[ident] && !inScopeClassNames[ident] {
				shouldExclude = true
				break
			}
		}

		if shouldExclude {
			excluded = append(excluded, inv)
		} else {
			included = append(included, inv)
		}
	}
	return included, excluded
}

// collectIdentifiersFromIR recursively walks an IR expression and collects
// all identifier-like names found (LocalVar names, field names, etc.).
func collectIdentifiersFromIR(expr me.Expression) map[string]bool {
	result := make(map[string]bool)
	walkIdentifiersIR(expr, result)
	return result
}

// walkIdentifiersIR recursively walks an IR expression and adds identifier-like
// names to the set. It dispatches to category-specific walkers.
func walkIdentifiersIR(expr me.Expression, result map[string]bool) {
	if expr == nil {
		return
	}
	if walkIdentifiersLeaf(expr, result) {
		return
	}
	if walkIdentifiersOperator(expr, result) {
		return
	}
	walkIdentifiersCompound(expr, result)
}

// walkIdentifiersLeaf handles leaf nodes and references. Returns true if handled.
func walkIdentifiersLeaf(expr me.Expression, result map[string]bool) bool {
	switch e := expr.(type) {
	case *me.LocalVar:
		result[e.Name] = true
	case *me.AttributeRef, *me.SelfRef, *me.PriorFieldValue:
		// No identifier names to collect.
	case *me.BoolLiteral, *me.IntLiteral, *me.RationalLiteral,
		*me.StringLiteral, *me.SetConstant, *me.NamedSetRef:
		// No expression children.
	default:
		return false
	}
	return true
}

// walkIdentifiersOperator handles unary/binary operators and comparisons. Returns true if handled.
func walkIdentifiersOperator(expr me.Expression, result map[string]bool) bool {
	switch e := expr.(type) {
	case *me.BinaryArith:
		walkBinaryIR(e.Left, e.Right, result)
	case *me.BinaryLogic:
		walkBinaryIR(e.Left, e.Right, result)
	case *me.Compare:
		walkBinaryIR(e.Left, e.Right, result)
	case *me.SetOp:
		walkBinaryIR(e.Left, e.Right, result)
	case *me.SetCompare:
		walkBinaryIR(e.Left, e.Right, result)
	case *me.BagOp:
		walkBinaryIR(e.Left, e.Right, result)
	case *me.BagCompare:
		walkBinaryIR(e.Left, e.Right, result)
	case *me.Membership:
		walkBinaryIR(e.Element, e.Set, result)
	case *me.Negate:
		walkIdentifiersIR(e.Expr, result)
	case *me.Not:
		walkIdentifiersIR(e.Expr, result)
	case *me.NextState:
		walkIdentifiersIR(e.Expr, result)
	default:
		return false
	}
	return true
}

// walkIdentifiersCompound handles compound expressions (collections, control flow, calls).
func walkIdentifiersCompound(expr me.Expression, result map[string]bool) {
	switch e := expr.(type) {
	case *me.SetLiteral:
		walkSliceIR(e.Elements, result)
	case *me.TupleLiteral:
		walkSliceIR(e.Elements, result)
	case *me.RecordLiteral:
		for _, f := range e.Fields {
			walkIdentifiersIR(f.Value, result)
		}
	case *me.FieldAccess:
		walkIdentifiersIR(e.Base, result)
	case *me.TupleIndex:
		walkBinaryIR(e.Tuple, e.Index, result)
	case *me.StringIndex:
		walkBinaryIR(e.Str, e.Index, result)
	case *me.RecordUpdate:
		walkIdentifiersIR(e.Base, result)
		for _, alt := range e.Alterations {
			walkIdentifiersIR(alt.Value, result)
		}
	case *me.StringConcat:
		walkSliceIR(e.Operands, result)
	case *me.TupleConcat:
		walkSliceIR(e.Operands, result)
	case *me.Quantifier:
		walkBinaryIR(e.Domain, e.Predicate, result)
	case *me.SetFilter:
		walkBinaryIR(e.Set, e.Predicate, result)
	case *me.SetRange:
		walkBinaryIR(e.Start, e.End, result)
	case *me.IfThenElse:
		walkIdentifiersIR(e.Condition, result)
		walkIdentifiersIR(e.Then, result)
		walkIdentifiersIR(e.Else, result)
	case *me.Case:
		for _, branch := range e.Branches {
			walkBinaryIR(branch.Condition, branch.Result, result)
		}
		walkIdentifiersIR(e.Otherwise, result)
	case *me.ActionCall:
		walkSliceIR(e.Args, result)
	case *me.GlobalCall:
		walkSliceIR(e.Args, result)
	case *me.BuiltinCall:
		walkSliceIR(e.Args, result)
	}
}

// walkBinaryIR walks two child expressions.
func walkBinaryIR(left, right me.Expression, result map[string]bool) {
	walkIdentifiersIR(left, result)
	walkIdentifiersIR(right, result)
}

// walkSliceIR walks a slice of child expressions.
func walkSliceIR(exprs []me.Expression, result map[string]bool) {
	for _, expr := range exprs {
		walkIdentifiersIR(expr, result)
	}
}
