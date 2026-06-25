package database

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/pkg/errors"
)

// parameterDataTypeKey extracts the data type key string for database storage.
// Returns nil if no data type is set.
func parameterDataTypeKey(param model_state.Parameter) *string {
	if param.DataType != nil {
		s := param.DataType.Key.String()
		return &s
	}
	return nil
}

// Populate a golang struct from a database row.
func scanQueryParameter(scanner Scanner, queryKeyPtr *identity.Key, param *model_state.Parameter, sortOrder *int) (err error) {
	var queryKeyStr string
	var parameterKeyStr string // Read but not used on the struct.
	var dataTypeRules sql.NullString
	var dataTypeKey sql.NullString

	if err = scanner.Scan(
		&queryKeyStr,
		&parameterKeyStr,
		&param.Name,
		sortOrder,
		&dataTypeRules,
		&dataTypeKey,
		&param.Nullable,
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

	param.Key, err = identity.NewParameterKey(*queryKeyPtr, parameterKeyStr)
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
		parsedKey, parseErr := identity.ParseKey(dataTypeKey.String)
		if parseErr != nil {
			return errors.Wrapf(parseErr, "failed to parse data type key '%s'", dataTypeKey.String)
		}
		param.DataType = &model_data_type.DataType{Key: parsedKey}
	}

	return nil
}

// LoadQueryParameter loads a query parameter from the database.
func LoadQueryParameter(dbOrTx DbOrTx, modelKey string, queryKey identity.Key, parameterKey string) (param model_state.Parameter, err error) {
	var loadedQueryKey identity.Key
	var sortOrder int

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanQueryParameter(scanner, &loadedQueryKey, &param, &sortOrder); err != nil {
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
			data_type_key   ,
			nullable
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
func UpdateQueryParameter(dbOrTx DbOrTx, modelKey string, queryKey identity.Key, sortOrder int, param model_state.Parameter) (err error) {
	paramKey := param.Key.SubKey

	// Update the data.
	err = dbExec(dbOrTx, `
		UPDATE
			query_parameter
		SET
			name            = $4 ,
			sort_order      = $5 ,
			data_type_rules = $6 ,
			data_type_key   = $7 ,
			nullable        = $8
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
		sortOrder,
		param.DataTypeRules,
		parameterDataTypeKey(param),
		param.Nullable)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveQueryParameter deletes a query parameter from the database.
func RemoveQueryParameter(dbOrTx DbOrTx, modelKey string, queryKey identity.Key, parameterKey string) (err error) {
	// Delete the data.
	err = dbExec(dbOrTx, `
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
			var sortOrder int
			if err = scanQueryParameter(scanner, &queryKey, &param, &sortOrder); err != nil {
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
			data_type_key   ,
			nullable
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
	var qb strings.Builder
	qb.WriteString(`INSERT INTO query_parameter (model_key, query_key, parameter_key, name, sort_order, data_type_rules, data_type_key, nullable) VALUES `)
	args := make([]any, 0, count*8)
	i := 0
	for queryKey, paramList := range params {
		for paramIdx, param := range paramList {
			if i > 0 {
				qb.WriteString(", ")
			}

			paramKey := param.Key.SubKey

			base := i * 8
			fmt.Fprintf(&qb, "($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8)
			args = append(args, modelKey, queryKey.String(), paramKey, param.Name, paramIdx, param.DataTypeRules, parameterDataTypeKey(param), param.Nullable)
			i++
		}
	}

	err = dbExec(dbOrTx, qb.String(), args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
