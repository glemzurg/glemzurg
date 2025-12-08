package database

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanStateAction(scanner Scanner, stateKeyPtr *string, stateAction *requirements.StateAction) (err error) {

	if err = scanner.Scan(
		stateKeyPtr,
		&stateAction.Key,
		&stateAction.ActionKey,
		&stateAction.When,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	return nil
}

// LoadStateAction loads a stateAction from the database
func LoadStateAction(dbOrTx DbOrTx, modelKey, stateActionKey string) (stateKey string, stateAction requirements.StateAction, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return "", requirements.StateAction{}, err
	}
	stateActionKey, err = requirements.PreenKey(stateActionKey)
	if err != nil {
		return "", requirements.StateAction{}, err
	}

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanStateAction(scanner, &stateKey, &stateAction); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			state_key        ,
			state_action_key ,
			action_key       ,
			action_when
		FROM
			state_action
		WHERE
			state_action_key = $2
		AND
			model_key = $1`,
		modelKey,
		stateActionKey)
	if err != nil {
		return "", requirements.StateAction{}, errors.WithStack(err)
	}

	return stateKey, stateAction, nil
}

// AddStateAction adds a stateAction to the database.
func AddStateAction(dbOrTx DbOrTx, modelKey, stateKey string, stateAction requirements.StateAction) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	stateKey, err = requirements.PreenKey(stateKey)
	if err != nil {
		return err
	}
	stateActionKey, err := requirements.PreenKey(stateAction.Key)
	if err != nil {
		return err
	}
	actionKey, err := requirements.PreenKey(stateAction.ActionKey)
	if err != nil {
		return err
	}

	// Add the data.
	_, err = dbExec(dbOrTx, `
		INSERT INTO state_action
			(
				model_key        ,
				state_key        ,
				state_action_key ,
				action_key       ,
				action_when
			)
		VALUES
			(
				$1,
				$2,
				$3,
				$4,
				$5
			)`,
		modelKey,
		stateKey,
		stateActionKey,
		actionKey,
		stateAction.When)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// UpdateStateAction updates a stateAction in the database.
func UpdateStateAction(dbOrTx DbOrTx, modelKey, stateKey string, stateAction requirements.StateAction) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	stateKey, err = requirements.PreenKey(stateKey)
	if err != nil {
		return err
	}
	stateActionKey, err := requirements.PreenKey(stateAction.Key)
	if err != nil {
		return err
	}
	actionKey, err := requirements.PreenKey(stateAction.ActionKey)
	if err != nil {
		return err
	}

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			state_action
		SET
			action_key  = $4 ,
			action_when = $5
		WHERE
			state_key = $2
		AND
			state_action_key = $3
		AND
			model_key = $1`,
		modelKey,
		stateKey,
		stateActionKey,
		actionKey,
		stateAction.When)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveStateAction deletes a stateAction from the database.
func RemoveStateAction(dbOrTx DbOrTx, modelKey, stateKey, stateActionKey string) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	stateKey, err = requirements.PreenKey(stateKey)
	if err != nil {
		return err
	}
	stateActionKey, err = requirements.PreenKey(stateActionKey)
	if err != nil {
		return err
	}

	// Delete the data.
	_, err = dbExec(dbOrTx, `
		DELETE FROM
			state_action
		WHERE
			state_key = $2
		AND
			state_action_key = $3
		AND
			model_key = $1`,
		modelKey,
		stateKey,
		stateActionKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryStateActions loads all stateAction from the database
func QueryStateActions(dbOrTx DbOrTx, modelKey string) (stateActions map[string][]requirements.StateAction, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return nil, err
	}

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var stateKey string
			var stateAction requirements.StateAction
			if err = scanStateAction(scanner, &stateKey, &stateAction); err != nil {
				return errors.WithStack(err)
			}
			if stateActions == nil {
				stateActions = map[string][]requirements.StateAction{}
			}
			classStateActions := stateActions[stateKey]
			classStateActions = append(classStateActions, stateAction)
			stateActions[stateKey] = classStateActions
			return nil
		},
		`SELECT
				state_key        ,
				state_action_key ,
				action_key       ,
				action_when
			FROM
				state_action
			WHERE
				model_key = $1
			ORDER BY state_key, state_action_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return stateActions, nil
}
