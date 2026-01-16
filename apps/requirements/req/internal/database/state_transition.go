package database

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanTransition(scanner Scanner, classKeyPtr *string, transition *model_state.Transition) (err error) {
	var fromStateKeyPtr, guardKeyPtr, actionKeyPtr, toStateKeyPtr *string

	if err = scanner.Scan(
		classKeyPtr,
		&transition.Key,
		&fromStateKeyPtr,
		&transition.EventKey,
		&guardKeyPtr,
		&actionKeyPtr,
		&toStateKeyPtr,
		&transition.UmlComment,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	if fromStateKeyPtr != nil {
		transition.FromStateKey = *fromStateKeyPtr
	}
	if guardKeyPtr != nil {
		transition.GuardKey = *guardKeyPtr
	}
	if actionKeyPtr != nil {
		transition.ActionKey = *actionKeyPtr
	}
	if toStateKeyPtr != nil {
		transition.ToStateKey = *toStateKeyPtr
	}

	return nil
}

// LoadTransition loads a transition from the database
func LoadTransition(dbOrTx DbOrTx, modelKey, transitionKey string) (classKey string, transition model_state.Transition, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return "", model_state.Transition{}, err
	}
	transitionKey, err = identity.PreenKey(transitionKey)
	if err != nil {
		return "", model_state.Transition{}, err
	}

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanTransition(scanner, &classKey, &transition); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			class_key      ,
			transition_key ,
			from_state_key ,
			event_key      ,
			guard_key      ,
			action_key     ,
			to_state_key   ,
			uml_comment
		FROM
			transition
		WHERE
			transition_key = $2
		AND
			model_key = $1`,
		modelKey,
		transitionKey)
	if err != nil {
		return "", model_state.Transition{}, errors.WithStack(err)
	}

	return classKey, transition, nil
}

// AddTransition adds a transition to the database.
func AddTransition(dbOrTx DbOrTx, modelKey, classKey string, transition model_state.Transition) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}
	classKey, err = identity.PreenKey(classKey)
	if err != nil {
		return err
	}
	transitionKey, err := identity.PreenKey(transition.Key)
	if err != nil {
		return err
	}
	// We may or may not have a from state.
	var fromStateKeyPtr *string
	if transition.FromStateKey != "" {
		fromStateKey, err := identity.PreenKey(transition.FromStateKey)
		if err != nil {
			return err
		}
		fromStateKeyPtr = &fromStateKey
	}
	eventKey, err := identity.PreenKey(transition.EventKey)
	if err != nil {
		return err
	}
	// We may or may not have a guard.
	var guardKeyPtr *string
	if transition.GuardKey != "" {
		guardKey, err := identity.PreenKey(transition.GuardKey)
		if err != nil {
			return err
		}
		guardKeyPtr = &guardKey
	}
	// We may or may not have an action.
	var actionKeyPtr *string
	if transition.ActionKey != "" {
		actionKey, err := identity.PreenKey(transition.ActionKey)
		if err != nil {
			return err
		}
		actionKeyPtr = &actionKey
	}
	// We may or may not have a to state.
	var toStateKeyPtr *string
	if transition.ToStateKey != "" {
		toStateKey, err := identity.PreenKey(transition.ToStateKey)
		if err != nil {
			return err
		}
		toStateKeyPtr = &toStateKey
	}

	// Add the data.
	_, err = dbExec(dbOrTx, `
			INSERT INTO transition
				(
					model_key      ,
					class_key      ,
					transition_key ,
					from_state_key ,
					event_key      ,
					guard_key      ,
					action_key     ,
					to_state_key   ,
					uml_comment
				)
			VALUES
				(
					$1,
					$2,
					$3,
					$4,
					$5,
					$6,
					$7,
					$8,
					$9
				)`,
		modelKey,
		classKey,
		transitionKey,
		fromStateKeyPtr,
		eventKey,
		guardKeyPtr,
		actionKeyPtr,
		toStateKeyPtr,
		transition.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// UpdateTransition updates a transition in the database.
func UpdateTransition(dbOrTx DbOrTx, modelKey, classKey string, transition model_state.Transition) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}
	classKey, err = identity.PreenKey(classKey)
	if err != nil {
		return err
	}
	transitionKey, err := identity.PreenKey(transition.Key)
	if err != nil {
		return err
	}
	// We may or may not have a from state.
	var fromStateKeyPtr *string
	if transition.FromStateKey != "" {
		fromStateKey, err := identity.PreenKey(transition.FromStateKey)
		if err != nil {
			return err
		}
		fromStateKeyPtr = &fromStateKey
	}
	eventKey, err := identity.PreenKey(transition.EventKey)
	if err != nil {
		return err
	}
	// We may or may not have a guard.
	var guardKeyPtr *string
	if transition.GuardKey != "" {
		guardKey, err := identity.PreenKey(transition.GuardKey)
		if err != nil {
			return err
		}
		guardKeyPtr = &guardKey
	}
	// We may or may not have an action.
	var actionKeyPtr *string
	if transition.ActionKey != "" {
		actionKey, err := identity.PreenKey(transition.ActionKey)
		if err != nil {
			return err
		}
		actionKeyPtr = &actionKey
	}
	// We may or may not have a to state.
	var toStateKeyPtr *string
	if transition.ToStateKey != "" {
		toStateKey, err := identity.PreenKey(transition.ToStateKey)
		if err != nil {
			return err
		}
		toStateKeyPtr = &toStateKey
	}

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			transition
		SET
			from_state_key = $4 ,
			event_key      = $5 ,
			guard_key      = $6 ,
			action_key     = $7 ,
			to_state_key   = $8 ,
			uml_comment    = $9
		WHERE
			class_key = $2
		AND
			transition_key = $3
		AND
			model_key = $1`,
		modelKey,
		classKey,
		transitionKey,
		fromStateKeyPtr,
		eventKey,
		guardKeyPtr,
		actionKeyPtr,
		toStateKeyPtr,
		transition.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveTransition deletes a transition from the database.
func RemoveTransition(dbOrTx DbOrTx, modelKey, classKey, transitionKey string) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}
	classKey, err = identity.PreenKey(classKey)
	if err != nil {
		return err
	}
	transitionKey, err = identity.PreenKey(transitionKey)
	if err != nil {
		return err
	}

	// Delete the data.
	_, err = dbExec(dbOrTx, `
		DELETE FROM
			transition
		WHERE
			class_key = $2
		AND
			transition_key = $3
		AND
			model_key = $1`,
		modelKey,
		classKey,
		transitionKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryTransitions loads all transition from the database
func QueryTransitions(dbOrTx DbOrTx, modelKey string) (transitions map[string][]model_state.Transition, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return nil, err
	}

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var classKey string
			var transition model_state.Transition
			if err = scanTransition(scanner, &classKey, &transition); err != nil {
				return errors.WithStack(err)
			}
			if transitions == nil {
				transitions = map[string][]model_state.Transition{}
			}
			classTransitions := transitions[classKey]
			classTransitions = append(classTransitions, transition)
			transitions[classKey] = classTransitions
			return nil
		},
		`SELECT
			class_key      ,
			transition_key ,
			from_state_key ,
			event_key      ,
			guard_key      ,
			action_key     ,
			to_state_key   ,
			uml_comment
		FROM
			transition
		WHERE
			model_key = $1
		ORDER BY class_key, transition_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return transitions, nil
}
