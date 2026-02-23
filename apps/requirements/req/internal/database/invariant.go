package database

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/pkg/errors"
)

// LoadInvariant loads an invariant logic key from the database.
func LoadInvariant(dbOrTx DbOrTx, modelKey string, logicKey identity.Key) (key identity.Key, err error) {

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
			invariant
		WHERE
			model_key = $1
		AND
			logic_key = $2`,
		modelKey,
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

// AddInvariant adds an invariant join row to the database.
// The logic row must already exist.
func AddInvariant(dbOrTx DbOrTx, modelKey string, logicKey identity.Key) (err error) {
	return AddInvariants(dbOrTx, modelKey, []identity.Key{logicKey})
}

// RemoveInvariant deletes an invariant join row from the database.
func RemoveInvariant(dbOrTx DbOrTx, modelKey string, logicKey identity.Key) (err error) {

	_, err = dbExec(dbOrTx, `
		DELETE FROM
			invariant
		WHERE
			model_key = $1
		AND
			logic_key = $2`,
		modelKey,
		logicKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryInvariants loads all invariant logic keys from the database for a given model.
func QueryInvariants(dbOrTx DbOrTx, modelKey string) (keys []identity.Key, err error) {

	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var logicKeyStr string
			if err = scanner.Scan(&logicKeyStr); err != nil {
				return errors.WithStack(err)
			}
			key, err := identity.ParseKey(logicKeyStr)
			if err != nil {
				return errors.WithStack(err)
			}
			keys = append(keys, key)
			return nil
		},
		`SELECT
			i.logic_key
		FROM
			invariant i
		JOIN
			logic l ON l.model_key = i.model_key AND l.logic_key = i.logic_key
		WHERE
			i.model_key = $1
		ORDER BY l.sort_order`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return keys, nil
}

// AddInvariants adds multiple invariant join rows to the database.
// The logic rows must already exist.
func AddInvariants(dbOrTx DbOrTx, modelKey string, logicKeys []identity.Key) (err error) {
	if len(logicKeys) == 0 {
		return nil
	}

	query := `INSERT INTO invariant (model_key, logic_key) VALUES `
	args := make([]interface{}, 0, len(logicKeys)*2)
	for i, logicKey := range logicKeys {
		if i > 0 {
			query += ", "
		}
		base := i * 2
		query += fmt.Sprintf("($%d, $%d)", base+1, base+2)
		args = append(args, modelKey, logicKey.String())
	}

	_, err = dbExec(dbOrTx, query, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
