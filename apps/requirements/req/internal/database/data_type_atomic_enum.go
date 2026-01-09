package database

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_data_type"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanAtomicEnum(scanner Scanner, dataTypeKeyPtr *string, atomicEnum *model_data_type.AtomicEnum) (err error) {
	if err = scanner.Scan(
		dataTypeKeyPtr,
		&atomicEnum.Value,
		&atomicEnum.SortOrder,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	return nil
}

// LoadAtomicEnums loads all atomic enums for a data type from the database
func LoadAtomicEnums(dbOrTx DbOrTx, modelKey, dataTypeKey string) (atomicEnums map[string][]model_data_type.AtomicEnum, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return nil, err
	}
	dataTypeKey, err = identity.PreenKey(dataTypeKey)
	if err != nil {
		return nil, err
	}

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var parentDataTypeKey string
			var atomicEnum model_data_type.AtomicEnum
			if err = scanAtomicEnum(scanner, &parentDataTypeKey, &atomicEnum); err != nil {
				return errors.WithStack(err)
			}
			if atomicEnums == nil {
				atomicEnums = map[string][]model_data_type.AtomicEnum{}
			}
			atomicEnums[parentDataTypeKey] = append(atomicEnums[parentDataTypeKey], atomicEnum)
			return nil
		},
		`SELECT
			data_type_key,
			value         ,
			sort_order
		FROM
			data_type_atomic_enum_value
		WHERE
			data_type_key = $2
		AND
			model_key = $1
		ORDER BY sort_order`,
		modelKey,
		dataTypeKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Treat missing enums as if a record was not found.
	if atomicEnums == nil {
		return nil, ErrNotFound
	}

	return atomicEnums, nil
}

// AddAtomicEnum adds an atomic enum to the database.
func AddAtomicEnum(dbOrTx DbOrTx, modelKey, dataTypeKey string, atomicEnum model_data_type.AtomicEnum) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}
	dataTypeKey, err = identity.PreenKey(dataTypeKey)
	if err != nil {
		return err
	}

	// Add the data.
	_, err = dbExec(dbOrTx, `
			INSERT INTO data_type_atomic_enum_value
				(
					model_key    ,
					data_type_key,
					value        ,
					sort_order
				)
			VALUES
				(
					$1,
					$2,
					$3,
					$4
				)`,
		modelKey,
		dataTypeKey,
		atomicEnum.Value,
		atomicEnum.SortOrder)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// UpdateAtomicEnum updates an atomic enum in the database.
func UpdateAtomicEnum(dbOrTx DbOrTx, modelKey, dataTypeKey, oldValue string, atomicEnum model_data_type.AtomicEnum) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}
	dataTypeKey, err = identity.PreenKey(dataTypeKey)
	if err != nil {
		return err
	}

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			data_type_atomic_enum_value
		SET
			value      = $4 ,
			sort_order = $5
		WHERE
			data_type_key = $2
		AND
			model_key = $1
		AND
			value = $3`,
		modelKey,
		dataTypeKey,
		oldValue,
		atomicEnum.Value,
		atomicEnum.SortOrder)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveAtomicEnum deletes an atomic enum from the database.
func RemoveAtomicEnum(dbOrTx DbOrTx, modelKey, dataTypeKey, value string) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}
	dataTypeKey, err = identity.PreenKey(dataTypeKey)
	if err != nil {
		return err
	}

	// Delete the data.
	_, err = dbExec(dbOrTx, `
		DELETE FROM
			data_type_atomic_enum_value
		WHERE
			data_type_key = $2
		AND
			model_key = $1
		AND
			value = $3`,
		modelKey,
		dataTypeKey,
		value)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryAtomicEnums loads all atomic enums from the database
func QueryAtomicEnums(dbOrTx DbOrTx, modelKey string) (atomicEnums map[string][]model_data_type.AtomicEnum, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return nil, err
	}

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var dataTypeKey string
			var atomicEnum model_data_type.AtomicEnum
			if err = scanner.Scan(
				&dataTypeKey,
				&atomicEnum.Value,
				&atomicEnum.SortOrder,
			); err != nil {
				return errors.WithStack(err)
			}
			if atomicEnums == nil {
				atomicEnums = map[string][]model_data_type.AtomicEnum{}
			}
			atomicEnums[dataTypeKey] = append(atomicEnums[dataTypeKey], atomicEnum)
			return nil
		},
		`SELECT
			data_type_key,
			value        ,
			sort_order
		FROM
			data_type_atomic_enum_value
		WHERE
			model_key = $1
		ORDER BY data_type_key, sort_order`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return atomicEnums, nil
}

// BulkInsertAtomicEnums inserts multiple atomic enums in a single SQL statement.
func BulkInsertAtomicEnums(dbOrTx DbOrTx, modelKey string, atomicEnums map[string][]model_data_type.AtomicEnum) (err error) {
	totalEnums := 0
	for _, enums := range atomicEnums {
		totalEnums += len(enums)
	}
	if totalEnums == 0 {
		return nil
	}

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}

	// Prepare the args
	args := make([]interface{}, 0, totalEnums*4)
	valueStrings := make([]string, 0, totalEnums)
	i := 0
	for dataTypeKey, enums := range atomicEnums {
		dataTypeKey, err = identity.PreenKey(dataTypeKey)
		if err != nil {
			return err
		}
		for _, enum := range enums {
			args = append(args, modelKey, dataTypeKey, enum.Value, enum.SortOrder)
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d)", i*4+1, i*4+2, i*4+3, i*4+4))
			i++
		}
	}

	// Build the query
	query := fmt.Sprintf(`
		INSERT INTO data_type_atomic_enum_value
			(
				model_key    ,
				data_type_key,
				value        ,
				sort_order
			)
		VALUES %s`, strings.Join(valueStrings, ", "))

	// Execute
	_, err = dbExec(dbOrTx, query, args...)
	return errors.WithStack(err)
}
