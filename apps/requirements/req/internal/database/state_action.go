package database

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanAction(scanner Scanner, classKeyPtr *identity.Key, action *model_state.Action) (err error) {
	var classKeyStr string
	var actionKeyStr string

	if err = scanner.Scan(
		&classKeyStr,
		&actionKeyStr,
		&action.Name,
		&action.Details,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	// Parse the class key string into an identity.Key.
	*classKeyPtr, err = identity.ParseKey(classKeyStr)
	if err != nil {
		return err
	}

	// Parse the action key string into an identity.Key.
	action.Key, err = identity.ParseKey(actionKeyStr)
	if err != nil {
		return err
	}

	return nil
}

// LoadAction loads a action from the database
func LoadAction(dbOrTx DbOrTx, modelKey string, actionKey identity.Key) (classKey identity.Key, action model_state.Action, err error) {

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
			details
		FROM
			action
		WHERE
			action_key = $2
		AND
			model_key = $1`,
		modelKey,
		actionKey.String())
	if err != nil {
		return identity.Key{}, model_state.Action{}, errors.WithStack(err)
	}

	return classKey, action, nil
}

// AddAction adds a action to the database.
func AddAction(dbOrTx DbOrTx, modelKey string, classKey identity.Key, action model_state.Action) (err error) {
	return AddActions(dbOrTx, modelKey, map[identity.Key][]model_state.Action{
		classKey: {action},
	})
}

// UpdateAction updates a action in the database.
func UpdateAction(dbOrTx DbOrTx, modelKey string, classKey identity.Key, action model_state.Action) (err error) {

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			action
		SET
			name    = $4 ,
			details = $5
		WHERE
			class_key = $2
		AND
			action_key = $3
		AND
			model_key = $1`,
		modelKey,
		classKey.String(),
		action.Key.String(),
		action.Name,
		action.Details)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveAction deletes a action from the database.
func RemoveAction(dbOrTx DbOrTx, modelKey string, classKey identity.Key, actionKey identity.Key) (err error) {

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
		classKey.String(),
		actionKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryActions loads all action from the database
func QueryActions(dbOrTx DbOrTx, modelKey string) (actions map[identity.Key][]model_state.Action, err error) {

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var classKey identity.Key
			var action model_state.Action
			if err = scanAction(scanner, &classKey, &action); err != nil {
				return errors.WithStack(err)
			}
			if actions == nil {
				actions = map[identity.Key][]model_state.Action{}
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
			details
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

// AddActions adds multiple actions to the database in a single insert.
func AddActions(dbOrTx DbOrTx, modelKey string, actions map[identity.Key][]model_state.Action) (err error) {
	// Count total actions.
	count := 0
	for _, acts := range actions {
		count += len(acts)
	}
	if count == 0 {
		return nil
	}

	// Build the bulk insert query.
	query := `INSERT INTO action (model_key, class_key, action_key, name, details) VALUES `
	args := make([]interface{}, 0, count*5)
	i := 0
	for classKey, actionList := range actions {
		for _, action := range actionList {
			if i > 0 {
				query += ", "
			}
			base := i * 5
			query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4, base+5)
			args = append(args, modelKey, classKey.String(), action.Key.String(), action.Name, action.Details)
			i++
		}
	}

	_, err = dbExec(dbOrTx, query, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
