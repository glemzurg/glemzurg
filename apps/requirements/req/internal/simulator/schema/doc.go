// Package schema holds the immutable description of a simulation surface for one run.
//
// It captures non-changing facts used while the simulator executes: which classes
// are in scope, their attributes, and association structure. [instance.State]
// holds a *Schema pointer for lookups and never mutates it.
//
// Schema is built once from the (typically surface-filtered) model before the run
// starts. Mutable run data lives in simulator/instance, not here.
package schema
