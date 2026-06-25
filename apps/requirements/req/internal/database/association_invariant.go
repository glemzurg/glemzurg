package database

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/pkg/errors"
)

// QueryAssociationInvariants loads association invariant logic keys grouped by association key.
func QueryAssociationInvariants(dbOrTx DbOrTx, modelKey string) (map[identity.Key][]identity.Key, error) {
	result := make(map[identity.Key][]identity.Key)

	err := dbQuery(
		dbOrTx,
		func(scanner Scanner) error {
			var associationKeyStr, logicKeyStr string
			if err := scanner.Scan(&associationKeyStr, &logicKeyStr); err != nil {
				return errors.WithStack(err)
			}
			associationKey, err := identity.ParseKey(associationKeyStr)
			if err != nil {
				return errors.WithStack(err)
			}
			logicKey, err := identity.ParseKey(logicKeyStr)
			if err != nil {
				return errors.WithStack(err)
			}
			result[associationKey] = append(result[associationKey], logicKey)
			return nil
		},
		`SELECT
			ai.association_key, ai.logic_key
		FROM
			association_invariant ai
		JOIN
			logic l ON l.model_key = ai.model_key AND l.logic_key = ai.logic_key
		WHERE
			ai.model_key = $1
		ORDER BY ai.association_key, l.sort_order, ai.logic_key`,
		modelKey,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return result, nil
}

// AddAssociationInvariants inserts association invariant join rows.
func AddAssociationInvariants(dbOrTx DbOrTx, modelKey string, associationInvariants map[identity.Key][]identity.Key) error {
	totalRows := 0
	for _, logicKeys := range associationInvariants {
		totalRows += len(logicKeys)
	}
	if totalRows == 0 {
		return nil
	}

	var qb strings.Builder
	qb.WriteString(`INSERT INTO association_invariant (model_key, association_key, logic_key) VALUES `)
	args := make([]any, 0, totalRows*3)
	first := true
	argIdx := 0
	for associationKey, logicKeys := range associationInvariants {
		for _, logicKey := range logicKeys {
			if !first {
				qb.WriteString(", ")
			}
			first = false
			fmt.Fprintf(&qb, "($%d, $%d, $%d)", argIdx+1, argIdx+2, argIdx+3)
			args = append(args, modelKey, associationKey.String(), logicKey.String())
			argIdx += 3
		}
	}

	if err := dbExec(dbOrTx, qb.String(), args...); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
