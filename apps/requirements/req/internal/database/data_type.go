package database

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_data_type"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanDataType(scanner Scanner, dataType *model_data_type.DataType) (err error) {

	if err = scanner.Scan(
		&dataType.Key,
		&dataType.CollectionType,
		&dataType.CollectionUnique,
		&dataType.CollectionMin,
		&dataType.CollectionMax,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	return nil
}

// LoadDataType loads a data type from the database
func LoadDataType(dbOrTx DbOrTx, modelKey, dataTypeKey string) (dataType model_data_type.DataType, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = preenKey(modelKey)
	if err != nil {
		return model_data_type.DataType{}, err
	}
	dataTypeKey, err = preenKey(dataTypeKey)
	if err != nil {
		return model_data_type.DataType{}, err
	}

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
			data_type_key     ,
			collection_type   ,
			collection_unique ,
			collection_min    ,
			collection_max
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

	// Keys should be preened so they collide correctly.
	modelKey, err = preenKey(modelKey)
	if err != nil {
		return err
	}
	dataTypeKey, err := preenKey(dataType.Key)
	if err != nil {
		return err
	}

	// Add the data.
	_, err = dbExec(dbOrTx, `
		INSERT INTO data_type
			(
				model_key         ,
				data_type_key     ,
				collection_type   ,
				collection_unique ,
				collection_min    ,
				collection_max
			)
		VALUES
			(
				$1,
				$2,
				$3,
				$4,
				$5,
				$6
			)`,
		modelKey,
		dataTypeKey,
		dataType.CollectionType,
		dataType.CollectionUnique,
		dataType.CollectionMin,
		dataType.CollectionMax)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// UpdateDataType updates a data type in the database.
func UpdateDataType(dbOrTx DbOrTx, modelKey string, dataType model_data_type.DataType) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = preenKey(modelKey)
	if err != nil {
		return err
	}
	dataTypeKey, err := preenKey(dataType.Key)
	if err != nil {
		return err
	}

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE data_type
		SET
			collection_type   = $3,
			collection_unique = $4,
			collection_min    = $5,
			collection_max    = $6
		WHERE
			data_type_key = $2
		AND
			model_key = $1`,
		modelKey,
		dataTypeKey,
		dataType.CollectionType,
		dataType.CollectionUnique,
		dataType.CollectionMin,
		dataType.CollectionMax)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// DeleteDataType deletes a data type from the database.
func DeleteDataType(dbOrTx DbOrTx, modelKey, dataTypeKey string) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = preenKey(modelKey)
	if err != nil {
		return err
	}
	dataTypeKey, err = preenKey(dataTypeKey)
	if err != nil {
		return err
	}

	// Delete the data.
	_, err = dbExec(dbOrTx, `
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
	modelKey, err = preenKey(modelKey)
	if err != nil {
		return nil, err
	}

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
			data_type_key     ,
			collection_type   ,
			collection_unique ,
			collection_min    ,
			collection_max
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
	modelKey, err = preenKey(modelKey)
	if err != nil {
		return err
	}

	// Prepare the args
	args := make([]interface{}, 0, len(dataTypes)*6)
	valueStrings := make([]string, 0, len(dataTypes))
	for i, dt := range dataTypes {
		dataTypeKey, err := preenKey(dt.Key)
		if err != nil {
			return err
		}
		args = append(args, modelKey, dataTypeKey, dt.CollectionType, dt.CollectionUnique, dt.CollectionMin, dt.CollectionMax)
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", i*6+1, i*6+2, i*6+3, i*6+4, i*6+5, i*6+6))
	}

	// Build the query
	query := fmt.Sprintf(`
		INSERT INTO data_type
			(
				model_key         ,
				data_type_key     ,
				collection_type   ,
				collection_unique ,
				collection_min    ,
				collection_max
			)
		VALUES %s`, strings.Join(valueStrings, ", "))

	// Execute
	_, err = dbExec(dbOrTx, query, args...)
	return errors.WithStack(err)
}
