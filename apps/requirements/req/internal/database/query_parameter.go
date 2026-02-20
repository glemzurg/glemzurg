package database

import (
	"database/sql"
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"

	"github.com/pkg/errors"
)

// parameterDataTypeKey extracts the data type key string for database storage.
// Returns nil if no data type is set.
func parameterDataTypeKey(param model_state.Parameter) *string {
	if param.DataType != nil {
		s := param.DataType.Key
		return &s
	}
	return nil
}

// Populate a golang struct from a database row.
func scanQueryParameter(scanner Scanner, queryKeyPtr *identity.Key, param *model_state.Parameter) (err error) {
	var queryKeyStr string
	var parameterKeyStr string // Read but not used on the struct (it's derived from Name via preenKey).
	var dataTypeRules sql.NullString
	var dataTypeKey sql.NullString

	if err = scanner.Scan(
		&queryKeyStr,
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

	// Parse the query key string into an identity.Key.
	*queryKeyPtr, err = identity.ParseKey(queryKeyStr)
	if err != nil {
		return err
	}

	// Set nullable fields.
	if dataTypeRules.Valid {
		param.DataTypeRules = dataTypeRules.String
	}

	return nil
}

// LoadQueryParameter loads a query parameter from the database.
func LoadQueryParameter(dbOrTx DbOrTx, modelKey string, queryKey identity.Key, parameterKey string) (param model_state.Parameter, err error) {

	var loadedQueryKey identity.Key

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanQueryParameter(scanner, &loadedQueryKey, &param); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			query_key       ,
			parameter_key   ,
			name            ,
			sort_order      ,
			data_type_rules ,
			data_type_key
		FROM
			query_parameter
		WHERE
			model_key     = $1
		AND
			query_key     = $2
		AND
			parameter_key = $3`,
		modelKey,
		queryKey.String(),
		parameterKey)
	if err != nil {
		return model_state.Parameter{}, errors.WithStack(err)
	}

	return param, nil
}

// AddQueryParameter adds a single query parameter to the database.
func AddQueryParameter(dbOrTx DbOrTx, modelKey string, queryKey identity.Key, param model_state.Parameter) (err error) {
	return AddQueryParameters(dbOrTx, modelKey, map[identity.Key][]model_state.Parameter{
		queryKey: {param},
	})
}

// UpdateQueryParameter updates a query parameter in the database.
func UpdateQueryParameter(dbOrTx DbOrTx, modelKey string, queryKey identity.Key, param model_state.Parameter) (err error) {

	paramKey, err := preenKey(param.Name)
	if err != nil {
		return errors.Wrapf(err, "parameter name '%s'", param.Name)
	}

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			query_parameter
		SET
			name            = $4 ,
			sort_order      = $5 ,
			data_type_rules = $6 ,
			data_type_key   = $7
		WHERE
			model_key     = $1
		AND
			query_key     = $2
		AND
			parameter_key = $3`,
		modelKey,
		queryKey.String(),
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

// RemoveQueryParameter deletes a query parameter from the database.
func RemoveQueryParameter(dbOrTx DbOrTx, modelKey string, queryKey identity.Key, parameterKey string) (err error) {

	// Delete the data.
	_, err = dbExec(dbOrTx, `
		DELETE FROM
			query_parameter
		WHERE
			model_key     = $1
		AND
			query_key     = $2
		AND
			parameter_key = $3`,
		modelKey,
		queryKey.String(),
		parameterKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryQueryParameters loads all query parameters from the database, grouped by query key.
func QueryQueryParameters(dbOrTx DbOrTx, modelKey string) (params map[identity.Key][]model_state.Parameter, err error) {

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var queryKey identity.Key
			var param model_state.Parameter
			if err = scanQueryParameter(scanner, &queryKey, &param); err != nil {
				return errors.WithStack(err)
			}
			if params == nil {
				params = map[identity.Key][]model_state.Parameter{}
			}
			queryParams := params[queryKey]
			queryParams = append(queryParams, param)
			params[queryKey] = queryParams
			return nil
		},
		`SELECT
			query_key       ,
			parameter_key   ,
			name            ,
			sort_order      ,
			data_type_rules ,
			data_type_key
		FROM
			query_parameter
		WHERE
			model_key = $1
		ORDER BY query_key, sort_order`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return params, nil
}

// AddQueryParameters adds multiple query parameters to the database in a single insert.
func AddQueryParameters(dbOrTx DbOrTx, modelKey string, params map[identity.Key][]model_state.Parameter) (err error) {
	// Count total parameters.
	count := 0
	for _, ps := range params {
		count += len(ps)
	}
	if count == 0 {
		return nil
	}

	// Build the bulk insert query.
	sqlQuery := `INSERT INTO query_parameter (model_key, query_key, parameter_key, name, sort_order, data_type_rules, data_type_key) VALUES `
	args := make([]interface{}, 0, count*7)
	i := 0
	for queryKey, paramList := range params {
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
			args = append(args, modelKey, queryKey.String(), paramKey, param.Name, param.SortOrder, param.DataTypeRules, parameterDataTypeKey(param))
			i++
		}
	}

	_, err = dbExec(dbOrTx, sqlQuery, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
