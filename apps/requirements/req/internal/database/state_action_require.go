package database

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/pkg/errors"
)

// Populate golang structs from a database row.
func scanActionRequire(scanner Scanner, actionKeyPtr *identity.Key, logicKeyPtr *identity.Key) (err error) {
	var actionKeyStr string
	var logicKeyStr string

	if err = scanner.Scan(&actionKeyStr, &logicKeyStr); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	*actionKeyPtr, err = identity.ParseKey(actionKeyStr)
	if err != nil {
		return err
	}

	*logicKeyPtr, err = identity.ParseKey(logicKeyStr)
	if err != nil {
		return err
	}

	return nil
}

// LoadActionRequire loads an action require join row from the database.
func LoadActionRequire(dbOrTx DbOrTx, modelKey string, actionKey identity.Key, logicKey identity.Key) (key identity.Key, err error) {

	var loadedActionKey identity.Key

	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanActionRequire(scanner, &loadedActionKey, &key); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			action_key,
			logic_key
		FROM
			action_require
		WHERE
			model_key  = $1
		AND
			action_key = $2
		AND
			logic_key  = $3`,
		modelKey,
		actionKey.String(),
		logicKey.String())
	if err != nil {
		return identity.Key{}, errors.WithStack(err)
	}

	return key, nil
}

// AddActionRequire adds a single action require join row to the database.
// The logic row must already exist.
func AddActionRequire(dbOrTx DbOrTx, modelKey string, actionKey identity.Key, logicKey identity.Key) (err error) {
	return AddActionRequires(dbOrTx, modelKey, map[identity.Key][]identity.Key{
		actionKey: {logicKey},
	})
}

// RemoveActionRequire deletes an action require join row from the database.
func RemoveActionRequire(dbOrTx DbOrTx, modelKey string, actionKey identity.Key, logicKey identity.Key) (err error) {

	_, err = dbExec(dbOrTx, `
		DELETE FROM
			action_require
		WHERE
			model_key  = $1
		AND
			action_key = $2
		AND
			logic_key  = $3`,
		modelKey,
		actionKey.String(),
		logicKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryActionRequires loads all action require logic keys from the database, grouped by action key.
func QueryActionRequires(dbOrTx DbOrTx, modelKey string) (requires map[identity.Key][]identity.Key, err error) {

	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var actionKey identity.Key
			var logicKey identity.Key
			if err = scanActionRequire(scanner, &actionKey, &logicKey); err != nil {
				return errors.WithStack(err)
			}
			if requires == nil {
				requires = map[identity.Key][]identity.Key{}
			}
			requires[actionKey] = append(requires[actionKey], logicKey)
			return nil
		},
		`SELECT
			ar.action_key,
			ar.logic_key
		FROM
			action_require ar
		JOIN
			logic l ON l.model_key = ar.model_key AND l.logic_key = ar.logic_key
		WHERE
			ar.model_key = $1
		ORDER BY ar.action_key, l.sort_order, ar.logic_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return requires, nil
}

// AddActionRequires adds multiple action require join rows to the database in a single insert.
// The logic rows must already exist.
func AddActionRequires(dbOrTx DbOrTx, modelKey string, requires map[identity.Key][]identity.Key) (err error) {
	// Count total rows.
	count := 0
	for _, keys := range requires {
		count += len(keys)
	}
	if count == 0 {
		return nil
	}

	query := `INSERT INTO action_require (model_key, action_key, logic_key) VALUES `
	args := make([]interface{}, 0, count*3)
	i := 0
	for actionKey, logicKeys := range requires {
		for _, logicKey := range logicKeys {
			if i > 0 {
				query += ", "
			}
			base := i * 3
			query += fmt.Sprintf("($%d, $%d, $%d)", base+1, base+2, base+3)
			args = append(args, modelKey, actionKey.String(), logicKey.String())
			i++
		}
	}

	_, err = dbExec(dbOrTx, query, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
