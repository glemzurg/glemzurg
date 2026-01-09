package database

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_use_case"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanUseCaseActor(scanner Scanner, useCaseKeyPtr, actorKeyPtr *string, actor *model_use_case.Actor) (err error) {
	if err = scanner.Scan(
		useCaseKeyPtr,
		actorKeyPtr,
		&actor.UmlComment,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	return nil
}

// LoadUseCaseActor loads a use case actor from the database
func LoadUseCaseActor(dbOrTx DbOrTx, modelKey, useCaseKey, actorKey string) (actor model_use_case.Actor, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return model_use_case.Actor{}, err
	}
	useCaseKey, err = requirements.PreenKey(useCaseKey)
	if err != nil {
		return model_use_case.Actor{}, err
	}
	actorKey, err = requirements.PreenKey(actorKey)
	if err != nil {
		return model_use_case.Actor{}, err
	}

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var unusedUseCaseKey, unusedActorKey string
			if err = scanUseCaseActor(scanner, &unusedUseCaseKey, &unusedActorKey, &actor); err != nil {
				return err
			}
			// Not using the keys since this code already has them.
			_, _ = unusedUseCaseKey, unusedActorKey
			return nil
		},
		`SELECT
			use_case_key ,
			actor_key    ,
			uml_comment
		FROM
			use_case_actor
		WHERE
			use_case_key = $2
		AND
			actor_key = $3
		AND
			model_key = $1`,
		modelKey,
		useCaseKey,
		actorKey)
	if err != nil {
		return model_use_case.Actor{}, errors.WithStack(err)
	}

	return actor, nil
}

// AddUseCaseActor adds a use case actor to the database.
func AddUseCaseActor(dbOrTx DbOrTx, modelKey, useCaseKey, actorKey string, actor model_use_case.Actor) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	useCaseKey, err = requirements.PreenKey(useCaseKey)
	if err != nil {
		return err
	}
	actorKey, err = requirements.PreenKey(actorKey)
	if err != nil {
		return err
	}

	// Add the data.
	_, err = dbExec(dbOrTx, `
			INSERT INTO use_case_actor
				(
					model_key    ,
					use_case_key ,
					actor_key    ,
					uml_comment
				)
			VALUES
				(
					$1,
					$2,
					$3,
					$4
				)`,
		modelKey,
		useCaseKey,
		actorKey,
		actor.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// UpdateUseCaseActor updates a use case actor in the database.
func UpdateUseCaseActor(dbOrTx DbOrTx, modelKey, useCaseKey, actorKey string, actor model_use_case.Actor) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	useCaseKey, err = requirements.PreenKey(useCaseKey)
	if err != nil {
		return err
	}
	actorKey, err = requirements.PreenKey(actorKey)
	if err != nil {
		return err
	}

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			use_case_actor
		SET
			uml_comment = $4
		WHERE
			use_case_key = $2
		AND
			actor_key = $3
		AND
			model_key = $1`,
		modelKey,
		useCaseKey,
		actorKey,
		actor.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveUseCaseActor deletes a use case actor from the database.
func RemoveUseCaseActor(dbOrTx DbOrTx, modelKey, useCaseKey, actorKey string) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	useCaseKey, err = requirements.PreenKey(useCaseKey)
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
			use_case_actor
		WHERE
			use_case_key = $2
		AND
			actor_key = $3
		AND
			model_key = $1`,
		modelKey,
		useCaseKey,
		actorKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryUseCaseActors loads all use case actors from the database
func QueryUseCaseActors(dbOrTx DbOrTx, modelKey string) (actors map[string]map[string]model_use_case.Actor, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return nil, err
	}

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var useCaseKey, actorKey string
			var actor model_use_case.Actor
			if err = scanUseCaseActor(scanner, &useCaseKey, &actorKey, &actor); err != nil {
				return errors.WithStack(err)
			}
			if actors == nil {
				actors = map[string]map[string]model_use_case.Actor{}
			}
			oneActors := actors[useCaseKey]
			if oneActors == nil {
				oneActors = map[string]model_use_case.Actor{}
			}
			oneActors[actorKey] = actor
			actors[useCaseKey] = oneActors
			return nil
		},
		`SELECT
			use_case_key ,
			actor_key    ,
			uml_comment
		FROM
			use_case_actor
		WHERE
			model_key = $1
		ORDER BY use_case_key, actor_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return actors, nil
}
