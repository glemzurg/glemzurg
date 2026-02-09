package state

import (
	"github.com/glemzurg/go-tlaplus/internal/identity"
	"github.com/glemzurg/go-tlaplus/internal/simulator/object"
)

// ClassInstance represents a single instance of a class in the simulation.
// Each instance has:
//   - A unique ID within the simulation
//   - A class key identifying which class it belongs to
//   - Attributes stored as a Record
type ClassInstance struct {
	// ID uniquely identifies this instance within the simulation
	ID InstanceID

	// ClassKey identifies the class this instance belongs to
	ClassKey identity.Key

	// Attributes holds the current attribute values for this instance
	Attributes *object.Record
}

// Clone creates a deep copy of the class instance.
func (i *ClassInstance) Clone() *ClassInstance {
	return &ClassInstance{
		ID:         i.ID,
		ClassKey:   i.ClassKey,
		Attributes: i.Attributes.Clone().(*object.Record),
	}
}

// GetAttribute returns the value of an attribute by name.
// Returns nil if the attribute doesn't exist.
func (i *ClassInstance) GetAttribute(name string) object.Object {
	return i.Attributes.Get(name)
}

// SetAttribute sets the value of an attribute.
func (i *ClassInstance) SetAttribute(name string, value object.Object) {
	i.Attributes.Set(name, value)
}

// HasAttribute returns true if the attribute exists.
func (i *ClassInstance) HasAttribute(name string) bool {
	return i.Attributes.Has(name)
}

// AttributeNames returns the list of attribute names.
func (i *ClassInstance) AttributeNames() []string {
	return i.Attributes.FieldNames()
}

// WithAttribute returns a new instance with the specified attribute updated.
// The original instance is not modified.
func (i *ClassInstance) WithAttribute(name string, value object.Object) *ClassInstance {
	return &ClassInstance{
		ID:         i.ID,
		ClassKey:   i.ClassKey,
		Attributes: i.Attributes.WithField(name, value),
	}
}
