// Package instance holds all mutable state for one simulation run.
//
// It owns class instances, binary association links, association-class host
// rows, state-machine positions, and the identity mappings needed to keep that
// world consistent. Action execution, expression evaluation, model loading, and
// TLA binding construction live in other packages and call into this one.
//
// The exported types and methods are the protocol callers may rely on. Storage
// maps, locks, and ID counters are unexported implementation details.
package instance
