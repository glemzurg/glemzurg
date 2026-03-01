package surface

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/parser"
)

// ScopeInvariants filters model invariants to those relevant to the given
// surface. An invariant is excluded if it references a class name that exists
// in the model but is NOT in the in-scope set.
//
// classNames is the set of in-scope class names.
// allClassNames is the set of ALL class names in the model.
func ScopeInvariants(invariants []model_logic.Logic, classNames map[string]bool) (included []model_logic.Logic, excluded []model_logic.Logic) {
	// Without knowing all class names, we can't detect out-of-scope references.
	// Include everything by default. Use ScopeInvariantsWithAllClasses for full filtering.
	for _, inv := range invariants {
		included = append(included, inv)
	}
	return included, excluded
}

// ScopeInvariantsWithAllClasses filters invariants using both in-scope and
// all class names. An invariant is excluded if it references a class name
// that exists in the model but is NOT in the in-scope set.
func ScopeInvariantsWithAllClasses(
	invariants []model_logic.Logic,
	inScopeClassNames map[string]bool,
	allClassNames map[string]bool,
) (included []model_logic.Logic, excluded []model_logic.Logic) {
	for _, inv := range invariants {
		expr, err := parser.ParseExpression(inv.Spec.Specification)
		if err != nil {
			// If we can't parse it, include it — better safe than sorry.
			included = append(included, inv)
			continue
		}

		identifiers := collectIdentifiers(expr)

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

// collectIdentifiers recursively walks an AST expression and collects
// all Identifier values found.
func collectIdentifiers(expr ast.Expression) map[string]bool {
	result := make(map[string]bool)
	walkIdentifiers(expr, result)
	return result
}

// walkIdentifiers recursively walks an AST and adds identifier values to the set.
func walkIdentifiers(expr ast.Expression, result map[string]bool) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.Identifier:
		result[e.Value] = true

	// Leaf nodes with no children containing identifiers.
	case *ast.NumberLiteral, *ast.StringLiteral,
		*ast.BooleanLiteral, *ast.SetConstant, *ast.ExistingValue,
		*ast.SetLiteralEnum, *ast.SetLiteralInt, *ast.SetRange:
		// No expression children.

	// Set literals with expression elements.
	case *ast.SetLiteral:
		for _, elem := range e.Elements {
			walkIdentifiers(elem, result)
		}

	// Dynamic range expressions.
	case *ast.SetRangeExpr:
		walkIdentifiers(e.Start, result)
		walkIdentifiers(e.End, result)

	// Binary operators.
	case *ast.BinaryArithmetic:
		walkIdentifiers(e.Left, result)
		walkIdentifiers(e.Right, result)
	case *ast.BinaryLogic:
		walkIdentifiers(e.Left, result)
		walkIdentifiers(e.Right, result)
	case *ast.BinaryComparison:
		walkIdentifiers(e.Left, result)
		walkIdentifiers(e.Right, result)
	case *ast.BinaryEquality:
		walkIdentifiers(e.Left, result)
		walkIdentifiers(e.Right, result)
	case *ast.BinarySetComparison:
		walkIdentifiers(e.Left, result)
		walkIdentifiers(e.Right, result)
	case *ast.BinarySetOperation:
		walkIdentifiers(e.Left, result)
		walkIdentifiers(e.Right, result)
	case *ast.BinaryBagComparison:
		walkIdentifiers(e.Left, result)
		walkIdentifiers(e.Right, result)
	case *ast.BinaryBagOperation:
		walkIdentifiers(e.Left, result)
		walkIdentifiers(e.Right, result)
	case *ast.Membership:
		walkIdentifiers(e.Left, result)
		walkIdentifiers(e.Right, result)
	case *ast.Fraction:
		walkIdentifiers(e.Numerator, result)
		walkIdentifiers(e.Denominator, result)

	// Unary operators.
	case *ast.UnaryLogic:
		walkIdentifiers(e.Right, result)
	case *ast.UnaryNegation:
		walkIdentifiers(e.Right, result)
	case *ast.Parenthesized:
		walkIdentifiers(e.Inner, result)
	case *ast.Primed:
		walkIdentifiers(e.Base, result)

	// Access and indexing.
	case *ast.FieldAccess:
		walkIdentifiers(e.Base, result)
	case *ast.TupleIndex:
		walkIdentifiers(e.Tuple, result)
		walkIdentifiers(e.Index, result)
	case *ast.StringIndex:
		walkIdentifiers(e.Str, result)
		walkIdentifiers(e.Index, result)

	// String concatenation.
	case *ast.StringConcat:
		for _, op := range e.Operands {
			walkIdentifiers(op, result)
		}

	// Tuple concatenation.
	case *ast.TupleConcat:
		for _, op := range e.Operands {
			walkIdentifiers(op, result)
		}

	// Quantifier (covers both ∀ and ∃).
	case *ast.Quantifier:
		walkIdentifiers(e.Membership, result)
		walkIdentifiers(e.Predicate, result)

	// Set filter (set comprehension).
	case *ast.SetFilter:
		walkIdentifiers(e.Membership, result)
		walkIdentifiers(e.Predicate, result)

	// Conditional.
	case *ast.IfThenElse:
		walkIdentifiers(e.Condition, result)
		walkIdentifiers(e.Then, result)
		walkIdentifiers(e.Else, result)

	// Function calls.
	case *ast.FunctionCall:
		for _, seg := range e.ScopePath {
			walkIdentifiers(seg, result)
		}
		walkIdentifiers(e.Name, result)
		for _, arg := range e.Args {
			walkIdentifiers(arg, result)
		}

	// Scoped calls.
	case *ast.ScopedCall:
		walkIdentifiers(e.Domain, result)
		walkIdentifiers(e.Subdomain, result)
		walkIdentifiers(e.Class, result)
		walkIdentifiers(e.FunctionName, result)
		walkIdentifiers(e.Parameter, result)

	// Builtin calls.
	case *ast.BuiltinCall:
		for _, arg := range e.Args {
			walkIdentifiers(arg, result)
		}

	// Tuple literal.
	case *ast.TupleLiteral:
		for _, elem := range e.Elements {
			walkIdentifiers(elem, result)
		}

	// Record instance.
	case *ast.RecordInstance:
		for _, binding := range e.Bindings {
			walkIdentifiers(binding.Expression, result)
		}

	// Record altered (EXCEPT).
	case *ast.RecordAltered:
		walkIdentifiers(e.Identifier, result)
		for _, alt := range e.Alterations {
			walkIdentifiers(alt.Expression, result)
		}

	// Case expression.
	case *ast.CaseExpr:
		for _, branch := range e.Branches {
			walkIdentifiers(branch.Condition, result)
			walkIdentifiers(branch.Result, result)
		}
		walkIdentifiers(e.Other, result)

	}
}
