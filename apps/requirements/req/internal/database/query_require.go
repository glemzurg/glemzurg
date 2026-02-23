package database

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/pkg/errors"
)

// Populate golang structs from a database row.
func scanQueryRequire(scanner Scanner, queryKeyPtr *identity.Key, logicKeyPtr *identity.Key) (err error) {
	var queryKeyStr string
	var logicKeyStr string

	if err = scanner.Scan(&queryKeyStr, &logicKeyStr); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	*queryKeyPtr, err = identity.ParseKey(queryKeyStr)
	if err != nil {
		return err
	}

	*logicKeyPtr, err = identity.ParseKey(logicKeyStr)
	if err != nil {
		return err
	}

	return nil
}

// LoadQueryRequire loads a query require join row from the database.
func LoadQueryRequire(dbOrTx DbOrTx, modelKey string, queryKey identity.Key, logicKey identity.Key) (key identity.Key, err error) {

	var loadedQueryKey identity.Key

	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanQueryRequire(scanner, &loadedQueryKey, &key); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			query_key,
			logic_key
		FROM
			query_require
		WHERE
			model_key = $1
		AND
			query_key = $2
		AND
			logic_key = $3`,
		modelKey,
		queryKey.String(),
		logicKey.String())
	if err != nil {
		return identity.Key{}, errors.WithStack(err)
	}

	return key, nil
}

// AddQueryRequire adds a single query require join row to the database.
// The logic row must already exist.
func AddQueryRequire(dbOrTx DbOrTx, modelKey string, queryKey identity.Key, logicKey identity.Key) (err error) {
	return AddQueryRequires(dbOrTx, modelKey, map[identity.Key][]identity.Key{
		queryKey: {logicKey},
	})
}

// RemoveQueryRequire deletes a query require join row from the database.
func RemoveQueryRequire(dbOrTx DbOrTx, modelKey string, queryKey identity.Key, logicKey identity.Key) (err error) {

	_, err = dbExec(dbOrTx, `
		DELETE FROM
			query_require
		WHERE
			model_key = $1
		AND
			query_key = $2
		AND
			logic_key = $3`,
		modelKey,
		queryKey.String(),
		logicKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryQueryRequires loads all query require logic keys from the database, grouped by query key.
func QueryQueryRequires(dbOrTx DbOrTx, modelKey string) (requires map[identity.Key][]identity.Key, err error) {

	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var queryKey identity.Key
			var logicKey identity.Key
			if err = scanQueryRequire(scanner, &queryKey, &logicKey); err != nil {
				return errors.WithStack(err)
			}
			if requires == nil {
				requires = map[identity.Key][]identity.Key{}
			}
			requires[queryKey] = append(requires[queryKey], logicKey)
			return nil
		},
		`SELECT
			qr.query_key,
			qr.logic_key
		FROM
			query_require qr
		JOIN
			logic l ON l.model_key = qr.model_key AND l.logic_key = qr.logic_key
		WHERE
			qr.model_key = $1
		ORDER BY qr.query_key, l.sort_order, qr.logic_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return requires, nil
}

// AddQueryRequires adds multiple query require join rows to the database in a single insert.
// The logic rows must already exist.
func AddQueryRequires(dbOrTx DbOrTx, modelKey string, requires map[identity.Key][]identity.Key) (err error) {
	// Count total rows.
	count := 0
	for _, keys := range requires {
		count += len(keys)
	}
	if count == 0 {
		return nil
	}

	query := `INSERT INTO query_require (model_key, query_key, logic_key) VALUES `
	args := make([]interface{}, 0, count*3)
	i := 0
	for queryKey, logicKeys := range requires {
		for _, logicKey := range logicKeys {
			if i > 0 {
				query += ", "
			}
			base := i * 3
			query += fmt.Sprintf("($%d, $%d, $%d)", base+1, base+2, base+3)
			args = append(args, modelKey, queryKey.String(), logicKey.String())
			i++
		}
	}

	_, err = dbExec(dbOrTx, query, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
