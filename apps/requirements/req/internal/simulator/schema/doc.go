// Package schema is the sole home of model facts for one simulation run.
//
// Data-flow gate:
//
//	core.Model ──NewFromModel──► *Schema ──► instance.State, engine, checkers, …
//
// After construction, the running simulator must not carry a separate *core.Model
// for the same run. Components either call Schema methods or, during migration,
// [Schema.CoreModel] to build run-local structures — then drop the model pointer.
//
// [instance.State] holds *Schema for static lookups; mutable world state stays in
// instance. Do not mutate the model after NewFromModel.
package schema
