// Package instance holds all mutable state for one simulation run.
//
// It owns class instances, binary association links, association-class host
// rows, state-machine positions, and the identity mappings needed to keep that
// world consistent. Static model facts live in [schema.Schema] (the sole model
// home for the run), passed into [NewState]; instance never mutates schema.
//
// Action execution, expression evaluation, model loading, and TLA binding
// construction live in other packages and call into this one.
//
// Callers iterate and query through protocol methods (ForEach*, Lookup*,
// ProjectToRelationContext, Snapshot) rather than dumping the full instance map.
// Storage maps, locks, and ID counters are unexported implementation details.
package instance
