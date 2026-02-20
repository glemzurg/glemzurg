package database

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/pkg/errors"
)

// Populate golang structs from a database row.
func scanActionSafety(scanner Scanner, actionKeyPtr *identity.Key, logicKeyPtr *identity.Key) (err error) {
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

// LoadActionSafety loads an action safety join row from the database.
func LoadActionSafety(dbOrTx DbOrTx, modelKey string, actionKey identity.Key, logicKey identity.Key) (key identity.Key, err error) {

	var loadedActionKey identity.Key

	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanActionSafety(scanner, &loadedActionKey, &key); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			action_key,
			logic_key
		FROM
			action_safety
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

// AddActionSafety adds a single action safety join row to the database.
// The logic row must already exist.
func AddActionSafety(dbOrTx DbOrTx, modelKey string, actionKey identity.Key, logicKey identity.Key) (err error) {
	return AddActionSafeties(dbOrTx, modelKey, map[identity.Key][]identity.Key{
		actionKey: {logicKey},
	})
}

// RemoveActionSafety deletes an action safety join row from the database.
func RemoveActionSafety(dbOrTx DbOrTx, modelKey string, actionKey identity.Key, logicKey identity.Key) (err error) {

	_, err = dbExec(dbOrTx, `
		DELETE FROM
			action_safety
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

// QueryActionSafeties loads all action safety logic keys from the database, grouped by action key.
func QueryActionSafeties(dbOrTx DbOrTx, modelKey string) (safeties map[identity.Key][]identity.Key, err error) {

	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var actionKey identity.Key
			var logicKey identity.Key
			if err = scanActionSafety(scanner, &actionKey, &logicKey); err != nil {
				return errors.WithStack(err)
			}
			if safeties == nil {
				safeties = map[identity.Key][]identity.Key{}
			}
			safeties[actionKey] = append(safeties[actionKey], logicKey)
			return nil
		},
		`SELECT
			action_key,
			logic_key
		FROM
			action_safety
		WHERE
			model_key = $1
		ORDER BY action_key, logic_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return safeties, nil
}

// AddActionSafeties adds multiple action safety join rows to the database in a single insert.
// The logic rows must already exist.
func AddActionSafeties(dbOrTx DbOrTx, modelKey string, safeties map[identity.Key][]identity.Key) (err error) {
	// Count total rows.
	count := 0
	for _, keys := range safeties {
		count += len(keys)
	}
	if count == 0 {
		return nil
	}

	query := `INSERT INTO action_safety (model_key, action_key, logic_key) VALUES `
	args := make([]interface{}, 0, count*3)
	i := 0
	for actionKey, logicKeys := range safeties {
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
