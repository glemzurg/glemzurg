package database

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/pkg/errors"
)

// LoadClassInvariant loads a class invariant logic key from the database.
func LoadClassInvariant(dbOrTx DbOrTx, modelKey string, classKey identity.Key, logicKey identity.Key) (key identity.Key, err error) {

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
			class_invariant
		WHERE
			model_key = $1
		AND
			class_key = $2
		AND
			logic_key = $3`,
		modelKey,
		classKey.String(),
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

// AddClassInvariant adds a class invariant join row to the database.
// The logic row must already exist.
func AddClassInvariant(dbOrTx DbOrTx, modelKey string, classKey identity.Key, logicKey identity.Key) (err error) {
	return AddClassInvariants(dbOrTx, modelKey, map[identity.Key][]identity.Key{classKey: {logicKey}})
}

// RemoveClassInvariant deletes a class invariant join row from the database.
func RemoveClassInvariant(dbOrTx DbOrTx, modelKey string, classKey identity.Key, logicKey identity.Key) (err error) {

	_, err = dbExec(dbOrTx, `
		DELETE FROM
			class_invariant
		WHERE
			model_key = $1
		AND
			class_key = $2
		AND
			logic_key = $3`,
		modelKey,
		classKey.String(),
		logicKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryClassInvariants loads all class invariant logic keys from the database,
// grouped by class key. Within each class, results are ordered by logic sort_order.
func QueryClassInvariants(dbOrTx DbOrTx, modelKey string) (result map[identity.Key][]identity.Key, err error) {

	result = make(map[identity.Key][]identity.Key)

	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var classKeyStr, logicKeyStr string
			if err = scanner.Scan(&classKeyStr, &logicKeyStr); err != nil {
				return errors.WithStack(err)
			}
			classKey, err := identity.ParseKey(classKeyStr)
			if err != nil {
				return errors.WithStack(err)
			}
			logicKey, err := identity.ParseKey(logicKeyStr)
			if err != nil {
				return errors.WithStack(err)
			}
			result[classKey] = append(result[classKey], logicKey)
			return nil
		},
		`SELECT
			ci.class_key, ci.logic_key
		FROM
			class_invariant ci
		JOIN
			logic l ON l.model_key = ci.model_key AND l.logic_key = ci.logic_key
		WHERE
			ci.model_key = $1
		ORDER BY ci.class_key, l.sort_order, ci.logic_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return result, nil
}

// AddClassInvariants adds multiple class invariant join rows to the database.
// The logic rows must already exist.
// The map is keyed by class key, with each value being a slice of logic keys.
func AddClassInvariants(dbOrTx DbOrTx, modelKey string, classInvariants map[identity.Key][]identity.Key) (err error) {
	// Count total rows.
	totalRows := 0
	for _, logicKeys := range classInvariants {
		totalRows += len(logicKeys)
	}
	if totalRows == 0 {
		return nil
	}

	query := `INSERT INTO class_invariant (model_key, class_key, logic_key) VALUES `
	args := make([]interface{}, 0, totalRows*3)
	first := true
	argIdx := 0
	for classKey, logicKeys := range classInvariants {
		for _, logicKey := range logicKeys {
			if !first {
				query += ", "
			}
			first = false
			query += fmt.Sprintf("($%d, $%d, $%d)", argIdx+1, argIdx+2, argIdx+3)
			args = append(args, modelKey, classKey.String(), logicKey.String())
			argIdx += 3
		}
	}

	_, err = dbExec(dbOrTx, query, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
