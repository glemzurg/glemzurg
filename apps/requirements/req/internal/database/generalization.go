package database

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanGeneralization(scanner Scanner, generalization *requirements.Generalization) (err error) {
	if err = scanner.Scan(
		&generalization.Key,
		&generalization.Name,
		&generalization.Details,
		&generalization.IsComplete,
		&generalization.IsStatic,
		&generalization.UmlComment,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	return nil
}

// LoadGeneralization loads a generalization from the database
func LoadGeneralization(dbOrTx DbOrTx, modelKey, generalizationKey string) (generalization requirements.Generalization, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return requirements.Generalization{}, err
	}
	generalizationKey, err = requirements.PreenKey(generalizationKey)
	if err != nil {
		return requirements.Generalization{}, err
	}

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanGeneralization(scanner, &generalization); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			generalization_key ,
			name               ,
			details            ,
			is_complete        ,
			is_static          ,
			uml_comment
		FROM
			generalization
		WHERE
			generalization_key = $2
		AND
			model_key = $1`,
		modelKey,
		generalizationKey)
	if err != nil {
		return requirements.Generalization{}, errors.WithStack(err)
	}

	return generalization, nil
}

// AddGeneralization adds a generalization to the database.
func AddGeneralization(dbOrTx DbOrTx, modelKey string, generalization requirements.Generalization) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	generalizationKey, err := requirements.PreenKey(generalization.Key)
	if err != nil {
		return err
	}

	// Add the data.
	_, err = dbExec(dbOrTx, `
			INSERT INTO generalization
				(
					model_key          ,
					generalization_key ,
					name               ,
					details            ,
					is_complete        ,
					is_static          ,
					uml_comment
				)
			VALUES
				(
					$1,
					$2,
					$3,
					$4,
					$5,
					$6,
					$7
				)`,
		modelKey,
		generalizationKey,
		generalization.Name,
		generalization.Details,
		generalization.IsComplete,
		generalization.IsStatic,
		generalization.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// UpdateGeneralization updates a generalization in the database.
func UpdateGeneralization(dbOrTx DbOrTx, modelKey string, generalization requirements.Generalization) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	generalizationKey, err := requirements.PreenKey(generalization.Key)
	if err != nil {
		return err
	}

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			generalization
		SET
			name        = $3 ,
			details     = $4 ,
			is_complete = $5 ,
			is_static   = $6 ,
			uml_comment = $7
		WHERE
			model_key = $1
		AND
			generalization_key = $2`,
		modelKey,
		generalizationKey,
		generalization.Name,
		generalization.Details,
		generalization.IsComplete,
		generalization.IsStatic,
		generalization.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveGeneralization deletes a generalization from the database.
func RemoveGeneralization(dbOrTx DbOrTx, modelKey, generalizationKey string) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	generalizationKey, err = requirements.PreenKey(generalizationKey)
	if err != nil {
		return err
	}

	// Delete the data.
	_, err = dbExec(dbOrTx, `
			DELETE FROM
				generalization
			WHERE
				model_key = $1
			AND
				generalization_key = $2`,
		modelKey,
		generalizationKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryGeneralizations loads all generalizations from the database
func QueryGeneralizations(dbOrTx DbOrTx, modelKey string) (generalizations []requirements.Generalization, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return nil, err
	}

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var generalization requirements.Generalization
			if err = scanGeneralization(scanner, &generalization); err != nil {
				return errors.WithStack(err)
			}
			generalizations = append(generalizations, generalization)
			return nil
		},
		`SELECT
			generalization_key ,
			name               ,
			details            ,
			is_complete        ,
			is_static          ,
			uml_comment
		FROM
			generalization
		WHERE
			model_key = $1
		ORDER BY generalization_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return generalizations, nil
}
