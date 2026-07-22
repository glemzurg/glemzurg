package state

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance"
)

// Compatibility aliases so existing simulator packages can migrate onto
// instance.State without a single-shot import rewrite. Prefer importing
// instance directly for new code.

// InstanceID uniquely identifies a class instance within a simulation.
type InstanceID = instance.ID

// ClassInstance is one live class instance (identity + class + attributes).
type ClassInstance = instance.Instance

// SimulationState is the mutable world for one simulation run.
type SimulationState = instance.State

// AssociationLink materializes one host association via an association-class instance.
type AssociationLink = instance.AssociationLink

// AssociationLinkTable indexes association-class host rows.
type AssociationLinkTable = instance.AssociationLinkTable

// NewSimulationState creates a new empty simulation state.
func NewSimulationState() *SimulationState {
	return instance.NewState()
}

// NewAssociationLinkTable creates an empty association link table.
func NewAssociationLinkTable() *AssociationLinkTable {
	return instance.NewAssociationLinkTable()
}
