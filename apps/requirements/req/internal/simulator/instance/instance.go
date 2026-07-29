package instance

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// Instance is one live class instance in a simulation run.
// Construct via [State.CreateInstance]; tests may use [NewInstance].
//
// Accessors [Instance.GetID], [Instance.GetClassKey], and [Instance.GetAttributes]
// are preferred. Fields stay exported because the package is named instance: methods
// named ID/ClassKey would collide with package-qualified types (instance.ID) at
// call sites that also name the variable instance.
//
// Attribute mutability: update attributes via [Instance.SetAttribute] or
// [State.UpdateInstanceField]. The attribute record pointer is shared with the
// evaluator for self-bindings; treat it as read-mostly outside those APIs.
type Instance struct {
	// ID uniquely identifies this instance within the simulation.
	ID ID

	// ClassKey identifies the class this instance belongs to.
	ClassKey identity.Key

	// Attributes holds the current attribute values for this instance.
	Attributes *object.Record
}

// NewInstance builds an instance with the given identity and attribute record.
// Prefer [State.CreateInstance] for live world membership.
func NewInstance(id ID, classKey identity.Key, attributes *object.Record) *Instance {
	if attributes == nil {
		attributes = object.NewRecord()
	}
	return &Instance{
		ID:         id,
		ClassKey:   classKey,
		Attributes: attributes,
	}
}

// GetID returns the instance identity.
func (i *Instance) GetID() ID {
	if i == nil {
		return 0
	}
	return i.ID
}

// GetClassKey returns the class key for this instance.
func (i *Instance) GetClassKey() identity.Key {
	if i == nil {
		return identity.Key{}
	}
	return i.ClassKey
}

// GetAttributes returns the attribute record (shared storage; see type comment).
func (i *Instance) GetAttributes() *object.Record {
	if i == nil {
		return nil
	}
	return i.Attributes
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

// withAttribute returns a new instance with the specified attribute updated.
// The original instance is not modified.
func (i *Instance) withAttribute(name string, value object.Object) *Instance {
	return &Instance{
		ID:         i.ID,
		ClassKey:   i.ClassKey,
		Attributes: i.Attributes.WithField(name, value),
	}
}
