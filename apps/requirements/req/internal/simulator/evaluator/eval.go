package evaluator

import (
	"fmt"

	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// Package-level constants for common values.
// Using empty set {} instead of NULL for TLA+.
var (
	TRUE      = object.NewBoolean(true)
	FALSE     = object.NewBoolean(false)
	EMPTY_SET = object.NewSet() // {} - used instead of NULL
)

// EvalResult encapsulates the result of evaluation.
// It includes both the evaluated value and metadata about state changes.
type EvalResult struct {
	// Value is the evaluated result.
	// For Assignments: this is the EmptySet (success indicator).
	// For Logic expressions: this is a Boolean.
	// For other expressions: this is the computed value.
	Value object.Object

	// PrimedBindings contains all variables that were assigned via prime (x' = ...).
	// This is populated after evaluating Assignment statements.
	PrimedBindings map[string]object.Object

	// Error indicates if evaluation failed.
	// When set, Value should be ignored.
	Error *object.Error
}

// NewEvalResult creates a successful result with a value.
func NewEvalResult(value object.Object) *EvalResult {
	return &EvalResult{
		Value:          value,
		PrimedBindings: make(map[string]object.Object),
	}
}

// NewEvalResultWithPrimed creates a result with primed bindings.
func NewEvalResultWithPrimed(value object.Object, primed map[string]object.Object) *EvalResult {
	return &EvalResult{
		Value:          value,
		PrimedBindings: primed,
	}
}

// NewEvalError creates an error result.
func NewEvalError(format string, args ...any) *EvalResult {
	return &EvalResult{
		Error: &object.Error{Message: fmt.Sprintf(format, args...)},
	}
}

// IsError checks if the result is an error.
func (r *EvalResult) IsError() bool {
	return r.Error != nil
}

// Success checks if evaluation succeeded (no error).
func (r *EvalResult) Success() bool {
	return r.Error == nil
}

// HasPrimedBindings checks if any variables were primed.
func (r *EvalResult) HasPrimedBindings() bool {
	return len(r.PrimedBindings) > 0
}

// Eval evaluates a model expression and returns the result.
// This is the primary evaluator entry point. It dispatches on the concrete
// logic_expression.Expression type to the appropriate handler.
//
//complexity:cyclo:warn=50,fail=50 A simple routing switch.
//complexity:fanout:warn=50,fail=50 A simple routing switch.
func Eval(node me.Expression, bindings *Bindings) *EvalResult {
	switch n := node.(type) {
	// === Literals ===
	case *me.IntLiteral:
		return evalIntLiteral(n)
	case *me.RationalLiteral:
		return evalRationalLiteral(n)
	case *me.BoolLiteral:
		return evalBoolLiteral(n)
	case *me.StringLiteral:
		return evalMEStringLiteral(n)
	case *me.TupleLiteral:
		return evalMETupleLiteral(n, bindings)
	case *me.SetLiteral:
		return evalMESetLiteral(n, bindings)
	case *me.SetConstant:
		return evalMESetConstant(n)
	case *me.SetRange:
		return evalMESetRange(n, bindings)
	case *me.RecordLiteral:
		return evalMERecordLiteral(n, bindings)

	// === References ===
	case *me.SelfRef:
		return evalSelfRef(bindings)
	case *me.AttributeRef:
		return evalAttributeRef(n, bindings)
	case *me.LocalVar:
		return evalLocalVar(n, bindings)
	case *me.PriorFieldValue:
		return evalPriorFieldValue(bindings)
	case *me.NextState:
		return evalNextState(n, bindings)
	case *me.NamedSetRef:
		return evalNamedSetRef(n, bindings)

	// === Binary operators ===
	case *me.BinaryArith:
		return evalMEBinaryArith(n, bindings)
	case *me.BinaryLogic:
		return evalMEBinaryLogic(n, bindings)
	case *me.Compare:
		return evalMECompare(n, bindings)
	case *me.SetOp:
		return evalMESetOp(n, bindings)
	case *me.SetCompare:
		return evalMESetCompare(n, bindings)
	case *me.BagOp:
		return evalMEBagOp(n, bindings)
	case *me.BagCompare:
		return evalMEBagCompare(n, bindings)
	case *me.Membership:
		return evalMEMembership(n, bindings)

	// === Unary operators ===
	case *me.Negate:
		return evalMENegate(n, bindings)
	case *me.Not:
		return evalMENot(n, bindings)

	// === Collections ===
	case *me.FieldAccess:
		return evalMEFieldAccess(n, bindings)
	case *me.TupleIndex:
		return evalMETupleIndex(n, bindings)
	case *me.RecordUpdate:
		return evalMERecordUpdate(n, bindings)
	case *me.StringIndex:
		return evalMEStringIndex(n, bindings)
	case *me.StringConcat:
		return evalMEStringConcat(n, bindings)
	case *me.TupleConcat:
		return evalMETupleConcat(n, bindings)

	// === Control flow ===
	case *me.IfThenElse:
		return evalMEIfThenElse(n, bindings)
	case *me.Case:
		return evalMECase(n, bindings)

	// === Quantifiers ===
	case *me.Quantifier:
		return evalMEQuantifier(n, bindings)
	case *me.SetFilter:
		return evalMESetFilter(n, bindings)

	// === Calls ===
	case *me.BuiltinCall:
		return evalMEBuiltinCall(n, bindings)
	case *me.GlobalCall:
		return evalMEGlobalCall(n, bindings)
	case *me.ActionCall:
		return evalMEActionCall(n, bindings)

	default:
		return NewEvalError("unknown model expression type: %T", node)
	}
}

// nativeBoolToBoolean converts a Go bool to the package Boolean constants.
func nativeBoolToBoolean(value bool) *object.Boolean {
	if value {
		return TRUE
	}
	return FALSE
}
