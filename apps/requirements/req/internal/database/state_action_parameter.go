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
func scanActionParameter(scanner Scanner, actionKeyPtr *identity.Key, param *model_state.Parameter) (err error) {
	var actionKeyStr string
	var parameterKeyStr string // Read but not used on the struct (it's derived from Name via preenKey).
	var dataTypeRules sql.NullString
	var dataTypeKey sql.NullString

	if err = scanner.Scan(
		&actionKeyStr,
		&parameterKeyStr,
		&param.Name,
		&param.SortOrder,
		&dataTypeRules,
		&dataTypeKey,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	// Parse the action key string into an identity.Key.
	*actionKeyPtr, err = identity.ParseKey(actionKeyStr)
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

// LoadActionParameter loads an action parameter from the database.
func LoadActionParameter(dbOrTx DbOrTx, modelKey string, actionKey identity.Key, parameterKey string) (param model_state.Parameter, err error) {

	var loadedActionKey identity.Key

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanActionParameter(scanner, &loadedActionKey, &param); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			action_key      ,
			parameter_key   ,
			name            ,
			sort_order      ,
			data_type_rules ,
			data_type_key
		FROM
			action_parameter
		WHERE
			model_key     = $1
		AND
			action_key    = $2
		AND
			parameter_key = $3`,
		modelKey,
		actionKey.String(),
		parameterKey)
	if err != nil {
		return model_state.Parameter{}, errors.WithStack(err)
	}

	return param, nil
}

// AddActionParameter adds a single action parameter to the database.
func AddActionParameter(dbOrTx DbOrTx, modelKey string, actionKey identity.Key, param model_state.Parameter) (err error) {
	return AddActionParameters(dbOrTx, modelKey, map[identity.Key][]model_state.Parameter{
		actionKey: {param},
	})
}

// UpdateActionParameter updates an action parameter in the database.
func UpdateActionParameter(dbOrTx DbOrTx, modelKey string, actionKey identity.Key, param model_state.Parameter) (err error) {

	paramKey, err := preenKey(param.Name)
	if err != nil {
		return errors.Wrapf(err, "parameter name '%s'", param.Name)
	}

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			action_parameter
		SET
			name            = $4 ,
			sort_order      = $5 ,
			data_type_rules = $6 ,
			data_type_key   = $7
		WHERE
			model_key     = $1
		AND
			action_key    = $2
		AND
			parameter_key = $3`,
		modelKey,
		actionKey.String(),
		paramKey,
		param.Name,
		param.SortOrder,
		param.DataTypeRules,
		parameterDataTypeKey(param))
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveActionParameter deletes an action parameter from the database.
func RemoveActionParameter(dbOrTx DbOrTx, modelKey string, actionKey identity.Key, parameterKey string) (err error) {

	// Delete the data.
	_, err = dbExec(dbOrTx, `
		DELETE FROM
			action_parameter
		WHERE
			model_key     = $1
		AND
			action_key    = $2
		AND
			parameter_key = $3`,
		modelKey,
		actionKey.String(),
		parameterKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryActionParameters loads all action parameters from the database, grouped by action key.
func QueryActionParameters(dbOrTx DbOrTx, modelKey string) (params map[identity.Key][]model_state.Parameter, err error) {

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var actionKey identity.Key
			var param model_state.Parameter
			if err = scanActionParameter(scanner, &actionKey, &param); err != nil {
				return errors.WithStack(err)
			}
			if params == nil {
				params = map[identity.Key][]model_state.Parameter{}
			}
			actionParams := params[actionKey]
			actionParams = append(actionParams, param)
			params[actionKey] = actionParams
			return nil
		},
		`SELECT
			action_key      ,
			parameter_key   ,
			name            ,
			sort_order      ,
			data_type_rules ,
			data_type_key
		FROM
			action_parameter
		WHERE
			model_key = $1
		ORDER BY action_key, sort_order`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return params, nil
}

// AddActionParameters adds multiple action parameters to the database in a single insert.
func AddActionParameters(dbOrTx DbOrTx, modelKey string, params map[identity.Key][]model_state.Parameter) (err error) {
	// Count total parameters.
	count := 0
	for _, ps := range params {
		count += len(ps)
	}
	if count == 0 {
		return nil
	}

	// Build the bulk insert query.
	sqlQuery := `INSERT INTO action_parameter (model_key, action_key, parameter_key, name, sort_order, data_type_rules, data_type_key) VALUES `
	args := make([]interface{}, 0, count*7)
	i := 0
	for actionKey, paramList := range params {
		for _, param := range paramList {
			if i > 0 {
				sqlQuery += ", "
			}

			paramKey, err := preenKey(param.Name)
			if err != nil {
				return errors.Wrapf(err, "parameter name '%s'", param.Name)
			}

			base := i * 7
			sqlQuery += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4, base+5, base+6, base+7)
			args = append(args, modelKey, actionKey.String(), paramKey, param.Name, param.SortOrder, param.DataTypeRules, parameterDataTypeKey(param))
			i++
		}
	}

	_, err = dbExec(dbOrTx, sqlQuery, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
