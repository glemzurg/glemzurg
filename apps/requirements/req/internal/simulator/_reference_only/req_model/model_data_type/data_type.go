package model_data_type

import (
	"sort"
	"strconv"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

const (
	CollectionTypeAtomic    = "atomic"    // An atomic type, not a collection.
	CollectionTypeOrdered   = "ordered"   // An ordered collection.
	CollectionTypeQueue     = "queue"     // A first in, first out queue.
	CollectionTypeRecord    = "record"    // A data type composed of fields of other data types..
	CollectionTypeStack     = "stack"     // A first in, last out stack.
	CollectionTypeUnordered = "unordered" // An unordered collection.
)

// DataType represents the main data type structure.
type DataType struct {
	Key              string
	CollectionType   string
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
			CollectionType: CollectionTypeAtomic,
			Atomic: &Atomic{
				ConstraintType: ConstraintTypeUnconstrained,
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
	return validation.ValidateStruct(&d,
		validation.Field(&d.Key, validation.Required),
		validation.Field(&d.CollectionType, validation.Required, validation.In(CollectionTypeAtomic, CollectionTypeStack, CollectionTypeUnordered, CollectionTypeOrdered, CollectionTypeQueue, CollectionTypeRecord)),
		validation.Field(&d.Atomic, validation.Required.When(d.CollectionType == CollectionTypeAtomic), validation.By(func(value interface{}) error {
			if a, ok := value.(*Atomic); ok && a != nil {
				return a.Validate()
			}
			return nil
		})),
		validation.Field(&d.RecordFields, validation.Required.When(d.CollectionType == CollectionTypeRecord), validation.By(func(value interface{}) error {
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
			if d.CollectionType == CollectionTypeStack || d.CollectionType == CollectionTypeUnordered || d.CollectionType == CollectionTypeOrdered || d.CollectionType == CollectionTypeQueue {
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
	case CollectionTypeRecord:
		result := "{\n"
		for _, field := range d.RecordFields {
			result += field.String() + "\n"
		}
		result += "}"
		return result
	case CollectionTypeAtomic:
		if d.Atomic == nil {
			panic("atomic is nil")
		}
		return d.Atomic.String()
	case CollectionTypeStack, CollectionTypeUnordered, CollectionTypeOrdered, CollectionTypeQueue:
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
		if collectionType == CollectionTypeUnordered || collectionType == CollectionTypeOrdered {
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
		if dt.CollectionType == CollectionTypeRecord {
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
