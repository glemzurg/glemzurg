package registry

import (
	"fmt"

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

// ScopeContext tracks the current scope during evaluation.
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

// AtClassScope creates a new scope context at the class level of the given definition.
func (sc *ScopeContext) AtClassScope(def *Definition) *ScopeContext {
	if def.Kind != KindClassFunction {
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
