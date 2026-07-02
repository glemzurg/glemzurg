package database

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

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

	// Parse the string column back into a typed identity.Key. The full DataType is
	// stitched in top_level_requirements.go from the data_type table; this stub
	// carries only the key.
	parsedKey, parseErr := identity.ParseKey(fieldDataTypeKey)
	if parseErr != nil {
		return errors.Wrapf(parseErr, "failed to parse field data type key '%s'", fieldDataTypeKey)
	}
	field.FieldDataType = &model_data_type.DataType{Key: parsedKey}

	return nil
}

// LoadDataTypeFields loads all fields for a data type from the database.
func LoadDataTypeFields(dbOrTx DbOrTx, modelKey, dataTypeKey string) (fields map[string][]model_data_type.Field, err error) {
	// Keys should be preened so they collide correctly.

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

// AddField adds a data type field to the database.
func AddField(dbOrTx DbOrTx, modelKey, dataTypeKey string, field model_data_type.Field) (err error) {
	if field.FieldDataType == nil {
		return errors.New("FieldDataType cannot be nil")
	}
	fieldDataTypeKey := field.FieldDataType.Key.String()

	// Add to the database.
	err = dbExec(
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

// UpdateField updates a data type field in the database.
func UpdateField(dbOrTx DbOrTx, modelKey, dataTypeKey string, field model_data_type.Field) (err error) {
	if field.FieldDataType == nil {
		return errors.New("FieldDataType cannot be nil")
	}
	fieldDataTypeKey := field.FieldDataType.Key.String()

	// Update the database.
	err = dbExec(
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

// RemoveField removes a data type field from the database.
func RemoveField(dbOrTx DbOrTx, modelKey, dataTypeKey, name string) (err error) {
	// Keys should be preened so they collide correctly.

	// Remove from the database.
	err = dbExec(
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

// QueryFields loads all data type fields for a model from the database.
func QueryFields(dbOrTx DbOrTx, modelKey string) (fields map[string][]model_data_type.Field, err error) {
	// Keys should be preened so they collide correctly.

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

	// Prepare the args
	args := make([]any, 0, totalFields*4)
	valueStrings := make([]string, 0, totalFields)
	i := 0
	for dataTypeKey, fields := range fieldMap {
		for _, field := range fields {
			if field.FieldDataType == nil {
				return errors.New("FieldDataType cannot be nil")
			}
			fieldDataTypeKey := field.FieldDataType.Key.String()
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
	err = dbExec(dbOrTx, query, args...)
	return errors.WithStack(err)
}
