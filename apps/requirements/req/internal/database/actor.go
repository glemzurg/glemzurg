package database

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanActor(scanner Scanner, actor *requirements.Actor) (err error) {
	if err = scanner.Scan(
		&actor.Key,
		&actor.Name,
		&actor.Details,
		&actor.Type,
		&actor.UmlComment,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	return nil
}

// LoadActor loads a actor from the database
func LoadActor(dbOrTx DbOrTx, modelKey, actorKey string) (actor requirements.Actor, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return requirements.Actor{}, err
	}
	actorKey, err = requirements.PreenKey(actorKey)
	if err != nil {
		return requirements.Actor{}, err
	}

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanActor(scanner, &actor); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			actor_key   ,
			name        ,
			details     ,
			actor_type  ,
			uml_comment
		FROM
			actor
		WHERE
			actor_key = $2
		AND
			model_key = $1`,
		modelKey,
		actorKey)
	if err != nil {
		return requirements.Actor{}, errors.WithStack(err)
	}

	return actor, nil
}

// AddActor adds a actor to the database.
func AddActor(dbOrTx DbOrTx, modelKey string, actor requirements.Actor) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	actorKey, err := requirements.PreenKey(actor.Key)
	if err != nil {
		return err
	}

	// Add the data.
	_, err = dbExec(dbOrTx, `
			INSERT INTO actor
				(
					model_key   ,
					actor_key   ,
					name        ,
					details     ,
					actor_type  ,
					uml_comment
				)
			VALUES
				(
					$1,
					$2,
					$3,
					$4,
					$5,
					$6
				)`,
		modelKey,
		actorKey,
		actor.Name,
		actor.Details,
		actor.Type,
		actor.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// UpdateActor updates a actor in the database.
func UpdateActor(dbOrTx DbOrTx, modelKey string, actor requirements.Actor) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	actorKey, err := requirements.PreenKey(actor.Key)
	if err != nil {
		return err
	}

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			actor
		SET
			name        = $3 ,
			details     = $4 ,
			actor_type  = $5 ,
			uml_comment = $6
		WHERE
			model_key = $1
		AND
			actor_key = $2`,
		modelKey,
		actorKey,
		actor.Name,
		actor.Details,
		actor.Type,
		actor.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveActor deletes a actor from the database.
func RemoveActor(dbOrTx DbOrTx, modelKey, actorKey string) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	actorKey, err = requirements.PreenKey(actorKey)
	if err != nil {
		return err
	}

	// Delete the data.
	_, err = dbExec(dbOrTx, `
			DELETE FROM
				actor
			WHERE
				model_key = $1
			AND
				actor_key = $2`,
		modelKey,
		actorKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryActors loads all actors from the database
func QueryActors(dbOrTx DbOrTx, modelKey string) (actors []requirements.Actor, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return nil, err
	}

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var actor requirements.Actor
			if err = scanActor(scanner, &actor); err != nil {
				return errors.WithStack(err)
			}
			actors = append(actors, actor)
			return nil
		},
		`SELECT
				actor_key   ,
				name        ,
				details     ,
				actor_type  ,
				uml_comment
			FROM
				actor
			WHERE
				model_key = $1
			ORDER BY actor_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return actors, nil
}
