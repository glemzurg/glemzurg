package database

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"

	"github.com/lib/pq"
	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanGlobalFunction(scanner Scanner, gf *model_logic.GlobalFunction) (err error) {
	var logicKeyStr string

	if err = scanner.Scan(
		&logicKeyStr,
		&gf.Name,
		pq.Array(&gf.Parameters),
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	// Parse the key string into an identity.Key.
	gf.Key, err = identity.ParseKey(logicKeyStr)
	if err != nil {
		return err
	}

	return nil
}

// LoadGlobalFunction loads a global function from the database.
// The returned GlobalFunction will not have Specification populated;
// that is stitched in top_level_requirements.go.
func LoadGlobalFunction(dbOrTx DbOrTx, modelKey string, logicKey identity.Key) (gf model_logic.GlobalFunction, err error) {

	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanGlobalFunction(scanner, &gf); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			logic_key  ,
			name       ,
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
		return model_logic.GlobalFunction{}, errors.WithStack(err)
	}

	return gf, nil
}

// AddGlobalFunction adds a global function row to the database.
// The logic row must already exist.
func AddGlobalFunction(dbOrTx DbOrTx, modelKey string, gf model_logic.GlobalFunction) (err error) {
	return AddGlobalFunctions(dbOrTx, modelKey, []model_logic.GlobalFunction{gf})
}

// UpdateGlobalFunction updates a global function row in the database.
func UpdateGlobalFunction(dbOrTx DbOrTx, modelKey string, gf model_logic.GlobalFunction) (err error) {

	_, err = dbExec(dbOrTx, `
		UPDATE
			global_function
		SET
			name       = $3 ,
			parameters = $4
		WHERE
			model_key = $1
		AND
			logic_key = $2`,
		modelKey,
		gf.Key.String(),
		gf.Name,
		pq.Array(gf.Parameters))
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

// QueryGlobalFunctions loads all global functions from the database for a given model.
// The returned GlobalFunctions will not have Specification populated;
// that is stitched in top_level_requirements.go.
func QueryGlobalFunctions(dbOrTx DbOrTx, modelKey string) (gfs []model_logic.GlobalFunction, err error) {

	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var gf model_logic.GlobalFunction
			if err = scanGlobalFunction(scanner, &gf); err != nil {
				return errors.WithStack(err)
			}
			gfs = append(gfs, gf)
			return nil
		},
		`SELECT
			logic_key  ,
			name       ,
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

	return gfs, nil
}

// AddGlobalFunctions adds multiple global function rows to the database in a single insert.
// The logic rows must already exist.
func AddGlobalFunctions(dbOrTx DbOrTx, modelKey string, gfs []model_logic.GlobalFunction) (err error) {
	if len(gfs) == 0 {
		return nil
	}

	query := `INSERT INTO global_function (model_key, logic_key, name, parameters) VALUES `
	args := make([]interface{}, 0, len(gfs)*4)
	for i, gf := range gfs {
		if i > 0 {
			query += ", "
		}
		base := i * 4
		query += fmt.Sprintf("($%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4)
		args = append(args,
			modelKey,
			gf.Key.String(),
			gf.Name,
			pq.Array(gf.Parameters))
	}

	_, err = dbExec(dbOrTx, query, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
