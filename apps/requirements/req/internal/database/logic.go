package database

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanLogic(scanner Scanner, logic *model_logic.Logic) (err error) {
	var keyStr string

	if err = scanner.Scan(
		&keyStr,
		&logic.Description,
		&logic.Notation,
		&logic.Specification,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	// Parse the key string into an identity.Key.
	logic.Key, err = identity.ParseKey(keyStr)
	if err != nil {
		return err
	}

	return nil
}

// LoadLogic loads a logic from the database.
func LoadLogic(dbOrTx DbOrTx, modelKey string, logicKey identity.Key) (logic model_logic.Logic, err error) {

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanLogic(scanner, &logic); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			logic_key     ,
			description   ,
			notation      ,
			specification
		FROM
			logic
		WHERE
			logic_key = $2
		AND
			model_key = $1`,
		modelKey,
		logicKey.String())
	if err != nil {
		return model_logic.Logic{}, errors.WithStack(err)
	}

	return logic, nil
}

// AddLogic adds a logic to the database.
func AddLogic(dbOrTx DbOrTx, modelKey string, logic model_logic.Logic) (err error) {
	return AddLogics(dbOrTx, modelKey, []model_logic.Logic{logic})
}

// UpdateLogic updates a logic in the database.
func UpdateLogic(dbOrTx DbOrTx, modelKey string, logic model_logic.Logic) (err error) {

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			logic
		SET
			description   = $3 ,
			notation      = $4 ,
			specification = $5
		WHERE
			model_key = $1
		AND
			logic_key = $2`,
		modelKey,
		logic.Key.String(),
		logic.Description,
		logic.Notation,
		logic.Specification)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveLogic deletes a logic from the database.
func RemoveLogic(dbOrTx DbOrTx, modelKey string, logicKey identity.Key) (err error) {

	// Delete the data.
	_, err = dbExec(dbOrTx, `
		DELETE FROM
			logic
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

// QueryLogics loads all logics from the database for a given model.
func QueryLogics(dbOrTx DbOrTx, modelKey string) (logics []model_logic.Logic, err error) {

	// Query the database.
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
			logic_key     ,
			description   ,
			notation      ,
			specification
		FROM
			logic
		WHERE
			model_key = $1
		ORDER BY logic_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return logics, nil
}

// AddLogics adds multiple logics to the database in a single insert.
func AddLogics(dbOrTx DbOrTx, modelKey string, logics []model_logic.Logic) (err error) {
	if len(logics) == 0 {
		return nil
	}

	// Build the bulk insert query.
	query := `INSERT INTO logic (model_key, logic_key, description, notation, specification) VALUES `
	args := make([]interface{}, 0, len(logics)*5)
	for i, logic := range logics {
		if i > 0 {
			query += ", "
		}
		base := i * 5
		query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4, base+5)
		args = append(args, modelKey, logic.Key.String(), logic.Description, logic.Notation, logic.Specification)
	}

	_, err = dbExec(dbOrTx, query, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
