package state

import (
	"github.com/glemzurg/go-tlaplus/internal/identity"
	"github.com/glemzurg/go-tlaplus/internal/simulator/evaluator"
	"github.com/glemzurg/go-tlaplus/internal/simulator/object"
)

// DerivedAttributeResolver computes derived attribute values for an instance.
// Implementations should evaluate TlaDerivationPolicy expressions on-demand.
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

// BuildGlobal creates a root bindings context with global state variables.
// This is suitable for evaluating model-level invariants.
func (b *BindingsBuilder) BuildGlobal() *evaluator.Bindings {
	bindings := evaluator.NewBindings()
	bindings.SetRelationContext(b.buildRelationContext())
	return bindings
}

// SetDerivedResolver sets the resolver used to compute derived attribute values.
func (b *BindingsBuilder) SetDerivedResolver(resolver DerivedAttributeResolver) {
	b.derivedResolver = resolver
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
	return bindings.WithSelfAndClass(attrs, instance.ClassKey.String())
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

	return bindings
}

// BuildWithClassInstances creates bindings that include all instances of classes
// as sets accessible by class name. This enables expressions like "∀ o ∈ Orders : ...".
func (b *BindingsBuilder) BuildWithClassInstances(classNameMap map[identity.Key]string) *evaluator.Bindings {
	bindings := evaluator.NewBindings()
	bindings.SetRelationContext(b.buildRelationContext())

	// Build sets for each class
	for classKey, className := range classNameMap {
		instances := b.state.InstancesByClass(classKey)

		// Create a set of all instance attribute records
		elements := make([]object.Object, len(instances))
		for i, instance := range instances {
			elements[i] = instance.Attributes
		}

		classSet := object.NewSet()
		for _, elem := range elements {
			classSet.Add(elem)
		}

		bindings.Set(className, classSet, evaluator.NamespaceGlobal)
	}

	return bindings
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

	// Sync the link table from simulation state
	// The relation context's link table is separate - we need to sync it
	b.syncLinks()

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
