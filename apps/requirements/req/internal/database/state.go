package database

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_state"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanState(scanner Scanner, classKeyPtr *string, state *model_state.State) (err error) {
	if err = scanner.Scan(
		classKeyPtr,
		&state.Key,
		&state.Name,
		&state.Details,
		&state.UmlComment,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	return nil
}

// LoadState loads a state from the database
func LoadState(dbOrTx DbOrTx, modelKey, stateKey string) (classKey string, state model_state.State, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return "", model_state.State{}, err
	}
	stateKey, err = requirements.PreenKey(stateKey)
	if err != nil {
		return "", model_state.State{}, err
	}

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
		stateKey)
	if err != nil {
		return "", model_state.State{}, errors.WithStack(err)
	}

	return classKey, state, nil
}

// AddState adds a state to the database.
func AddState(dbOrTx DbOrTx, modelKey, classKey string, state model_state.State) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	classKey, err = requirements.PreenKey(classKey)
	if err != nil {
		return err
	}
	stateKey, err := requirements.PreenKey(state.Key)
	if err != nil {
		return err
	}

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
		classKey,
		stateKey,
		state.Name,
		state.Details,
		state.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// UpdateState updates a state in the database.
func UpdateState(dbOrTx DbOrTx, modelKey, classKey string, state model_state.State) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	classKey, err = requirements.PreenKey(classKey)
	if err != nil {
		return err
	}
	stateKey, err := requirements.PreenKey(state.Key)
	if err != nil {
		return err
	}

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
		classKey,
		stateKey,
		state.Name,
		state.Details,
		state.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveState deletes a state from the database.
func RemoveState(dbOrTx DbOrTx, modelKey, classKey, stateKey string) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	classKey, err = requirements.PreenKey(classKey)
	if err != nil {
		return err
	}
	stateKey, err = requirements.PreenKey(stateKey)
	if err != nil {
		return err
	}

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
		classKey,
		stateKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryStates loads all state from the database
func QueryStates(dbOrTx DbOrTx, modelKey string) (states map[string][]model_state.State, err error) {

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
			var state model_state.State
			if err = scanState(scanner, &classKey, &state); err != nil {
				return errors.WithStack(err)
			}
			if states == nil {
				states = map[string][]model_state.State{}
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
