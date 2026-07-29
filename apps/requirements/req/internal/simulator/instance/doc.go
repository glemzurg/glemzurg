// Package instance holds all mutable state for one simulation run.
//
// It owns class instances, binary association links, association-class host
// rows, state-machine positions, and the identity mappings needed to keep that
// world consistent. Static model facts live in [schema.Schema] (the sole model
// home for the run), passed into [NewState]; instance never mutates schema.
//
// State correctness checks (authored invariants, data types, indexes, multiplicity,
// association structural rules) live here: they examine live world data and ask
// schema only keyed questions about the subject under check.
//
// Liveness obligations are installed once at [NewState] from schema; hit collection
// and [State.CheckLiveness] compare against that contract without further schema walks.
//
// Action execution, expression evaluation, model loading, and TLA binding
// construction live in other packages and call into this one.
//
// Callers iterate and query through protocol methods (ForEach*, Lookup*,
// ProjectToRelationContext, Snapshot) rather than dumping the full instance map.
// Storage maps, locks, and ID counters are unexported implementation details.
package instance
