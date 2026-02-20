package database

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/pkg/errors"
)

// Populate golang structs from a database row.
func scanActionGuarantee(scanner Scanner, actionKeyPtr *identity.Key, logicKeyPtr *identity.Key) (err error) {
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

// LoadActionGuarantee loads an action guarantee join row from the database.
func LoadActionGuarantee(dbOrTx DbOrTx, modelKey string, actionKey identity.Key, logicKey identity.Key) (key identity.Key, err error) {

	var loadedActionKey identity.Key

	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanActionGuarantee(scanner, &loadedActionKey, &key); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			action_key,
			logic_key
		FROM
			action_guarantee
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

// AddActionGuarantee adds a single action guarantee join row to the database.
// The logic row must already exist.
func AddActionGuarantee(dbOrTx DbOrTx, modelKey string, actionKey identity.Key, logicKey identity.Key) (err error) {
	return AddActionGuarantees(dbOrTx, modelKey, map[identity.Key][]identity.Key{
		actionKey: {logicKey},
	})
}

// RemoveActionGuarantee deletes an action guarantee join row from the database.
func RemoveActionGuarantee(dbOrTx DbOrTx, modelKey string, actionKey identity.Key, logicKey identity.Key) (err error) {

	_, err = dbExec(dbOrTx, `
		DELETE FROM
			action_guarantee
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

// QueryActionGuarantees loads all action guarantee logic keys from the database, grouped by action key.
func QueryActionGuarantees(dbOrTx DbOrTx, modelKey string) (guarantees map[identity.Key][]identity.Key, err error) {

	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var actionKey identity.Key
			var logicKey identity.Key
			if err = scanActionGuarantee(scanner, &actionKey, &logicKey); err != nil {
				return errors.WithStack(err)
			}
			if guarantees == nil {
				guarantees = map[identity.Key][]identity.Key{}
			}
			guarantees[actionKey] = append(guarantees[actionKey], logicKey)
			return nil
		},
		`SELECT
			action_key,
			logic_key
		FROM
			action_guarantee
		WHERE
			model_key = $1
		ORDER BY action_key, logic_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return guarantees, nil
}

// AddActionGuarantees adds multiple action guarantee join rows to the database in a single insert.
// The logic rows must already exist.
func AddActionGuarantees(dbOrTx DbOrTx, modelKey string, guarantees map[identity.Key][]identity.Key) (err error) {
	// Count total rows.
	count := 0
	for _, keys := range guarantees {
		count += len(keys)
	}
	if count == 0 {
		return nil
	}

	query := `INSERT INTO action_guarantee (model_key, action_key, logic_key) VALUES `
	args := make([]interface{}, 0, count*3)
	i := 0
	for actionKey, logicKeys := range guarantees {
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
