package data_type

import (
	"strconv"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
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
	RecordFields     []Field
}

// New creates a new DataType by parsing the input text.
func New(key, text string) (dataType *DataType, err error) {

	// Parse the data type.
	dataTypeAny, err := Parse("", []byte(text))
	if err != nil {
		return nil, nil // Not an error, just cannot parse.
	}

	// Case to the data type.
	dataType, ok := dataTypeAny.(*DataType)
	if !ok {
		return nil, errors.Errorf("parsed data type is not of type *DataType")
	}

	// Set the key.
	dataType.Key = key

	// Set the name.
	dataType.Name = dataType.String()

	// Validate the data type.
	if err = dataType.Validate(); err != nil {
		return nil, err
	}

	return dataType, nil
}

// Validate validates the DataType struct.
func (d DataType) Validate() error {
	return validation.ValidateStruct(&d,
		validation.Field(&d.Key, validation.Required),
		validation.Field(&d.Name, validation.Required),
		validation.Field(&d.CollectionType, validation.Required, validation.In(_COLLECTION_TYPE_ATOMIC, _COLLECTION_TYPE_STACK, _COLLECTION_TYPE_UNORDERED, _COLLECTION_TYPE_ORDERED, _COLLECTION_TYPE_QUEUE, _COLLECTION_TYPE_RECORD)),
		validation.Field(&d.Atomic, validation.Required, validation.By(func(value interface{}) error {
			if a, ok := value.(*Atomic); ok && a != nil {
				return a.Validate()
			}
			return nil
		})),
		validation.Field(&d.RecordFields, validation.Required.When(d.CollectionType == _COLLECTION_TYPE_RECORD), validation.By(func(value interface{}) error {
			if fields, ok := value.([]Field); ok {
				for _, f := range fields {
					if err := f.Validate(); err != nil {
						return err
					}
				}
			}
			return nil
		})),
		validation.Field(&d.CollectionMin, validation.By(func(value interface{}) error {
			if d.CollectionType == _COLLECTION_TYPE_STACK || d.CollectionType == _COLLECTION_TYPE_UNORDERED || d.CollectionType == _COLLECTION_TYPE_ORDERED || d.CollectionType == _COLLECTION_TYPE_QUEUE {
				if value == nil {
					return errors.New("cannot be blank")
				}
				return nil
			}
			return nil
		}), validation.Min(0)),
		validation.Field(&d.CollectionMax, validation.Min(0)),
	)
}

// String returns a string representation of the DataType.
func (d DataType) String() string {
	switch d.CollectionType {
	case _COLLECTION_TYPE_RECORD:
		result := "{\n"
		for _, field := range d.RecordFields {
			result += field.String() + "\n"
		}
		result += "}"
		return result
	case _COLLECTION_TYPE_ATOMIC:
		if d.Atomic == nil {
			panic("atomic is nil")
		}
		return d.Atomic.String()
	case _COLLECTION_TYPE_STACK, _COLLECTION_TYPE_UNORDERED, _COLLECTION_TYPE_ORDERED, _COLLECTION_TYPE_QUEUE:
		name := ""
		if d.CollectionUnique != nil && *d.CollectionUnique {
			name += "unique "
		}
		if d.CollectionMin != nil && (*d.CollectionMin != 0 || d.CollectionMax != nil) {
			name += strconv.Itoa(*d.CollectionMin)
			if d.CollectionMax != nil {
				name += "-" + strconv.Itoa(*d.CollectionMax)
			} else {
				name += "+"
			}
			name += " "
		}

		collectionType := d.CollectionType
		if collectionType == _COLLECTION_TYPE_UNORDERED || collectionType == _COLLECTION_TYPE_ORDERED {
			collectionType = collectionType + " collection"
		}

		name += collectionType + " of "
		if d.Atomic != nil {
			name += d.Atomic.String()
		}
		return name
	default:
		panic("unsupported collection type: '" + d.CollectionType + "'")
	}
}
