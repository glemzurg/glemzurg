package database

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/pkg/errors"
)

// collectionMinForDB maps CollectionMin=0 to nil (SQL NULL) to satisfy CHECK (collection_min > 0).
// A zero minimum means "no minimum" and is stored as NULL in the database.
func collectionMinForDB(minVal *int) *int {
	if minVal != nil && *minVal == 0 {
		return nil
	}
	return minVal
}

// Populate a golang struct from a database row.
func scanDataType(scanner Scanner, dataType *model_data_type.DataType) (err error) {
	var tsNotation *string
	var tsSpecification *string
	var elementDataTypeKey *string
	var dataTypeKeyStr string

	if err = scanner.Scan(
		&dataTypeKeyStr,
		&dataType.CollectionType,
		&dataType.CollectionUnique,
		&dataType.CollectionMin,
		&dataType.CollectionMax,
		&elementDataTypeKey,
		&tsNotation,
		&tsSpecification,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	// Parse the data_type_key column back into a typed identity.Key.
	dataType.Key, err = identity.ParseKey(dataTypeKeyStr)
	if err != nil {
		return errors.Wrapf(err, "failed to parse data type key '%s'", dataTypeKeyStr)
	}

	// Stub composite collection element; full type is stitched in ReconstructNestedDataTypes.
	if elementDataTypeKey != nil && *elementDataTypeKey != "" {
		elementKey, parseErr := identity.ParseKey(*elementDataTypeKey)
		if parseErr != nil {
			return errors.Wrapf(parseErr, "failed to parse element data type key '%s'", *elementDataTypeKey)
		}
		dataType.ElementDataType = &model_data_type.DataType{Key: elementKey}
	}

	// Reconstitute TypeSpec if present.
	if tsNotation != nil && *tsNotation != "" {
		spec := ""
		if tsSpecification != nil {
			spec = *tsSpecification
		}
		ts, err := logic_spec.NewTypeSpec(*tsNotation, spec, nil)
		if err != nil {
			return err
		}
		dataType.TypeSpec = &ts
	}

	return nil
}

// LoadDataType loads a data type from the database.
func LoadDataType(dbOrTx DbOrTx, modelKey, dataTypeKey string) (dataType model_data_type.DataType, err error) {
	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanDataType(scanner, &dataType); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			data_type_key           ,
			collection_type         ,
			collection_unique       ,
			collection_min          ,
			collection_max          ,
			element_data_type_key   ,
			type_spec_notation      ,
			type_spec_specification
		FROM
			data_type
		WHERE
			data_type_key = $2
		AND
			model_key = $1`,
		modelKey,
		dataTypeKey)
	if err != nil {
		return model_data_type.DataType{}, errors.WithStack(err)
	}

	return dataType, nil
}

// AddDataType adds a data type to the database.
func AddDataType(dbOrTx DbOrTx, modelKey string, dataType model_data_type.DataType) (err error) {
	dataTypeKey := dataType.Key.String()

	// Extract type spec fields.
	var tsNotation *string
	var tsSpecification *string
	if dataType.TypeSpec != nil {
		tsNotation = &dataType.TypeSpec.Notation
		tsSpecification = &dataType.TypeSpec.Specification
	}

	var elementDataTypeKey *string
	if dataType.ElementDataType != nil && dataType.ElementDataType.Key.KeyType != "" {
		ks := dataType.ElementDataType.Key.String()
		elementDataTypeKey = &ks
	}

	// Add the data.
	err = dbExec(dbOrTx, `
		INSERT INTO data_type
			(
				model_key               ,
				data_type_key           ,
				collection_type         ,
				collection_unique       ,
				collection_min          ,
				collection_max          ,
				element_data_type_key   ,
				type_spec_notation      ,
				type_spec_specification
			)
		VALUES
			(
				$1,
				$2,
				$3,
				$4,
				$5,
				$6,
				$7,
				$8,
				$9
			)`,
		modelKey,
		dataTypeKey,
		dataType.CollectionType,
		dataType.CollectionUnique,
		collectionMinForDB(dataType.CollectionMin),
		dataType.CollectionMax,
		elementDataTypeKey,
		tsNotation,
		tsSpecification)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// UpdateDataType updates a data type in the database.
func UpdateDataType(dbOrTx DbOrTx, modelKey string, dataType model_data_type.DataType) (err error) {
	dataTypeKey := dataType.Key.String()

	// Extract type spec fields.
	var tsNotation *string
	var tsSpecification *string
	if dataType.TypeSpec != nil {
		tsNotation = &dataType.TypeSpec.Notation
		tsSpecification = &dataType.TypeSpec.Specification
	}

	var elementDataTypeKey *string
	if dataType.ElementDataType != nil && dataType.ElementDataType.Key.KeyType != "" {
		ks := dataType.ElementDataType.Key.String()
		elementDataTypeKey = &ks
	}

	// Update the data.
	err = dbExec(dbOrTx, `
		UPDATE data_type
		SET
			collection_type          = $3,
			collection_unique        = $4,
			collection_min           = $5,
			collection_max           = $6,
			element_data_type_key    = $7,
			type_spec_notation       = $8,
			type_spec_specification  = $9
		WHERE
			data_type_key = $2
		AND
			model_key = $1`,
		modelKey,
		dataTypeKey,
		dataType.CollectionType,
		dataType.CollectionUnique,
		collectionMinForDB(dataType.CollectionMin),
		dataType.CollectionMax,
		elementDataTypeKey,
		tsNotation,
		tsSpecification)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// DeleteDataType deletes a data type from the database.
func DeleteDataType(dbOrTx DbOrTx, modelKey, dataTypeKey string) (err error) {
	// Keys should be preened so they collide correctly.

	// Delete the data.
	err = dbExec(dbOrTx, `
		DELETE FROM data_type
		WHERE
			data_type_key = $2
		AND
			model_key = $1`,
		modelKey,
		dataTypeKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryDataTypes lists all data types for a model from the database.
func QueryDataTypes(dbOrTx DbOrTx, modelKey string) (dataTypes []model_data_type.DataType, err error) {
	// Keys should be preened so they collide correctly.

	// Query the database.
	dataTypes = []model_data_type.DataType{}
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var dataType model_data_type.DataType
			if err = scanDataType(scanner, &dataType); err != nil {
				return err
			}
			dataTypes = append(dataTypes, dataType)
			return nil
		},
		`SELECT
			data_type_key           ,
			collection_type         ,
			collection_unique       ,
			collection_min          ,
			collection_max          ,
			element_data_type_key   ,
			type_spec_notation      ,
			type_spec_specification
		FROM
			data_type
		WHERE
			model_key = $1
		ORDER BY data_type_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return dataTypes, nil
}

// BulkInsertDataTypes inserts multiple data types in a single SQL statement.
func BulkInsertDataTypes(dbOrTx DbOrTx, modelKey string, dataTypes []model_data_type.DataType) (err error) {
	if len(dataTypes) == 0 {
		return nil
	}

	// Keys should be preened so they collide correctly.

	// Prepare the args
	args := make([]any, 0, len(dataTypes)*9)
	valueStrings := make([]string, 0, len(dataTypes))
	for i, dt := range dataTypes {
		dataTypeKey := dt.Key.String()
		var tsNotation *string
		var tsSpecification *string
		if dt.TypeSpec != nil {
			tsNotation = &dt.TypeSpec.Notation
			tsSpecification = &dt.TypeSpec.Specification
		}
		var elementDataTypeKey *string
		if dt.ElementDataType != nil && dt.ElementDataType.Key.KeyType != "" {
			ks := dt.ElementDataType.Key.String()
			elementDataTypeKey = &ks
		}
		args = append(args, modelKey, dataTypeKey, dt.CollectionType, dt.CollectionUnique, collectionMinForDB(dt.CollectionMin), dt.CollectionMax, elementDataTypeKey, tsNotation, tsSpecification)
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", i*9+1, i*9+2, i*9+3, i*9+4, i*9+5, i*9+6, i*9+7, i*9+8, i*9+9))
	}

	// Build the query
	query := fmt.Sprintf(`
		INSERT INTO data_type
			(
				model_key               ,
				data_type_key           ,
				collection_type         ,
				collection_unique       ,
				collection_min          ,
				collection_max          ,
				element_data_type_key   ,
				type_spec_notation      ,
				type_spec_specification
			)
		VALUES %s`, strings.Join(valueStrings, ", "))

	// Execute
	err = dbExec(dbOrTx, query, args...)
	return errors.WithStack(err)
}
