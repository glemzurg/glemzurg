package data_type

const (
	_COLLECTION_TYPE_ATOMIC    = "atomic"    // An atomic type, not a collection.
	_COLLECTION_TYPE_ORDERED   = "ordered"   // An ordered collection.
	_COLLECTION_TYPE_QUEUE     = "queue"     // A first in, first out queue.
	_COLLECTION_TYPE_RECORD    = "record"    // A data type composed of fields of other data types..
	_COLLECTION_TYPE_STACK     = "stack"     // A first in, last out stack.
	_COLLECTION_TYPE_UNORDERED = "unordered" // An unordered collection.
)

// DataType represents the main data type structure.
type DataType struct {
	Key              string
	Name             string
	Details          string
	CollectionType   string
	CollectionUnique *bool
	CollectionMin    *int
	CollectionMax    *int
	Atomic           *Atomic
}
