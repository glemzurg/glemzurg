package database

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanGeneralization(scanner Scanner, generalization *model_class.Generalization) (err error) {
	var keyStr string

	if err = scanner.Scan(
		&keyStr,
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

	// Parse the key string into an identity.Key.
	generalization.Key, err = identity.ParseKey(keyStr)
	if err != nil {
		return err
	}

	return nil
}

// LoadGeneralization loads a generalization from the database
func LoadGeneralization(dbOrTx DbOrTx, modelKey string, generalizationKey identity.Key) (generalization model_class.Generalization, err error) {

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
		generalizationKey.String())
	if err != nil {
		return model_class.Generalization{}, errors.WithStack(err)
	}

	return generalization, nil
}

// AddGeneralization adds a generalization to the database.
func AddGeneralization(dbOrTx DbOrTx, modelKey string, generalization model_class.Generalization) (err error) {

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
		generalization.Key.String(),
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
func UpdateGeneralization(dbOrTx DbOrTx, modelKey string, generalization model_class.Generalization) (err error) {

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
		generalization.Key.String(),
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
func RemoveGeneralization(dbOrTx DbOrTx, modelKey string, generalizationKey identity.Key) (err error) {

	// Delete the data.
	_, err = dbExec(dbOrTx, `
			DELETE FROM
				generalization
			WHERE
				model_key = $1
			AND
				generalization_key = $2`,
		modelKey,
		generalizationKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryGeneralizations loads all generalizations from the database
func QueryGeneralizations(dbOrTx DbOrTx, modelKey string) (generalizations []model_class.Generalization, err error) {

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var generalization model_class.Generalization
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

// AddGeneralizations adds multiple generalizations to the database in a single insert.
func AddGeneralizations(dbOrTx DbOrTx, modelKey string, generalizations []model_class.Generalization) (err error) {
	if len(generalizations) == 0 {
		return nil
	}

	// Build the bulk insert query.
	query := `INSERT INTO generalization (model_key, generalization_key, name, details, is_complete, is_static, uml_comment) VALUES `
	args := make([]interface{}, 0, len(generalizations)*7)
	for i, gen := range generalizations {
		if i > 0 {
			query += ", "
		}
		base := i * 7
		query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4, base+5, base+6, base+7)
		args = append(args, modelKey, gen.Key.String(), gen.Name, gen.Details, gen.IsComplete, gen.IsStatic, gen.UmlComment)
	}

	_, err = dbExec(dbOrTx, query, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
