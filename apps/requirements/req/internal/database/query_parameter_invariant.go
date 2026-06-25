package database

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/pkg/errors"
)

// Populate golang structs from a database row.
func scanQueryParameterInvariant(scanner Scanner, queryKeyPtr *identity.Key, parameterKeyPtr *string, logicKeyPtr *identity.Key) (err error) {
	var queryKeyStr string
	var parameterKeyStr string
	var logicKeyStr string

	if err = scanner.Scan(&queryKeyStr, &parameterKeyStr, &logicKeyStr); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	*queryKeyPtr, err = identity.ParseKey(queryKeyStr)
	if err != nil {
		return err
	}

	*parameterKeyPtr = parameterKeyStr

	*logicKeyPtr, err = identity.ParseKey(logicKeyStr)
	if err != nil {
		return err
	}

	return nil
}

// LoadQueryParameterInvariant loads a query parameter invariant join row from the database.
func LoadQueryParameterInvariant(dbOrTx DbOrTx, modelKey string, queryKey identity.Key, parameterKey string, logicKey identity.Key) (key identity.Key, err error) {
	var loadedQueryKey identity.Key
	var loadedParameterKey string

	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanQueryParameterInvariant(scanner, &loadedQueryKey, &loadedParameterKey, &key); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			query_key,
			parameter_key,
			logic_key
		FROM
			query_parameter_invariant
		WHERE
			model_key     = $1
		AND
			query_key     = $2
		AND
			parameter_key = $3
		AND
			logic_key     = $4`,
		modelKey,
		queryKey.String(),
		parameterKey,
		logicKey.String())
	if err != nil {
		return identity.Key{}, errors.WithStack(err)
	}

	return key, nil
}

// AddQueryParameterInvariant adds a single query parameter invariant join row to the database.
// The logic row must already exist.
func AddQueryParameterInvariant(dbOrTx DbOrTx, modelKey string, queryKey identity.Key, parameterKey string, logicKey identity.Key) (err error) {
	return AddQueryParameterInvariants(dbOrTx, modelKey, map[identity.Key]map[string][]identity.Key{
		queryKey: {parameterKey: {logicKey}},
	})
}

// RemoveQueryParameterInvariant deletes a query parameter invariant join row from the database.
func RemoveQueryParameterInvariant(dbOrTx DbOrTx, modelKey string, queryKey identity.Key, parameterKey string, logicKey identity.Key) (err error) {
	err = dbExec(dbOrTx, `
		DELETE FROM
			query_parameter_invariant
		WHERE
			model_key     = $1
		AND
			query_key     = $2
		AND
			parameter_key = $3
		AND
			logic_key     = $4`,
		modelKey,
		queryKey.String(),
		parameterKey,
		logicKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryQueryParameterInvariants loads all query parameter invariant logic keys from the database,
// grouped by full parameter identity key. Within each parameter, results are ordered by logic sort_order.
func QueryQueryParameterInvariants(dbOrTx DbOrTx, modelKey string) (result map[identity.Key][]identity.Key, err error) {
	result = make(map[identity.Key][]identity.Key)

	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var queryKey identity.Key
			var parameterSubKey string
			var logicKey identity.Key
			if err = scanQueryParameterInvariant(scanner, &queryKey, &parameterSubKey, &logicKey); err != nil {
				return errors.WithStack(err)
			}
			paramKey, err := identity.NewParameterKey(queryKey, parameterSubKey)
			if err != nil {
				return errors.WithStack(err)
			}
			result[paramKey] = append(result[paramKey], logicKey)
			return nil
		},
		`SELECT
			qpi.query_key,
			qpi.parameter_key,
			qpi.logic_key
		FROM
			query_parameter_invariant qpi
		JOIN
			logic l ON l.model_key = qpi.model_key AND l.logic_key = qpi.logic_key
		WHERE
			qpi.model_key = $1
		ORDER BY qpi.query_key, qpi.parameter_key, l.sort_order, qpi.logic_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return result, nil
}

// AddQueryParameterInvariants adds multiple query parameter invariant join rows to the database.
// The logic rows must already exist.
// The outer map is keyed by query key; inner maps are keyed by parameter subkey.
func AddQueryParameterInvariants(dbOrTx DbOrTx, modelKey string, invariants map[identity.Key]map[string][]identity.Key) (err error) {
	totalRows := 0
	for _, paramInvariants := range invariants {
		for _, logicKeys := range paramInvariants {
			totalRows += len(logicKeys)
		}
	}
	if totalRows == 0 {
		return nil
	}

	var qb strings.Builder
	qb.WriteString(`INSERT INTO query_parameter_invariant (model_key, query_key, parameter_key, logic_key) VALUES `)
	args := make([]any, 0, totalRows*4)
	i := 0
	for queryKey, paramInvariants := range invariants {
		for parameterKey, logicKeys := range paramInvariants {
			for _, logicKey := range logicKeys {
				if i > 0 {
					qb.WriteString(", ")
				}
				base := i * 4
				fmt.Fprintf(&qb, "($%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4)
				args = append(args, modelKey, queryKey.String(), parameterKey, logicKey.String())
				i++
			}
		}
	}

	err = dbExec(dbOrTx, qb.String(), args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
