package instance

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// ExpressionBindings builds evaluator bindings over the live world for invariant checks.
// Implemented by state.BindingsBuilder; defined here so instance checkers do not import state.
type ExpressionBindings interface {
	BuildWithClassInstances(classNameMap map[identity.Key]string) *evaluator.Bindings
	BuildForInstance(inst *Instance) *evaluator.Bindings
	BuildForInstanceWithVariables(inst *Instance, additional map[string]object.Object) *evaluator.Bindings
}
