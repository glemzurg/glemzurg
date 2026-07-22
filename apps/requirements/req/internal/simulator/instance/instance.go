package instance

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// Instance is one live class instance in a simulation run.
// Construct via [State.CreateInstance]; tests may use struct literals.
//
// Fields are exported for the initial migration so existing call sites can move
// packages without accessor churn. Callers should still prefer attribute helpers
// over ad-hoc map surgery on Attributes.
type Instance struct {
	// ID uniquely identifies this instance within the simulation.
	ID ID

	// ClassKey identifies the class this instance belongs to.
	ClassKey identity.Key

	// Attributes holds the current attribute values for this instance.
	Attributes *object.Record
}

// Clone creates a deep copy of the class instance.
func (i *Instance) Clone() *Instance {
	return &Instance{
		ID:         i.ID,
		ClassKey:   i.ClassKey,
		Attributes: i.Attributes.Clone().(*object.Record),
	}
}

// GetAttribute returns the value of an attribute by name.
// Returns nil if the attribute does not exist.
func (i *Instance) GetAttribute(name string) object.Object {
	return i.Attributes.Get(name)
}

// SetAttribute sets the value of an attribute.
func (i *Instance) SetAttribute(name string, value object.Object) {
	i.Attributes.Set(name, value)
}

// HasAttribute reports whether the attribute exists.
func (i *Instance) HasAttribute(name string) bool {
	return i.Attributes.Has(name)
}

// AttributeNames returns the list of attribute names.
func (i *Instance) AttributeNames() []string {
	return i.Attributes.FieldNames()
}

// WithAttribute returns a new instance with the specified attribute updated.
// The original instance is not modified.
func (i *Instance) WithAttribute(name string, value object.Object) *Instance {
	return &Instance{
		ID:         i.ID,
		ClassKey:   i.ClassKey,
		Attributes: i.Attributes.WithField(name, value),
	}
}
