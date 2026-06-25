package database

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/pkg/errors"
)

// Populate golang structs from a database row.
func scanActionParameterInvariant(scanner Scanner, actionKeyPtr *identity.Key, parameterKeyPtr *string, logicKeyPtr *identity.Key) (err error) {
	var actionKeyStr string
	var parameterKeyStr string
	var logicKeyStr string

	if err = scanner.Scan(&actionKeyStr, &parameterKeyStr, &logicKeyStr); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	*actionKeyPtr, err = identity.ParseKey(actionKeyStr)
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

// LoadActionParameterInvariant loads an action parameter invariant join row from the database.
func LoadActionParameterInvariant(dbOrTx DbOrTx, modelKey string, actionKey identity.Key, parameterKey string, logicKey identity.Key) (key identity.Key, err error) {
	var loadedActionKey identity.Key
	var loadedParameterKey string

	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanActionParameterInvariant(scanner, &loadedActionKey, &loadedParameterKey, &key); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			action_key,
			parameter_key,
			logic_key
		FROM
			action_parameter_invariant
		WHERE
			model_key     = $1
		AND
			action_key    = $2
		AND
			parameter_key = $3
		AND
			logic_key     = $4`,
		modelKey,
		actionKey.String(),
		parameterKey,
		logicKey.String())
	if err != nil {
		return identity.Key{}, errors.WithStack(err)
	}

	return key, nil
}

// AddActionParameterInvariant adds a single action parameter invariant join row to the database.
// The logic row must already exist.
func AddActionParameterInvariant(dbOrTx DbOrTx, modelKey string, actionKey identity.Key, parameterKey string, logicKey identity.Key) (err error) {
	return AddActionParameterInvariants(dbOrTx, modelKey, map[identity.Key]map[string][]identity.Key{
		actionKey: {parameterKey: {logicKey}},
	})
}

// RemoveActionParameterInvariant deletes an action parameter invariant join row from the database.
func RemoveActionParameterInvariant(dbOrTx DbOrTx, modelKey string, actionKey identity.Key, parameterKey string, logicKey identity.Key) (err error) {
	err = dbExec(dbOrTx, `
		DELETE FROM
			action_parameter_invariant
		WHERE
			model_key     = $1
		AND
			action_key    = $2
		AND
			parameter_key = $3
		AND
			logic_key     = $4`,
		modelKey,
		actionKey.String(),
		parameterKey,
		logicKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryActionParameterInvariants loads all action parameter invariant logic keys from the database,
// grouped by full parameter identity key. Within each parameter, results are ordered by logic sort_order.
func QueryActionParameterInvariants(dbOrTx DbOrTx, modelKey string) (result map[identity.Key][]identity.Key, err error) {
	result = make(map[identity.Key][]identity.Key)

	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var actionKey identity.Key
			var parameterSubKey string
			var logicKey identity.Key
			if err = scanActionParameterInvariant(scanner, &actionKey, &parameterSubKey, &logicKey); err != nil {
				return errors.WithStack(err)
			}
			paramKey, err := identity.NewParameterKey(actionKey, parameterSubKey)
			if err != nil {
				return errors.WithStack(err)
			}
			result[paramKey] = append(result[paramKey], logicKey)
			return nil
		},
		`SELECT
			api.action_key,
			api.parameter_key,
			api.logic_key
		FROM
			action_parameter_invariant api
		JOIN
			logic l ON l.model_key = api.model_key AND l.logic_key = api.logic_key
		WHERE
			api.model_key = $1
		ORDER BY api.action_key, api.parameter_key, l.sort_order, api.logic_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return result, nil
}

// AddActionParameterInvariants adds multiple action parameter invariant join rows to the database.
// The logic rows must already exist.
// The outer map is keyed by action key; inner maps are keyed by parameter subkey.
func AddActionParameterInvariants(dbOrTx DbOrTx, modelKey string, invariants map[identity.Key]map[string][]identity.Key) (err error) {
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
	qb.WriteString(`INSERT INTO action_parameter_invariant (model_key, action_key, parameter_key, logic_key) VALUES `)
	args := make([]any, 0, totalRows*4)
	i := 0
	for actionKey, paramInvariants := range invariants {
		for parameterKey, logicKeys := range paramInvariants {
			for _, logicKey := range logicKeys {
				if i > 0 {
					qb.WriteString(", ")
				}
				base := i * 4
				fmt.Fprintf(&qb, "($%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4)
				args = append(args, modelKey, actionKey.String(), parameterKey, logicKey.String())
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
