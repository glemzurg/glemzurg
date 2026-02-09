package evaluator

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/ast"
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
func NewEvalError(format string, args ...interface{}) *EvalResult {
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

// Eval evaluates a TLA+ AST node and returns the result.
// Valid root nodes are:
// - *ast.Assignment: Primes a binding, returns EvalResult with PrimedBindings populated
// - Logic nodes: Returns Boolean for assertion checks
// Other nodes can be evaluated but are typically sub-expressions.
func Eval(node ast.Node, bindings *Bindings) *EvalResult {
	switch n := node.(type) {

	// === Root Nodes ===
	case *ast.Assignment:
		return evalAssignment(n, bindings)

	// === Literals ===
	case *ast.NumberLiteral:
		return evalNumberLiteral(n)
	case *ast.NumericPrefixExpression:
		return evalNumericPrefixExpression(n, bindings)
	case *ast.FractionExpr:
		return evalFractionExpr(n, bindings)
	case *ast.ParenExpr:
		return evalParenExpr(n, bindings)
	case *ast.StringLiteral:
		return evalStringLiteral(n)
	case *ast.BooleanLiteral:
		return evalBooleanLiteral(n)
	case *ast.TupleLiteral:
		return evalTupleLiteral(n, bindings)
	case *ast.SetLiteralInt:
		return evalSetLiteralInt(n)
	case *ast.SetLiteralEnum:
		return evalSetLiteralEnum(n)
	case *ast.SetLiteral:
		return evalSetLiteral(n, bindings)
	case *ast.SetRange:
		return evalSetRange(n, bindings)
	case *ast.SetRangeExpr:
		return evalSetRangeExpr(n, bindings)
	case *ast.SetConstant:
		return evalSetConstant(n)
	case *ast.RecordInstance:
		return evalRecordInstance(n, bindings)

	// === Identifiers ===
	case *ast.Identifier:
		return evalIdentifier(n, bindings)
	case *ast.FieldIdentifier:
		return evalFieldIdentifier(n, bindings)
	case *ast.ExistingValue:
		return evalExistingValue(bindings)
	case *ast.Primed:
		return evalPrimed(n, bindings)

	// === Arithmetic ===
	case *ast.RealInfixExpression:
		return evalRealInfix(n, bindings)

	// === Logic ===
	case *ast.LogicInfixExpression:
		return evalLogicInfix(n, bindings)
	case *ast.LogicPrefixExpression:
		return evalLogicPrefix(n, bindings)
	case *ast.LogicRealComparison:
		return evalLogicRealComparison(n, bindings)
	case *ast.LogicMembership:
		return evalLogicMembership(n, bindings)
	case *ast.LogicBoundQuantifier:
		return evalLogicBoundQuantifier(n, bindings)
	case *ast.LogicInfixSet:
		return evalLogicInfixSet(n, bindings)
	case *ast.LogicInfixBag:
		return evalLogicInfixBag(n, bindings)
	case *ast.LogicEquality:
		return evalLogicEquality(n, bindings)

	// === Sets ===
	case *ast.SetInfix:
		return evalSetInfix(n, bindings)
	case *ast.SetConditional:
		return evalSetConditional(n, bindings)

	// === Bags ===
	case *ast.BagInfix:
		return evalBagInfix(n, bindings)

	// === Tuples/Sequences ===
	case *ast.ExpressionTupleIndex:
		return evalTupleIndex(n, bindings)
	case *ast.TupleInfixExpression:
		return evalTupleInfix(n, bindings)

	// === Builtins ===
	case *ast.BuiltinCall:
		return evalBuiltinCall(n, bindings)

	// === Records ===
	case *ast.RecordAltered:
		return evalRecordAltered(n, bindings)

	// === Control Flow ===
	case *ast.ExpressionIfElse:
		return evalIfElse(n, bindings)
	case *ast.ExpressionCase:
		return evalCase(n, bindings)

	// === Calls ===
	case *ast.CallExpression:
		return evalCallExpression(n, bindings)
	case *ast.FunctionCall:
		return evalFunctionCall(n, bindings)

	// === String operations ===
	case *ast.StringIndex:
		return evalStringIndex(n, bindings)
	case *ast.StringInfixExpression:
		return evalStringInfix(n, bindings)

	default:
		return NewEvalError("unknown node type: %T", node)
	}
}

// nativeBoolToBoolean converts a Go bool to the package Boolean constants.
func nativeBoolToBoolean(value bool) *object.Boolean {
	if value {
		return TRUE
	}
	return FALSE
}

// evalBuiltinCall evaluates a builtin function call.
func evalBuiltinCall(node *ast.BuiltinCall, bindings *Bindings) *EvalResult {
	// Evaluate all arguments
	args := make([]object.Object, len(node.Args))
	for i, argExpr := range node.Args {
		result := Eval(argExpr, bindings)
		if result.IsError() {
			return result
		}
		args[i] = result.Value
	}

	// Look up and call the builtin
	fn, ok := LookupBuiltin(node.Name)
	if !ok {
		return NewEvalError("unknown builtin: %s", node.Name)
	}
	return fn(args)
}
