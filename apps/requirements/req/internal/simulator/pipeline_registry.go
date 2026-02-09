package simulator

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/registry"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/typechecker"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/types"
)

// RegistryPipeline extends Pipeline with custom definition support.
// It manages a registry of TLA+ definitions and handles scoped resolution.
type RegistryPipeline struct {
	typeChecker    *typechecker.TypeChecker
	registry       *registry.Registry
	tcAdapter      *registry.TypeCheckerAdapter
	runtimeAdapter *registry.RuntimeAdapter
	relationCtx    *evaluator.RelationContext // Optional relation context for association traversal
}

// NewRegistryPipeline creates a new pipeline with registry support.
func NewRegistryPipeline() *RegistryPipeline {
	reg := registry.NewRegistry()
	tc := typechecker.NewTypeChecker()

	// Set up type checker adapter
	tcAdapter := registry.NewTypeCheckerAdapter(reg)
	tc.SetRegistry(tcAdapter)

	// Set up runtime adapter
	runtimeAdapter := registry.NewRuntimeAdapter(reg)

	return &RegistryPipeline{
		typeChecker:    tc,
		registry:       reg,
		tcAdapter:      tcAdapter,
		runtimeAdapter: runtimeAdapter,
	}
}

// Registry returns the underlying registry for direct access.
func (p *RegistryPipeline) Registry() *registry.Registry {
	return p.registry
}

// SetRelationContext sets the relation context for association traversal.
// This must be called before evaluating expressions that access relation fields.
func (p *RegistryPipeline) SetRelationContext(ctx *evaluator.RelationContext) {
	p.relationCtx = ctx
}

// RelationContext returns the current relation context, or nil if not set.
func (p *RegistryPipeline) RelationContext() *evaluator.RelationContext {
	return p.relationCtx
}

// RegisterClassFunction registers a class function definition.
func (p *RegistryPipeline) RegisterClassFunction(
	domain, subdomain, class, name string,
	body ast.Expression,
	params []registry.Parameter,
) (*registry.Definition, error) {
	return p.registry.RegisterClassFunction(domain, subdomain, class, name, body, params)
}

// RegisterGlobalFunction registers a global function definition.
func (p *RegistryPipeline) RegisterGlobalFunction(
	name string,
	body ast.Expression,
	params []registry.Parameter,
) (*registry.Definition, error) {
	return p.registry.RegisterGlobalFunction(name, body, params)
}

// RebuildDefinitions re-type-checks definitions after changes.
// Uses fail-fast approach: all errors are collected before returning.
func (p *RegistryPipeline) RebuildDefinitions(changedKeys []registry.DefinitionKey) error {
	// Invalidate changed definitions and their dependents
	invalidated := p.registry.InvalidateMultiple(changedKeys)

	// Create type check function
	typeCheckFn := func(def *registry.Definition, tc *typechecker.TypeChecker, scopeCtx *registry.ScopeContext) error {
		return p.typeCheckDefinition(def, scopeCtx)
	}

	// Rebuild with incremental strategy
	rebuildErr := p.registry.Rebuild(registry.IncrementalRebuild, invalidated, p.typeChecker, typeCheckFn)
	if rebuildErr != nil {
		return rebuildErr
	}

	return nil
}

// RebuildAll re-type-checks all definitions from scratch.
func (p *RegistryPipeline) RebuildAll() error {
	typeCheckFn := func(def *registry.Definition, tc *typechecker.TypeChecker, scopeCtx *registry.ScopeContext) error {
		return p.typeCheckDefinition(def, scopeCtx)
	}

	rebuildErr := p.registry.Rebuild(registry.FullRebuild, nil, p.typeChecker, typeCheckFn)
	if rebuildErr != nil {
		return rebuildErr
	}

	return nil
}

// typeCheckDefinition type-checks a single definition.
func (p *RegistryPipeline) typeCheckDefinition(def *registry.Definition, scopeCtx *registry.ScopeContext) error {
	// Set up dependency tracking
	depRecorder := registry.NewDependencyRecorder(p.registry)
	p.typeChecker.SetDependencyTracker(depRecorder, string(def.Key))
	defer p.typeChecker.ClearDependencyTracker()

	// Set scope for resolution
	p.typeChecker.SetScope(
		int(scopeCtx.Level),
		scopeCtx.Domain,
		scopeCtx.Subdomain,
		scopeCtx.Class,
	)

	// Create type environment with parameters
	env := p.typeChecker.Env().Extend()
	for _, param := range def.Parameters {
		env.BindMono(param.Name, param.Type)
	}

	// Type check the body
	typed, err := p.typeChecker.Check(def.Body)
	if err != nil {
		return fmt.Errorf("type error: %w", err)
	}

	// Update definition with typed body and return type
	return p.registry.SetTypedBody(def.Key, typed, typed.Type)
}

// TypeCheck performs type checking on an AST node at the given scope.
func (p *RegistryPipeline) TypeCheck(node ast.Expression, scopeLevel int, domain, subdomain, class string) (*typechecker.TypedNode, error) {
	// Set scope for resolution
	p.typeChecker.SetScope(scopeLevel, domain, subdomain, class)

	typed, err := p.typeChecker.Check(node)
	if err != nil {
		return nil, fmt.Errorf("type error: %w", err)
	}
	return typed, nil
}

// TypeCheckAtGlobalScope performs type checking at global scope.
func (p *RegistryPipeline) TypeCheckAtGlobalScope(node ast.Expression) (*typechecker.TypedNode, error) {
	return p.TypeCheck(node, int(registry.ScopeLevelGlobal), "", "", "")
}

// TypeCheckAtClassScope performs type checking at class scope.
func (p *RegistryPipeline) TypeCheckAtClassScope(node ast.Expression, domain, subdomain, class string) (*typechecker.TypedNode, error) {
	return p.TypeCheck(node, int(registry.ScopeLevelClass), domain, subdomain, class)
}

// Eval performs type checking and evaluation of an AST node at the given scope.
func (p *RegistryPipeline) Eval(node ast.Expression, bindings *evaluator.Bindings, scopeLevel int, domain, subdomain, class string) *evaluator.EvalResult {
	// Phase 1: Type check
	typed, err := p.TypeCheck(node, scopeLevel, domain, subdomain, class)
	if err != nil {
		return evaluator.NewEvalError("%s", err.Error())
	}

	// Phase 2: Set up eval context for registry-based calls
	ctx := &evaluator.EvalContext{
		Registry:   p.runtimeAdapter,
		ScopeLevel: scopeLevel,
		Domain:     domain,
		Subdomain:  subdomain,
		Class:      class,
	}

	// Phase 3: Ensure bindings has relation context if available
	if p.relationCtx != nil && bindings.RelationContext() == nil {
		bindings.SetRelationContext(p.relationCtx)
	}

	// Phase 4: Evaluate with context
	return evaluator.EvalTypedWithContext(typed, bindings, ctx)
}

// EvalAtGlobalScope performs type checking and evaluation at global scope.
func (p *RegistryPipeline) EvalAtGlobalScope(node ast.Expression, bindings *evaluator.Bindings) *evaluator.EvalResult {
	return p.Eval(node, bindings, int(registry.ScopeLevelGlobal), "", "", "")
}

// EvalAtClassScope performs type checking and evaluation at class scope.
func (p *RegistryPipeline) EvalAtClassScope(node ast.Expression, bindings *evaluator.Bindings, domain, subdomain, class string) *evaluator.EvalResult {
	return p.Eval(node, bindings, int(registry.ScopeLevelClass), domain, subdomain, class)
}

// DeclareVariable adds a variable with a known type to the type environment.
func (p *RegistryPipeline) DeclareVariable(name string, typ types.Type) {
	p.typeChecker.DeclareVariable(name, typ)
}

// CompileAtScope compiles an expression at the given scope.
type CompiledWithScope struct {
	typed      *typechecker.TypedNode
	scopeLevel int
	domain     string
	subdomain  string
	class      string
	pipeline   *RegistryPipeline
}

// CompileAtScope type-checks an expression and returns a compiled form.
func (p *RegistryPipeline) CompileAtScope(node ast.Expression, scopeLevel int, domain, subdomain, class string) (*CompiledWithScope, error) {
	typed, err := p.TypeCheck(node, scopeLevel, domain, subdomain, class)
	if err != nil {
		return nil, err
	}
	return &CompiledWithScope{
		typed:      typed,
		scopeLevel: scopeLevel,
		domain:     domain,
		subdomain:  subdomain,
		class:      class,
		pipeline:   p,
	}, nil
}

// Eval evaluates a compiled expression with the given bindings.
func (c *CompiledWithScope) Eval(bindings *evaluator.Bindings) *evaluator.EvalResult {
	ctx := &evaluator.EvalContext{
		Registry:   c.pipeline.runtimeAdapter,
		ScopeLevel: c.scopeLevel,
		Domain:     c.domain,
		Subdomain:  c.subdomain,
		Class:      c.class,
	}
	return evaluator.EvalTypedWithContext(c.typed, bindings, ctx)
}

// Type returns the inferred type of the compiled expression.
func (c *CompiledWithScope) Type() types.Type {
	return c.typed.Type
}
