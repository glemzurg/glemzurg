package model_data_type

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

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
	Key              string `validate:"required"`
	CollectionType   string `validate:"required,oneof=atomic ordered queue record stack unordered"`
	CollectionUnique *bool
	CollectionMin    *int
	CollectionMax    *int
	Atomic           *Atomic
	RecordFields     []Field
}

// New creates a new DataType by parsing the input text.
func New(key, text string) (dataType *DataType, err error) {

	// If this is blank then it is an unconstrained data type.
	if strings.TrimSpace(text) == "" {

		dataType = &DataType{
			CollectionType: _COLLECTION_TYPE_ATOMIC,
			Atomic: &Atomic{
				ConstraintType: _CONSTRAINT_TYPE_UNCONSTRAINED,
			},
		}

	} else {

		// Simplify the text to have easier to parse whitespace.
		// All data types are a single line anyway.
		text = strings.TrimSpace(normalizeWhitespace(text))

		// Parse the data type.
		dataTypeAny, err := Parse("", []byte(text))
		if err != nil {

			// Is this the parser error?
			if el, ok := err.(errList); ok {
				// Overwrite the err with a error calling code can use.
				err = &CannotParseError{
					err:   errors.WithStack(el),
					input: text,
				}
			}

			return nil, err
		}

		// Case to the data type.
		var ok bool
		dataType, ok = dataTypeAny.(*DataType)
		if !ok {
			return nil, errors.Errorf("parsed data type is not of type *DataType")
		}
	}

	// Set the key.
	dataType.Key = key

	// Validate the data type.
	if err = dataType.Validate(); err != nil {
		return nil, err
	}

	return dataType, nil
}

// Validate validates the DataType struct.
func (d DataType) Validate() error {
	// Validate struct tags (Key required, CollectionType required + oneof).
	if err := _validate.Struct(d); err != nil {
		return err
	}

	// Atomic: required when atomic; validate if present.
	if d.CollectionType == _COLLECTION_TYPE_ATOMIC {
		if d.Atomic == nil {
			return fmt.Errorf("Atomic: cannot be blank.")
		}
		if err := d.Atomic.Validate(); err != nil {
			return fmt.Errorf("Atomic: (%s).", err.Error())
		}
	} else if d.Atomic != nil {
		if err := d.Atomic.Validate(); err != nil {
			return fmt.Errorf("Atomic: (%s).", err.Error())
		}
	}

	// RecordFields: required when record; each field validated.
	if d.CollectionType == _COLLECTION_TYPE_RECORD {
		if len(d.RecordFields) == 0 {
			return fmt.Errorf("RecordFields: cannot be blank.")
		}
		for _, f := range d.RecordFields {
			if err := f.Validate(); err != nil {
				return fmt.Errorf("RecordFields: (%s).", err.Error())
			}
		}
	} else {
		for _, f := range d.RecordFields {
			if err := f.Validate(); err != nil {
				return fmt.Errorf("RecordFields: (%s).", err.Error())
			}
		}
	}

	// Collection field rules.
	isCollection := d.CollectionType == _COLLECTION_TYPE_STACK ||
		d.CollectionType == _COLLECTION_TYPE_UNORDERED ||
		d.CollectionType == _COLLECTION_TYPE_ORDERED ||
		d.CollectionType == _COLLECTION_TYPE_QUEUE

	if isCollection {
		// CollectionUnique is required for collections.
		if d.CollectionUnique == nil {
			return fmt.Errorf("CollectionUnique: cannot be blank.")
		}
		// CollectionMin if set must be >= 1.
		if d.CollectionMin != nil && *d.CollectionMin < 1 {
			return fmt.Errorf("CollectionMin: must be no less than 1.")
		}
		// CollectionMax if set must be >= 1.
		if d.CollectionMax != nil && *d.CollectionMax < 1 {
			return fmt.Errorf("CollectionMax: must be no less than 1.")
		}
		// If both defined, max >= min.
		if d.CollectionMin != nil && d.CollectionMax != nil && *d.CollectionMax < *d.CollectionMin {
			return fmt.Errorf("CollectionMax: must be no less than CollectionMin.")
		}
	} else {
		// Non-collections must not have collection fields.
		if d.CollectionUnique != nil {
			return fmt.Errorf("CollectionUnique: must be blank.")
		}
		if d.CollectionMin != nil {
			return fmt.Errorf("CollectionMin: must be blank.")
		}
		if d.CollectionMax != nil {
			return fmt.Errorf("CollectionMax: must be blank.")
		}
	}

	return nil
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
		if d.CollectionMin != nil || d.CollectionMax != nil {
			if d.CollectionMin != nil {
				name += strconv.Itoa(*d.CollectionMin)
			} else {
				name += "0"
			}
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

// UnpackNested unpacks all nested datatypes in the RecordFields data structures.
// Each nested datatype is given a key based on the path: root Key + "/" + field name + ...
// The order of the list is the deepest children first (post-order traversal), with the root last.
func (d DataType) UnpackNested() []DataType {
	var result []DataType
	for _, field := range d.RecordFields {
		child := field.FieldDataType
		if child != nil {
			// Set the key for the child
			child.Key = d.Key + "/" + field.Name
			// Recurse on the child
			nested := child.UnpackNested()
			result = append(result, nested...)
		}
	}
	// Add the root itself last
	result = append(result, d)
	return result
}

// SortDataTypesByKeyLengthDesc sorts a slice of DataType by key length in descending order (longest keys first).
func SortDataTypesByKeyLengthDesc(dataTypes []DataType) {
	sort.Slice(dataTypes, func(i, j int) bool {

		// Sort by key length descending first.
		// Children data types have longer keys than their parents.
		// So we can insert them first to sensure there are no foreign key violations.
		if len(dataTypes[i].Key) != len(dataTypes[j].Key) {
			return len(dataTypes[i].Key) > len(dataTypes[j].Key)
		}

		// In case of a tie, sorty by key lexicographically.
		return dataTypes[i].Key < dataTypes[j].Key
	})
}

// ExtractDatabaseObjects walks a slice of DataType and extracts database-suitable objects.
// Returns maps keyed by datatype key containing atomic enums, atomic spans, atomics, and fields.
func ExtractDatabaseObjects(dataTypes []DataType) (map[string][]Field, map[string]Atomic, map[string]AtomicSpan, map[string][]AtomicEnum) {
	atomicEnumMap := make(map[string][]AtomicEnum)
	atomicSpanMap := make(map[string]AtomicSpan)
	atomicMap := make(map[string]Atomic)
	fieldMap := make(map[string][]Field)

	for _, d := range dataTypes {

		// Collect fields for records.
		if len(d.RecordFields) > 0 {
			fieldMap[d.Key] = d.RecordFields
		}

		// Collect Atomics and their parts.
		if d.Atomic != nil {
			if d.Atomic.Span != nil {
				atomicSpanMap[d.Key] = *d.Atomic.Span
			}
			if len(d.Atomic.Enums) > 0 {
				atomicEnumMap[d.Key] = d.Atomic.Enums
			}
			atomicMap[d.Key] = *d.Atomic
		}
	}

	return fieldMap, atomicMap, atomicSpanMap, atomicEnumMap
}

// ReconstituteDataTypes takes a slice of DataType (with keys and collection types set) and maps of components,
// and reconstitutes the DataTypes by attaching the Atomics, Fields, AtomicSpans, and AtomicEnums from the maps.
// Returns a flat list sorted by key length descending.
func ReconstituteDataTypes(dataTypes []DataType, fieldMap map[string][]Field, atomicMap map[string]Atomic, atomicSpanMap map[string]AtomicSpan, atomicEnumMap map[string][]AtomicEnum) []DataType {
	result := make([]DataType, len(dataTypes))
	for i, dt := range dataTypes {
		result[i] = dt // copy the base DataType
		if fields, ok := fieldMap[dt.Key]; ok {
			result[i].RecordFields = fields
		}
		if atomic, ok := atomicMap[dt.Key]; ok {
			result[i].Atomic = &atomic
			// Attach spans and enums to the atomic
			if span, ok := atomicSpanMap[dt.Key]; ok {
				result[i].Atomic.Span = &span
			}
			if enums, ok := atomicEnumMap[dt.Key]; ok {
				result[i].Atomic.Enums = enums
			}
		}
	}
	// Sort by key length descending
	SortDataTypesByKeyLengthDesc(result)
	return result
}

// FlattenDataTypes takes a slice of DataType, calls UnpackNested on each, collects all into a single slice, sorts by key length descending, and returns it.
func FlattenDataTypes(dataTypes []DataType) []DataType {
	var result []DataType
	for _, dt := range dataTypes {
		result = append(result, dt.UnpackNested()...)
	}
	SortDataTypesByKeyLengthDesc(result)
	return result
}

// ReconstructNestedDataTypes takes a flat slice of DataType (output of FlattenDataTypes), rebuilds the nested structure by attaching child DataTypes to their parent records, and returns the root DataTypes.
func ReconstructNestedDataTypes(flatDataTypes []DataType) []DataType {
	// Create a map for quick lookup
	dtMap := make(map[string]*DataType)
	for i := range flatDataTypes {
		dtMap[flatDataTypes[i].Key] = &flatDataTypes[i]
	}

	// Track which DataTypes are children (attached to a parent)
	isChild := make(map[string]bool)

	// Attach nested DataTypes to fields
	for _, dt := range flatDataTypes {
		if dt.CollectionType == _COLLECTION_TYPE_RECORD {
			for i := range dt.RecordFields {
				field := &dt.RecordFields[i]
				if field.FieldDataType != nil && field.FieldDataType.Key != "" {
					if nested, ok := dtMap[field.FieldDataType.Key]; ok {
						field.FieldDataType = nested
						isChild[field.FieldDataType.Key] = true
					}
				}
			}
		}
	}

	// Collect roots (DataTypes that are not children
	var roots []DataType
	for _, dt := range flatDataTypes {
		if !isChild[dt.Key] {
			roots = append(roots, dt)
		}
	}

	SortDataTypesByKeyLengthDesc(roots)

	return roots
}
