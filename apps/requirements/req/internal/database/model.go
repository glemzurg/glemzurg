package database

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanModel(scanner Scanner, model *req_model.Model) (err error) {
	if err = scanner.Scan(
		&model.Key,
		&model.Name,
		&model.Details,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	return nil
}

// LoadModel loads a model from the database
func LoadModel(dbOrTx DbOrTx, modelKey string) (model req_model.Model, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return req_model.Model{}, err
	}

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanModel(scanner, &model); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			model_key   ,
			name        ,
			details
		FROM
			model
		WHERE
			model_key = $1`,
		modelKey)
	if err != nil {
		return req_model.Model{}, errors.WithStack(err)
	}

	return model, nil
}

// AddModel adds a model to the database.
func AddModel(dbOrTx DbOrTx, model req_model.Model) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err := identity.PreenKey(model.Key)
	if err != nil {
		return err
	}

	// Add the data.
	_, err = dbExec(dbOrTx, `
		INSERT INTO model
			(
				model_key ,
				name      ,
				details
			)
		VALUES
			(
				$1,
				$2,
				$3
			)`,
		modelKey,
		model.Name,
		model.Details)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// UpdateModel updates a model in the database.
func UpdateModel(dbOrTx DbOrTx, model req_model.Model) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err := identity.PreenKey(model.Key)
	if err != nil {
		return err
	}

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			model
		SET
			name    = $2 ,
			details = $3
		WHERE
			model_key = $1`,
		modelKey,
		model.Name,
		model.Details)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveModel deletes a model from the database.
func RemoveModel(dbOrTx DbOrTx, modelKey string) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}

	// Delete the data.
	_, err = dbExec(dbOrTx, `
		DELETE FROM
			model
		WHERE
			model_key = $1`,
		modelKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryModels loads all models from the database
func QueryModels(dbOrTx DbOrTx) (models []req_model.Model, err error) {

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var model req_model.Model
			if err = scanModel(scanner, &model); err != nil {
				return errors.WithStack(err)
			}
			models = append(models, model)
			return nil
		},
		`SELECT
			model_key   ,
			name        ,
			details
		FROM
			model
		ORDER BY model_key`)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return models, nil
}
