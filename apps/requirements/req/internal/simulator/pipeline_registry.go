package simulator

import (
	me "github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_expression"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/registry"
)

// RegistryPipeline manages a registry of definitions and handles scoped evaluation.
type RegistryPipeline struct {
	registry       *registry.Registry
	runtimeAdapter *registry.RuntimeAdapter
	relationCtx    *evaluator.RelationContext // Optional relation context for association traversal
}

// NewRegistryPipeline creates a new pipeline with registry support.
func NewRegistryPipeline() *RegistryPipeline {
	reg := registry.NewRegistry()
	runtimeAdapter := registry.NewRuntimeAdapter(reg)

	return &RegistryPipeline{
		registry:       reg,
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
	body me.Expression,
	params []registry.Parameter,
) (*registry.Definition, error) {
	return p.registry.RegisterClassFunction(domain, subdomain, class, name, body, params)
}

// RegisterGlobalFunction registers a global function definition.
func (p *RegistryPipeline) RegisterGlobalFunction(
	name string,
	body me.Expression,
	params []registry.Parameter,
) (*registry.Definition, error) {
	return p.registry.RegisterGlobalFunction(name, body, params)
}

// RebuildDefinitions re-validates definitions after changes.
func (p *RegistryPipeline) RebuildDefinitions(changedKeys []registry.DefinitionKey) error {
	invalidated := p.registry.InvalidateMultiple(changedKeys)

	validateFn := func(def *registry.Definition, scopeCtx *registry.ScopeContext) error {
		// IR bodies are already validated during lowering; nothing to do.
		return nil
	}

	return p.registry.Rebuild(registry.IncrementalRebuild, invalidated, validateFn)
}

// RebuildAll re-validates all definitions from scratch.
func (p *RegistryPipeline) RebuildAll() error {
	validateFn := func(def *registry.Definition, scopeCtx *registry.ScopeContext) error {
		// IR bodies are already validated during lowering; nothing to do.
		return nil
	}

	return p.registry.Rebuild(registry.FullRebuild, nil, validateFn)
}

// EvalIR evaluates an IR expression with registry context at the given scope.
func (p *RegistryPipeline) EvalIR(expr me.Expression, bindings *evaluator.Bindings, scopeLevel int, domain, subdomain, class string) *evaluator.EvalResult {
	ctx := &evaluator.EvalContext{
		IRRegistry: p.runtimeAdapter,
		ScopeLevel: scopeLevel,
		Domain:     domain,
		Subdomain:  subdomain,
		Class:      class,
	}

	if p.relationCtx != nil && bindings.RelationContext() == nil {
		bindings.SetRelationContext(p.relationCtx)
	}

	return evaluator.EvalWithContext(expr, bindings, ctx)
}

// EvalIRAtGlobalScope evaluates an IR expression at global scope.
func (p *RegistryPipeline) EvalIRAtGlobalScope(expr me.Expression, bindings *evaluator.Bindings) *evaluator.EvalResult {
	return p.EvalIR(expr, bindings, int(registry.ScopeLevelGlobal), "", "", "")
}

// EvalIRAtClassScope evaluates an IR expression at class scope.
func (p *RegistryPipeline) EvalIRAtClassScope(expr me.Expression, bindings *evaluator.Bindings, domain, subdomain, class string) *evaluator.EvalResult {
	return p.EvalIR(expr, bindings, int(registry.ScopeLevelClass), domain, subdomain, class)
}
