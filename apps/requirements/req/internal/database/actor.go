package database

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanActor(scanner Scanner, actor *model_actor.Actor) (err error) {
	var keyStr string

	if err = scanner.Scan(
		&keyStr,
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

	// Parse the key string into an identity.Key.
	actor.Key, err = identity.ParseKey(keyStr)
	if err != nil {
		return err
	}

	return nil
}

// LoadActor loads a actor from the database
func LoadActor(dbOrTx DbOrTx, modelKey string, actorKey identity.Key) (actor model_actor.Actor, err error) {

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
		actorKey.String())
	if err != nil {
		return model_actor.Actor{}, errors.WithStack(err)
	}

	return actor, nil
}

// AddActor adds a actor to the database.
func AddActor(dbOrTx DbOrTx, modelKey string, actor model_actor.Actor) (err error) {
	return AddActors(dbOrTx, modelKey, []model_actor.Actor{actor})
}

// UpdateActor updates a actor in the database.
func UpdateActor(dbOrTx DbOrTx, modelKey string, actor model_actor.Actor) (err error) {

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
		actor.Key.String(),
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
func RemoveActor(dbOrTx DbOrTx, modelKey string, actorKey identity.Key) (err error) {

	// Delete the data.
	_, err = dbExec(dbOrTx, `
			DELETE FROM
				actor
			WHERE
				model_key = $1
			AND
				actor_key = $2`,
		modelKey,
		actorKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryActors loads all actors from the database
func QueryActors(dbOrTx DbOrTx, modelKey string) (actors []model_actor.Actor, err error) {

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var actor model_actor.Actor
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

// AddActors adds multiple actors to the database in a single insert.
func AddActors(dbOrTx DbOrTx, modelKey string, actors []model_actor.Actor) (err error) {
	if len(actors) == 0 {
		return nil
	}

	// Build the bulk insert query.
	query := `INSERT INTO actor (model_key, actor_key, name, details, actor_type, uml_comment) VALUES `
	args := make([]interface{}, 0, len(actors)*6)
	for i, actor := range actors {
		if i > 0 {
			query += ", "
		}
		base := i * 6
		query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4, base+5, base+6)
		args = append(args, modelKey, actor.Key.String(), actor.Name, actor.Details, actor.Type, actor.UmlComment)
	}

	_, err = dbExec(dbOrTx, query, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
