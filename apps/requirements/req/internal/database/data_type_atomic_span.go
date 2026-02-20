package database

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_data_type"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanAtomicSpan(scanner Scanner, dataTypeKeyPtr *string, atomicSpan *model_data_type.AtomicSpan) (err error) {
	if err = scanner.Scan(
		dataTypeKeyPtr,
		&atomicSpan.LowerType,
		&atomicSpan.LowerValue,
		&atomicSpan.LowerDenominator,
		&atomicSpan.HigherType,
		&atomicSpan.HigherValue,
		&atomicSpan.HigherDenominator,
		&atomicSpan.Units,
		&atomicSpan.Precision,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	return nil
}

// LoadAtomicSpan loads an atomic span from the database
func LoadAtomicSpan(dbOrTx DbOrTx, modelKey, dataTypeKey string) (parentDataTypePtr string, atomicSpan model_data_type.AtomicSpan, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = PreenKey(modelKey)
	if err != nil {
		return "", model_data_type.AtomicSpan{}, err
	}
	dataTypeKey, err = PreenKey(dataTypeKey)
	if err != nil {
		return "", model_data_type.AtomicSpan{}, err
	}

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanAtomicSpan(scanner, &parentDataTypePtr, &atomicSpan); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			data_type_key,
			lower_type,
			lower_value,
			lower_denominator,
			higher_type,
			higher_value,
			higher_denominator,
			units,
			precision
		FROM
			data_type_atomic_span
		WHERE
			data_type_key = $2
		AND
			model_key = $1`,
		modelKey,
		dataTypeKey)
	if err != nil {
		return "", model_data_type.AtomicSpan{}, errors.WithStack(err)
	}

	return parentDataTypePtr, atomicSpan, nil
}

// AddAtomicSpan adds an atomic span to the database
func AddAtomicSpan(dbOrTx DbOrTx, modelKey, dataTypeKey string, atomicSpan model_data_type.AtomicSpan) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = PreenKey(modelKey)
	if err != nil {
		return err
	}
	dataTypeKey, err = PreenKey(dataTypeKey)
	if err != nil {
		return err
	}

	// Add to the database.
	_, err = dbExec(
		dbOrTx,
		`INSERT INTO data_type_atomic_span
			(
				model_key,
				data_type_key,
				lower_type,
				lower_value,
				lower_denominator,
				higher_type,
				higher_value,
				higher_denominator,
				units,
				precision
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
				$9,
				$10
			)`,
		modelKey,
		dataTypeKey,
		atomicSpan.LowerType,
		atomicSpan.LowerValue,
		atomicSpan.LowerDenominator,
		atomicSpan.HigherType,
		atomicSpan.HigherValue,
		atomicSpan.HigherDenominator,
		atomicSpan.Units,
		atomicSpan.Precision)
	return err
}

// UpdateAtomicSpan updates an atomic span in the database
func UpdateAtomicSpan(dbOrTx DbOrTx, modelKey, dataTypeKey string, atomicSpan model_data_type.AtomicSpan) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = PreenKey(modelKey)
	if err != nil {
		return err
	}
	dataTypeKey, err = PreenKey(dataTypeKey)
	if err != nil {
		return err
	}

	// Update the database.
	_, err = dbExec(
		dbOrTx,
		`UPDATE data_type_atomic_span SET
			lower_type = $3,
			lower_value = $4,
			lower_denominator = $5,
			higher_type = $6,
			higher_value = $7,
			higher_denominator = $8,
			units = $9,
			precision = $10
		WHERE
			data_type_key = $2
		AND
			model_key = $1`,
		modelKey,
		dataTypeKey,
		atomicSpan.LowerType,
		atomicSpan.LowerValue,
		atomicSpan.LowerDenominator,
		atomicSpan.HigherType,
		atomicSpan.HigherValue,
		atomicSpan.HigherDenominator,
		atomicSpan.Units,
		atomicSpan.Precision)
	return err
}

// RemoveAtomicSpan removes an atomic span from the database
func RemoveAtomicSpan(dbOrTx DbOrTx, modelKey, dataTypeKey string) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = PreenKey(modelKey)
	if err != nil {
		return err
	}
	dataTypeKey, err = PreenKey(dataTypeKey)
	if err != nil {
		return err
	}

	// Remove from the database.
	_, err = dbExec(
		dbOrTx,
		`DELETE FROM data_type_atomic_span
		WHERE
			data_type_key = $2
		AND
			model_key = $1`,
		modelKey,
		dataTypeKey)
	return err
}

// QueryAtomicSpans loads all atomic spans for a model from the database
func QueryAtomicSpans(dbOrTx DbOrTx, modelKey string) (atomicSpans map[string]model_data_type.AtomicSpan, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = PreenKey(modelKey)
	if err != nil {
		return nil, err
	}

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var parentDataTypeKey string
			var atomicSpan model_data_type.AtomicSpan
			if err = scanAtomicSpan(scanner, &parentDataTypeKey, &atomicSpan); err != nil {
				return errors.WithStack(err)
			}
			if atomicSpans == nil {
				atomicSpans = map[string]model_data_type.AtomicSpan{}
			}
			atomicSpans[parentDataTypeKey] = atomicSpan
			return nil
		},
		`SELECT
			data_type_key,
			lower_type,
			lower_value,
			lower_denominator,
			higher_type,
			higher_value,
			higher_denominator,
			units,
			precision
		FROM
			data_type_atomic_span
		WHERE
			model_key = $1`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return atomicSpans, nil
}

// BulkInsertAtomicSpans inserts multiple atomic spans in a single SQL statement.
func BulkInsertAtomicSpans(dbOrTx DbOrTx, modelKey string, atomicSpans map[string]model_data_type.AtomicSpan) (err error) {
	if len(atomicSpans) == 0 {
		return nil
	}

	// Keys should be preened so they collide correctly.
	modelKey, err = PreenKey(modelKey)
	if err != nil {
		return err
	}

	// Prepare the args
	args := make([]interface{}, 0, len(atomicSpans)*10)
	valueStrings := make([]string, 0, len(atomicSpans))
	i := 0
	for dataTypeKey, span := range atomicSpans {
		dataTypeKey, err = PreenKey(dataTypeKey)
		if err != nil {
			return err
		}
		args = append(args, modelKey, dataTypeKey, span.LowerType, span.LowerValue, span.LowerDenominator, span.HigherType, span.HigherValue, span.HigherDenominator, span.Units, span.Precision)
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", i*10+1, i*10+2, i*10+3, i*10+4, i*10+5, i*10+6, i*10+7, i*10+8, i*10+9, i*10+10))
		i++
	}

	// Build the query
	query := fmt.Sprintf(`
		INSERT INTO data_type_atomic_span
			(
				model_key,
				data_type_key,
				lower_type,
				lower_value,
				lower_denominator,
				higher_type,
				higher_value,
				higher_denominator,
				units,
				precision
			)
		VALUES %s`, strings.Join(valueStrings, ", "))

	// Execute
	_, err = dbExec(dbOrTx, query, args...)
	return errors.WithStack(err)
}
