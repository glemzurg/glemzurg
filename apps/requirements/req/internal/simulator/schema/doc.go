// Package schema is the sole home of model facts, run scope, and static simulator
// indexes for one simulation run.
//
// Data-flow gate:
//
//	full *core.Model + RunScope ──New──► *Schema ──► instance, engine, checkers, …
//
// After construction, the running simulator must not carry a separate *core.Model
// or include-list as authority for the same run. Schema presents:
//
//   - Keyed lookups: Class / Association / ClassSim / … as (*T, inScope bool, err)
//     with out-of-scope = (nil, false, nil)
//   - Scoped bulk indexes: ScopedAssociations, AssociationsWithUniqueness,
//     AllClassSims, ClassIndexes, DerivedAttributes, extents, boundary edges
//   - Factories: NewEvalContext
//
// Static indexes (association graph, class simulation metadata, attribute/derived
// projections) are built once at New. Runtime Check* and execution stay outside
// this package.
//
// [instance.State] holds *Schema for static lookups; mutable world state stays in
// instance. Do not mutate the model after New.
package schema
