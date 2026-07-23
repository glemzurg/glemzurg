// Package schema is the sole home of model facts for one simulation run.
//
// Data-flow gate:
//
//	*core.Model ──New──► *Schema ──► instance.State, engine, checkers, …
//
// After construction, the running simulator must not carry a separate *core.Model
// for the same run. Components use Schema methods (Class, Association, ForEach*,
// NamedSets, …) and values built from schema (catalog, checkers, eval context).
// Lookups return model tree types ([model_class.Class], [model_class.Association]),
// not parallel schema DTOs. The owned model pointer is private.
//
// [instance.State] holds *Schema for static lookups; mutable world state stays in
// instance. Do not mutate the model after New.
package schema
