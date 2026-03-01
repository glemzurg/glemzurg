package registry

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/ast"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// ScopeLevel represents the depth of the current binding scope.
type ScopeLevel int

const (
	// ScopeLevelGlobal is above all domains (can call with full Domain!Subdomain!Class!Func path).
	ScopeLevelGlobal ScopeLevel = 0
	// ScopeLevelDomain is at a domain (can call with Subdomain!Class!Func path).
	ScopeLevelDomain ScopeLevel = 1
	// ScopeLevelSubdomain is at a subdomain (can call with Class!Func path).
	ScopeLevelSubdomain ScopeLevel = 2
	// ScopeLevelClass is at a class (can call with just Func path).
	ScopeLevelClass ScopeLevel = 3
)

func (l ScopeLevel) String() string {
	switch l {
	case ScopeLevelGlobal:
		return "global"
	case ScopeLevelDomain:
		return "domain"
	case ScopeLevelSubdomain:
		return "subdomain"
	case ScopeLevelClass:
		return "class"
	default:
		return fmt.Sprintf("unknown(%d)", l)
	}
}

// ScopeContext tracks the current scope during type checking and evaluation.
// This enables deterministic resolution of definition names based on call syntax.
type ScopeContext struct {
	Level      ScopeLevel     // Current scope depth
	Domain     string         // Set if Level >= ScopeLevelDomain
	Subdomain  string         // Set if Level >= ScopeLevelSubdomain
	Class      string         // Set if Level >= ScopeLevelClass
	SelfRecord *object.Record // For class methods (optional)
	Registry   *Registry      // Reference to the definition registry
}

// NewGlobalScopeContext creates a scope context at global level.
func NewGlobalScopeContext(registry *Registry) *ScopeContext {
	return &ScopeContext{
		Level:    ScopeLevelGlobal,
		Registry: registry,
	}
}

// NewDomainScopeContext creates a scope context at domain level.
func NewDomainScopeContext(registry *Registry, domain string) *ScopeContext {
	return &ScopeContext{
		Level:    ScopeLevelDomain,
		Domain:   domain,
		Registry: registry,
	}
}

// NewSubdomainScopeContext creates a scope context at subdomain level.
func NewSubdomainScopeContext(registry *Registry, domain, subdomain string) *ScopeContext {
	return &ScopeContext{
		Level:     ScopeLevelSubdomain,
		Domain:    domain,
		Subdomain: subdomain,
		Registry:  registry,
	}
}

// NewClassScopeContext creates a scope context at class level.
func NewClassScopeContext(registry *Registry, domain, subdomain, class string) *ScopeContext {
	return &ScopeContext{
		Level:     ScopeLevelClass,
		Domain:    domain,
		Subdomain: subdomain,
		Class:     class,
		Registry:  registry,
	}
}

// WithSelf creates a copy of this context with the self record set.
func (sc *ScopeContext) WithSelf(self *object.Record) *ScopeContext {
	return &ScopeContext{
		Level:      sc.Level,
		Domain:     sc.Domain,
		Subdomain:  sc.Subdomain,
		Class:      sc.Class,
		SelfRecord: self,
		Registry:   sc.Registry,
	}
}

// ScopePath returns the current scope as a ScopePath.
// Returns empty string if not at class level.
func (sc *ScopeContext) ScopePath() ScopePath {
	if sc.Level < ScopeLevelClass {
		return ""
	}
	return ScopePath(sc.Domain + "!" + sc.Subdomain + "!" + sc.Class)
}

// ResolveCall resolves a CallExpression to a fully-qualified DefinitionKey.
//
// Resolution is deterministic based on call structure and current scope level:
//
//	FuncName()                        -> {scope}!FuncName (requires class scope)
//	Class!FuncName()                  -> {scope}!Class!FuncName (requires subdomain scope)
//	Subdomain!Class!FuncName()        -> {scope}!Subdomain!Class!FuncName (requires domain scope)
//	Domain!Subdomain!Class!FuncName() -> Domain!Subdomain!Class!FuncName (requires global scope)
//	_FuncName()                       -> _FuncName (global function, any scope)
//
// Returns an error if the scope level doesn't match the call structure.
func (sc *ScopeContext) ResolveCall(call *ast.CallExpression) (DefinitionKey, *Definition, error) {
	// Handle global function calls (_FuncName)
	if call.ModelScope {
		return sc.resolveGlobalCall(call)
	}

	// Handle class function calls based on how many scope parts are specified
	return sc.resolveClassCall(call)
}

// resolveGlobalCall handles calls to global functions (_FuncName).
func (sc *ScopeContext) resolveGlobalCall(call *ast.CallExpression) (DefinitionKey, *Definition, error) {
	if call.Domain != nil || call.Subdomain != nil || call.Class != nil {
		// This looks like a builtin call (_Module!Func), not a global function
		// Builtins are handled elsewhere, return an appropriate error
		return "", nil, fmt.Errorf("builtin function calls (_Module!Func) are handled by the builtin system, not the registry")
	}

	localName := call.FunctionName.Value
	key := DefinitionKey("_" + localName)

	def, ok := sc.Registry.GetGlobal(localName)
	if !ok {
		return "", nil, fmt.Errorf("undefined global function: %s", key)
	}

	return key, def, nil
}

// resolveClassCall handles calls to class functions with various scope depths.
func (sc *ScopeContext) resolveClassCall(call *ast.CallExpression) (DefinitionKey, *Definition, error) {
	// Count how many scope parts are provided in the call
	callDepth := sc.countCallDepth(call)

	// Determine required scope level based on call depth
	// callDepth 0: FuncName() - need class scope (level 3)
	// callDepth 1: Class!FuncName() - need subdomain scope (level 2)
	// callDepth 2: Subdomain!Class!FuncName() - need domain scope (level 1)
	// callDepth 3: Domain!Subdomain!Class!FuncName() - need global scope (level 0)
	requiredLevel := ScopeLevelClass - ScopeLevel(callDepth)

	if sc.Level != requiredLevel {
		return "", nil, fmt.Errorf(
			"scope mismatch: call with %d scope parts requires %s scope, but current scope is %s",
			callDepth, requiredLevel, sc.Level,
		)
	}

	// Build the fully qualified key
	key := sc.buildClassFunctionKey(call, callDepth)

	def, ok := sc.Registry.Get(key)
	if !ok {
		return "", nil, fmt.Errorf("undefined class function: %s", key)
	}

	return key, def, nil
}

// countCallDepth counts how many scope parts are specified in the call.
// 0 = FuncName(), 1 = Class!FuncName(), 2 = Subdomain!Class!FuncName(), 3 = Domain!Subdomain!Class!FuncName()
func (sc *ScopeContext) countCallDepth(call *ast.CallExpression) int {
	count := 0
	if call.Domain != nil {
		count++
	}
	if call.Subdomain != nil {
		count++
	}
	if call.Class != nil {
		count++
	}
	return count
}

// buildClassFunctionKey builds the fully-qualified key for a class function call.
func (sc *ScopeContext) buildClassFunctionKey(call *ast.CallExpression, callDepth int) DefinitionKey {
	var parts []string

	switch callDepth {
	case 0:
		// FuncName() - prepend full scope from context
		parts = []string{sc.Domain, sc.Subdomain, sc.Class, call.FunctionName.Value}
	case 1:
		// Class!FuncName() - prepend domain and subdomain from context
		parts = []string{sc.Domain, sc.Subdomain, call.Class.Value, call.FunctionName.Value}
	case 2:
		// Subdomain!Class!FuncName() - prepend domain from context
		parts = []string{sc.Domain, call.Subdomain.Value, call.Class.Value, call.FunctionName.Value}
	case 3:
		// Domain!Subdomain!Class!FuncName() - use all parts from call
		parts = []string{call.Domain.Value, call.Subdomain.Value, call.Class.Value, call.FunctionName.Value}
	}

	// Join with ! separator
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += "!"
		}
		result += part
	}

	return DefinitionKey(result)
}

// AtClassScope creates a new scope context at the class level of the given definition.
// This is used when entering a class function to execute its body.
func (sc *ScopeContext) AtClassScope(def *Definition) *ScopeContext {
	if def.Kind != KindClassFunction {
		// Global functions don't change the scope context meaningfully,
		// but we keep the registry reference
		return &ScopeContext{
			Level:     sc.Level,
			Domain:    sc.Domain,
			Subdomain: sc.Subdomain,
			Class:     sc.Class,
			Registry:  sc.Registry,
		}
	}

	parts := def.Scope.Parts()
	if len(parts) != 3 {
		// Invalid scope, return unchanged
		return sc
	}

	return &ScopeContext{
		Level:     ScopeLevelClass,
		Domain:    parts[0],
		Subdomain: parts[1],
		Class:     parts[2],
		Registry:  sc.Registry,
	}
}
