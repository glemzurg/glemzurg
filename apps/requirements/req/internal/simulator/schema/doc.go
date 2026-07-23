// Package schema is the sole home of model facts for one simulation run.
//
// Data-flow gate:
//
//	*core.Model ──New──► *Schema ──► instance.State, engine, checkers, …
//
// After construction, the running simulator must not carry a separate *core.Model
// for the same run. Components either call Schema methods or, during migration,
// [Schema.CoreModel] to build run-local structures — then drop the model pointer.
//
// Lookups return model tree types ([model_class.Class], [model_class.Association],
// …), not parallel schema DTOs.
//
// [instance.State] holds *Schema for static lookups; mutable world state stays in
// instance. Do not mutate the model after New.
package schema
