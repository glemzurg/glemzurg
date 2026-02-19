package database

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_data_type"

	"github.com/pkg/errors"
)

// AddTopLevelDataTypes adds a map of DataType to the database, handling nested structures.
func AddTopLevelDataTypes(dbOrTx DbOrTx, modelKey string, dataTypes map[string]model_data_type.DataType) error {
	// Convert map to slice for flattening
	var dataTypeSlice []model_data_type.DataType
	for _, dt := range dataTypes {
		dataTypeSlice = append(dataTypeSlice, dt)
	}

	// Flatten the data types
	flatDataTypes := model_data_type.FlattenDataTypes(dataTypeSlice)

	// Extract database objects
	fieldMap, atomicMap, atomicSpanMap, atomicEnumMap := model_data_type.ExtractDatabaseObjects(flatDataTypes)

	// Insert in order: data_type, data_type_atomic, data_type_atomic_enum_value, data_type_atomic_span, data_type_field

	// Insert data_type
	if err := BulkInsertDataTypes(dbOrTx, modelKey, flatDataTypes); err != nil {
		return errors.WithStack(err)
	}

	// Insert data_type_atomic
	if err := BulkInsertAtomics(dbOrTx, modelKey, atomicMap); err != nil {
		return errors.WithStack(err)
	}

	// Insert data_type_atomic_enum_value
	if err := BulkInsertAtomicEnums(dbOrTx, modelKey, atomicEnumMap); err != nil {
		return errors.WithStack(err)
	}

	// Insert data_type_atomic_span
	if err := BulkInsertAtomicSpans(dbOrTx, modelKey, atomicSpanMap); err != nil {
		return errors.WithStack(err)
	}

	// Insert data_type_field
	if err := BulkInsertFields(dbOrTx, modelKey, fieldMap); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// LoadTopLevelDataTypes loads all DataType for a model from the database, reconstructing nested structures.
func LoadTopLevelDataTypes(dbOrTx DbOrTx, modelKey string) (map[string]model_data_type.DataType, error) {
	// Load all data_type rows
	baseDataTypes, err := QueryDataTypes(dbOrTx, modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Load all atomic
	atomicMap, err := QueryAtomics(dbOrTx, modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Load all atomic spans
	atomicSpanMap, err := QueryAtomicSpans(dbOrTx, modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Load all atomic enums
	atomicEnumMap, err := QueryAtomicEnums(dbOrTx, modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Load all fields
	fieldMap, err := QueryFields(dbOrTx, modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Reconstitute
	reconstituted := model_data_type.ReconstituteDataTypes(baseDataTypes, fieldMap, atomicMap, atomicSpanMap, atomicEnumMap)

	// Reconstruct nested
	reconstructedFlat := model_data_type.ReconstructNestedDataTypes(reconstituted)

	// Convert to map
	dataTypeMap := make(map[string]model_data_type.DataType)
	for _, dt := range reconstructedFlat {
		dataTypeMap[dt.Key] = dt
	}

	return dataTypeMap, nil
}
