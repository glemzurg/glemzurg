package simulator

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
)

// AssociationConfig contains the configuration for an association.
// This is used to build a RelationContext from model data.
type AssociationConfig struct {
	// Key is the full identity.Key.String() for the association
	Key string

	// Name is the display name (e.g., "Lines")
	Name string

	// FromClassKey is the identity.Key.String() for the "from" class
	FromClassKey string

	// ToClassKey is the identity.Key.String() for the "to" class
	ToClassKey string

	// FromMultiplicity contains the cardinality on the "from" side
	FromMultiplicity evaluator.Multiplicity

	// ToMultiplicity contains the cardinality on the "to" side
	ToMultiplicity evaluator.Multiplicity
}

// BuildRelationContext creates a RelationContext from a list of associations.
// This populates the metadata but does not create any links.
func BuildRelationContext(associations []AssociationConfig) *evaluator.RelationContext {
	ctx := evaluator.NewRelationContext()

	for _, assoc := range associations {
		ctx.AddAssociation(
			evaluator.AssociationKey(assoc.Key),
			assoc.Name,
			assoc.FromClassKey,
			assoc.ToClassKey,
			assoc.FromMultiplicity,
			assoc.ToMultiplicity,
		)
	}

	return ctx
}

// RelationContextBuilder provides a fluent API for building RelationContext.
type RelationContextBuilder struct {
	associations []AssociationConfig
}

// NewRelationContextBuilder creates a new builder.
func NewRelationContextBuilder() *RelationContextBuilder {
	return &RelationContextBuilder{
		associations: make([]AssociationConfig, 0),
	}
}

// AddAssociation adds an association to the builder.
func (b *RelationContextBuilder) AddAssociation(config AssociationConfig) *RelationContextBuilder {
	b.associations = append(b.associations, config)
	return b
}

// Build creates the RelationContext from the configured associations.
func (b *RelationContextBuilder) Build() *evaluator.RelationContext {
	return BuildRelationContext(b.associations)
}
