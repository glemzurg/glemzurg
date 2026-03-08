package convert

import (
	"fmt"
	"math/big"
	"strings"

	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
)

// RaiseContext provides the inverse name-resolution mappings for raising
// model_expression trees back to TLA+ AST nodes.
type RaiseContext struct {
	// AttributeNames maps attribute identity keys to their display names.
	AttributeNames map[identity.Key]string

	// ActionNames maps same-class action identity keys to their names.
	ActionNames map[identity.Key]string

	// QueryNames maps same-class query identity keys to their names.
	QueryNames map[identity.Key]string

	// GlobalFunctions maps global function identity keys to their names (with leading _).
	GlobalFunctions map[identity.Key]string

	// NamedSets maps named set identity keys to their display names.
	NamedSets map[identity.Key]string

	// ActionScopePaths maps cross-class action identity keys to their
	// fully scoped path (e.g., "Domain!Subdomain!Class!ActionName").
	ActionScopePaths map[identity.Key]string
}

// Raise converts a model_expression.Expression tree into a TLA+ AST expression.
// This is the inverse of Lower().
//
//complexity:cyclo:warn=60,fail=60 Simple routing switch.
//complexity:fanout:warn=60,fail=60 Simple routing switch.
func Raise(expr me.Expression, ctx *RaiseContext) (ast.Expression, error) {
	if expr == nil {
		return nil, fmt.Errorf("cannot raise nil expression")
	}

	switch e := expr.(type) {
	// --- Literals ---
	case *me.BoolLiteral:
		return &ast.BooleanLiteral{Value: e.Value}, nil

	case *me.IntLiteral:
		return raiseIntLiteral(e)

	case *me.RationalLiteral:
		return raiseRationalLiteral(e)

	case *me.StringLiteral:
		return &ast.StringLiteral{Value: e.Value}, nil

	// --- Collections ---
	case *me.SetLiteral:
		return raiseSetLiteral(e, ctx)

	case *me.TupleLiteral:
		return raiseTupleLiteral(e, ctx)

	case *me.RecordLiteral:
		return raiseRecordLiteral(e, ctx)

	case *me.SetConstant:
		return raiseSetConstant(e)

	case *me.SetRange:
		return raiseSetRange(e, ctx)

	// --- References ---
	case *me.SelfRef:
		return &ast.Identifier{Value: "self"}, nil

	case *me.AttributeRef:
		return raiseAttributeRef(e, ctx)

	case *me.LocalVar:
		return &ast.Identifier{Value: e.Name}, nil

	case *me.PriorFieldValue:
		return &ast.ExistingValue{}, nil

	case *me.NextState:
		return raiseNextState(e, ctx)

	case *me.NamedSetRef:
		return raiseNamedSetRef(e, ctx)

	// --- Unary operators ---
	case *me.Negate:
		return raiseNegate(e, ctx)

	case *me.Not:
		return raiseNot(e, ctx)

	// --- Binary operators ---
	case *me.BinaryArith:
		return raiseBinaryArith(e, ctx)

	case *me.BinaryLogic:
		return raiseBinaryLogic(e, ctx)

	case *me.Compare:
		return raiseCompare(e, ctx)

	case *me.SetOp:
		return raiseSetOp(e, ctx)

	case *me.SetCompare:
		return raiseSetCompare(e, ctx)

	case *me.BagOp:
		return raiseBagOp(e, ctx)

	case *me.BagCompare:
		return raiseBagCompare(e, ctx)

	case *me.Membership:
		return raiseMembership(e, ctx)

	// --- Concatenation ---
	case *me.StringConcat:
		return raiseStringConcat(e, ctx)

	case *me.TupleConcat:
		return raiseTupleConcat(e, ctx)

	// --- Field/Index access ---
	case *me.FieldAccess:
		return raiseFieldAccess(e, ctx)

	case *me.TupleIndex:
		return raiseTupleIndex(e, ctx)

	case *me.StringIndex:
		return raiseStringIndex(e, ctx)

	// --- Record alteration ---
	case *me.RecordUpdate:
		return raiseRecordUpdate(e, ctx)

	// --- Control flow ---
	case *me.IfThenElse:
		return raiseIfThenElse(e, ctx)

	case *me.Case:
		return raiseCase(e, ctx)

	// --- Quantifiers ---
	case *me.Quantifier:
		return raiseQuantifier(e, ctx)

	case *me.SetFilter:
		return raiseSetFilter(e, ctx)

	// --- Calls ---
	case *me.ActionCall:
		return raiseActionCall(e, ctx)

	case *me.GlobalCall:
		return raiseGlobalCall(e, ctx)

	case *me.BuiltinCall:
		return raiseBuiltinCall(e, ctx)

	default:
		return nil, fmt.Errorf("unsupported model_expression node type: %T", expr)
	}
}

// --- Operator enum → Unicode string tables ---

var raiseArithOp = map[me.ArithOp]string{
	me.ArithAdd: "+",
	me.ArithSub: "-",
	me.ArithMul: "*",
	me.ArithDiv: "÷",
	me.ArithMod: "%",
	me.ArithPow: "^",
}

var raiseLogicOp = map[me.LogicOp]string{
	me.LogicAnd:     "∧",
	me.LogicOr:      "∨",
	me.LogicImplies: "⇒",
	me.LogicEquiv:   "≡",
}

var raiseCompareOp = map[me.CompareOp]string{
	me.CompareLt:  "<",
	me.CompareGt:  ">",
	me.CompareLte: "≤",
	me.CompareGte: "≥",
	me.CompareEq:  "=",
	me.CompareNeq: "≠",
}

var raiseSetOpMap = map[me.SetOpKind]string{
	me.SetUnion:      "∪",
	me.SetIntersect:  "∩",
	me.SetDifference: `\`,
}

var raiseSetCompareOp = map[me.SetCompareOp]string{
	me.SetCompareSubsetEq:   "⊆",
	me.SetCompareSubset:     "⊂",
	me.SetCompareSupersetEq: "⊇",
	me.SetCompareSuperset:   "⊃",
}

var raiseBagOpMap = map[me.BagOpKind]string{
	me.BagSum:        "⊕",
	me.BagDifference: "⊖",
}

var raiseBagCompareOp = map[me.BagCompareOp]string{
	me.BagCompareProperSubBag: "⊏",
	me.BagCompareSubBag:       "⊑",
	me.BagCompareProperSupBag: "⊐",
	me.BagCompareSupBag:       "⊒",
}

var raiseSetConstantKind = map[me.SetConstantKind]string{
	me.SetConstantNat:     "Nat",
	me.SetConstantInt:     "Int",
	me.SetConstantReal:    "Real",
	me.SetConstantBoolean: "BOOLEAN",
}

var raiseQuantifierKind = map[me.QuantifierKind]string{
	me.QuantifierForall: "∀",
	me.QuantifierExists: "∃",
}

// --- Literal raising ---

// raiseIntLiteral raises an IntLiteral to a NumberLiteral AST node.
// Negative values are emitted as UnaryNegation{NumberLiteral{abs(value)}}.
func raiseIntLiteral(e *me.IntLiteral) (ast.Expression, error) {
	if e.Value.Sign() < 0 {
		abs := new(big.Int).Abs(e.Value)
		return ast.NewNegation(ast.NewNumberLiteral(abs.String())), nil
	}
	return ast.NewNumberLiteral(e.Value.String()), nil
}

// raiseRationalLiteral raises a RationalLiteral to a Fraction AST node.
// If the value is an integer (denominator=1), it emits just a NumberLiteral.
// Negative rationals are handled by emitting negative numerator or
// wrapping in negation as appropriate.
func raiseRationalLiteral(e *me.RationalLiteral) (ast.Expression, error) {
	if e.Value.IsInt() {
		num := e.Value.Num()
		if num.Sign() < 0 {
			abs := new(big.Int).Abs(num)
			return ast.NewNegation(ast.NewNumberLiteral(abs.String())), nil
		}
		return ast.NewNumberLiteral(num.String()), nil
	}

	num := new(big.Int).Set(e.Value.Num())
	den := new(big.Int).Set(e.Value.Denom())

	// big.Rat normalizes so denom is always positive.
	// Handle negative numerator: emit as -(|num|/den).
	if num.Sign() < 0 {
		abs := new(big.Int).Abs(num)
		return ast.NewNegation(ast.NewFraction(
			ast.NewNumberLiteral(abs.String()),
			ast.NewNumberLiteral(den.String()),
		)), nil
	}

	return ast.NewFraction(
		ast.NewNumberLiteral(num.String()),
		ast.NewNumberLiteral(den.String()),
	), nil
}

// --- Collection raising ---

func raiseSetLiteral(e *me.SetLiteral, ctx *RaiseContext) (ast.Expression, error) {
	elems := make([]ast.Expression, len(e.Elements))
	for i, elem := range e.Elements {
		raised, err := Raise(elem, ctx)
		if err != nil {
			return nil, fmt.Errorf("SetLiteral.Elements[%d]: %w", i, err)
		}
		elems[i] = raised
	}
	return &ast.SetLiteral{Elements: elems}, nil
}

func raiseTupleLiteral(e *me.TupleLiteral, ctx *RaiseContext) (ast.Expression, error) {
	elems := make([]ast.Expression, len(e.Elements))
	for i, elem := range e.Elements {
		raised, err := Raise(elem, ctx)
		if err != nil {
			return nil, fmt.Errorf("TupleLiteral.Elements[%d]: %w", i, err)
		}
		elems[i] = raised
	}
	return &ast.TupleLiteral{Elements: elems}, nil
}

func raiseRecordLiteral(e *me.RecordLiteral, ctx *RaiseContext) (ast.Expression, error) {
	bindings := make([]*ast.FieldBinding, len(e.Fields))
	for i, f := range e.Fields {
		val, err := Raise(f.Value, ctx)
		if err != nil {
			return nil, fmt.Errorf("RecordLiteral.Fields[%d]: %w", i, err)
		}
		bindings[i] = &ast.FieldBinding{
			Field:      &ast.Identifier{Value: f.Name},
			Expression: val,
		}
	}
	return &ast.RecordInstance{Bindings: bindings}, nil
}

func raiseSetConstant(e *me.SetConstant) (ast.Expression, error) {
	val, ok := raiseSetConstantKind[e.Kind]
	if !ok {
		return nil, fmt.Errorf("unknown SetConstantKind: %q", e.Kind)
	}
	// Emit as Identifier rather than ast.SetConstant because the parser
	// produces Identifier nodes for Nat/Int/Real/BOOLEAN (they are not
	// reserved keywords in the PEG grammar).
	return &ast.Identifier{Value: val}, nil
}

func raiseSetRange(e *me.SetRange, ctx *RaiseContext) (ast.Expression, error) {
	start, err := Raise(e.Start, ctx)
	if err != nil {
		return nil, fmt.Errorf("SetRange.Start: %w", err)
	}
	end, err := Raise(e.End, ctx)
	if err != nil {
		return nil, fmt.Errorf("SetRange.End: %w", err)
	}
	return &ast.SetRangeExpr{Start: start, End: end}, nil
}

// --- Reference raising ---

func raiseAttributeRef(e *me.AttributeRef, ctx *RaiseContext) (ast.Expression, error) {
	name, ok := ctx.AttributeNames[e.AttributeKey]
	if !ok {
		return nil, fmt.Errorf("unresolved attribute key: %v", e.AttributeKey)
	}
	return &ast.Identifier{Value: name}, nil
}

func raiseNextState(e *me.NextState, ctx *RaiseContext) (ast.Expression, error) {
	inner, err := Raise(e.Expr, ctx)
	if err != nil {
		return nil, fmt.Errorf("NextState: %w", err)
	}
	return &ast.Primed{Base: inner}, nil
}

func raiseNamedSetRef(e *me.NamedSetRef, ctx *RaiseContext) (ast.Expression, error) {
	name, ok := ctx.NamedSets[e.SetKey]
	if !ok {
		return nil, fmt.Errorf("unresolved named set key: %v", e.SetKey)
	}
	return &ast.Identifier{Value: name}, nil
}

// --- Unary operator raising ---

func raiseNegate(e *me.Negate, ctx *RaiseContext) (ast.Expression, error) {
	inner, err := Raise(e.Expr, ctx)
	if err != nil {
		return nil, fmt.Errorf("negate: %w", err)
	}
	return ast.NewNegation(inner), nil
}

func raiseNot(e *me.Not, ctx *RaiseContext) (ast.Expression, error) {
	inner, err := Raise(e.Expr, ctx)
	if err != nil {
		return nil, fmt.Errorf("not: %w", err)
	}
	return &ast.UnaryLogic{Operator: "¬", Right: inner}, nil
}

// --- Binary operator raising ---

func raiseBinaryArith(e *me.BinaryArith, ctx *RaiseContext) (ast.Expression, error) {
	op, ok := raiseArithOp[e.Op]
	if !ok {
		return nil, fmt.Errorf("unknown ArithOp: %q", e.Op)
	}
	left, err := Raise(e.Left, ctx)
	if err != nil {
		return nil, fmt.Errorf("BinaryArith.Left: %w", err)
	}
	right, err := Raise(e.Right, ctx)
	if err != nil {
		return nil, fmt.Errorf("BinaryArith.Right: %w", err)
	}
	return &ast.BinaryArithmetic{Operator: op, Left: left, Right: right}, nil
}

func raiseBinaryLogic(e *me.BinaryLogic, ctx *RaiseContext) (ast.Expression, error) {
	op, ok := raiseLogicOp[e.Op]
	if !ok {
		return nil, fmt.Errorf("unknown LogicOp: %q", e.Op)
	}
	left, err := Raise(e.Left, ctx)
	if err != nil {
		return nil, fmt.Errorf("BinaryLogic.Left: %w", err)
	}
	right, err := Raise(e.Right, ctx)
	if err != nil {
		return nil, fmt.Errorf("BinaryLogic.Right: %w", err)
	}
	return &ast.BinaryLogic{Operator: op, Left: left, Right: right}, nil
}

// raiseCompare raises a Compare node. Compare with eq/neq → BinaryEquality;
// Compare with lt/gt/lte/gte → BinaryComparison. This matches the different
// precedence levels in the PEG grammar.
func raiseCompare(e *me.Compare, ctx *RaiseContext) (ast.Expression, error) {
	op, ok := raiseCompareOp[e.Op]
	if !ok {
		return nil, fmt.Errorf("unknown CompareOp: %q", e.Op)
	}
	left, err := Raise(e.Left, ctx)
	if err != nil {
		return nil, fmt.Errorf("Compare.Left: %w", err)
	}
	right, err := Raise(e.Right, ctx)
	if err != nil {
		return nil, fmt.Errorf("Compare.Right: %w", err)
	}

	// Equality operators go through BinaryEquality (different precedence level).
	if e.Op == me.CompareEq || e.Op == me.CompareNeq {
		return &ast.BinaryEquality{Operator: op, Left: left, Right: right}, nil
	}
	return &ast.BinaryComparison{Operator: op, Left: left, Right: right}, nil
}

func raiseSetOp(e *me.SetOp, ctx *RaiseContext) (ast.Expression, error) {
	op, ok := raiseSetOpMap[e.Op]
	if !ok {
		return nil, fmt.Errorf("unknown SetOpKind: %q", e.Op)
	}
	left, err := Raise(e.Left, ctx)
	if err != nil {
		return nil, fmt.Errorf("SetOp.Left: %w", err)
	}
	right, err := Raise(e.Right, ctx)
	if err != nil {
		return nil, fmt.Errorf("SetOp.Right: %w", err)
	}
	return &ast.BinarySetOperation{Operator: op, Left: left, Right: right}, nil
}

func raiseSetCompare(e *me.SetCompare, ctx *RaiseContext) (ast.Expression, error) {
	op, ok := raiseSetCompareOp[e.Op]
	if !ok {
		return nil, fmt.Errorf("unknown SetCompareOp: %q", e.Op)
	}
	left, err := Raise(e.Left, ctx)
	if err != nil {
		return nil, fmt.Errorf("SetCompare.Left: %w", err)
	}
	right, err := Raise(e.Right, ctx)
	if err != nil {
		return nil, fmt.Errorf("SetCompare.Right: %w", err)
	}
	return &ast.BinarySetComparison{Operator: op, Left: left, Right: right}, nil
}

func raiseBagOp(e *me.BagOp, ctx *RaiseContext) (ast.Expression, error) {
	op, ok := raiseBagOpMap[e.Op]
	if !ok {
		return nil, fmt.Errorf("unknown BagOpKind: %q", e.Op)
	}
	left, err := Raise(e.Left, ctx)
	if err != nil {
		return nil, fmt.Errorf("BagOp.Left: %w", err)
	}
	right, err := Raise(e.Right, ctx)
	if err != nil {
		return nil, fmt.Errorf("BagOp.Right: %w", err)
	}
	return &ast.BinaryBagOperation{Operator: op, Left: left, Right: right}, nil
}

func raiseBagCompare(e *me.BagCompare, ctx *RaiseContext) (ast.Expression, error) {
	op, ok := raiseBagCompareOp[e.Op]
	if !ok {
		return nil, fmt.Errorf("unknown BagCompareOp: %q", e.Op)
	}
	left, err := Raise(e.Left, ctx)
	if err != nil {
		return nil, fmt.Errorf("BagCompare.Left: %w", err)
	}
	right, err := Raise(e.Right, ctx)
	if err != nil {
		return nil, fmt.Errorf("BagCompare.Right: %w", err)
	}
	return &ast.BinaryBagComparison{Operator: op, Left: left, Right: right}, nil
}

func raiseMembership(e *me.Membership, ctx *RaiseContext) (ast.Expression, error) {
	element, err := Raise(e.Element, ctx)
	if err != nil {
		return nil, fmt.Errorf("Membership.Element: %w", err)
	}
	set, err := Raise(e.Set, ctx)
	if err != nil {
		return nil, fmt.Errorf("Membership.Set: %w", err)
	}
	op := "∈"
	if e.Negated {
		op = "∉"
	}
	return &ast.Membership{Operator: op, Left: element, Right: set}, nil
}

// --- Concatenation raising ---

func raiseStringConcat(e *me.StringConcat, ctx *RaiseContext) (ast.Expression, error) {
	operands := make([]ast.Expression, len(e.Operands))
	for i, op := range e.Operands {
		raised, err := Raise(op, ctx)
		if err != nil {
			return nil, fmt.Errorf("StringConcat.Operands[%d]: %w", i, err)
		}
		operands[i] = raised
	}
	return &ast.TupleConcat{Operator: "∘", Operands: operands}, nil
}

func raiseTupleConcat(e *me.TupleConcat, ctx *RaiseContext) (ast.Expression, error) {
	operands := make([]ast.Expression, len(e.Operands))
	for i, op := range e.Operands {
		raised, err := Raise(op, ctx)
		if err != nil {
			return nil, fmt.Errorf("TupleConcat.Operands[%d]: %w", i, err)
		}
		operands[i] = raised
	}
	return &ast.TupleConcat{Operator: "∘", Operands: operands}, nil
}

// --- Field/Index access raising ---

func raiseFieldAccess(e *me.FieldAccess, ctx *RaiseContext) (ast.Expression, error) {
	base, err := Raise(e.Base, ctx)
	if err != nil {
		return nil, fmt.Errorf("FieldAccess.Base: %w", err)
	}
	return &ast.FieldAccess{Base: base, Member: e.Field}, nil
}

func raiseTupleIndex(e *me.TupleIndex, ctx *RaiseContext) (ast.Expression, error) {
	tuple, err := Raise(e.Tuple, ctx)
	if err != nil {
		return nil, fmt.Errorf("TupleIndex.Tuple: %w", err)
	}
	index, err := Raise(e.Index, ctx)
	if err != nil {
		return nil, fmt.Errorf("TupleIndex.Index: %w", err)
	}
	return &ast.TupleIndex{Tuple: tuple, Index: index}, nil
}

func raiseStringIndex(e *me.StringIndex, ctx *RaiseContext) (ast.Expression, error) {
	str, err := Raise(e.Str, ctx)
	if err != nil {
		return nil, fmt.Errorf("StringIndex.Str: %w", err)
	}
	index, err := Raise(e.Index, ctx)
	if err != nil {
		return nil, fmt.Errorf("StringIndex.Index: %w", err)
	}
	return &ast.StringIndex{Str: str, Index: index}, nil
}

// --- Record alteration raising ---

func raiseRecordUpdate(e *me.RecordUpdate, ctx *RaiseContext) (ast.Expression, error) {
	// The base can be any expression (identifier, another RecordAltered, etc.).
	base, err := Raise(e.Base, ctx)
	if err != nil {
		return nil, fmt.Errorf("RecordUpdate.Base: %w", err)
	}

	alts := make([]*ast.FieldAlteration, len(e.Alterations))
	for i, alt := range e.Alterations {
		val, err := Raise(alt.Value, ctx)
		if err != nil {
			return nil, fmt.Errorf("RecordUpdate.Alterations[%d]: %w", i, err)
		}
		alts[i] = &ast.FieldAlteration{
			Field:      &ast.FieldAccess{Member: alt.Field},
			Expression: val,
		}
	}
	return &ast.RecordAltered{Base: base, Alterations: alts}, nil
}

// --- Control flow raising ---

func raiseIfThenElse(e *me.IfThenElse, ctx *RaiseContext) (ast.Expression, error) {
	cond, err := Raise(e.Condition, ctx)
	if err != nil {
		return nil, fmt.Errorf("IfThenElse.Condition: %w", err)
	}
	then, err := Raise(e.Then, ctx)
	if err != nil {
		return nil, fmt.Errorf("IfThenElse.Then: %w", err)
	}
	elseExpr, err := Raise(e.Else, ctx)
	if err != nil {
		return nil, fmt.Errorf("IfThenElse.Else: %w", err)
	}
	return &ast.IfThenElse{Condition: cond, Then: then, Else: elseExpr}, nil
}

func raiseCase(e *me.Case, ctx *RaiseContext) (ast.Expression, error) {
	branches := make([]*ast.CaseBranch, len(e.Branches))
	for i, branch := range e.Branches {
		cond, err := Raise(branch.Condition, ctx)
		if err != nil {
			return nil, fmt.Errorf("Case.Branches[%d].Condition: %w", i, err)
		}
		result, err := Raise(branch.Result, ctx)
		if err != nil {
			return nil, fmt.Errorf("Case.Branches[%d].Result: %w", i, err)
		}
		branches[i] = &ast.CaseBranch{Condition: cond, Result: result}
	}
	var other ast.Expression
	if e.Otherwise != nil {
		var err error
		other, err = Raise(e.Otherwise, ctx)
		if err != nil {
			return nil, fmt.Errorf("Case.Otherwise: %w", err)
		}
	}
	return &ast.CaseExpr{Branches: branches, Other: other}, nil
}

// --- Quantifier raising ---

func raiseQuantifier(e *me.Quantifier, ctx *RaiseContext) (ast.Expression, error) {
	quantifier, ok := raiseQuantifierKind[e.Kind]
	if !ok {
		return nil, fmt.Errorf("unknown QuantifierKind: %q", e.Kind)
	}

	domain, err := Raise(e.Domain, ctx)
	if err != nil {
		return nil, fmt.Errorf("Quantifier.Domain: %w", err)
	}
	predicate, err := Raise(e.Predicate, ctx)
	if err != nil {
		return nil, fmt.Errorf("Quantifier.Predicate: %w", err)
	}

	// Reconstruct the Membership binding: variable ∈ domain.
	membership := &ast.Membership{
		Operator: "∈",
		Left:     &ast.Identifier{Value: e.Variable},
		Right:    domain,
	}

	return &ast.Quantifier{
		Quantifier: quantifier,
		Membership: membership,
		Predicate:  predicate,
	}, nil
}

func raiseSetFilter(e *me.SetFilter, ctx *RaiseContext) (ast.Expression, error) {
	set, err := Raise(e.Set, ctx)
	if err != nil {
		return nil, fmt.Errorf("SetFilter.Set: %w", err)
	}
	predicate, err := Raise(e.Predicate, ctx)
	if err != nil {
		return nil, fmt.Errorf("SetFilter.Predicate: %w", err)
	}

	// Reconstruct the Membership binding: variable ∈ set.
	membership := &ast.Membership{
		Operator: "∈",
		Left:     &ast.Identifier{Value: e.Variable},
		Right:    set,
	}

	return &ast.SetFilter{
		Membership: membership,
		Predicate:  predicate,
	}, nil
}

// --- Call raising ---

func raiseActionCall(e *me.ActionCall, ctx *RaiseContext) (ast.Expression, error) {
	args := make([]ast.Expression, len(e.Args))
	for i, arg := range e.Args {
		raised, err := Raise(arg, ctx)
		if err != nil {
			return nil, fmt.Errorf("ActionCall.Args[%d]: %w", i, err)
		}
		args[i] = raised
	}

	// Try same-class action first.
	if name, ok := ctx.ActionNames[e.ActionKey]; ok {
		return &ast.FunctionCall{
			Name: &ast.Identifier{Value: name},
			Args: args,
		}, nil
	}

	// Try same-class query.
	if name, ok := ctx.QueryNames[e.ActionKey]; ok {
		return &ast.FunctionCall{
			Name: &ast.Identifier{Value: name},
			Args: args,
		}, nil
	}

	// Cross-class action: split scope path.
	if scopePath, ok := ctx.ActionScopePaths[e.ActionKey]; ok {
		parts := strings.Split(scopePath, "!")
		if len(parts) < 2 {
			return &ast.FunctionCall{
				Name: &ast.Identifier{Value: scopePath},
				Args: args,
			}, nil
		}
		scopeIdents := make([]*ast.Identifier, len(parts)-1)
		for i, p := range parts[:len(parts)-1] {
			scopeIdents[i] = &ast.Identifier{Value: p}
		}
		return &ast.FunctionCall{
			ScopePath: scopeIdents,
			Name:      &ast.Identifier{Value: parts[len(parts)-1]},
			Args:      args,
		}, nil
	}

	return nil, fmt.Errorf("unresolved action key: %v", e.ActionKey)
}

func raiseGlobalCall(e *me.GlobalCall, ctx *RaiseContext) (ast.Expression, error) {
	args := make([]ast.Expression, len(e.Args))
	for i, arg := range e.Args {
		raised, err := Raise(arg, ctx)
		if err != nil {
			return nil, fmt.Errorf("GlobalCall.Args[%d]: %w", i, err)
		}
		args[i] = raised
	}

	name, ok := ctx.GlobalFunctions[e.FunctionKey]
	if !ok {
		return nil, fmt.Errorf("unresolved global function key: %v", e.FunctionKey)
	}

	return &ast.FunctionCall{
		Name: &ast.Identifier{Value: name},
		Args: args,
	}, nil
}

func raiseBuiltinCall(e *me.BuiltinCall, ctx *RaiseContext) (ast.Expression, error) {
	args := make([]ast.Expression, len(e.Args))
	for i, arg := range e.Args {
		raised, err := Raise(arg, ctx)
		if err != nil {
			return nil, fmt.Errorf("BuiltinCall.Args[%d]: %w", i, err)
		}
		args[i] = raised
	}

	return &ast.FunctionCall{
		ScopePath: []*ast.Identifier{{Value: e.Module}},
		Name:      &ast.Identifier{Value: e.Function},
		Args:      args,
	}, nil
}
