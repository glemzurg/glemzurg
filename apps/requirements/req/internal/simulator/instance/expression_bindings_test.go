package instance

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// testBindings is a minimal ExpressionBindings for checker tests (avoids importing state).
type testBindings struct {
	sim    *State
	relCtx *evaluator.RelationContext
}

func newTestBindings(sim *State) *testBindings {
	return &testBindings{
		sim:    sim,
		relCtx: evaluator.NewRelationContext(),
	}
}

// AddAssociation registers association metadata for navigation (test helper).
func (b *testBindings) AddAssociation(
	assocKey identity.Key,
	name string,
	fromClassKey, toClassKey identity.Key,
	fromMult, toMult evaluator.Multiplicity,
) {
	if b.relCtx == nil {
		b.relCtx = evaluator.NewRelationContext()
	}
	b.relCtx.AddAssociation(
		evaluator.AssociationKey(assocKey.String()),
		name,
		fromClassKey.String(),
		toClassKey.String(),
		fromMult,
		toMult,
	)
}

// RelationContext returns the shared relation context for tests that create links.
func (b *testBindings) RelationContext() *evaluator.RelationContext {
	if b.relCtx == nil {
		b.relCtx = evaluator.NewRelationContext()
	}
	return b.relCtx
}

func (b *testBindings) BuildWithClassInstances(classNameMap map[identity.Key]string) *evaluator.Bindings {
	bindings := evaluator.NewBindings()
	if b.relCtx != nil {
		bindings.SetRelationContext(b.relCtx)
	}
	if b.sim == nil {
		return bindings
	}
	for classKey, name := range classNameMap {
		var elems []object.Object
		for _, inst := range b.sim.InstancesByClass(classKey) {
			elems = append(elems, inst.GetAttributes())
		}
		bindings.Set(name, object.NewSetFromElements(elems), evaluator.NamespaceGlobal)
	}
	return bindings
}

func (b *testBindings) BuildForInstance(inst *Instance) *evaluator.Bindings {
	bindings := evaluator.NewBindings()
	if b.relCtx != nil {
		bindings.SetRelationContext(b.relCtx)
	}
	if inst == nil {
		return bindings
	}
	if b.relCtx != nil {
		id := evaluator.ObjectID(inst.GetID())
		b.relCtx.EnsureInstance(id, inst.GetAttributes())
		b.relCtx.RegisterClassKey(id, inst.GetClassKey().String())
	}
	if inst.GetAttributes() != nil {
		return bindings.WithSelfAndClass(inst.GetAttributes(), inst.GetClassKey().String())
	}
	return bindings
}

func (b *testBindings) BuildForInstanceWithVariables(inst *Instance, additional map[string]object.Object) *evaluator.Bindings {
	bindings := b.BuildForInstance(inst)
	for k, v := range additional {
		bindings.Set(k, v, evaluator.NamespaceLocal)
	}
	return bindings
}
