// Package simulator provides the main compilation and execution pipeline
// for TLA+ specifications.
//
// The pipeline follows standard compiler architecture with three phases:
//  1. Parsing: Build untyped AST (handled by parser, not in this package)
//  2. Type Checking: Traverse AST, infer types, produce typed AST
//  3. Evaluation: Execute typed AST with guaranteed type safety
package simulator

import (
	"fmt"

	"github.com/glemzurg/go-tlaplus/internal/simulator/ast"
	"github.com/glemzurg/go-tlaplus/internal/simulator/evaluator"
	"github.com/glemzurg/go-tlaplus/internal/simulator/typechecker"
	"github.com/glemzurg/go-tlaplus/internal/simulator/types"
)

// Pipeline orchestrates type checking and evaluation of TLA+ AST nodes.
type Pipeline struct {
	typeChecker *typechecker.TypeChecker
}

// NewPipeline creates a new compilation/evaluation pipeline.
func NewPipeline() *Pipeline {
	return &Pipeline{
		typeChecker: typechecker.NewTypeChecker(),
	}
}

// TypeCheck performs type checking on an AST node without evaluation.
// Returns the typed AST node with inferred types attached.
func (p *Pipeline) TypeCheck(node ast.Expression) (*typechecker.TypedNode, error) {
	typed, err := p.typeChecker.Check(node)
	if err != nil {
		return nil, fmt.Errorf("type error: %w", err)
	}
	return typed, nil
}

// Eval performs type checking and evaluation of an AST node.
// This is the main entry point for executing TLA+ expressions.
func (p *Pipeline) Eval(node ast.Expression, bindings *evaluator.Bindings) *evaluator.EvalResult {
	// Phase 1: Type check
	typed, err := p.TypeCheck(node)
	if err != nil {
		return evaluator.NewEvalError("%s", err.Error())
	}

	// Phase 2: Evaluate typed AST
	return evaluator.EvalTyped(typed, bindings)
}

// DeclareVariable adds a variable with a known type to the type environment.
// This is used to declare variables before type checking expressions that use them.
func (p *Pipeline) DeclareVariable(name string, typ types.Type) {
	p.typeChecker.DeclareVariable(name, typ)
}

// InferType returns the inferred type of an expression without evaluating it.
func (p *Pipeline) InferType(node ast.Expression) (types.Type, error) {
	typed, err := p.TypeCheck(node)
	if err != nil {
		return nil, err
	}
	return typed.Type, nil
}

// Compile performs type checking and returns a compiled representation
// that can be evaluated multiple times with different bindings.
type CompiledExpr struct {
	typed *typechecker.TypedNode
}

// Compile type-checks an expression and returns a compiled form.
func (p *Pipeline) Compile(node ast.Expression) (*CompiledExpr, error) {
	typed, err := p.TypeCheck(node)
	if err != nil {
		return nil, err
	}
	return &CompiledExpr{typed: typed}, nil
}

// Eval evaluates a compiled expression with the given bindings.
func (c *CompiledExpr) Eval(bindings *evaluator.Bindings) *evaluator.EvalResult {
	return evaluator.EvalTyped(c.typed, bindings)
}

// Type returns the inferred type of the compiled expression.
func (c *CompiledExpr) Type() types.Type {
	return c.typed.Type
}

// Run is a convenience function that creates a pipeline and evaluates an expression.
// For repeated evaluations, use NewPipeline() and Compile() instead.
func Run(node ast.Expression, bindings *evaluator.Bindings) *evaluator.EvalResult {
	pipeline := NewPipeline()
	return pipeline.Eval(node, bindings)
}

// Check is a convenience function that performs type checking only.
// Returns the inferred type or an error.
func Check(node ast.Expression) (types.Type, error) {
	pipeline := NewPipeline()
	return pipeline.InferType(node)
}
