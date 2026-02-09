package object

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// ObjectType represents the runtime type of an object.
// These are used for hashing and type identification during AST evaluation.
type ObjectType string

const (
	TypeNumber  ObjectType = "Number"
	TypeBoolean ObjectType = "Boolean"
	TypeString  ObjectType = "String"
	TypeSet     ObjectType = "Set"
	TypeBag     ObjectType = "Bag"
	TypeTuple   ObjectType = "Tuple"
	TypeRecord  ObjectType = "Record"
	TypeError   ObjectType = "Error"
)

// Object is a runtime value in the simulator.
// These are pure mathematical objects as TLA+ sees them.
type Object interface {
	// Type returns the runtime object type.
	Type() ObjectType

	// Inspect returns a human-readable string representation of the value.
	Inspect() string

	// SetValue assigns the value from another object to this one.
	// Returns an error if the source object's type is not compatible.
	SetValue(source Object) error

	// Clone creates a deep copy of this object.
	Clone() Object
}

// hashValue computes a hash from an object's type and value representation.
// This is used to create unique keys for objects in collections.
// The ~ delimiter is escaped in the inspect output to prevent collisions
// where a crafted string value could mimic another type's representation.
func hashValue(obj Object) string {
	h := sha256.New()
	h.Write([]byte(obj.Type()))
	h.Write([]byte("~"))
	// Escape ~ as ~~ to prevent delimiter injection attacks
	escaped := strings.ReplaceAll(obj.Inspect(), "~", "~~")
	h.Write([]byte(escaped))
	hashBytes := h.Sum(nil)
	return hex.EncodeToString(hashBytes)
}
