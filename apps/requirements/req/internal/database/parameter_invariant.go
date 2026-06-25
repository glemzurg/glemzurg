package database

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/pkg/errors"
)

// LoadParameterInvariant loads a parameter invariant logic key from the database.
func LoadParameterInvariant(dbOrTx DbOrTx, modelKey string, parameterKey identity.Key, logicKey identity.Key) (key identity.Key, err error) {
	var logicKeyStr string
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanner.Scan(&logicKeyStr); err != nil {
				if err.Error() == _POSTGRES_NOT_FOUND {
					err = ErrNotFound
				}
				return err
			}
			return nil
		},
		`SELECT
			logic_key
		FROM
			parameter_invariant
		WHERE
			model_key = $1
		AND
			parameter_key = $2
		AND
			logic_key = $3`,
		modelKey,
		parameterKey.String(),
		logicKey.String())
	if err != nil {
		return identity.Key{}, errors.WithStack(err)
	}

	key, err = identity.ParseKey(logicKeyStr)
	if err != nil {
		return identity.Key{}, errors.WithStack(err)
	}

	return key, nil
}

// AddParameterInvariant adds a parameter invariant join row to the database.
// The logic row must already exist.
func AddParameterInvariant(dbOrTx DbOrTx, modelKey string, parameterKey identity.Key, logicKey identity.Key) (err error) {
	return AddParameterInvariants(dbOrTx, modelKey, map[identity.Key][]identity.Key{parameterKey: {logicKey}})
}

// RemoveParameterInvariant deletes a parameter invariant join row from the database.
func RemoveParameterInvariant(dbOrTx DbOrTx, modelKey string, parameterKey identity.Key, logicKey identity.Key) (err error) {
	err = dbExec(dbOrTx, `
		DELETE FROM
			parameter_invariant
		WHERE
			model_key = $1
		AND
			parameter_key = $2
		AND
			logic_key = $3`,
		modelKey,
		parameterKey.String(),
		logicKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryParameterInvariants loads all parameter invariant logic keys from the database,
// grouped by parameter key. Within each parameter, results are ordered by logic sort_order.
func QueryParameterInvariants(dbOrTx DbOrTx, modelKey string) (result map[identity.Key][]identity.Key, err error) {
	result = make(map[identity.Key][]identity.Key)

	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var paramKeyStr, logicKeyStr string
			if err = scanner.Scan(&paramKeyStr, &logicKeyStr); err != nil {
				return errors.WithStack(err)
			}
			paramKey, err := identity.ParseKey(paramKeyStr)
			if err != nil {
				return errors.WithStack(err)
			}
			logicKey, err := identity.ParseKey(logicKeyStr)
			if err != nil {
				return errors.WithStack(err)
			}
			result[paramKey] = append(result[paramKey], logicKey)
			return nil
		},
		`SELECT
			pi.parameter_key, pi.logic_key
		FROM
			parameter_invariant pi
		JOIN
			logic l ON l.model_key = pi.model_key AND l.logic_key = pi.logic_key
		WHERE
			pi.model_key = $1
		ORDER BY pi.parameter_key, l.sort_order, pi.logic_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return result, nil
}

// AddParameterInvariants adds multiple parameter invariant join rows to the database.
// The logic rows must already exist.
// The map is keyed by parameter key, with each value being a slice of logic keys.
func AddParameterInvariants(dbOrTx DbOrTx, modelKey string, paramInvariants map[identity.Key][]identity.Key) (err error) {
	totalRows := 0
	for _, logicKeys := range paramInvariants {
		totalRows += len(logicKeys)
	}
	if totalRows == 0 {
		return nil
	}

	var queryBuilder strings.Builder
	queryBuilder.WriteString(`INSERT INTO parameter_invariant (model_key, parameter_key, logic_key) VALUES `)
	args := make([]any, 0, totalRows*3)
	first := true
	argIdx := 0
	for paramKey, logicKeys := range paramInvariants {
		for _, logicKey := range logicKeys {
			if !first {
				queryBuilder.WriteString(", ")
			}
			first = false
			fmt.Fprintf(&queryBuilder, "($%d, $%d, $%d)", argIdx+1, argIdx+2, argIdx+3)

			args = append(args, modelKey, paramKey.String(), logicKey.String())
			argIdx += 3
		}
	}

	err = dbExec(dbOrTx, queryBuilder.String(), args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
