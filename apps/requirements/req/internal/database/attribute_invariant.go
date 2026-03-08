package database

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/pkg/errors"
)

// LoadAttributeInvariant loads an attribute invariant logic key from the database.
func LoadAttributeInvariant(dbOrTx DbOrTx, modelKey string, attributeKey identity.Key, logicKey identity.Key) (key identity.Key, err error) {
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
			attribute_invariant
		WHERE
			model_key = $1
		AND
			attribute_key = $2
		AND
			logic_key = $3`,
		modelKey,
		attributeKey.String(),
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

// AddAttributeInvariant adds an attribute invariant join row to the database.
// The logic row must already exist.
func AddAttributeInvariant(dbOrTx DbOrTx, modelKey string, attributeKey identity.Key, logicKey identity.Key) (err error) {
	return AddAttributeInvariants(dbOrTx, modelKey, map[identity.Key][]identity.Key{attributeKey: {logicKey}})
}

// RemoveAttributeInvariant deletes an attribute invariant join row from the database.
func RemoveAttributeInvariant(dbOrTx DbOrTx, modelKey string, attributeKey identity.Key, logicKey identity.Key) (err error) {
	err = dbExec(dbOrTx, `
		DELETE FROM
			attribute_invariant
		WHERE
			model_key = $1
		AND
			attribute_key = $2
		AND
			logic_key = $3`,
		modelKey,
		attributeKey.String(),
		logicKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryAttributeInvariants loads all attribute invariant logic keys from the database,
// grouped by attribute key. Within each attribute, results are ordered by logic sort_order.
func QueryAttributeInvariants(dbOrTx DbOrTx, modelKey string) (result map[identity.Key][]identity.Key, err error) {
	result = make(map[identity.Key][]identity.Key)

	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var attrKeyStr, logicKeyStr string
			if err = scanner.Scan(&attrKeyStr, &logicKeyStr); err != nil {
				return errors.WithStack(err)
			}
			attrKey, err := identity.ParseKey(attrKeyStr)
			if err != nil {
				return errors.WithStack(err)
			}
			logicKey, err := identity.ParseKey(logicKeyStr)
			if err != nil {
				return errors.WithStack(err)
			}
			result[attrKey] = append(result[attrKey], logicKey)
			return nil
		},
		`SELECT
			ai.attribute_key, ai.logic_key
		FROM
			attribute_invariant ai
		JOIN
			logic l ON l.model_key = ai.model_key AND l.logic_key = ai.logic_key
		WHERE
			ai.model_key = $1
		ORDER BY ai.attribute_key, l.sort_order, ai.logic_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return result, nil
}

// AddAttributeInvariants adds multiple attribute invariant join rows to the database.
// The logic rows must already exist.
// The map is keyed by attribute key, with each value being a slice of logic keys.
func AddAttributeInvariants(dbOrTx DbOrTx, modelKey string, attrInvariants map[identity.Key][]identity.Key) (err error) {
	// Count total rows.
	totalRows := 0
	for _, logicKeys := range attrInvariants {
		totalRows += len(logicKeys)
	}
	if totalRows == 0 {
		return nil
	}

	var queryBuilder strings.Builder
	queryBuilder.WriteString(`INSERT INTO attribute_invariant (model_key, attribute_key, logic_key) VALUES `)
	args := make([]any, 0, totalRows*3)
	first := true
	argIdx := 0
	for attrKey, logicKeys := range attrInvariants {
		for _, logicKey := range logicKeys {
			if !first {
				queryBuilder.WriteString(", ")
			}
			first = false
			queryBuilder.WriteString(fmt.Sprintf("($%d, $%d, $%d)", argIdx+1, argIdx+2, argIdx+3))

			args = append(args, modelKey, attrKey.String(), logicKey.String())
			argIdx += 3
		}
	}

	err = dbExec(dbOrTx, queryBuilder.String(), args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
