package database

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_data_type"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanField(scanner Scanner, dataTypeKeyPtr *string, field *model_data_type.Field) (err error) {
	var fieldDataTypeKey string
	if err = scanner.Scan(
		dataTypeKeyPtr,
		&field.Name,
		&fieldDataTypeKey,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	// Set the FieldDataType to a partial DataType with just the key.
	field.FieldDataType = &model_data_type.DataType{Key: fieldDataTypeKey}

	return nil
}

// LoadDataTypeFields loads all fields for a data type from the database
func LoadDataTypeFields(dbOrTx DbOrTx, modelKey, dataTypeKey string) (fields map[string][]model_data_type.Field, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = preenKey(modelKey)
	if err != nil {
		return nil, err
	}
	dataTypeKey, err = preenKey(dataTypeKey)
	if err != nil {
		return nil, err
	}

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var parentDataTypeKey string
			var field model_data_type.Field
			if err = scanField(scanner, &parentDataTypeKey, &field); err != nil {
				return errors.WithStack(err)
			}
			if fields == nil {
				fields = map[string][]model_data_type.Field{}
			}
			fields[parentDataTypeKey] = append(fields[parentDataTypeKey], field)
			return nil
		},
		`SELECT
			data_type_key,
			name,
			field_data_type_key
		FROM
			data_type_field
		WHERE
			data_type_key = $2
		AND
			model_key = $1
		ORDER BY name`,
		modelKey,
		dataTypeKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Treat no rows found as record not found.
	if len(fields) == 0 {
		return nil, ErrNotFound
	}

	return fields, nil
}

// AddField adds a data type field to the database
func AddField(dbOrTx DbOrTx, modelKey, dataTypeKey string, field model_data_type.Field) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = preenKey(modelKey)
	if err != nil {
		return err
	}
	dataTypeKey, err = preenKey(dataTypeKey)
	if err != nil {
		return err
	}
	if field.FieldDataType == nil {
		return errors.New("FieldDataType cannot be nil")
	}
	fieldDataTypeKey, err := preenKey(field.FieldDataType.Key)
	if err != nil {
		return err
	}

	// Add to the database.
	_, err = dbExec(
		dbOrTx,
		`INSERT INTO data_type_field
			(
				model_key,
				data_type_key,
				name,
				field_data_type_key
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
		field.Name,
		fieldDataTypeKey)
	return err
}

// UpdateField updates a data type field in the database
func UpdateField(dbOrTx DbOrTx, modelKey, dataTypeKey string, field model_data_type.Field) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = preenKey(modelKey)
	if err != nil {
		return err
	}
	dataTypeKey, err = preenKey(dataTypeKey)
	if err != nil {
		return err
	}
	if field.FieldDataType == nil {
		return errors.New("FieldDataType cannot be nil")
	}
	fieldDataTypeKey, err := preenKey(field.FieldDataType.Key)
	if err != nil {
		return err
	}

	// Update the database.
	_, err = dbExec(
		dbOrTx,
		`UPDATE data_type_field SET
			field_data_type_key = $4
		WHERE
			data_type_key = $2
		AND
			model_key = $1
		AND
			name = $3`,
		modelKey,
		dataTypeKey,
		field.Name,
		fieldDataTypeKey)
	return err
}

// RemoveField removes a data type field from the database
func RemoveField(dbOrTx DbOrTx, modelKey, dataTypeKey, name string) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = preenKey(modelKey)
	if err != nil {
		return err
	}
	dataTypeKey, err = preenKey(dataTypeKey)
	if err != nil {
		return err
	}

	// Remove from the database.
	_, err = dbExec(
		dbOrTx,
		`DELETE FROM data_type_field
		WHERE
			data_type_key = $2
		AND
			model_key = $1
		AND
			name = $3`,
		modelKey,
		dataTypeKey,
		name)
	return err
}

// QueryFields loads all data type fields for a model from the database
func QueryFields(dbOrTx DbOrTx, modelKey string) (fields map[string][]model_data_type.Field, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = preenKey(modelKey)
	if err != nil {
		return nil, err
	}

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var parentDataTypeKey string
			var field model_data_type.Field
			if err = scanField(scanner, &parentDataTypeKey, &field); err != nil {
				return errors.WithStack(err)
			}
			if fields == nil {
				fields = map[string][]model_data_type.Field{}
			}
			fields[parentDataTypeKey] = append(fields[parentDataTypeKey], field)
			return nil
		},
		`SELECT
			data_type_key,
			name,
			field_data_type_key
		FROM
			data_type_field
		WHERE
			model_key = $1
		ORDER BY data_type_key, name`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return fields, nil
}

// BulkInsertFields inserts multiple fields in a single SQL statement.
func BulkInsertFields(dbOrTx DbOrTx, modelKey string, fieldMap map[string][]model_data_type.Field) (err error) {
	totalFields := 0
	for _, fields := range fieldMap {
		totalFields += len(fields)
	}
	if totalFields == 0 {
		return nil
	}

	// Keys should be preened so they collide correctly.
	modelKey, err = preenKey(modelKey)
	if err != nil {
		return err
	}

	// Prepare the args
	args := make([]interface{}, 0, totalFields*4)
	valueStrings := make([]string, 0, totalFields)
	i := 0
	for dataTypeKey, fields := range fieldMap {
		dataTypeKey, err = preenKey(dataTypeKey)
		if err != nil {
			return err
		}
		for _, field := range fields {
			if field.FieldDataType == nil {
				return errors.New("FieldDataType cannot be nil")
			}
			fieldDataTypeKey, err := preenKey(field.FieldDataType.Key)
			if err != nil {
				return err
			}
			args = append(args, modelKey, dataTypeKey, field.Name, fieldDataTypeKey)
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d)", i*4+1, i*4+2, i*4+3, i*4+4))
			i++
		}
	}

	// Build the query
	query := fmt.Sprintf(`
		INSERT INTO data_type_field
			(
				model_key,
				data_type_key,
				name,
				field_data_type_key
			)
		VALUES %s`, strings.Join(valueStrings, ", "))

	// Execute
	_, err = dbExec(dbOrTx, query, args...)
	return errors.WithStack(err)
}
