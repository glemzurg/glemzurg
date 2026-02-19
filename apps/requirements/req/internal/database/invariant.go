package database

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"

	"github.com/pkg/errors"
)

// LoadInvariant loads an invariant (as its Logic) from the database.
func LoadInvariant(dbOrTx DbOrTx, modelKey string, logicKey identity.Key) (logic model_logic.Logic, err error) {

	// Query the database by joining invariant with logic.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanLogic(scanner, &logic); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			l.logic_key     ,
			l.description   ,
			l.notation      ,
			l.specification
		FROM
			invariant i
		JOIN
			logic l ON l.model_key = i.model_key AND l.logic_key = i.logic_key
		WHERE
			i.logic_key = $2
		AND
			i.model_key = $1`,
		modelKey,
		logicKey.String())
	if err != nil {
		return model_logic.Logic{}, errors.WithStack(err)
	}

	return logic, nil
}

// AddInvariant adds an invariant to the database.
// This inserts the logic row and the invariant join row.
func AddInvariant(dbOrTx DbOrTx, modelKey string, logic model_logic.Logic) (err error) {
	return AddInvariants(dbOrTx, modelKey, []model_logic.Logic{logic})
}

// RemoveInvariant deletes an invariant from the database.
// This removes the invariant join row and the logic row.
func RemoveInvariant(dbOrTx DbOrTx, modelKey string, logicKey identity.Key) (err error) {

	// Delete the invariant join row first.
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

	// Delete the logic row.
	err = RemoveLogic(dbOrTx, modelKey, logicKey)
	if err != nil {
		return err
	}

	return nil
}

// QueryInvariants loads all invariants (as Logic structs) from the database for a given model.
func QueryInvariants(dbOrTx DbOrTx, modelKey string) (logics []model_logic.Logic, err error) {

	// Query the database by joining invariant with logic.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var logic model_logic.Logic
			if err = scanLogic(scanner, &logic); err != nil {
				return errors.WithStack(err)
			}
			logics = append(logics, logic)
			return nil
		},
		`SELECT
			l.logic_key     ,
			l.description   ,
			l.notation      ,
			l.specification
		FROM
			invariant i
		JOIN
			logic l ON l.model_key = i.model_key AND l.logic_key = i.logic_key
		WHERE
			i.model_key = $1
		ORDER BY l.logic_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return logics, nil
}

// AddInvariants adds multiple invariants to the database.
// This inserts the logic rows and the invariant join rows.
func AddInvariants(dbOrTx DbOrTx, modelKey string, logics []model_logic.Logic) (err error) {
	if len(logics) == 0 {
		return nil
	}

	// First, insert the logic rows.
	if err = AddLogics(dbOrTx, modelKey, logics); err != nil {
		return err
	}

	// Then, insert the invariant join rows.
	query := `INSERT INTO invariant (model_key, logic_key) VALUES `
	args := make([]interface{}, 0, len(logics)*2)
	for i, logic := range logics {
		if i > 0 {
			query += ", "
		}
		base := i * 2
		query += fmt.Sprintf("($%d, $%d)", base+1, base+2)
		args = append(args, modelKey, logic.Key.String())
	}

	_, err = dbExec(dbOrTx, query, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
