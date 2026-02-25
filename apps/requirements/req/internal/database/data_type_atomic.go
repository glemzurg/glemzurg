package database

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_data_type"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanAtomic(scanner Scanner, dataTypeKeyPtr *string, atomic *model_data_type.Atomic) (err error) {
	if err = scanner.Scan(
		dataTypeKeyPtr,
		&atomic.ConstraintType,
		&atomic.Reference,
		&atomic.EnumOrdered,
		&atomic.ObjectClassKey,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	return nil
}

// LoadAtomic loads an atomic from the database
func LoadAtomic(dbOrTx DbOrTx, modelKey, dataTypeKey string) (parentDataTypePtr string, atomic model_data_type.Atomic, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = preenKey(modelKey)
	if err != nil {
		return "", model_data_type.Atomic{}, err
	}
	dataTypeKey, err = preenKey(dataTypeKey)
	if err != nil {
		return "", model_data_type.Atomic{}, err
	}

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanAtomic(scanner, &parentDataTypePtr, &atomic); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			data_type_key  ,
			constraint_type ,
			reference       ,
			enum_ordered    ,
			object_class_key
		FROM
			data_type_atomic
		WHERE
			data_type_key = $2
		AND
			model_key = $1`,
		modelKey,
		dataTypeKey)
	if err != nil {
		return "", model_data_type.Atomic{}, errors.WithStack(err)
	}

	return parentDataTypePtr, atomic, nil
}

// AddAtomic adds an atomic to the database.
func AddAtomic(dbOrTx DbOrTx, modelKey, dataTypeKey string, atomic model_data_type.Atomic) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = preenKey(modelKey)
	if err != nil {
		return err
	}
	dataTypeKey, err = preenKey(dataTypeKey)
	if err != nil {
		return err
	}

	// Add the data.
	_, err = dbExec(dbOrTx, `
			INSERT INTO data_type_atomic
				(
					model_key       ,
					data_type_key   ,
					constraint_type ,
					reference       ,
					enum_ordered    ,
					object_class_key
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
		atomic.ConstraintType,
		atomic.Reference,
		atomic.EnumOrdered,
		atomic.ObjectClassKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// UpdateAtomic updates an atomic in the database.
func UpdateAtomic(dbOrTx DbOrTx, modelKey, dataTypeKey string, atomic model_data_type.Atomic) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = preenKey(modelKey)
	if err != nil {
		return err
	}
	dataTypeKey, err = preenKey(dataTypeKey)
	if err != nil {
		return err
	}

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			data_type_atomic
		SET
			constraint_type  = $3 ,
			reference        = $4 ,
			enum_ordered     = $5 ,
			object_class_key = $6
		WHERE
			data_type_key = $2
		AND
			model_key = $1`,
		modelKey,
		dataTypeKey,
		atomic.ConstraintType,
		atomic.Reference,
		atomic.EnumOrdered,
		atomic.ObjectClassKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveAtomic deletes an atomic from the database.
func RemoveAtomic(dbOrTx DbOrTx, modelKey, dataTypeKey string) (err error) {

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
		DELETE FROM
			data_type_atomic
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

// QueryAtomics loads all atomics from the database
func QueryAtomics(dbOrTx DbOrTx, modelKey string) (atomics map[string]model_data_type.Atomic, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = preenKey(modelKey)
	if err != nil {
		return nil, err
	}

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var dataTypeKey string
			var atomic model_data_type.Atomic
			if err = scanAtomic(scanner, &dataTypeKey, &atomic); err != nil {
				return errors.WithStack(err)
			}
			if atomics == nil {
				atomics = map[string]model_data_type.Atomic{}
			}
			atomics[dataTypeKey] = atomic
			return nil
		},
		`SELECT
			data_type_key  ,
			constraint_type ,
			reference       ,
			enum_ordered    ,
			object_class_key
		FROM
			data_type_atomic
		WHERE
			model_key = $1
		ORDER BY data_type_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return atomics, nil
}

// BulkInsertAtomics inserts multiple atomics in a single SQL statement.
func BulkInsertAtomics(dbOrTx DbOrTx, modelKey string, atomics map[string]model_data_type.Atomic) (err error) {
	if len(atomics) == 0 {
		return nil
	}

	// Keys should be preened so they collide correctly.
	modelKey, err = preenKey(modelKey)
	if err != nil {
		return err
	}

	// Prepare the args
	args := make([]interface{}, 0, len(atomics)*6)
	valueStrings := make([]string, 0, len(atomics))
	i := 0
	for dataTypeKey, atomic := range atomics {
		dataTypeKey, err = preenKey(dataTypeKey)
		if err != nil {
			return err
		}
		args = append(args, modelKey, dataTypeKey, atomic.ConstraintType, atomic.Reference, atomic.EnumOrdered, atomic.ObjectClassKey)
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", i*6+1, i*6+2, i*6+3, i*6+4, i*6+5, i*6+6))
		i++
	}

	// Build the query
	query := fmt.Sprintf(`
		INSERT INTO data_type_atomic
			(
				model_key       ,
				data_type_key   ,
				constraint_type ,
				reference       ,
				enum_ordered    ,
				object_class_key
			)
		VALUES %s`, strings.Join(valueStrings, ", "))

	// Execute
	_, err = dbExec(dbOrTx, query, args...)
	return errors.WithStack(err)
}
