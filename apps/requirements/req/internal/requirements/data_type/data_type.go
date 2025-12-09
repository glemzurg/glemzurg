package data_type

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

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

// New creates a new DataType by parsing the input text.
func New(text string) (dataType *DataType, err error) {

	// parser := NewParser(text)

	// // Parse the data type.
	// dataType, err = parser.Parse()
	// if err != nil {
	// 	return nil, err
	// }

	// // Validate the data type.
	// if err = dataType.Validate(); err != nil {
	// 	return nil, err
	// }

	return dataType, nil
}

// Validate validates the DataType struct.
func (d DataType) Validate() error {
	return validation.ValidateStruct(&d,
		validation.Field(&d.Key, validation.Required),
		validation.Field(&d.Name, validation.Required),
		validation.Field(&d.CollectionType, validation.Required, validation.In(_COLLECTION_TYPE_ATOMIC)),
		validation.Field(&d.Atomic, validation.Required.When(d.CollectionType == _COLLECTION_TYPE_ATOMIC), validation.By(func(value interface{}) error {
			if a, ok := value.(*Atomic); ok && a != nil {
				return a.Validate()
			}
			return nil
		})),
	)
}
