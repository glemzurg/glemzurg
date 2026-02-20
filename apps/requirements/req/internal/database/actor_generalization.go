package database

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanActorGeneralization(scanner Scanner, generalization *model_actor.Generalization) (err error) {
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

// LoadActorGeneralization loads an actor generalization from the database.
func LoadActorGeneralization(dbOrTx DbOrTx, modelKey string, generalizationKey identity.Key) (generalization model_actor.Generalization, err error) {

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanActorGeneralization(scanner, &generalization); err != nil {
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
			actor_generalization
		WHERE
			generalization_key = $2
		AND
			model_key = $1`,
		modelKey,
		generalizationKey.String())
	if err != nil {
		return model_actor.Generalization{}, errors.WithStack(err)
	}

	return generalization, nil
}

// AddActorGeneralization adds an actor generalization to the database.
func AddActorGeneralization(dbOrTx DbOrTx, modelKey string, generalization model_actor.Generalization) (err error) {
	return AddActorGeneralizations(dbOrTx, modelKey, []model_actor.Generalization{generalization})
}

// UpdateActorGeneralization updates an actor generalization in the database.
func UpdateActorGeneralization(dbOrTx DbOrTx, modelKey string, generalization model_actor.Generalization) (err error) {

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			actor_generalization
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

// RemoveActorGeneralization deletes an actor generalization from the database.
func RemoveActorGeneralization(dbOrTx DbOrTx, modelKey string, generalizationKey identity.Key) (err error) {

	// Delete the data.
	_, err = dbExec(dbOrTx, `
			DELETE FROM
				actor_generalization
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

// QueryActorGeneralizations loads all actor generalizations from the database.
func QueryActorGeneralizations(dbOrTx DbOrTx, modelKey string) (generalizations []model_actor.Generalization, err error) {

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var generalization model_actor.Generalization
			if err = scanActorGeneralization(scanner, &generalization); err != nil {
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
			actor_generalization
		WHERE
			model_key = $1
		ORDER BY generalization_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return generalizations, nil
}

// AddActorGeneralizations adds multiple actor generalizations to the database in a single insert.
func AddActorGeneralizations(dbOrTx DbOrTx, modelKey string, generalizations []model_actor.Generalization) (err error) {
	if len(generalizations) == 0 {
		return nil
	}

	// Build the bulk insert query.
	query := `INSERT INTO actor_generalization (model_key, generalization_key, name, details, is_complete, is_static, uml_comment) VALUES `
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
