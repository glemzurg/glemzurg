package database

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanStateAction(scanner Scanner, stateKeyPtr *identity.Key, stateAction *model_state.StateAction) (err error) {
	var stateKeyStr string
	var stateActionKeyStr string
	var actionKeyStr string

	if err = scanner.Scan(
		&stateKeyStr,
		&stateActionKeyStr,
		&actionKeyStr,
		&stateAction.When,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	// Parse the state key string into an identity.Key.
	*stateKeyPtr, err = identity.ParseKey(stateKeyStr)
	if err != nil {
		return err
	}

	// Parse the state action key string into an identity.Key.
	stateAction.Key, err = identity.ParseKey(stateActionKeyStr)
	if err != nil {
		return err
	}

	// Parse the action key string into an identity.Key.
	stateAction.ActionKey, err = identity.ParseKey(actionKeyStr)
	if err != nil {
		return err
	}

	return nil
}

// LoadStateAction loads a stateAction from the database
func LoadStateAction(dbOrTx DbOrTx, modelKey string, stateActionKey identity.Key) (stateKey identity.Key, stateAction model_state.StateAction, err error) {

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
		stateActionKey.String())
	if err != nil {
		return identity.Key{}, model_state.StateAction{}, errors.WithStack(err)
	}

	return stateKey, stateAction, nil
}

// AddStateAction adds a stateAction to the database.
func AddStateAction(dbOrTx DbOrTx, modelKey string, stateKey identity.Key, stateAction model_state.StateAction) (err error) {
	return AddStateActions(dbOrTx, modelKey, map[identity.Key][]model_state.StateAction{
		stateKey: {stateAction},
	})
}

// UpdateStateAction updates a stateAction in the database.
func UpdateStateAction(dbOrTx DbOrTx, modelKey string, stateKey identity.Key, stateAction model_state.StateAction) (err error) {

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
		stateKey.String(),
		stateAction.Key.String(),
		stateAction.ActionKey.String(),
		stateAction.When)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveStateAction deletes a stateAction from the database.
func RemoveStateAction(dbOrTx DbOrTx, modelKey string, stateKey identity.Key, stateActionKey identity.Key) (err error) {

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
		stateKey.String(),
		stateActionKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryStateActions loads all stateAction from the database
func QueryStateActions(dbOrTx DbOrTx, modelKey string) (stateActions map[identity.Key][]model_state.StateAction, err error) {

	stateActions = make(map[identity.Key][]model_state.StateAction)

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var stateKey identity.Key
			var stateAction model_state.StateAction
			if err = scanStateAction(scanner, &stateKey, &stateAction); err != nil {
				return errors.WithStack(err)
			}
			stateActions[stateKey] = append(stateActions[stateKey], stateAction)
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

// AddStateActions adds multiple state actions to the database in a single insert.
func AddStateActions(dbOrTx DbOrTx, modelKey string, stateActions map[identity.Key][]model_state.StateAction) (err error) {
	// Count total state actions.
	count := 0
	for _, sas := range stateActions {
		count += len(sas)
	}
	if count == 0 {
		return nil
	}

	// Build the bulk insert query.
	query := `INSERT INTO state_action (model_key, state_key, state_action_key, action_key, action_when) VALUES `
	args := make([]interface{}, 0, count*5)
	i := 0
	for stateKey, saList := range stateActions {
		for _, sa := range saList {
			if i > 0 {
				query += ", "
			}
			base := i * 5
			query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4, base+5)
			args = append(args, modelKey, stateKey.String(), sa.Key.String(), sa.ActionKey.String(), sa.When)
			i++
		}
	}

	_, err = dbExec(dbOrTx, query, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
