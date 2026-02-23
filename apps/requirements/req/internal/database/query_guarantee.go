package database

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/pkg/errors"
)

// Populate golang structs from a database row.
func scanQueryGuarantee(scanner Scanner, queryKeyPtr *identity.Key, logicKeyPtr *identity.Key) (err error) {
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

// LoadQueryGuarantee loads a query guarantee join row from the database.
func LoadQueryGuarantee(dbOrTx DbOrTx, modelKey string, queryKey identity.Key, logicKey identity.Key) (key identity.Key, err error) {

	var loadedQueryKey identity.Key

	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanQueryGuarantee(scanner, &loadedQueryKey, &key); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			query_key,
			logic_key
		FROM
			query_guarantee
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

// AddQueryGuarantee adds a single query guarantee join row to the database.
// The logic row must already exist.
func AddQueryGuarantee(dbOrTx DbOrTx, modelKey string, queryKey identity.Key, logicKey identity.Key) (err error) {
	return AddQueryGuarantees(dbOrTx, modelKey, map[identity.Key][]identity.Key{
		queryKey: {logicKey},
	})
}

// RemoveQueryGuarantee deletes a query guarantee join row from the database.
func RemoveQueryGuarantee(dbOrTx DbOrTx, modelKey string, queryKey identity.Key, logicKey identity.Key) (err error) {

	_, err = dbExec(dbOrTx, `
		DELETE FROM
			query_guarantee
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

// QueryQueryGuarantees loads all query guarantee logic keys from the database, grouped by query key.
func QueryQueryGuarantees(dbOrTx DbOrTx, modelKey string) (guarantees map[identity.Key][]identity.Key, err error) {

	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var queryKey identity.Key
			var logicKey identity.Key
			if err = scanQueryGuarantee(scanner, &queryKey, &logicKey); err != nil {
				return errors.WithStack(err)
			}
			if guarantees == nil {
				guarantees = map[identity.Key][]identity.Key{}
			}
			guarantees[queryKey] = append(guarantees[queryKey], logicKey)
			return nil
		},
		`SELECT
			qg.query_key,
			qg.logic_key
		FROM
			query_guarantee qg
		JOIN
			logic l ON l.model_key = qg.model_key AND l.logic_key = qg.logic_key
		WHERE
			qg.model_key = $1
		ORDER BY qg.query_key, l.sort_order`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return guarantees, nil
}

// AddQueryGuarantees adds multiple query guarantee join rows to the database in a single insert.
// The logic rows must already exist.
func AddQueryGuarantees(dbOrTx DbOrTx, modelKey string, guarantees map[identity.Key][]identity.Key) (err error) {
	// Count total rows.
	count := 0
	for _, keys := range guarantees {
		count += len(keys)
	}
	if count == 0 {
		return nil
	}

	query := `INSERT INTO query_guarantee (model_key, query_key, logic_key) VALUES `
	args := make([]interface{}, 0, count*3)
	i := 0
	for queryKey, logicKeys := range guarantees {
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
