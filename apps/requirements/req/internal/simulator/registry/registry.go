// Package registry provides a registry for custom TLA+ definitions with scoped resolution.
//
// The registry supports three types of functions:
//   - Class functions: Live at Domain!Subdomain!Class scope
//   - Global functions: Start with underscore (_FuncName)
//   - Built-in functions: Use _Module!FuncName syntax (handled elsewhere)
//
// Definitions are stored with fully-qualified keys and support dependency tracking
// for incremental rebuilds when types change.
package registry

import (
	"fmt"
	"strings"
	"sync"

	"github.com/glemzurg/go-tlaplus/internal/simulator/ast"
	"github.com/glemzurg/go-tlaplus/internal/simulator/typechecker"
	"github.com/glemzurg/go-tlaplus/internal/simulator/types"
)

// DefinitionKey is a fully qualified function name.
// Examples:
//   - "_IsoCurrency" (global function)
//   - "DomainA!SubdomainB!ClassC!New" (class function)
type DefinitionKey string

// IsGlobal returns true if this is a global function key (starts with _).
func (k DefinitionKey) IsGlobal() bool {
	return len(k) > 0 && k[0] == '_'
}

// LocalName returns the function name without scope prefix.
func (k DefinitionKey) LocalName() string {
	s := string(k)
	if k.IsGlobal() {
		// _FuncName -> FuncName
		return s[1:]
	}
	// Domain!Subdomain!Class!FuncName -> FuncName
	parts := strings.Split(s, "!")
	return parts[len(parts)-1]
}

// DefinitionKind distinguishes between class functions and global functions.
type DefinitionKind int

const (
	// KindClassFunction is a function that lives at a class scope.
	KindClassFunction DefinitionKind = iota
	// KindGlobalFunction is a function that starts with underscore.
	KindGlobalFunction
)

func (k DefinitionKind) String() string {
	switch k {
	case KindClassFunction:
		return "class"
	case KindGlobalFunction:
		return "global"
	default:
		return "unknown"
	}
}

// ScopePath represents a hierarchical scope identifier.
// For class functions, always has exactly 3 parts: Domain!Subdomain!Class
// For global functions, this is empty.
type ScopePath string

// ParseScopePath creates a ScopePath from domain, subdomain, and class names.
// All three must be provided for class scope, or all empty for global scope.
func ParseScopePath(domain, subdomain, class string) (ScopePath, error) {
	allEmpty := domain == "" && subdomain == "" && class == ""
	allSet := domain != "" && subdomain != "" && class != ""

	if !allEmpty && !allSet {
		return "", fmt.Errorf("scope path must have all three components (domain, subdomain, class) or none")
	}

	if allEmpty {
		return "", nil
	}

	return ScopePath(domain + "!" + subdomain + "!" + class), nil
}

// Parts returns the scope components [domain, subdomain, class].
// Returns nil for empty/global scope.
func (s ScopePath) Parts() []string {
	if s == "" {
		return nil
	}
	return strings.Split(string(s), "!")
}

// Domain returns the domain component, or empty string if global scope.
func (s ScopePath) Domain() string {
	parts := s.Parts()
	if len(parts) < 1 {
		return ""
	}
	return parts[0]
}

// Subdomain returns the subdomain component, or empty string if global scope.
func (s ScopePath) Subdomain() string {
	parts := s.Parts()
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}

// Class returns the class component, or empty string if global scope.
func (s ScopePath) Class() string {
	parts := s.Parts()
	if len(parts) < 3 {
		return ""
	}
	return parts[2]
}

// Parameter represents a named parameter with its type.
type Parameter struct {
	Name string
	Type types.Type
}

// Definition represents a registered custom TLA+ block.
type Definition struct {
	Key        DefinitionKey
	Kind       DefinitionKind
	Scope      ScopePath              // Full path (empty for global)
	LocalName  string                 // Just "Func"
	Body       ast.Expression         // Untyped AST
	Parameters []Parameter            // Ordered list of typed parameters (can be empty)
	ReturnType types.Type             // Inferred return type (nil until type-checked)
	TypedBody  *typechecker.TypedNode // Cached typed AST (nil = needs recheck)
	Version    uint64                 // Incremented on modification
	DependsOn  []DefinitionKey        // Definitions this depends on
	DependedBy []DefinitionKey        // Definitions that depend on this
}

// NeedsTypeCheck returns true if the definition needs type checking.
func (d *Definition) NeedsTypeCheck() bool {
	return d.TypedBody == nil
}

// Registry manages all custom TLA+ definitions with scoped resolution.
type Registry struct {
	mu          sync.RWMutex
	definitions map[DefinitionKey]*Definition
	globals     map[string]*Definition // Quick lookup for global functions by local name
	version     uint64                 // Global version counter
}

// NewRegistry creates an empty definition registry.
func NewRegistry() *Registry {
	return &Registry{
		definitions: make(map[DefinitionKey]*Definition),
		globals:     make(map[string]*Definition),
	}
}

// RegisterClassFunction registers a class function definition.
// The key will be Domain!Subdomain!Class!name.
func (r *Registry) RegisterClassFunction(
	domain, subdomain, class, name string,
	body ast.Expression,
	params []Parameter,
) (*Definition, error) {
	scope, err := ParseScopePath(domain, subdomain, class)
	if err != nil {
		return nil, err
	}
	if scope == "" {
		return nil, fmt.Errorf("class function requires domain, subdomain, and class")
	}

	key := DefinitionKey(string(scope) + "!" + name)

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicate
	if _, exists := r.definitions[key]; exists {
		return nil, fmt.Errorf("definition already exists: %s", key)
	}

	def := &Definition{
		Key:        key,
		Kind:       KindClassFunction,
		Scope:      scope,
		LocalName:  name,
		Body:       body,
		Parameters: params,
		Version:    1,
	}

	r.definitions[key] = def
	r.version++

	return def, nil
}

// RegisterGlobalFunction registers a global function definition.
// The key will be _name.
func (r *Registry) RegisterGlobalFunction(
	name string,
	body ast.Expression,
	params []Parameter,
) (*Definition, error) {
	if name == "" {
		return nil, fmt.Errorf("global function name cannot be empty")
	}

	key := DefinitionKey("_" + name)

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicate
	if _, exists := r.definitions[key]; exists {
		return nil, fmt.Errorf("definition already exists: %s", key)
	}

	def := &Definition{
		Key:        key,
		Kind:       KindGlobalFunction,
		Scope:      "",
		LocalName:  name,
		Body:       body,
		Parameters: params,
		Version:    1,
	}

	r.definitions[key] = def
	r.globals[name] = def
	r.version++

	return def, nil
}

// Get retrieves a definition by its fully-qualified key.
func (r *Registry) Get(key DefinitionKey) (*Definition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	def, ok := r.definitions[key]
	return def, ok
}

// GetGlobal retrieves a global function by its local name (without underscore prefix).
func (r *Registry) GetGlobal(localName string) (*Definition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	def, ok := r.globals[localName]
	return def, ok
}

// Update updates an existing definition's body and parameters.
// This clears the typed body and increments the version, requiring re-type-checking.
func (r *Registry) Update(key DefinitionKey, body ast.Expression, params []Parameter) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	def, ok := r.definitions[key]
	if !ok {
		return fmt.Errorf("definition not found: %s", key)
	}

	def.Body = body
	def.Parameters = params
	def.TypedBody = nil
	def.ReturnType = nil
	def.Version++
	r.version++

	return nil
}

// Delete removes a definition from the registry.
// Returns an error if the definition doesn't exist.
func (r *Registry) Delete(key DefinitionKey) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	def, ok := r.definitions[key]
	if !ok {
		return fmt.Errorf("definition not found: %s", key)
	}

	delete(r.definitions, key)
	if def.Kind == KindGlobalFunction {
		delete(r.globals, def.LocalName)
	}
	r.version++

	return nil
}

// All returns all definitions in the registry.
func (r *Registry) All() []*Definition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*Definition, 0, len(r.definitions))
	for _, def := range r.definitions {
		result = append(result, def)
	}
	return result
}

// Version returns the current global version counter.
// This increments on any modification to the registry.
func (r *Registry) Version() uint64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.version
}

// Count returns the number of definitions in the registry.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.definitions)
}
