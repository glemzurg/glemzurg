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
// names to the set.
func walkIdentifiersIR(expr me.Expression, result map[string]bool) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	// References that carry names.
	case *me.LocalVar:
		result[e.Name] = true
	case *me.AttributeRef:
		// AttributeRef keys contain class/attribute hierarchy — not simple identifiers.
	case *me.SelfRef:
		// No name.
	case *me.PriorFieldValue:
		// Field name, not a class reference.

	// Leaf nodes.
	case *me.BoolLiteral, *me.IntLiteral, *me.RationalLiteral,
		*me.StringLiteral, *me.SetConstant, *me.NamedSetRef:
		// No expression children.

	// Set literal.
	case *me.SetLiteral:
		for _, elem := range e.Elements {
			walkIdentifiersIR(elem, result)
		}

	// Tuple literal.
	case *me.TupleLiteral:
		for _, elem := range e.Elements {
			walkIdentifiersIR(elem, result)
		}

	// Record literal.
	case *me.RecordLiteral:
		for _, f := range e.Fields {
			walkIdentifiersIR(f.Value, result)
		}

	// Binary operators.
	case *me.BinaryArith:
		walkIdentifiersIR(e.Left, result)
		walkIdentifiersIR(e.Right, result)
	case *me.BinaryLogic:
		walkIdentifiersIR(e.Left, result)
		walkIdentifiersIR(e.Right, result)
	case *me.Compare:
		walkIdentifiersIR(e.Left, result)
		walkIdentifiersIR(e.Right, result)
	case *me.SetOp:
		walkIdentifiersIR(e.Left, result)
		walkIdentifiersIR(e.Right, result)
	case *me.SetCompare:
		walkIdentifiersIR(e.Left, result)
		walkIdentifiersIR(e.Right, result)
	case *me.BagOp:
		walkIdentifiersIR(e.Left, result)
		walkIdentifiersIR(e.Right, result)
	case *me.BagCompare:
		walkIdentifiersIR(e.Left, result)
		walkIdentifiersIR(e.Right, result)
	case *me.Membership:
		walkIdentifiersIR(e.Element, result)
		walkIdentifiersIR(e.Set, result)

	// Unary operators.
	case *me.Negate:
		walkIdentifiersIR(e.Expr, result)
	case *me.Not:
		walkIdentifiersIR(e.Expr, result)
	case *me.NextState:
		walkIdentifiersIR(e.Expr, result)

	// Access and indexing.
	case *me.FieldAccess:
		walkIdentifiersIR(e.Base, result)
	case *me.TupleIndex:
		walkIdentifiersIR(e.Tuple, result)
		walkIdentifiersIR(e.Index, result)
	case *me.StringIndex:
		walkIdentifiersIR(e.Str, result)
		walkIdentifiersIR(e.Index, result)
	case *me.RecordUpdate:
		walkIdentifiersIR(e.Base, result)
		for _, alt := range e.Alterations {
			walkIdentifiersIR(alt.Value, result)
		}

	// Concatenation.
	case *me.StringConcat:
		for _, op := range e.Operands {
			walkIdentifiersIR(op, result)
		}
	case *me.TupleConcat:
		for _, op := range e.Operands {
			walkIdentifiersIR(op, result)
		}

	// Quantifiers.
	case *me.Quantifier:
		walkIdentifiersIR(e.Domain, result)
		walkIdentifiersIR(e.Predicate, result)
	case *me.SetFilter:
		walkIdentifiersIR(e.Set, result)
		walkIdentifiersIR(e.Predicate, result)
	case *me.SetRange:
		walkIdentifiersIR(e.Start, result)
		walkIdentifiersIR(e.End, result)

	// Conditional.
	case *me.IfThenElse:
		walkIdentifiersIR(e.Condition, result)
		walkIdentifiersIR(e.Then, result)
		walkIdentifiersIR(e.Else, result)
	case *me.Case:
		for _, branch := range e.Branches {
			walkIdentifiersIR(branch.Condition, result)
			walkIdentifiersIR(branch.Result, result)
		}
		walkIdentifiersIR(e.Otherwise, result)

	// Calls.
	case *me.ActionCall:
		for _, arg := range e.Args {
			walkIdentifiersIR(arg, result)
		}
	case *me.GlobalCall:
		for _, arg := range e.Args {
			walkIdentifiersIR(arg, result)
		}
	case *me.BuiltinCall:
		for _, arg := range e.Args {
			walkIdentifiersIR(arg, result)
		}
	}
}
