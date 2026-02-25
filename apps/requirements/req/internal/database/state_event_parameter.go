package database

import (
	"database/sql"
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanEventParameter(scanner Scanner, eventKeyPtr *identity.Key, param *model_state.Parameter, sortOrder *int) (err error) {
	var eventKeyStr string
	var parameterKeyStr string // Read but not used on the struct (it's derived from Name via preenKey).
	var dataTypeRules sql.NullString
	var dataTypeKey sql.NullString

	if err = scanner.Scan(
		&eventKeyStr,
		&parameterKeyStr,
		&param.Name,
		sortOrder,
		&dataTypeRules,
		&dataTypeKey,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	// Parse the event key string into an identity.Key.
	*eventKeyPtr, err = identity.ParseKey(eventKeyStr)
	if err != nil {
		return err
	}

	// Set nullable fields.
	if dataTypeRules.Valid {
		param.DataTypeRules = dataTypeRules.String
	}

	// Create a stub DataType with just the key if present.
	// The full DataType is stitched in top_level_requirements.go from the data_type table.
	if dataTypeKey.Valid {
		param.DataType = &model_data_type.DataType{Key: dataTypeKey.String}
	}

	return nil
}

// LoadEventParameter loads an event parameter from the database.
func LoadEventParameter(dbOrTx DbOrTx, modelKey string, eventKey identity.Key, parameterKey string) (param model_state.Parameter, err error) {

	var loadedEventKey identity.Key
	var sortOrder int

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanEventParameter(scanner, &loadedEventKey, &param, &sortOrder); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			event_key       ,
			parameter_key   ,
			name            ,
			sort_order      ,
			data_type_rules ,
			data_type_key
		FROM
			event_parameter
		WHERE
			model_key     = $1
		AND
			event_key     = $2
		AND
			parameter_key = $3`,
		modelKey,
		eventKey.String(),
		parameterKey)
	if err != nil {
		return model_state.Parameter{}, errors.WithStack(err)
	}

	return param, nil
}

// AddEventParameter adds a single event parameter to the database.
func AddEventParameter(dbOrTx DbOrTx, modelKey string, eventKey identity.Key, param model_state.Parameter) (err error) {
	return AddEventParameters(dbOrTx, modelKey, map[identity.Key][]model_state.Parameter{
		eventKey: {param},
	})
}

// UpdateEventParameter updates an event parameter in the database.
func UpdateEventParameter(dbOrTx DbOrTx, modelKey string, eventKey identity.Key, sortOrder int, param model_state.Parameter) (err error) {

	paramKey, err := preenKey(param.Name)
	if err != nil {
		return errors.Wrapf(err, "parameter name '%s'", param.Name)
	}

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			event_parameter
		SET
			name            = $4 ,
			sort_order      = $5 ,
			data_type_rules = $6 ,
			data_type_key   = $7
		WHERE
			model_key     = $1
		AND
			event_key     = $2
		AND
			parameter_key = $3`,
		modelKey,
		eventKey.String(),
		paramKey,
		param.Name,
		sortOrder,
		param.DataTypeRules,
		parameterDataTypeKey(param))
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveEventParameter deletes an event parameter from the database.
func RemoveEventParameter(dbOrTx DbOrTx, modelKey string, eventKey identity.Key, parameterKey string) (err error) {

	// Delete the data.
	_, err = dbExec(dbOrTx, `
		DELETE FROM
			event_parameter
		WHERE
			model_key     = $1
		AND
			event_key     = $2
		AND
			parameter_key = $3`,
		modelKey,
		eventKey.String(),
		parameterKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryEventParameters loads all event parameters from the database, grouped by event key.
func QueryEventParameters(dbOrTx DbOrTx, modelKey string) (params map[identity.Key][]model_state.Parameter, err error) {

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var eventKey identity.Key
			var param model_state.Parameter
			var sortOrder int
			if err = scanEventParameter(scanner, &eventKey, &param, &sortOrder); err != nil {
				return errors.WithStack(err)
			}
			if params == nil {
				params = map[identity.Key][]model_state.Parameter{}
			}
			eventParams := params[eventKey]
			eventParams = append(eventParams, param)
			params[eventKey] = eventParams
			return nil
		},
		`SELECT
			event_key       ,
			parameter_key   ,
			name            ,
			sort_order      ,
			data_type_rules ,
			data_type_key
		FROM
			event_parameter
		WHERE
			model_key = $1
		ORDER BY event_key, sort_order`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return params, nil
}

// AddEventParameters adds multiple event parameters to the database in a single insert.
func AddEventParameters(dbOrTx DbOrTx, modelKey string, params map[identity.Key][]model_state.Parameter) (err error) {
	// Count total parameters.
	count := 0
	for _, ps := range params {
		count += len(ps)
	}
	if count == 0 {
		return nil
	}

	// Build the bulk insert query.
	sqlQuery := `INSERT INTO event_parameter (model_key, event_key, parameter_key, name, sort_order, data_type_rules, data_type_key) VALUES `
	args := make([]interface{}, 0, count*7)
	i := 0
	for eventKey, paramList := range params {
		for paramIdx, param := range paramList {
			if i > 0 {
				sqlQuery += ", "
			}

			paramKey, err := preenKey(param.Name)
			if err != nil {
				return errors.Wrapf(err, "parameter name '%s'", param.Name)
			}

			base := i * 7
			sqlQuery += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4, base+5, base+6, base+7)
			args = append(args, modelKey, eventKey.String(), paramKey, param.Name, paramIdx, param.DataTypeRules, parameterDataTypeKey(param))
			i++
		}
	}

	_, err = dbExec(dbOrTx, sqlQuery, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
