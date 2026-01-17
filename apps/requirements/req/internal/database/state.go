package database

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanState(scanner Scanner, classKeyPtr *identity.Key, state *model_state.State) (err error) {
	var classKeyStr string
	var stateKeyStr string

	if err = scanner.Scan(
		&classKeyStr,
		&stateKeyStr,
		&state.Name,
		&state.Details,
		&state.UmlComment,
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

	// Parse the state key string into an identity.Key.
	state.Key, err = identity.ParseKey(stateKeyStr)
	if err != nil {
		return err
	}

	return nil
}

// LoadState loads a state from the database
func LoadState(dbOrTx DbOrTx, modelKey string, stateKey identity.Key) (classKey identity.Key, state model_state.State, err error) {

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanState(scanner, &classKey, &state); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			class_key   ,
			state_key   ,
			name        ,
			details     ,
			uml_comment
		FROM
			state
		WHERE
			state_key = $2
		AND
			model_key = $1`,
		modelKey,
		stateKey.String())
	if err != nil {
		return identity.Key{}, model_state.State{}, errors.WithStack(err)
	}

	return classKey, state, nil
}

// AddState adds a state to the database.
func AddState(dbOrTx DbOrTx, modelKey string, classKey identity.Key, state model_state.State) (err error) {

	// Add the data.
	_, err = dbExec(dbOrTx, `
			INSERT INTO state
				(
					model_key   ,
					class_key   ,
					state_key   ,
					name        ,
					details     ,
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
		classKey.String(),
		state.Key.String(),
		state.Name,
		state.Details,
		state.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// UpdateState updates a state in the database.
func UpdateState(dbOrTx DbOrTx, modelKey string, classKey identity.Key, state model_state.State) (err error) {

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			state
		SET
			name                  = $4 ,
			details               = $5 ,
			uml_comment           = $6
		WHERE
			class_key = $2
		AND
			state_key = $3
		AND
			model_key = $1`,
		modelKey,
		classKey.String(),
		state.Key.String(),
		state.Name,
		state.Details,
		state.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveState deletes a state from the database.
func RemoveState(dbOrTx DbOrTx, modelKey string, classKey identity.Key, stateKey identity.Key) (err error) {

	// Delete the data.
	_, err = dbExec(dbOrTx, `
		DELETE FROM
			state
		WHERE
			class_key = $2
		AND
			state_key = $3
		AND
			model_key = $1`,
		modelKey,
		classKey.String(),
		stateKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryStates loads all state from the database
func QueryStates(dbOrTx DbOrTx, modelKey string) (states map[identity.Key][]model_state.State, err error) {

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var classKey identity.Key
			var state model_state.State
			if err = scanState(scanner, &classKey, &state); err != nil {
				return errors.WithStack(err)
			}
			if states == nil {
				states = map[identity.Key][]model_state.State{}
			}
			classStates := states[classKey]
			classStates = append(classStates, state)
			states[classKey] = classStates
			return nil
		},
		`SELECT
			class_key   ,
			state_key   ,
			name        ,
			details     ,
			uml_comment
		FROM
			state
		WHERE
			model_key = $1
		ORDER BY class_key, state_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return states, nil
}
