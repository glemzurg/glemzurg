package instance

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// Instance is one live class instance in a simulation run.
// Construct via [State.CreateInstance]; tests may use [NewInstance].
//
// Use accessors [Instance.GetID], [Instance.GetClassKey], and [Instance.GetAttributes].
// Fields are unexported; the package name instance would make methods named ID/ClassKey
// collide with package-qualified types (instance.ID) at call sites that also name the
// variable instance, so Get* names are used instead.
//
// Attribute mutability: update attributes via [Instance.SetAttribute] or
// [State.UpdateInstanceField]. The attribute record pointer is shared with the
// evaluator for self-bindings; treat it as read-mostly outside those APIs.
//
// Primed attributes (x') are next-state values recorded before commit; they do
// not change [Instance.GetAttribute] until applied into Attributes.
type Instance struct {
	// ID uniquely identifies this instance within the simulation.
	id ID

	// ClassKey identifies the class this instance belongs to.
	classKey identity.Key

	// Attributes holds the current attribute values for this instance.
	attributes *object.Record

	// primed holds next-state attribute values (field' = …) until commit.
	// Nil when no primes are outstanding for this instance.
	primed map[string]object.Object
}

// NewInstance builds an instance with the given identity and attribute record.
// Prefer [State.CreateInstance] for live world membership.
func NewInstance(id ID, classKey identity.Key, attributes *object.Record) *Instance {
	if attributes == nil {
		attributes = object.NewRecord()
	}
	return &Instance{
		id:         id,
		classKey:   classKey,
		attributes: attributes,
	}
}

// GetID returns the instance identity.
func (i *Instance) GetID() ID {
	if i == nil {
		return 0
	}
	return i.id
}

// GetClassKey returns the class key for this instance.
func (i *Instance) GetClassKey() identity.Key {
	if i == nil {
		return identity.Key{}
	}
	return i.classKey
}

// GetAttributes returns the attribute record (shared storage; see type comment).
func (i *Instance) GetAttributes() *object.Record {
	if i == nil {
		return nil
	}
	return i.attributes
}

// Clone creates a deep copy of the class instance, including any primed values.
func (i *Instance) Clone() *Instance {
	clone := &Instance{
		id:         i.id,
		classKey:   i.classKey,
		attributes: i.attributes.Clone().(*object.Record),
	}
	if len(i.primed) > 0 {
		clone.primed = make(map[string]object.Object, len(i.primed))
		for name, value := range i.primed {
			clone.primed[name] = value.Clone()
		}
	}
	return clone
}

// GetAttribute returns the current (unprimed) value of an attribute by name.
// Returns nil if the attribute does not exist. Primed next-state values are
// not visible here; use [Instance.GetPrimedAttribute].
func (i *Instance) GetAttribute(name string) object.Object {
	return i.attributes.Get(name)
}

// SetAttribute sets the current (unprimed) value of an attribute.
func (i *Instance) SetAttribute(name string, value object.Object) {
	i.attributes.Set(name, value)
}

// SetPrimedAttribute records a next-state (primed) value for an attribute.
// Current unprimed storage is unchanged so both values remain readable.
func (i *Instance) SetPrimedAttribute(name string, value object.Object) {
	if i.primed == nil {
		i.primed = make(map[string]object.Object)
	}
	i.primed[name] = object.NormalizeSimulatorValue(value).Clone()
}

// GetPrimedAttribute returns the next-state value for an attribute when one
// was recorded via [Instance.SetPrimedAttribute]. The bool is false when no
// prime is outstanding for name.
func (i *Instance) GetPrimedAttribute(name string) (object.Object, bool) {
	if i == nil || i.primed == nil {
		return nil, false
	}
	value, ok := i.primed[name]
	return value, ok
}

// ClearPrimedAttributes drops all outstanding next-state values.
func (i *Instance) ClearPrimedAttributes() {
	if i == nil {
		return
	}
	i.primed = nil
}

// HasAttribute reports whether the attribute exists in current storage.
func (i *Instance) HasAttribute(name string) bool {
	return i.attributes.Has(name)
}

// AttributeNames returns the list of attribute names.
func (i *Instance) AttributeNames() []string {
	return i.attributes.FieldNames()
}

// withAttribute returns a new instance with the specified attribute updated.
// The original instance is not modified. Primed values are not copied.
func (i *Instance) withAttribute(name string, value object.Object) *Instance {
	return &Instance{
		id:         i.id,
		classKey:   i.classKey,
		attributes: i.attributes.WithField(name, value),
	}
}
