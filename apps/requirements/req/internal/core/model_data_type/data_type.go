package model_data_type

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	pkgerrors "github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
)

const (
	COLLECTION_TYPE_ATOMIC    = "atomic"    // An atomic type, not a collection.
	COLLECTION_TYPE_ORDERED   = "ordered"   // An ordered collection.
	COLLECTION_TYPE_QUEUE     = "queue"     // A first in, first out queue.
	COLLECTION_TYPE_RECORD    = "record"    // A data type composed of fields of other data types..
	COLLECTION_TYPE_STACK     = "stack"     // A first in, last out stack.
	COLLECTION_TYPE_UNORDERED = "unordered" // An unordered collection.
)

var _validCollectionTypes = map[string]bool{
	COLLECTION_TYPE_ATOMIC:    true,
	COLLECTION_TYPE_ORDERED:   true,
	COLLECTION_TYPE_QUEUE:     true,
	COLLECTION_TYPE_RECORD:    true,
	COLLECTION_TYPE_STACK:     true,
	COLLECTION_TYPE_UNORDERED: true,
}

// DataType represents the main data type structure.
type DataType struct {
	Key              string
	CollectionType   string
	CollectionUnique *bool
	CollectionMin    *int
	CollectionMax    *int
	Atomic           *Atomic
	RecordFields     []Field
	TypeSpec         *logic_spec.TypeSpec // Optional precise type specification.
}

// New creates a new DataType by parsing the input text.
func New(key, text string, typeSpec *logic_spec.TypeSpec) (dataType *DataType, err error) {
	// If this is blank then it is an unconstrained data type.
	if strings.TrimSpace(text) == "" {
		dataType = &DataType{
			CollectionType: COLLECTION_TYPE_ATOMIC,
			Atomic: &Atomic{
				ConstraintType: CONSTRAINT_TYPE_UNCONSTRAINED,
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
			var el errList
			if errors.As(err, &el) {
				// Overwrite the err with a error calling code can use.
				err = &CannotParseError{
					err:   pkgerrors.WithStack(el),
					input: text,
				}
			}

			return nil, err
		}

		// Case to the data type.
		var ok bool
		dataType, ok = dataTypeAny.(*DataType)
		if !ok {
			return nil, pkgerrors.Errorf("parsed data type is not of type *DataType")
		}
	}

	// Set the key and optional type spec.
	dataType.Key = key
	dataType.TypeSpec = typeSpec

	// Validate the data type.
	ctx := coreerr.NewContext("datatype", key)
	if err = dataType.Validate(ctx); err != nil {
		return nil, err
	}

	return dataType, nil
}

// Validate validates the DataType struct.
func (d DataType) Validate(ctx *coreerr.ValidationContext) error {
	// Key: required.
	if d.Key == "" {
		return coreerr.New(ctx, coreerr.DtypeKeyRequired, "Key is required", "Key")
	}
	// CollectionType: required and must be valid.
	if d.CollectionType == "" {
		return coreerr.NewWithValues(ctx, coreerr.DtypeCollectiontypeRequired, "CollectionType is required", "CollectionType", "", "one of: atomic, ordered, queue, record, stack, unordered")
	}
	if !_validCollectionTypes[d.CollectionType] {
		return coreerr.NewWithValues(ctx, coreerr.DtypeCollectiontypeInvalid, "CollectionType is not a valid value", "CollectionType", d.CollectionType, "one of: atomic, ordered, queue, record, stack, unordered")
	}
	if err := d.validateAtomic(ctx); err != nil {
		return err
	}
	if err := d.validateRecordFields(ctx); err != nil {
		return err
	}
	if err := d.validateCollectionFields(ctx); err != nil {
		return err
	}
	if d.TypeSpec != nil {
		if err := d.TypeSpec.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (d DataType) validateAtomic(ctx *coreerr.ValidationContext) error {
	if d.CollectionType == COLLECTION_TYPE_ATOMIC && d.Atomic == nil {
		return coreerr.New(ctx, coreerr.DtypeAtomicRequired, "atomic is required for atomic collection type", "Atomic")
	}
	if d.Atomic != nil {
		childCtx := ctx.Child("atomic", "")
		if err := d.Atomic.Validate(childCtx); err != nil {
			return err
		}
	}
	return nil
}

func (d DataType) validateRecordFields(ctx *coreerr.ValidationContext) error {
	if d.CollectionType == COLLECTION_TYPE_RECORD && len(d.RecordFields) == 0 {
		return coreerr.New(ctx, coreerr.DtypeRecordfieldsRequired, "record fields are required for record collection type", "RecordFields")
	}
	for i := range d.RecordFields {
		childCtx := ctx.Child("field", fmt.Sprintf("%d", i))
		if err := d.RecordFields[i].Validate(childCtx); err != nil {
			return err
		}
	}
	return nil
}

func (d DataType) validateCollectionFields(ctx *coreerr.ValidationContext) error {
	isCollection := d.CollectionType == COLLECTION_TYPE_STACK ||
		d.CollectionType == COLLECTION_TYPE_UNORDERED ||
		d.CollectionType == COLLECTION_TYPE_ORDERED ||
		d.CollectionType == COLLECTION_TYPE_QUEUE

	if isCollection {
		if d.CollectionUnique == nil {
			return coreerr.New(ctx, coreerr.DtypeColluniqRequired, "collection unique is required for collection types", "CollectionUnique")
		}
		if d.CollectionMin != nil && *d.CollectionMin < 1 {
			return coreerr.NewWithValues(ctx, coreerr.DtypeCollminTooSmall, "collection min must be at least 1", "CollectionMin", fmt.Sprintf("%d", *d.CollectionMin), "at least 1")
		}
		if d.CollectionMax != nil && *d.CollectionMax < 1 {
			return coreerr.NewWithValues(ctx, coreerr.DtypeCollmaxTooSmall, "collection max must be at least 1", "CollectionMax", fmt.Sprintf("%d", *d.CollectionMax), "at least 1")
		}
		if d.CollectionMin != nil && d.CollectionMax != nil && *d.CollectionMax < *d.CollectionMin {
			return coreerr.NewWithValues(ctx, coreerr.DtypeCollmaxLessThanMin, "collection max must be at least collection min", "CollectionMax", fmt.Sprintf("%d", *d.CollectionMax), fmt.Sprintf("at least %d", *d.CollectionMin))
		}
		return nil
	}
	if d.CollectionUnique != nil {
		return coreerr.New(ctx, coreerr.DtypeColluniqMustBeBlank, "collection unique must be nil for non-collection types", "CollectionUnique")
	}
	if d.CollectionMin != nil {
		return coreerr.New(ctx, coreerr.DtypeCollminMustBeBlank, "collection min must be nil for non-collection types", "CollectionMin")
	}
	if d.CollectionMax != nil {
		return coreerr.New(ctx, coreerr.DtypeCollmaxMustBeBlank, "collection max must be nil for non-collection types", "CollectionMax")
	}
	return nil
}

// String returns a string representation of the DataType.
func (d DataType) String() string {
	switch d.CollectionType {
	case COLLECTION_TYPE_RECORD:
		var b strings.Builder
		b.WriteString("{\n")
		for _, field := range d.RecordFields {
			b.WriteString(field.String())
			b.WriteString("\n")
		}
		b.WriteString("}")
		return b.String()
	case COLLECTION_TYPE_ATOMIC:
		if d.Atomic == nil {
			panic("atomic is nil")
		}
		return d.Atomic.String()
	case COLLECTION_TYPE_STACK, COLLECTION_TYPE_UNORDERED, COLLECTION_TYPE_ORDERED, COLLECTION_TYPE_QUEUE:
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
		if collectionType == COLLECTION_TYPE_UNORDERED || collectionType == COLLECTION_TYPE_ORDERED {
			collectionType += " collection"
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
		// So we can insert them first to ensure there are no foreign key violations.
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
		if dt.CollectionType == COLLECTION_TYPE_RECORD {
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
