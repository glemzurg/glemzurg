package database

import (
	"database/sql"
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/lib/pq"
	"github.com/pkg/errors"
)

// globalFunctionRow holds the columns stored in the global_function table.
// The logic (Specification) is stored separately in the logic table and
// stitched together in top_level_requirements.go.
type globalFunctionRow struct {
	LogicKey   identity.Key
	Name       string
	Comment    string
	Parameters []string
}

// Populate a golang struct from a database row.
func scanGlobalFunction(scanner Scanner, row *globalFunctionRow) (err error) {
	var logicKeyStr string
	var comment sql.NullString

	if err = scanner.Scan(
		&logicKeyStr,
		&row.Name,
		&comment,
		pq.Array(&row.Parameters),
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	// Parse the key string into an identity.Key.
	row.LogicKey, err = identity.ParseKey(logicKeyStr)
	if err != nil {
		return err
	}

	// Handle nullable comment.
	if comment.Valid {
		row.Comment = comment.String
	}

	return nil
}

// LoadGlobalFunction loads a global function row from the database.
func LoadGlobalFunction(dbOrTx DbOrTx, modelKey string, logicKey identity.Key) (row globalFunctionRow, err error) {

	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanGlobalFunction(scanner, &row); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			logic_key  ,
			name       ,
			comment    ,
			parameters
		FROM
			global_function
		WHERE
			model_key = $1
		AND
			logic_key = $2`,
		modelKey,
		logicKey.String())
	if err != nil {
		return globalFunctionRow{}, errors.WithStack(err)
	}

	return row, nil
}

// AddGlobalFunction adds a global function row to the database.
// The logic row must already exist.
func AddGlobalFunction(dbOrTx DbOrTx, modelKey string, row globalFunctionRow) (err error) {
	return AddGlobalFunctions(dbOrTx, modelKey, []globalFunctionRow{row})
}

// UpdateGlobalFunction updates a global function row in the database.
func UpdateGlobalFunction(dbOrTx DbOrTx, modelKey string, row globalFunctionRow) (err error) {

	_, err = dbExec(dbOrTx, `
		UPDATE
			global_function
		SET
			name       = $3 ,
			comment    = $4 ,
			parameters = $5
		WHERE
			model_key = $1
		AND
			logic_key = $2`,
		modelKey,
		row.LogicKey.String(),
		row.Name,
		sql.NullString{String: row.Comment, Valid: row.Comment != ""},
		pq.Array(row.Parameters))
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveGlobalFunction deletes a global function row from the database.
func RemoveGlobalFunction(dbOrTx DbOrTx, modelKey string, logicKey identity.Key) (err error) {

	_, err = dbExec(dbOrTx, `
		DELETE FROM
			global_function
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

// QueryGlobalFunctions loads all global function rows from the database for a given model.
func QueryGlobalFunctions(dbOrTx DbOrTx, modelKey string) (rows []globalFunctionRow, err error) {

	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var row globalFunctionRow
			if err = scanGlobalFunction(scanner, &row); err != nil {
				return errors.WithStack(err)
			}
			rows = append(rows, row)
			return nil
		},
		`SELECT
			logic_key  ,
			name       ,
			comment    ,
			parameters
		FROM
			global_function
		WHERE
			model_key = $1
		ORDER BY logic_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return rows, nil
}

// AddGlobalFunctions adds multiple global function rows to the database in a single insert.
// The logic rows must already exist.
func AddGlobalFunctions(dbOrTx DbOrTx, modelKey string, rows []globalFunctionRow) (err error) {
	if len(rows) == 0 {
		return nil
	}

	query := `INSERT INTO global_function (model_key, logic_key, name, comment, parameters) VALUES `
	args := make([]interface{}, 0, len(rows)*5)
	for i, row := range rows {
		if i > 0 {
			query += ", "
		}
		base := i * 5
		query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4, base+5)
		args = append(args,
			modelKey,
			row.LogicKey.String(),
			row.Name,
			sql.NullString{String: row.Comment, Valid: row.Comment != ""},
			pq.Array(row.Parameters))
	}

	_, err = dbExec(dbOrTx, query, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
