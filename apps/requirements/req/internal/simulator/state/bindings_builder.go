package state

import (
	"fmt"
	"maps"
	"math"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// DerivedAttributeResolver computes derived attribute values for an instance.
// Implementations should evaluate DerivationPolicy expressions on-demand.
type DerivedAttributeResolver interface {
	// ResolveDerived evaluates all derived attributes for the given instance
	// and returns a map of attribute name -> computed value.
	ResolveDerived(instance *ClassInstance) (map[string]object.Object, error)
}

// BindingsBuilder creates evaluator.Bindings from simulation state.
// It adapts the simulation state into the format expected by the evaluator.
type BindingsBuilder struct {
	state *SimulationState

	// relationContext is shared across all bindings created by this builder
	relationCtx *evaluator.RelationContext

	// derivedResolver computes derived attribute values on-demand (optional)
	derivedResolver DerivedAttributeResolver

	// namedSetValues holds pre-evaluated model named sets keyed by nset SubKey.
	namedSetValues map[string]object.Object
}

// NewBindingsBuilder creates a new bindings builder for the given simulation state.
func NewBindingsBuilder(state *SimulationState) *BindingsBuilder {
	return &BindingsBuilder{
		state:       state,
		relationCtx: evaluator.NewRelationContext(),
	}
}

// NewBindingsBuilderWithRelations creates a bindings builder with a pre-configured
// relation context containing association metadata.
func NewBindingsBuilderWithRelations(state *SimulationState, relationCtx *evaluator.RelationContext) *BindingsBuilder {
	return &BindingsBuilder{
		state:       state,
		relationCtx: relationCtx,
	}
}

// NamedSetValues returns pre-evaluated named set values keyed by set SubKey.
func (b *BindingsBuilder) NamedSetValues() map[string]object.Object {
	if len(b.namedSetValues) == 0 {
		return nil
	}
	copyMap := make(map[string]object.Object, len(b.namedSetValues))
	maps.Copy(copyMap, b.namedSetValues)
	return copyMap
}

// RegisterNamedSets evaluates and caches model-level named sets for expression lookup.
func (b *BindingsBuilder) RegisterNamedSets(model *core.Model) error {
	b.namedSetValues = make(map[string]object.Object)
	if len(model.NamedSets) == 0 {
		return nil
	}

	evalBindings := evaluator.NewBindings()
	for _, ns := range model.NamedSets {
		if ns.Spec.Expression == nil {
			return fmt.Errorf("named set %q has no lowered expression", ns.Name)
		}
		result := evaluator.Eval(ns.Spec.Expression, evalBindings)
		if result.IsError() {
			return fmt.Errorf("named set %q: %s", ns.Name, result.Error.Inspect())
		}
		b.namedSetValues[ns.Key.SubKey] = result.Value
	}
	return nil
}

func (b *BindingsBuilder) applyNamedSets(bindings *evaluator.Bindings) {
	for name, value := range b.namedSetValues {
		bindings.Set(name, value, evaluator.NamespaceGlobal)
	}
}

// BuildGlobal creates a root bindings context with global state variables.
// This is suitable for evaluating model-level invariants.
func (b *BindingsBuilder) BuildGlobal() *evaluator.Bindings {
	bindings := evaluator.NewBindings()
	bindings.SetRelationContext(b.buildRelationContext())
	b.applyNamedSets(bindings)
	return bindings
}

// SetDerivedResolver sets the resolver used to compute derived attribute values.
func (b *BindingsBuilder) SetDerivedResolver(resolver DerivedAttributeResolver) {
	b.derivedResolver = resolver
}

// BuildForInstanceBase creates bindings for an instance without resolving derived attributes.
// Use this when evaluating a DerivationPolicy to avoid recursive derived resolution.
func (b *BindingsBuilder) BuildForInstanceBase(instance *ClassInstance) *evaluator.Bindings {
	bindings := evaluator.NewBindings()
	bindings.SetRelationContext(b.buildRelationContext())
	child := bindings.WithSelfAndClass(instance.Attributes, instance.ClassKey.String())
	b.applyNamedSets(child)
	return child
}

// BuildForInstance creates bindings with "self" set to the given instance.
// This is suitable for evaluating action requires/guarantees.
// If a DerivedAttributeResolver is set, derived attributes are computed on-demand
// and injected into the self record.
func (b *BindingsBuilder) BuildForInstance(instance *ClassInstance) *evaluator.Bindings {
	bindings := evaluator.NewBindings()
	bindings.SetRelationContext(b.buildRelationContext())

	attrs := b.resolveAttributes(instance)

	// Create a child scope with self set
	child := bindings.WithSelfAndClass(attrs, instance.ClassKey.String())
	b.applyNamedSets(child)
	return child
}

// BuildForInstanceWithVariables creates bindings with "self" and additional variables.
// The variables map contains name -> value pairs to add to the binding scope.
// This is useful for action parameters.
func (b *BindingsBuilder) BuildForInstanceWithVariables(
	instance *ClassInstance,
	variables map[string]object.Object,
) *evaluator.Bindings {
	bindings := b.BuildForInstance(instance)

	// Add variables to the scope
	for name, value := range variables {
		bindings.Set(name, value, evaluator.NamespaceLocal)
	}

	b.applyNamedSets(bindings)
	return bindings
}

// Class extent elements bound into TLA as records [id |-> N, data |-> attrs].
// id is the engine instance identity (number); data is the attribute record.
// Association links use id; authors read attributes via x.data.field (or self as data in instance scope).
const (
	ClassExtentIDField   = "id"
	ClassExtentDataField = "data"
)

// BuildWithClassInstances creates bindings that include all instances of classes
// as sets accessible by class name. This enables expressions like "∀ o ∈ Orders : ...".
func (b *BindingsBuilder) BuildWithClassInstances(classNameMap map[identity.Key]string) *evaluator.Bindings {
	bindings := evaluator.NewBindings()
	bindings.SetRelationContext(b.buildRelationContext())
	b.bindClassInstanceSets(bindings, classNameMap)
	b.applyNamedSets(bindings)
	return bindings
}

// bindClassInstanceSets adds one set per class name. Each element is [id, data].
func (b *BindingsBuilder) bindClassInstanceSets(bindings *evaluator.Bindings, classNameMap map[identity.Key]string) {
	for classKey, className := range classNameMap {
		bindings.Set(className, classInstanceExtentSet(b.state.InstancesByClass(classKey)), evaluator.NamespaceGlobal)
	}
}

// classInstanceExtentSet builds the TLA class extent: a set of [id |-> id, data |-> attributes].
// Distinct ids keep instances separate even when attribute data is identical.
func classInstanceExtentSet(instances []*ClassInstance) *object.Set {
	classSet := object.NewSet()
	for _, instance := range instances {
		classSet.Add(ClassExtentElement(instance.ID, instance.Attributes))
	}
	return classSet
}

// ClassExtentElement builds one class-extent record [id |-> id, data |-> attrs].
// data is a clone so evaluation cannot mutate persisted instance attributes through the extent.
func ClassExtentElement(id InstanceID, attrs *object.Record) *object.Record {
	data := attrs
	if data != nil {
		data = data.Clone().(*object.Record)
	} else {
		data = object.NewRecord()
	}
	return object.NewRecordFromFields(map[string]object.Object{
		ClassExtentIDField:   object.NewNatural(instanceIDAsInt64(id)),
		ClassExtentDataField: data,
	})
}

// InstanceIDFromExtentElement returns the engine id from a class-extent [id, data] record.
func InstanceIDFromExtentElement(elem *object.Record) (InstanceID, bool) {
	if elem == nil {
		return 0, false
	}
	idVal := elem.Get(ClassExtentIDField)
	if idVal == nil {
		return 0, false
	}
	n, ok := idVal.(*object.Number)
	if !ok || n.Sign() < 0 {
		return 0, false
	}
	v := n.Rat().Num().Int64()
	if v < 0 {
		return 0, false
	}
	return InstanceID(uint64(v)), true
}

// DataFromExtentElement returns the data record from a class-extent element, or elem itself if flat.
func DataFromExtentElement(elem *object.Record) *object.Record {
	if elem == nil {
		return nil
	}
	if data, ok := elem.Get(ClassExtentDataField).(*object.Record); ok && data != nil {
		return data
	}
	return elem
}

// BuildWithClassInstancesForInstance combines BuildWithClassInstances and BuildForInstance.
// Creates bindings with class instance sets and "self" set to the given instance.
func (b *BindingsBuilder) BuildWithClassInstancesForInstance(
	classNameMap map[identity.Key]string,
	instance *ClassInstance,
) *evaluator.Bindings {
	bindings := b.BuildWithClassInstances(classNameMap)

	attrs := b.resolveAttributes(instance)

	// Create a child scope with self set
	return bindings.WithSelfAndClass(attrs, instance.ClassKey.String())
}

// BuildWithClassInstancesForInstanceWithVariables combines class instance sets, self, and parameters.
func (b *BindingsBuilder) BuildWithClassInstancesForInstanceWithVariables(
	classNameMap map[identity.Key]string,
	instance *ClassInstance,
	variables map[string]object.Object,
) *evaluator.Bindings {
	bindings := b.BuildWithClassInstancesForInstance(classNameMap, instance)
	for name, value := range variables {
		bindings.Set(name, value, evaluator.NamespaceLocal)
	}
	return bindings
}

// resolveAttributes returns the instance's attributes with derived values injected.
// If no DerivedAttributeResolver is set, returns the original attributes unchanged.
func (b *BindingsBuilder) resolveAttributes(instance *ClassInstance) *object.Record {
	if b.derivedResolver == nil {
		return instance.Attributes
	}

	derived, err := b.derivedResolver.ResolveDerived(instance)
	if err != nil || len(derived) == 0 {
		return instance.Attributes
	}

	// Clone the attributes to avoid modifying the persisted instance.
	attrs := instance.Attributes.Clone().(*object.Record)
	for name, value := range derived {
		attrs.Set(name, value)
	}
	return attrs
}

// buildRelationContext builds or returns the relation context with current link state.
func (b *BindingsBuilder) buildRelationContext() *evaluator.RelationContext {
	if b.relationCtx == nil {
		b.relationCtx = evaluator.NewRelationContext()
	}

	b.syncLinks()
	b.syncAssociationLinks()

	return b.relationCtx
}

// syncLinks synchronizes links from simulation state to the relation context.
// This ensures the evaluator sees the current link state.
func (b *BindingsBuilder) syncLinks() {
	// Clear existing links in relation context
	b.relationCtx.Links().Clear()

	// Copy links from simulation state
	for _, instance := range b.state.AllInstances() {
		objID := evaluator.ObjectID(instance.ID)

		// Get all forward links from this instance
		links := b.state.links.GetAllForward(objID)
		for _, link := range links {
			// We need to map InstanceIDs to record pointers for the relation context
			fromInstance := b.state.GetInstance(InstanceID(link.FromID))
			toInstance := b.state.GetInstance(InstanceID(link.ToID))

			if fromInstance != nil && toInstance != nil {
				b.relationCtx.CreateLink(link.AssociationKey, fromInstance.Attributes, toInstance.Attributes)
			}
		}
	}
}

func (b *BindingsBuilder) syncAssociationLinks() {
	b.relationCtx.ClearAssociationClassRows()

	for _, link := range b.state.AssociationLinks().AllLinks() {
		fromInstance := b.state.GetInstance(link.FromEndpointID)
		linkInstance := b.state.GetInstance(link.LinkInstanceID)
		toInstance := b.state.GetInstance(link.ToEndpointID)
		if fromInstance == nil || linkInstance == nil || toInstance == nil {
			continue
		}

		hostKey := evaluator.AssociationKey(link.HostAssocKey.String())
		b.relationCtx.CreateLink(hostKey, fromInstance.Attributes, toInstance.Attributes)
		b.relationCtx.AddAssociationClassRow(
			hostKey,
			fromInstance.Attributes,
			toInstance.Attributes,
			linkInstance.Attributes,
		)
	}
}

// AddAssociationClassHost registers a host association materialized by association-class instances.
func (b *BindingsBuilder) AddAssociationClassHost(
	assocKey identity.Key,
	name string,
	endpoints evaluator.AssociationHostEndpoints,
	linkClassName string,
	mults evaluator.AssociationHostMultiplicities,
) {
	b.relationCtx.AddAssociationClassHost(
		evaluator.AssociationKey(assocKey.String()),
		name,
		endpoints,
		linkClassName,
		mults,
	)
}

// State returns the underlying simulation state.
func (b *BindingsBuilder) State() *SimulationState {
	return b.state
}

// RelationContext returns the relation context used by this builder.
func (b *BindingsBuilder) RelationContext() *evaluator.RelationContext {
	return b.relationCtx
}

// AddAssociation registers an association with the relation context.
// This must be called for each association before evaluating expressions
// that traverse associations.
func (b *BindingsBuilder) AddAssociation(
	assocKey identity.Key,
	name string,
	fromClassKey identity.Key,
	toClassKey identity.Key,
	fromMultiplicity evaluator.Multiplicity,
	toMultiplicity evaluator.Multiplicity,
) {
	b.relationCtx.AddAssociation(
		evaluator.AssociationKey(assocKey.String()),
		name,
		fromClassKey.String(),
		toClassKey.String(),
		fromMultiplicity,
		toMultiplicity,
	)
}

// ApplyPrimedBindings applies primed bindings from an evaluation result
// back to the simulation state. This is how guarantees with primed assignments
// modify the simulation state.
func (b *BindingsBuilder) ApplyPrimedBindings(
	instance *ClassInstance,
	bindings *evaluator.Bindings,
) error {
	primedBindings := bindings.GetPrimedBindings()

	for name, value := range primedBindings {
		// Check if this is a self field modification (self.field' = value)
		// In the bindings, self fields appear as the field name directly
		if instance.HasAttribute(name) {
			if err := b.state.UpdateInstanceField(instance.ID, name, value); err != nil {
				return err
			}
		}
	}

	return nil
}

// instanceIDAsInt64 converts a simulation instance id for embedding in TLA records.
// Instance ids are small sequential values; values above MaxInt64 are clamped.
func instanceIDAsInt64(id InstanceID) int64 {
	if id > math.MaxInt64 {
		return math.MaxInt64
	}
	return int64(id)
}
