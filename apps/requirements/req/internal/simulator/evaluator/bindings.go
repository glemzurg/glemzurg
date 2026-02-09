package evaluator

import (
	"github.com/glemzurg/go-tlaplus/internal/simulator/object"
)

// Namespace indicates the type of binding scope.
type Namespace string

const (
	NamespaceGlobal Namespace = "global" // Global state variables
	NamespaceLocal  Namespace = "local"  // Local scope (quantifier variables, etc.)
	NamespaceReturn Namespace = "return" // Return values from block calls
)

// BindingEntry represents a single binding with metadata.
type BindingEntry struct {
	Value       object.Object // The current value (unprimed)
	PrimedValue object.Object // The next-state value (primed), nil if not primed
	Namespace   Namespace     // Which namespace this belongs to
	Primed      bool          // Whether this entry has been primed (x' = ...)
}

// Bindings manages variable bindings for TLA+ evaluation.
// It supports:
// - Hierarchical scoping (outer bindings)
// - "self" record for model_class scope
// - Tracking which variables have been primed
// - Namespace categorization (global vs return)
// - Relation context for association traversal
type Bindings struct {
	store map[string]*BindingEntry // Variable name to entry
	outer *Bindings                // Parent scope (nil for root)
	self  *object.Record           // Optional "self" record for model_class methods

	// selfClassKey is the identity.Key.String() of the class for self.
	// Empty string if not in a class scope.
	selfClassKey string

	// relationCtx provides access to association metadata and link state.
	// Shared across scopes; nil if no relations are configured.
	relationCtx *RelationContext

	// existingValue is set when evaluating EXCEPT expressions
	// to provide the @ reference to the current field value.
	existingValue object.Object
}

// NewBindings creates a new root bindings context.
func NewBindings() *Bindings {
	return &Bindings{
		store: make(map[string]*BindingEntry),
	}
}

// NewEnclosedBindings creates a child scope that inherits from outer.
func NewEnclosedBindings(outer *Bindings) *Bindings {
	b := NewBindings()
	b.outer = outer
	// Inherit from outer if present
	if outer != nil {
		b.self = outer.self
		b.selfClassKey = outer.selfClassKey
		b.relationCtx = outer.relationCtx
	}
	return b
}

// WithSelf creates a new scope with the given self record.
func (b *Bindings) WithSelf(self *object.Record) *Bindings {
	child := NewEnclosedBindings(b)
	child.self = self
	return child
}

// WithSelfAndClass creates a new scope with the given self record and class key.
// The classKey should be the identity.Key.String() for the class.
func (b *Bindings) WithSelfAndClass(self *object.Record, classKey string) *Bindings {
	child := NewEnclosedBindings(b)
	child.self = self
	child.selfClassKey = classKey
	return child
}

// Self returns the current "self" record, or nil if not in model_class scope.
func (b *Bindings) Self() *object.Record {
	return b.self
}

// SelfClassKey returns the class key for the self record.
// Returns empty string if not in a class scope.
func (b *Bindings) SelfClassKey() string {
	return b.selfClassKey
}

// SetSelfClassKey sets the class key for the current self record.
func (b *Bindings) SetSelfClassKey(classKey string) {
	b.selfClassKey = classKey
}

// RelationContext returns the relation context, searching up the scope chain.
// Returns nil if no relation context is configured.
func (b *Bindings) RelationContext() *RelationContext {
	if b.relationCtx != nil {
		return b.relationCtx
	}
	if b.outer != nil {
		return b.outer.RelationContext()
	}
	return nil
}

// SetRelationContext sets the relation context for this scope.
func (b *Bindings) SetRelationContext(ctx *RelationContext) {
	b.relationCtx = ctx
}

// Get retrieves a binding by name, searching up the scope chain.
// Returns the value, namespace, and whether it was found.
func (b *Bindings) Get(name string) (object.Object, Namespace, bool) {
	if entry, ok := b.store[name]; ok {
		return entry.Value, entry.Namespace, true
	}
	if b.outer != nil {
		return b.outer.Get(name)
	}
	return nil, "", false
}

// GetValue retrieves just the value by name (convenience method).
func (b *Bindings) GetValue(name string) (object.Object, bool) {
	val, _, found := b.Get(name)
	return val, found
}

// Set creates or updates a binding in the current scope.
func (b *Bindings) Set(name string, value object.Object, ns Namespace) {
	b.store[name] = &BindingEntry{
		Value:     value.Clone(), // Always clone to ensure immutability
		Namespace: ns,
		Primed:    false,
	}
}

// SetPrimed sets a variable as primed (x' = ...).
// If the variable doesn't exist, it creates it in the global namespace.
// This marks the entry as "altered" for tracking state changes.
// The primed value is stored separately from the current value, so:
// - x returns the current (unprimed) value
// - x' returns the primed (next-state) value
func (b *Bindings) SetPrimed(name string, value object.Object) {
	// Look for existing entry
	if entry, ok := b.store[name]; ok {
		// Update existing entry with primed value
		entry.PrimedValue = value.Clone()
		entry.Primed = true
		return
	}

	// Look in outer scope for namespace
	ns := NamespaceGlobal
	if _, existingNs, found := b.Get(name); found {
		ns = existingNs
	}

	// Create new entry - when priming without current value, both are set
	b.store[name] = &BindingEntry{
		Value:       value.Clone(), // Also set current value for new entries
		PrimedValue: value.Clone(),
		Namespace:   ns,
		Primed:      true,
	}
}

// IsPrimed checks if a variable has been primed in this scope.
func (b *Bindings) IsPrimed(name string) bool {
	if entry, ok := b.store[name]; ok {
		return entry.Primed
	}
	return false
}

// GetPrimedValue retrieves the primed (next-state) value for a variable.
// Returns the primed value and true if the variable has been primed.
// Returns nil and false if not primed or not found.
func (b *Bindings) GetPrimedValue(name string) (object.Object, bool) {
	if entry, ok := b.store[name]; ok && entry.Primed {
		return entry.PrimedValue, true
	}
	if b.outer != nil {
		return b.outer.GetPrimedValue(name)
	}
	return nil, false
}

// GetPrimedNames returns all variable names that have been primed.
func (b *Bindings) GetPrimedNames() []string {
	var names []string
	for name, entry := range b.store {
		if entry.Primed {
			names = append(names, name)
		}
	}
	return names
}

// GetPrimedBindings returns a map of all primed variables and their new values.
func (b *Bindings) GetPrimedBindings() map[string]object.Object {
	result := make(map[string]object.Object)
	for name, entry := range b.store {
		if entry.Primed {
			result[name] = entry.PrimedValue
		}
	}
	return result
}

// GetByNamespace returns all bindings in a specific namespace.
func (b *Bindings) GetByNamespace(ns Namespace) map[string]object.Object {
	result := make(map[string]object.Object)
	for name, entry := range b.store {
		if entry.Namespace == ns {
			result[name] = entry.Value
		}
	}
	return result
}

// SetExistingValue sets the @ reference for EXCEPT expressions.
func (b *Bindings) SetExistingValue(value object.Object) {
	b.existingValue = value
}

// GetExistingValue returns the @ reference, or nil if not set.
func (b *Bindings) GetExistingValue() object.Object {
	if b.existingValue != nil {
		return b.existingValue
	}
	if b.outer != nil {
		return b.outer.GetExistingValue()
	}
	return nil
}

// Clone creates a deep copy of the bindings (without outer reference).
func (b *Bindings) Clone() *Bindings {
	clone := NewBindings()
	for name, entry := range b.store {
		cloneEntry := &BindingEntry{
			Value:     entry.Value.Clone(),
			Namespace: entry.Namespace,
			Primed:    entry.Primed,
		}
		if entry.PrimedValue != nil {
			cloneEntry.PrimedValue = entry.PrimedValue.Clone()
		}
		clone.store[name] = cloneEntry
	}
	if b.self != nil {
		clone.self = b.self.Clone().(*object.Record)
	}
	clone.selfClassKey = b.selfClassKey
	clone.relationCtx = b.relationCtx // Shared reference (not cloned)
	return clone
}

// Names returns all variable names in the current scope.
func (b *Bindings) Names() []string {
	names := make([]string, 0, len(b.store))
	for name := range b.store {
		names = append(names, name)
	}
	return names
}
