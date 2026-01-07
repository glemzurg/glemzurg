package database

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_state"

	"github.com/lib/pq"
	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanAction(scanner Scanner, classKeyPtr *string, action *model_state.Action) (err error) {
	if err = scanner.Scan(
		classKeyPtr,
		&action.Key,
		&action.Name,
		&action.Details,
		pq.Array(&action.Requires),
		pq.Array(&action.Guarantees),
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	return nil
}

// LoadAction loads a action from the database
func LoadAction(dbOrTx DbOrTx, modelKey, actionKey string) (classKey string, action model_state.Action, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return "", model_state.Action{}, err
	}
	actionKey, err = requirements.PreenKey(actionKey)
	if err != nil {
		return "", model_state.Action{}, err
	}

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanAction(scanner, &classKey, &action); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			class_key  ,
			action_key ,
			name       ,
			details    ,
			requires   ,
			guarantees
		FROM
			action
		WHERE
			action_key = $2
		AND
			model_key = $1`,
		modelKey,
		actionKey)
	if err != nil {
		return "", model_state.Action{}, errors.WithStack(err)
	}

	return classKey, action, nil
}

// AddAction adds a action to the database.
func AddAction(dbOrTx DbOrTx, modelKey, classKey string, action model_state.Action) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	classKey, err = requirements.PreenKey(classKey)
	if err != nil {
		return err
	}
	actionKey, err := requirements.PreenKey(action.Key)
	if err != nil {
		return err
	}

	// Add the data.
	_, err = dbExec(dbOrTx, `
			INSERT INTO action
				(
					model_key  ,
					class_key  ,
					action_key ,
					name       ,
					details    ,
			        requires   ,
			        guarantees
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
		classKey,
		actionKey,
		action.Name,
		action.Details,
		pq.Array(action.Requires),
		pq.Array(action.Guarantees))
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// UpdateAction updates a action in the database.
func UpdateAction(dbOrTx DbOrTx, modelKey, classKey string, action model_state.Action) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	classKey, err = requirements.PreenKey(classKey)
	if err != nil {
		return err
	}
	actionKey, err := requirements.PreenKey(action.Key)
	if err != nil {
		return err
	}

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			action
		SET
			name       = $4 ,
			details    = $5 ,
			requires   = $6 ,
			guarantees = $7
		WHERE
			class_key = $2
		AND
			action_key = $3
		AND
			model_key = $1`,
		modelKey,
		classKey,
		actionKey,
		action.Name,
		action.Details,
		pq.Array(action.Requires),
		pq.Array(action.Guarantees))
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveAction deletes a action from the database.
func RemoveAction(dbOrTx DbOrTx, modelKey, classKey, actionKey string) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	classKey, err = requirements.PreenKey(classKey)
	if err != nil {
		return err
	}
	actionKey, err = requirements.PreenKey(actionKey)
	if err != nil {
		return err
	}

	// Delete the data.
	_, err = dbExec(dbOrTx, `
		DELETE FROM
			action
		WHERE
			class_key = $2
		AND
			action_key = $3
		AND
			model_key = $1`,
		modelKey,
		classKey,
		actionKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryActions loads all action from the database
func QueryActions(dbOrTx DbOrTx, modelKey string) (actions map[string][]model_state.Action, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return nil, err
	}

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var classKey string
			var action model_state.Action
			if err = scanAction(scanner, &classKey, &action); err != nil {
				return errors.WithStack(err)
			}
			if actions == nil {
				actions = map[string][]model_state.Action{}
			}
			classActions := actions[classKey]
			classActions = append(classActions, action)
			actions[classKey] = classActions
			return nil
		},
		`SELECT
			class_key  ,
			action_key ,
			name       ,
			details    ,
			requires   ,
			guarantees
		FROM
			action
		WHERE
			model_key = $1
		ORDER BY class_key, action_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return actions, nil
}
