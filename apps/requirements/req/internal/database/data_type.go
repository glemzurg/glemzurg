package database

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/data_type"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanDataType(scanner Scanner, dataType *data_type.DataType) (err error) {

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
func LoadDataType(dbOrTx DbOrTx, modelKey, dataTypeKey string) (dataType data_type.DataType, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return data_type.DataType{}, err
	}
	dataTypeKey, err = requirements.PreenKey(dataTypeKey)
	if err != nil {
		return data_type.DataType{}, err
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
		return data_type.DataType{}, errors.WithStack(err)
	}

	return dataType, nil
}

// AddDataType adds a data type to the database.
func AddDataType(dbOrTx DbOrTx, modelKey string, dataType data_type.DataType) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	dataTypeKey, err := requirements.PreenKey(dataType.Key)
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
func UpdateDataType(dbOrTx DbOrTx, modelKey string, dataType data_type.DataType) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	dataTypeKey, err := requirements.PreenKey(dataType.Key)
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
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	dataTypeKey, err = requirements.PreenKey(dataTypeKey)
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

// ListDataTypes lists all data types for a model from the database.
func ListDataTypes(dbOrTx DbOrTx, modelKey string) (dataTypes []data_type.DataType, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return nil, err
	}

	// Query the database.
	dataTypes = []data_type.DataType{}
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var dataType data_type.DataType
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
