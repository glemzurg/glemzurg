package registry

// AddDependency records that 'from' depends on 'to'.
// This is called during type checking when a call expression is resolved.
func (r *Registry) AddDependency(from, to DefinitionKey) {
	r.mu.Lock()
	defer r.mu.Unlock()

	fromDef, ok := r.definitions[from]
	if !ok {
		return
	}
	toDef, ok := r.definitions[to]
	if !ok {
		return
	}

	// Add to DependsOn list if not already present
	if !containsKey(fromDef.DependsOn, to) {
		fromDef.DependsOn = append(fromDef.DependsOn, to)
	}

	// Add to DependedBy list if not already present
	if !containsKey(toDef.DependedBy, from) {
		toDef.DependedBy = append(toDef.DependedBy, from)
	}
}

// ClearDependencies removes all dependency information for a definition.
// This is called before re-type-checking a definition.
func (r *Registry) ClearDependencies(key DefinitionKey) {
	r.mu.Lock()
	defer r.mu.Unlock()

	def, ok := r.definitions[key]
	if !ok {
		return
	}

	// Remove this definition from the DependedBy lists of its dependencies
	for _, depKey := range def.DependsOn {
		if depDef, ok := r.definitions[depKey]; ok {
			depDef.DependedBy = removeKey(depDef.DependedBy, key)
		}
	}

	// Clear the DependsOn list
	def.DependsOn = nil
}

// GetDependents returns all definitions that directly depend on the given key.
func (r *Registry) GetDependents(key DefinitionKey) []DefinitionKey {
	r.mu.RLock()
	defer r.mu.RUnlock()

	def, ok := r.definitions[key]
	if !ok {
		return nil
	}

	// Return a copy to avoid mutation issues
	result := make([]DefinitionKey, len(def.DependedBy))
	copy(result, def.DependedBy)
	return result
}

// GetDependencies returns all definitions that the given key depends on.
func (r *Registry) GetDependencies(key DefinitionKey) []DefinitionKey {
	r.mu.RLock()
	defer r.mu.RUnlock()

	def, ok := r.definitions[key]
	if !ok {
		return nil
	}

	// Return a copy to avoid mutation issues
	result := make([]DefinitionKey, len(def.DependsOn))
	copy(result, def.DependsOn)
	return result
}

// FindTransitiveDependents returns all definitions that transitively depend on the given key.
// This includes direct dependents and their dependents, recursively.
func (r *Registry) FindTransitiveDependents(key DefinitionKey) []DefinitionKey {
	r.mu.RLock()
	defer r.mu.RUnlock()

	visited := make(map[DefinitionKey]struct{})
	var result []DefinitionKey

	var visit func(k DefinitionKey)
	visit = func(k DefinitionKey) {
		def, ok := r.definitions[k]
		if !ok {
			return
		}

		for _, depKey := range def.DependedBy {
			if _, seen := visited[depKey]; seen {
				continue
			}
			visited[depKey] = struct{}{}
			result = append(result, depKey)
			visit(depKey)
		}
	}

	visit(key)
	return result
}

// InvalidationSet tracks which definitions need re-type-checking.
type InvalidationSet struct {
	Keys     []DefinitionKey
	versions map[DefinitionKey]uint64
}

// NewInvalidationSet creates an empty invalidation set.
func NewInvalidationSet() *InvalidationSet {
	return &InvalidationSet{
		versions: make(map[DefinitionKey]uint64),
	}
}

// Add adds a key to the invalidation set.
func (s *InvalidationSet) Add(key DefinitionKey, version uint64) {
	if _, exists := s.versions[key]; exists {
		return // Already in set
	}
	s.Keys = append(s.Keys, key)
	s.versions[key] = version
}

// Contains returns true if the key is in the invalidation set.
func (s *InvalidationSet) Contains(key DefinitionKey) bool {
	_, exists := s.versions[key]
	return exists
}

// Merge combines another invalidation set into this one.
func (s *InvalidationSet) Merge(other *InvalidationSet) {
	if other == nil {
		return
	}
	for _, key := range other.Keys {
		s.Add(key, other.versions[key])
	}
}

// InvalidateDefinition marks a definition as needing re-type-check.
// Returns the set of all affected definitions (including transitively dependent).
func (r *Registry) InvalidateDefinition(key DefinitionKey) *InvalidationSet {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := NewInvalidationSet()

	def, ok := r.definitions[key]
	if !ok {
		return result
	}

	// Clear typed body and increment version
	def.TypedBody = nil
	def.ReturnType = nil
	def.Version++
	r.version++

	result.Add(key, def.Version)

	// Find all transitive dependents (need to release lock for this)
	// We'll collect them first, then invalidate
	var dependents []DefinitionKey
	var collectDependents func(k DefinitionKey)
	collectDependents = func(k DefinitionKey) {
		d, ok := r.definitions[k]
		if !ok {
			return
		}
		for _, depKey := range d.DependedBy {
			if !result.Contains(depKey) {
				dependents = append(dependents, depKey)
				result.Add(depKey, 0) // Placeholder version
				collectDependents(depKey)
			}
		}
	}
	collectDependents(key)

	// Invalidate all dependents
	for _, depKey := range dependents {
		if depDef, ok := r.definitions[depKey]; ok {
			depDef.TypedBody = nil
			depDef.ReturnType = nil
			depDef.Version++
			result.versions[depKey] = depDef.Version
		}
	}

	return result
}

// InvalidateMultiple marks multiple definitions as needing re-type-check.
// Returns the combined set of all affected definitions.
func (r *Registry) InvalidateMultiple(keys []DefinitionKey) *InvalidationSet {
	result := NewInvalidationSet()
	for _, key := range keys {
		inv := r.InvalidateDefinition(key)
		result.Merge(inv)
	}
	return result
}

// Helper functions

func containsKey(slice []DefinitionKey, key DefinitionKey) bool {
	for _, k := range slice {
		if k == key {
			return true
		}
	}
	return false
}

func removeKey(slice []DefinitionKey, key DefinitionKey) []DefinitionKey {
	result := make([]DefinitionKey, 0, len(slice))
	for _, k := range slice {
		if k != key {
			result = append(result, k)
		}
	}
	return result
}
