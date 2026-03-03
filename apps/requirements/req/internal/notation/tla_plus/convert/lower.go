package convert

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_expression"
)

// LowerContext provides class-level context for semantic resolution during lowering.
// Identity keys are pre-constructed from the model hierarchy and passed in, not built during lowering.
type LowerContext struct {
	// ClassKey is the identity key for the current class (used to resolve same-class references).
	ClassKey identity.Key

	// AttributeNames maps attribute names to their identity keys within the current class.
	AttributeNames map[string]identity.Key

	// ActionNames maps action names to their identity keys within the current class.
	ActionNames map[string]identity.Key

	// QueryNames maps query names to their identity keys within the current class.
	QueryNames map[string]identity.Key

	// GlobalFunctions maps global function names (with leading underscore) to their identity keys.
	GlobalFunctions map[string]identity.Key

	// NamedSets maps named set names to their identity keys.
	NamedSets map[string]identity.Key

	// AllActions maps fully scoped action names (Domain!Subdomain!Class!Action) to identity keys
	// for cross-class action calls.
	AllActions map[string]identity.Key

	// Parameters is the set of parameter names bound by the enclosing action/query/global function.
	Parameters map[string]bool

	// localVars tracks quantifier-bound variables in scope. Managed internally.
	localVars map[string]bool

	// exceptField tracks the current field name inside a RecordAltered alteration
	// so ExistingValue (@) can be lowered to PriorFieldValue.
	exceptField string
}

// Lower converts a TLA+ AST expression into a notation-independent model expression.
// The LowerContext provides semantic resolution: identifiers are resolved to AttributeRef,
// SelfRef, LocalVar, etc. based on what names are in scope.
func Lower(expr ast.Expression, ctx *LowerContext) (me.Expression, error) {
	if expr == nil {
		return nil, fmt.Errorf("cannot lower nil expression")
	}

	switch e := expr.(type) {
	// --- Literals ---
	case *ast.BooleanLiteral:
		return lowerBooleanLiteral(e)
	case *ast.NumberLiteral:
		return lowerNumberLiteral(e)
	case *ast.StringLiteral:
		return lowerStringLiteral(e)
	case *ast.Fraction:
		return lowerFraction(e, ctx)

	// --- Collections ---
	case *ast.SetLiteral:
		return lowerSetLiteral(e, ctx)
	case *ast.SetLiteralEnum:
		return lowerSetLiteralEnum(e)
	case *ast.SetLiteralInt:
		return lowerSetLiteralInt(e)
	case *ast.SetConstant:
		return lowerSetConstant(e)
	case *ast.SetRange:
		return lowerSetRange(e)
	case *ast.SetRangeExpr:
		return lowerSetRangeExpr(e, ctx)
	case *ast.TupleLiteral:
		return lowerTupleLiteral(e, ctx)
	case *ast.RecordInstance:
		return lowerRecordInstance(e, ctx)

	// --- References ---
	case *ast.Identifier:
		return lowerIdentifier(e, ctx)
	case *ast.ExistingValue:
		return lowerExistingValue(ctx)

	// --- Unary operators ---
	case *ast.UnaryNegation:
		return lowerUnaryNegation(e, ctx)
	case *ast.UnaryLogic:
		return lowerUnaryLogic(e, ctx)
	case *ast.Primed:
		return lowerPrimed(e, ctx)

	// --- Binary operators ---
	case *ast.BinaryArithmetic:
		return lowerBinaryArithmetic(e, ctx)
	case *ast.BinaryLogic:
		return lowerBinaryLogic(e, ctx)
	case *ast.BinaryEquality:
		return lowerBinaryEquality(e, ctx)
	case *ast.BinaryComparison:
		return lowerBinaryComparison(e, ctx)
	case *ast.BinarySetOperation:
		return lowerBinarySetOperation(e, ctx)
	case *ast.BinarySetComparison:
		return lowerBinarySetComparison(e, ctx)
	case *ast.BinaryBagOperation:
		return lowerBinaryBagOperation(e, ctx)
	case *ast.BinaryBagComparison:
		return lowerBinaryBagComparison(e, ctx)
	case *ast.Membership:
		return lowerMembership(e, ctx)

	// --- String/Tuple concatenation ---
	case *ast.StringConcat:
		return lowerStringConcat(e, ctx)
	case *ast.TupleConcat:
		return lowerTupleConcat(e, ctx)

	// --- Indexing and field access ---
	case *ast.TupleIndex:
		return lowerTupleIndex(e, ctx)
	case *ast.StringIndex:
		return lowerStringIndex(e, ctx)
	case *ast.FieldAccess:
		return lowerFieldAccess(e, ctx)

	// --- Record alteration ---
	case *ast.RecordAltered:
		return lowerRecordAltered(e, ctx)

	// --- Control flow ---
	case *ast.IfThenElse:
		return lowerIfThenElse(e, ctx)
	case *ast.CaseExpr:
		return lowerCaseExpr(e, ctx)

	// --- Quantifiers ---
	case *ast.Quantifier:
		return lowerQuantifier(e, ctx)
	case *ast.SetFilter:
		return lowerSetFilter(e, ctx)

	// --- Calls ---
	case *ast.FunctionCall:
		return lowerFunctionCall(e, ctx)
	case *ast.BuiltinCall:
		return lowerBuiltinCall(e, ctx)
	case *ast.ScopedCall:
		return lowerScopedCall(e, ctx)

	// --- Grouping ---
	case *ast.Parenthesized:
		return Lower(e.Inner, ctx)

	// --- Type expressions are not value expressions ---
	case *ast.RecordTypeExpr:
		return nil, fmt.Errorf("RecordTypeExpr is a type expression, not a value expression")
	case *ast.CartesianProduct:
		return nil, fmt.Errorf("CartesianProduct is a type expression, not a value expression")

	default:
		return nil, fmt.Errorf("unsupported AST node type: %T", expr)
	}
}

// --- Literal lowering ---

func lowerBooleanLiteral(e *ast.BooleanLiteral) (*me.BoolLiteral, error) {
	return &me.BoolLiteral{Value: e.Value}, nil
}

func lowerNumberLiteral(e *ast.NumberLiteral) (*me.IntLiteral, error) {
	// NumberLiteral stores digits as strings with a base. We parse to *big.Int.
	// Only integers go through NumberLiteral (decimals are Fraction of two NumberLiterals).
	if e.HasDecimalPoint {
		return nil, fmt.Errorf("NumberLiteral with decimal point should be represented as a Fraction, not lowered directly")
	}

	v := new(big.Int)
	_, ok := v.SetString(e.IntegerPart, int(e.Base))
	if !ok {
		return nil, fmt.Errorf("failed to parse integer from NumberLiteral: %s (base %d)", e.IntegerPart, e.Base)
	}
	return &me.IntLiteral{Value: v}, nil
}

func lowerStringLiteral(e *ast.StringLiteral) (*me.StringLiteral, error) {
	return &me.StringLiteral{Value: e.Value}, nil
}

func lowerFraction(e *ast.Fraction, ctx *LowerContext) (me.Expression, error) {
	num, err := Lower(e.Numerator, ctx)
	if err != nil {
		return nil, fmt.Errorf("Fraction.Numerator: %w", err)
	}
	den, err := Lower(e.Denominator, ctx)
	if err != nil {
		return nil, fmt.Errorf("Fraction.Denominator: %w", err)
	}

	// If both sides are IntLiterals, collapse to a RationalLiteral.
	numInt, numOk := num.(*me.IntLiteral)
	denInt, denOk := den.(*me.IntLiteral)
	if numOk && denOk {
		rat := new(big.Rat).SetFrac(numInt.Value, denInt.Value)
		return &me.RationalLiteral{Value: rat}, nil
	}

	// Otherwise, treat as arithmetic division.
	return &me.BinaryArith{Op: me.ArithDiv, Left: num, Right: den}, nil
}

// --- Collection lowering ---

func lowerSetLiteral(e *ast.SetLiteral, ctx *LowerContext) (*me.SetLiteral, error) {
	elems := make([]me.Expression, len(e.Elements))
	for i, elem := range e.Elements {
		lowered, err := Lower(elem, ctx)
		if err != nil {
			return nil, fmt.Errorf("SetLiteral.Elements[%d]: %w", i, err)
		}
		elems[i] = lowered
	}
	return &me.SetLiteral{Elements: elems}, nil
}

func lowerSetLiteralEnum(e *ast.SetLiteralEnum) (*me.SetLiteral, error) {
	elems := make([]me.Expression, len(e.Values))
	for i, v := range e.Values {
		elems[i] = &me.StringLiteral{Value: v}
	}
	return &me.SetLiteral{Elements: elems}, nil
}

func lowerSetLiteralInt(e *ast.SetLiteralInt) (*me.SetLiteral, error) {
	elems := make([]me.Expression, len(e.Values))
	for i, v := range e.Values {
		elems[i] = &me.IntLiteral{Value: big.NewInt(int64(v))}
	}
	return &me.SetLiteral{Elements: elems}, nil
}

func lowerSetConstant(e *ast.SetConstant) (*me.SetConstant, error) {
	switch e.Value {
	case ast.SetConstantNat:
		return &me.SetConstant{Kind: me.SetConstantNat}, nil
	case ast.SetConstantInt:
		return &me.SetConstant{Kind: me.SetConstantInt}, nil
	case ast.SetConstantReal:
		return &me.SetConstant{Kind: me.SetConstantReal}, nil
	case ast.SetConstantBoolean:
		return &me.SetConstant{Kind: me.SetConstantBoolean}, nil
	default:
		return nil, fmt.Errorf("unknown set constant: %s", e.Value)
	}
}

func lowerSetRange(e *ast.SetRange) (*me.SetRange, error) {
	return &me.SetRange{
		Start: &me.IntLiteral{Value: big.NewInt(int64(e.Start))},
		End:   &me.IntLiteral{Value: big.NewInt(int64(e.End))},
	}, nil
}

func lowerSetRangeExpr(e *ast.SetRangeExpr, ctx *LowerContext) (*me.SetRange, error) {
	start, err := Lower(e.Start, ctx)
	if err != nil {
		return nil, fmt.Errorf("SetRangeExpr.Start: %w", err)
	}
	end, err := Lower(e.End, ctx)
	if err != nil {
		return nil, fmt.Errorf("SetRangeExpr.End: %w", err)
	}
	return &me.SetRange{Start: start, End: end}, nil
}

func lowerTupleLiteral(e *ast.TupleLiteral, ctx *LowerContext) (*me.TupleLiteral, error) {
	elems := make([]me.Expression, len(e.Elements))
	for i, elem := range e.Elements {
		lowered, err := Lower(elem, ctx)
		if err != nil {
			return nil, fmt.Errorf("TupleLiteral.Elements[%d]: %w", i, err)
		}
		elems[i] = lowered
	}
	return &me.TupleLiteral{Elements: elems}, nil
}

func lowerRecordInstance(e *ast.RecordInstance, ctx *LowerContext) (*me.RecordLiteral, error) {
	fields := make([]me.RecordField, len(e.Bindings))
	for i, binding := range e.Bindings {
		val, err := Lower(binding.Expression, ctx)
		if err != nil {
			return nil, fmt.Errorf("RecordInstance.Bindings[%d]: %w", i, err)
		}
		fields[i] = me.RecordField{
			Name:  binding.Field.Value,
			Value: val,
		}
	}
	return &me.RecordLiteral{Fields: fields}, nil
}

// --- Reference lowering ---

func lowerIdentifier(e *ast.Identifier, ctx *LowerContext) (me.Expression, error) {
	name := e.Value

	// "self" → SelfRef.
	if name == "self" {
		return &me.SelfRef{}, nil
	}

	// Check quantifier-bound local variables first (innermost scope).
	if ctx.localVars != nil && ctx.localVars[name] {
		return &me.LocalVar{Name: name}, nil
	}

	// Check parameter names.
	if ctx.Parameters != nil && ctx.Parameters[name] {
		return &me.LocalVar{Name: name}, nil
	}

	// Check attribute names → AttributeRef.
	if key, ok := ctx.AttributeNames[name]; ok {
		return &me.AttributeRef{AttributeKey: key}, nil
	}

	// Check named sets → NamedSetRef.
	if key, ok := ctx.NamedSets[name]; ok {
		return &me.NamedSetRef{SetKey: key}, nil
	}

	return nil, fmt.Errorf("unresolved identifier: %q", name)
}

func lowerExistingValue(ctx *LowerContext) (*me.PriorFieldValue, error) {
	if ctx.exceptField == "" {
		return nil, fmt.Errorf("ExistingValue (@) used outside of EXCEPT context")
	}
	return &me.PriorFieldValue{Field: ctx.exceptField}, nil
}

// --- Unary operator lowering ---

func lowerUnaryNegation(e *ast.UnaryNegation, ctx *LowerContext) (*me.Negate, error) {
	inner, err := Lower(e.Right, ctx)
	if err != nil {
		return nil, fmt.Errorf("UnaryNegation: %w", err)
	}
	return &me.Negate{Expr: inner}, nil
}

func lowerUnaryLogic(e *ast.UnaryLogic, ctx *LowerContext) (*me.Not, error) {
	inner, err := Lower(e.Right, ctx)
	if err != nil {
		return nil, fmt.Errorf("UnaryLogic: %w", err)
	}
	return &me.Not{Expr: inner}, nil
}

func lowerPrimed(e *ast.Primed, ctx *LowerContext) (*me.NextState, error) {
	inner, err := Lower(e.Base, ctx)
	if err != nil {
		return nil, fmt.Errorf("Primed: %w", err)
	}
	return &me.NextState{Expr: inner}, nil
}

// --- Binary operator lowering ---

var arithOpMap = map[string]me.ArithOp{
	"+": me.ArithAdd,
	"-": me.ArithSub,
	"*": me.ArithMul,
	"÷": me.ArithDiv,
	"^": me.ArithPow,
	"%": me.ArithMod,
}

func lowerBinaryArithmetic(e *ast.BinaryArithmetic, ctx *LowerContext) (*me.BinaryArith, error) {
	op, ok := arithOpMap[e.Operator]
	if !ok {
		return nil, fmt.Errorf("unknown arithmetic operator: %q", e.Operator)
	}
	left, err := Lower(e.Left, ctx)
	if err != nil {
		return nil, fmt.Errorf("BinaryArithmetic.Left: %w", err)
	}
	right, err := Lower(e.Right, ctx)
	if err != nil {
		return nil, fmt.Errorf("BinaryArithmetic.Right: %w", err)
	}
	return &me.BinaryArith{Op: op, Left: left, Right: right}, nil
}

var logicOpMap = map[string]me.LogicOp{
	"∧": me.LogicAnd,
	"∨": me.LogicOr,
	"⇒": me.LogicImplies,
	"≡": me.LogicEquiv,
}

func lowerBinaryLogic(e *ast.BinaryLogic, ctx *LowerContext) (*me.BinaryLogic, error) {
	op, ok := logicOpMap[e.Operator]
	if !ok {
		return nil, fmt.Errorf("unknown logic operator: %q", e.Operator)
	}
	left, err := Lower(e.Left, ctx)
	if err != nil {
		return nil, fmt.Errorf("BinaryLogic.Left: %w", err)
	}
	right, err := Lower(e.Right, ctx)
	if err != nil {
		return nil, fmt.Errorf("BinaryLogic.Right: %w", err)
	}
	return &me.BinaryLogic{Op: op, Left: left, Right: right}, nil
}

func lowerBinaryEquality(e *ast.BinaryEquality, ctx *LowerContext) (*me.Compare, error) {
	var op me.CompareOp
	switch e.Operator {
	case "=":
		op = me.CompareEq
	case "≠":
		op = me.CompareNeq
	default:
		return nil, fmt.Errorf("unknown equality operator: %q", e.Operator)
	}
	left, err := Lower(e.Left, ctx)
	if err != nil {
		return nil, fmt.Errorf("BinaryEquality.Left: %w", err)
	}
	right, err := Lower(e.Right, ctx)
	if err != nil {
		return nil, fmt.Errorf("BinaryEquality.Right: %w", err)
	}
	return &me.Compare{Op: op, Left: left, Right: right}, nil
}

var compareOpMap = map[string]me.CompareOp{
	"<": me.CompareLt,
	">": me.CompareGt,
	"≤": me.CompareLte,
	"≥": me.CompareGte,
}

func lowerBinaryComparison(e *ast.BinaryComparison, ctx *LowerContext) (*me.Compare, error) {
	op, ok := compareOpMap[e.Operator]
	if !ok {
		return nil, fmt.Errorf("unknown comparison operator: %q", e.Operator)
	}
	left, err := Lower(e.Left, ctx)
	if err != nil {
		return nil, fmt.Errorf("BinaryComparison.Left: %w", err)
	}
	right, err := Lower(e.Right, ctx)
	if err != nil {
		return nil, fmt.Errorf("BinaryComparison.Right: %w", err)
	}
	return &me.Compare{Op: op, Left: left, Right: right}, nil
}

var setOpMap = map[string]me.SetOpKind{
	"∪":  me.SetUnion,
	"∩":  me.SetIntersect,
	"\\": me.SetDifference,
}

func lowerBinarySetOperation(e *ast.BinarySetOperation, ctx *LowerContext) (*me.SetOp, error) {
	op, ok := setOpMap[e.Operator]
	if !ok {
		return nil, fmt.Errorf("unknown set operation: %q", e.Operator)
	}
	left, err := Lower(e.Left, ctx)
	if err != nil {
		return nil, fmt.Errorf("BinarySetOperation.Left: %w", err)
	}
	right, err := Lower(e.Right, ctx)
	if err != nil {
		return nil, fmt.Errorf("BinarySetOperation.Right: %w", err)
	}
	return &me.SetOp{Op: op, Left: left, Right: right}, nil
}

var setCompareOpMap = map[string]me.SetCompareOp{
	"⊆": me.SetCompareSubsetEq,
	"⊂": me.SetCompareSubset,
	"⊇": me.SetCompareSupersetEq,
	"⊃": me.SetCompareSuperset,
}

func lowerBinarySetComparison(e *ast.BinarySetComparison, ctx *LowerContext) (me.Expression, error) {
	// BinarySetComparison can also use = and ≠ for set equality.
	switch e.Operator {
	case "=":
		left, err := Lower(e.Left, ctx)
		if err != nil {
			return nil, fmt.Errorf("BinarySetComparison.Left: %w", err)
		}
		right, err := Lower(e.Right, ctx)
		if err != nil {
			return nil, fmt.Errorf("BinarySetComparison.Right: %w", err)
		}
		return &me.Compare{Op: me.CompareEq, Left: left, Right: right}, nil
	case "≠":
		left, err := Lower(e.Left, ctx)
		if err != nil {
			return nil, fmt.Errorf("BinarySetComparison.Left: %w", err)
		}
		right, err := Lower(e.Right, ctx)
		if err != nil {
			return nil, fmt.Errorf("BinarySetComparison.Right: %w", err)
		}
		return &me.Compare{Op: me.CompareNeq, Left: left, Right: right}, nil
	}

	op, ok := setCompareOpMap[e.Operator]
	if !ok {
		return nil, fmt.Errorf("unknown set comparison operator: %q", e.Operator)
	}
	left, err := Lower(e.Left, ctx)
	if err != nil {
		return nil, fmt.Errorf("BinarySetComparison.Left: %w", err)
	}
	right, err := Lower(e.Right, ctx)
	if err != nil {
		return nil, fmt.Errorf("BinarySetComparison.Right: %w", err)
	}
	return &me.SetCompare{Op: op, Left: left, Right: right}, nil
}

var bagOpMap = map[string]me.BagOpKind{
	"⊕": me.BagSum,
	"⊖": me.BagDifference,
}

func lowerBinaryBagOperation(e *ast.BinaryBagOperation, ctx *LowerContext) (*me.BagOp, error) {
	op, ok := bagOpMap[e.Operator]
	if !ok {
		return nil, fmt.Errorf("unknown bag operation: %q", e.Operator)
	}
	left, err := Lower(e.Left, ctx)
	if err != nil {
		return nil, fmt.Errorf("BinaryBagOperation.Left: %w", err)
	}
	right, err := Lower(e.Right, ctx)
	if err != nil {
		return nil, fmt.Errorf("BinaryBagOperation.Right: %w", err)
	}
	return &me.BagOp{Op: op, Left: left, Right: right}, nil
}

var bagCompareOpMap = map[string]me.BagCompareOp{
	"⊏": me.BagCompareProperSubBag,
	"⊑": me.BagCompareSubBag,
	"⊐": me.BagCompareProperSupBag,
	"⊒": me.BagCompareSupBag,
}

func lowerBinaryBagComparison(e *ast.BinaryBagComparison, ctx *LowerContext) (*me.BagCompare, error) {
	op, ok := bagCompareOpMap[e.Operator]
	if !ok {
		return nil, fmt.Errorf("unknown bag comparison: %q", e.Operator)
	}
	left, err := Lower(e.Left, ctx)
	if err != nil {
		return nil, fmt.Errorf("BinaryBagComparison.Left: %w", err)
	}
	right, err := Lower(e.Right, ctx)
	if err != nil {
		return nil, fmt.Errorf("BinaryBagComparison.Right: %w", err)
	}
	return &me.BagCompare{Op: op, Left: left, Right: right}, nil
}

func lowerMembership(e *ast.Membership, ctx *LowerContext) (*me.Membership, error) {
	element, err := Lower(e.Left, ctx)
	if err != nil {
		return nil, fmt.Errorf("Membership.Element: %w", err)
	}
	set, err := Lower(e.Right, ctx)
	if err != nil {
		return nil, fmt.Errorf("Membership.Set: %w", err)
	}
	negated := e.Operator == "∉"
	return &me.Membership{Element: element, Set: set, Negated: negated}, nil
}

// --- Concatenation lowering ---

func lowerStringConcat(e *ast.StringConcat, ctx *LowerContext) (*me.StringConcat, error) {
	operands := make([]me.Expression, len(e.Operands))
	for i, op := range e.Operands {
		lowered, err := Lower(op, ctx)
		if err != nil {
			return nil, fmt.Errorf("StringConcat.Operands[%d]: %w", i, err)
		}
		operands[i] = lowered
	}
	return &me.StringConcat{Operands: operands}, nil
}

func lowerTupleConcat(e *ast.TupleConcat, ctx *LowerContext) (*me.TupleConcat, error) {
	operands := make([]me.Expression, len(e.Operands))
	for i, op := range e.Operands {
		lowered, err := Lower(op, ctx)
		if err != nil {
			return nil, fmt.Errorf("TupleConcat.Operands[%d]: %w", i, err)
		}
		operands[i] = lowered
	}
	return &me.TupleConcat{Operands: operands}, nil
}

// --- Indexing and field access lowering ---

func lowerTupleIndex(e *ast.TupleIndex, ctx *LowerContext) (*me.TupleIndex, error) {
	tuple, err := Lower(e.Tuple, ctx)
	if err != nil {
		return nil, fmt.Errorf("TupleIndex.Tuple: %w", err)
	}
	index, err := Lower(e.Index, ctx)
	if err != nil {
		return nil, fmt.Errorf("TupleIndex.Index: %w", err)
	}
	return &me.TupleIndex{Tuple: tuple, Index: index}, nil
}

func lowerStringIndex(e *ast.StringIndex, ctx *LowerContext) (*me.StringIndex, error) {
	str, err := Lower(e.Str, ctx)
	if err != nil {
		return nil, fmt.Errorf("StringIndex.Str: %w", err)
	}
	index, err := Lower(e.Index, ctx)
	if err != nil {
		return nil, fmt.Errorf("StringIndex.Index: %w", err)
	}
	return &me.StringIndex{Str: str, Index: index}, nil
}

func lowerFieldAccess(e *ast.FieldAccess, ctx *LowerContext) (*me.FieldAccess, error) {
	base := e.GetBase()
	if base == nil {
		// nil base means existing value (@) in EXCEPT context.
		if ctx.exceptField == "" {
			return nil, fmt.Errorf("FieldAccess with nil base outside of EXCEPT context")
		}
		return &me.FieldAccess{
			Base:  &me.PriorFieldValue{Field: ctx.exceptField},
			Field: e.Member,
		}, nil
	}
	loweredBase, err := Lower(base, ctx)
	if err != nil {
		return nil, fmt.Errorf("FieldAccess.Base: %w", err)
	}
	return &me.FieldAccess{Base: loweredBase, Field: e.Member}, nil
}

// --- Record alteration lowering ---

func lowerRecordAltered(e *ast.RecordAltered, ctx *LowerContext) (*me.RecordUpdate, error) {
	// The identifier is the record being altered.
	base, err := Lower(e.Identifier, ctx)
	if err != nil {
		return nil, fmt.Errorf("RecordAltered.Identifier: %w", err)
	}

	alts := make([]me.FieldAlteration, len(e.Alterations))
	for i, alt := range e.Alterations {
		fieldName := alt.Field.Member

		// Create a child context with the except field set for ExistingValue resolution.
		childCtx := *ctx
		childCtx.exceptField = fieldName

		val, err := Lower(alt.Expression, &childCtx)
		if err != nil {
			return nil, fmt.Errorf("RecordAltered.Alterations[%d]: %w", i, err)
		}
		alts[i] = me.FieldAlteration{
			Field: fieldName,
			Value: val,
		}
	}
	return &me.RecordUpdate{Base: base, Alterations: alts}, nil
}

// --- Control flow lowering ---

func lowerIfThenElse(e *ast.IfThenElse, ctx *LowerContext) (*me.IfThenElse, error) {
	cond, err := Lower(e.Condition, ctx)
	if err != nil {
		return nil, fmt.Errorf("IfThenElse.Condition: %w", err)
	}
	then, err := Lower(e.Then, ctx)
	if err != nil {
		return nil, fmt.Errorf("IfThenElse.Then: %w", err)
	}
	elseExpr, err := Lower(e.Else, ctx)
	if err != nil {
		return nil, fmt.Errorf("IfThenElse.Else: %w", err)
	}
	return &me.IfThenElse{Condition: cond, Then: then, Else: elseExpr}, nil
}

func lowerCaseExpr(e *ast.CaseExpr, ctx *LowerContext) (*me.Case, error) {
	branches := make([]me.CaseBranch, len(e.Branches))
	for i, branch := range e.Branches {
		cond, err := Lower(branch.Condition, ctx)
		if err != nil {
			return nil, fmt.Errorf("CaseExpr.Branches[%d].Condition: %w", i, err)
		}
		result, err := Lower(branch.Result, ctx)
		if err != nil {
			return nil, fmt.Errorf("CaseExpr.Branches[%d].Result: %w", i, err)
		}
		branches[i] = me.CaseBranch{Condition: cond, Result: result}
	}
	var otherwise me.Expression
	if e.Other != nil {
		var err error
		otherwise, err = Lower(e.Other, ctx)
		if err != nil {
			return nil, fmt.Errorf("CaseExpr.Other: %w", err)
		}
	}
	return &me.Case{Branches: branches, Otherwise: otherwise}, nil
}

// --- Quantifier lowering ---

// extractMembershipBinding decomposes an AST Membership node used as a quantifier binding
// into variable name and domain expression.
func extractMembershipBinding(expr ast.Expression, ctx *LowerContext) (string, me.Expression, error) {
	mem, ok := expr.(*ast.Membership)
	if !ok {
		return "", nil, fmt.Errorf("expected Membership for quantifier binding, got %T", expr)
	}
	ident, ok := mem.Left.(*ast.Identifier)
	if !ok {
		return "", nil, fmt.Errorf("expected Identifier on left side of quantifier binding, got %T", mem.Left)
	}
	domain, err := Lower(mem.Right, ctx)
	if err != nil {
		return "", nil, fmt.Errorf("quantifier domain: %w", err)
	}
	return ident.Value, domain, nil
}

// withLocalVar returns a copy of the context with the given variable name added to local vars.
func withLocalVar(ctx *LowerContext, name string) *LowerContext {
	child := *ctx
	child.localVars = make(map[string]bool)
	if ctx.localVars != nil {
		for k, v := range ctx.localVars {
			child.localVars[k] = v
		}
	}
	child.localVars[name] = true
	return &child
}

func lowerQuantifier(e *ast.Quantifier, ctx *LowerContext) (*me.Quantifier, error) {
	var kind me.QuantifierKind
	switch e.Quantifier {
	case "∀":
		kind = me.QuantifierForall
	case "∃":
		kind = me.QuantifierExists
	default:
		return nil, fmt.Errorf("unknown quantifier: %q", e.Quantifier)
	}

	varName, domain, err := extractMembershipBinding(e.Membership, ctx)
	if err != nil {
		return nil, fmt.Errorf("Quantifier: %w", err)
	}

	childCtx := withLocalVar(ctx, varName)
	predicate, err := Lower(e.Predicate, childCtx)
	if err != nil {
		return nil, fmt.Errorf("Quantifier.Predicate: %w", err)
	}
	return &me.Quantifier{Kind: kind, Variable: varName, Domain: domain, Predicate: predicate}, nil
}

func lowerSetFilter(e *ast.SetFilter, ctx *LowerContext) (*me.SetFilter, error) {
	varName, set, err := extractMembershipBinding(e.Membership, ctx)
	if err != nil {
		return nil, fmt.Errorf("SetFilter: %w", err)
	}

	childCtx := withLocalVar(ctx, varName)
	predicate, err := Lower(e.Predicate, childCtx)
	if err != nil {
		return nil, fmt.Errorf("SetFilter.Predicate: %w", err)
	}
	return &me.SetFilter{Variable: varName, Set: set, Predicate: predicate}, nil
}

// --- Call lowering ---

func lowerFunctionCall(e *ast.FunctionCall, ctx *LowerContext) (me.Expression, error) {
	if e.IsGlobalOrBuiltin() {
		return lowerGlobalOrBuiltinFunctionCall(e, ctx)
	}
	return lowerClassActionCall(e, ctx)
}

func lowerGlobalOrBuiltinFunctionCall(e *ast.FunctionCall, ctx *LowerContext) (me.Expression, error) {
	// Lower arguments.
	args := make([]me.Expression, len(e.Args))
	for i, arg := range e.Args {
		lowered, err := Lower(arg, ctx)
		if err != nil {
			return nil, fmt.Errorf("FunctionCall.Args[%d]: %w", i, err)
		}
		args[i] = lowered
	}

	if len(e.ScopePath) > 0 {
		// Built-in module call: _Module!Function(args...)
		module := e.ScopePath[0].Value
		function := e.Name.Value
		return &me.BuiltinCall{Module: module, Function: function, Args: args}, nil
	}

	// Global function call: _FunctionName(args...)
	name := e.Name.Value
	key, ok := ctx.GlobalFunctions[name]
	if !ok {
		return nil, fmt.Errorf("unresolved global function: %q", name)
	}
	return &me.GlobalCall{FunctionKey: key, Args: args}, nil
}

func lowerClassActionCall(e *ast.FunctionCall, ctx *LowerContext) (me.Expression, error) {
	// Lower arguments.
	args := make([]me.Expression, len(e.Args))
	for i, arg := range e.Args {
		lowered, err := Lower(arg, ctx)
		if err != nil {
			return nil, fmt.Errorf("FunctionCall.Args[%d]: %w", i, err)
		}
		args[i] = lowered
	}

	if len(e.ScopePath) == 0 {
		// Same-class action/query call: ActionName(args...)
		name := e.Name.Value

		// Check actions first.
		if key, ok := ctx.ActionNames[name]; ok {
			return &me.ActionCall{ActionKey: key, Args: args}, nil
		}
		// Then queries.
		if key, ok := ctx.QueryNames[name]; ok {
			return &me.ActionCall{ActionKey: key, Args: args}, nil
		}
		return nil, fmt.Errorf("unresolved action/query: %q", name)
	}

	// Cross-class action call: Domain!Subdomain!Class!Action or shorter scope paths.
	fullName := e.FullName()
	if key, ok := ctx.AllActions[fullName]; ok {
		return &me.ActionCall{ActionKey: key, Args: args}, nil
	}
	return nil, fmt.Errorf("unresolved cross-class action: %q", fullName)
}

func lowerBuiltinCall(e *ast.BuiltinCall, ctx *LowerContext) (*me.BuiltinCall, error) {
	// BuiltinCall.Name is "_Module!Function" format.
	parts := strings.SplitN(e.Name, "!", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid builtin call name format: %q (expected _Module!Function)", e.Name)
	}

	args := make([]me.Expression, len(e.Args))
	for i, arg := range e.Args {
		lowered, err := Lower(arg, ctx)
		if err != nil {
			return nil, fmt.Errorf("BuiltinCall.Args[%d]: %w", i, err)
		}
		args[i] = lowered
	}

	return &me.BuiltinCall{Module: parts[0], Function: parts[1], Args: args}, nil
}

func lowerScopedCall(e *ast.ScopedCall, ctx *LowerContext) (me.Expression, error) {
	// Lower the parameter expression.
	param, err := Lower(e.Parameter, ctx)
	if err != nil {
		return nil, fmt.Errorf("ScopedCall.Parameter: %w", err)
	}
	args := []me.Expression{param}

	if e.ModelScope {
		// Model scope: _FunctionName(param)
		name := "_" + e.FunctionName.Value
		key, ok := ctx.GlobalFunctions[name]
		if !ok {
			return nil, fmt.Errorf("unresolved model-scope function: %q", name)
		}
		return &me.GlobalCall{FunctionKey: key, Args: args}, nil
	}

	// Build the scoped name for lookup in AllActions.
	var parts []string
	if e.Domain != nil {
		parts = append(parts, e.Domain.Value)
	}
	if e.Subdomain != nil {
		parts = append(parts, e.Subdomain.Value)
	}
	if e.Class != nil {
		parts = append(parts, e.Class.Value)
	}
	parts = append(parts, e.FunctionName.Value)
	fullName := strings.Join(parts, "!")

	// Try same-class first if no explicit scope.
	if e.Domain == nil && e.Subdomain == nil && e.Class == nil {
		name := e.FunctionName.Value
		if key, ok := ctx.ActionNames[name]; ok {
			return &me.ActionCall{ActionKey: key, Args: args}, nil
		}
		if key, ok := ctx.QueryNames[name]; ok {
			return &me.ActionCall{ActionKey: key, Args: args}, nil
		}
	}

	// Cross-class lookup.
	if key, ok := ctx.AllActions[fullName]; ok {
		return &me.ActionCall{ActionKey: key, Args: args}, nil
	}
	return nil, fmt.Errorf("unresolved scoped call: %q", fullName)
}
