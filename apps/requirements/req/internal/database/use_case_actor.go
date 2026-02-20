package database

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanUseCaseActor(scanner Scanner, useCaseKeyPtr, actorKeyPtr *identity.Key, actor *model_use_case.Actor) (err error) {
	var useCaseKeyStr string
	var actorKeyStr string

	if err = scanner.Scan(
		&useCaseKeyStr,
		&actorKeyStr,
		&actor.UmlComment,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	// Parse the use case key string into an identity.Key.
	*useCaseKeyPtr, err = identity.ParseKey(useCaseKeyStr)
	if err != nil {
		return err
	}

	// Parse the actor key string into an identity.Key.
	*actorKeyPtr, err = identity.ParseKey(actorKeyStr)
	if err != nil {
		return err
	}

	return nil
}

// LoadUseCaseActor loads a use case actor from the database
func LoadUseCaseActor(dbOrTx DbOrTx, modelKey string, useCaseKey identity.Key, actorKey identity.Key) (actor model_use_case.Actor, err error) {

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var unusedUseCaseKey, unusedActorKey identity.Key
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
		useCaseKey.String(),
		actorKey.String())
	if err != nil {
		return model_use_case.Actor{}, errors.WithStack(err)
	}

	return actor, nil
}

// AddUseCaseActor adds a use case actor to the database.
func AddUseCaseActor(dbOrTx DbOrTx, modelKey string, useCaseKey identity.Key, actorKey identity.Key, actor model_use_case.Actor) (err error) {
	return AddUseCaseActors(dbOrTx, modelKey, map[identity.Key]map[identity.Key]model_use_case.Actor{
		useCaseKey: {actorKey: actor},
	})
}

// UpdateUseCaseActor updates a use case actor in the database.
func UpdateUseCaseActor(dbOrTx DbOrTx, modelKey string, useCaseKey identity.Key, actorKey identity.Key, actor model_use_case.Actor) (err error) {

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
		useCaseKey.String(),
		actorKey.String(),
		actor.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveUseCaseActor deletes a use case actor from the database.
func RemoveUseCaseActor(dbOrTx DbOrTx, modelKey string, useCaseKey identity.Key, actorKey identity.Key) (err error) {

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
		useCaseKey.String(),
		actorKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryUseCaseActors loads all use case actors from the database
func QueryUseCaseActors(dbOrTx DbOrTx, modelKey string) (actors map[identity.Key]map[identity.Key]model_use_case.Actor, err error) {

	actors = make(map[identity.Key]map[identity.Key]model_use_case.Actor)

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var useCaseKey, actorKey identity.Key
			var actor model_use_case.Actor
			if err = scanUseCaseActor(scanner, &useCaseKey, &actorKey, &actor); err != nil {
				return errors.WithStack(err)
			}
			oneActors := actors[useCaseKey]
			if oneActors == nil {
				oneActors = map[identity.Key]model_use_case.Actor{}
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

// AddUseCaseActors adds multiple use case actors to the database in a single insert.
func AddUseCaseActors(dbOrTx DbOrTx, modelKey string, actors map[identity.Key]map[identity.Key]model_use_case.Actor) (err error) {
	// Count total actors.
	count := 0
	for _, actorMap := range actors {
		count += len(actorMap)
	}
	if count == 0 {
		return nil
	}

	// Build the bulk insert query.
	query := `INSERT INTO use_case_actor (model_key, use_case_key, actor_key, uml_comment) VALUES `
	args := make([]interface{}, 0, count*4)
	i := 0
	for useCaseKey, actorMap := range actors {
		for actorKey, actor := range actorMap {
			if i > 0 {
				query += ", "
			}
			base := i * 4
			query += fmt.Sprintf("($%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4)
			args = append(args, modelKey, useCaseKey.String(), actorKey.String(), actor.UmlComment)
			i++
		}
	}

	_, err = dbExec(dbOrTx, query, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
