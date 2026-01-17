package database

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanTransition(scanner Scanner, classKeyPtr *identity.Key, transition *model_state.Transition) (err error) {
	var classKeyStr string
	var transitionKeyStr string
	var fromStateKeyPtr, guardKeyPtr, actionKeyPtr, toStateKeyPtr *string
	var eventKeyStr string

	if err = scanner.Scan(
		&classKeyStr,
		&transitionKeyStr,
		&fromStateKeyPtr,
		&eventKeyStr,
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

	// Parse the class key string into an identity.Key.
	*classKeyPtr, err = identity.ParseKey(classKeyStr)
	if err != nil {
		return err
	}

	// Parse the transition key string into an identity.Key.
	transition.Key, err = identity.ParseKey(transitionKeyStr)
	if err != nil {
		return err
	}

	// Parse the event key string into an identity.Key (required).
	transition.EventKey, err = identity.ParseKey(eventKeyStr)
	if err != nil {
		return err
	}

	// Parse optional keys.
	if fromStateKeyPtr != nil {
		fromStateKey, err := identity.ParseKey(*fromStateKeyPtr)
		if err != nil {
			return err
		}
		transition.FromStateKey = &fromStateKey
	}
	if guardKeyPtr != nil {
		guardKey, err := identity.ParseKey(*guardKeyPtr)
		if err != nil {
			return err
		}
		transition.GuardKey = &guardKey
	}
	if actionKeyPtr != nil {
		actionKey, err := identity.ParseKey(*actionKeyPtr)
		if err != nil {
			return err
		}
		transition.ActionKey = &actionKey
	}
	if toStateKeyPtr != nil {
		toStateKey, err := identity.ParseKey(*toStateKeyPtr)
		if err != nil {
			return err
		}
		transition.ToStateKey = &toStateKey
	}

	return nil
}

// LoadTransition loads a transition from the database
func LoadTransition(dbOrTx DbOrTx, modelKey string, transitionKey identity.Key) (classKey identity.Key, transition model_state.Transition, err error) {

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
		transitionKey.String())
	if err != nil {
		return identity.Key{}, model_state.Transition{}, errors.WithStack(err)
	}

	return classKey, transition, nil
}

// AddTransition adds a transition to the database.
func AddTransition(dbOrTx DbOrTx, modelKey string, classKey identity.Key, transition model_state.Transition) (err error) {

	// We may or may not have a from state.
	var fromStateKeyPtr *string
	if transition.FromStateKey != nil {
		s := transition.FromStateKey.String()
		fromStateKeyPtr = &s
	}
	// We may or may not have a guard.
	var guardKeyPtr *string
	if transition.GuardKey != nil {
		s := transition.GuardKey.String()
		guardKeyPtr = &s
	}
	// We may or may not have an action.
	var actionKeyPtr *string
	if transition.ActionKey != nil {
		s := transition.ActionKey.String()
		actionKeyPtr = &s
	}
	// We may or may not have a to state.
	var toStateKeyPtr *string
	if transition.ToStateKey != nil {
		s := transition.ToStateKey.String()
		toStateKeyPtr = &s
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
		classKey.String(),
		transition.Key.String(),
		fromStateKeyPtr,
		transition.EventKey.String(),
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
func UpdateTransition(dbOrTx DbOrTx, modelKey string, classKey identity.Key, transition model_state.Transition) (err error) {

	// We may or may not have a from state.
	var fromStateKeyPtr *string
	if transition.FromStateKey != nil {
		s := transition.FromStateKey.String()
		fromStateKeyPtr = &s
	}
	// We may or may not have a guard.
	var guardKeyPtr *string
	if transition.GuardKey != nil {
		s := transition.GuardKey.String()
		guardKeyPtr = &s
	}
	// We may or may not have an action.
	var actionKeyPtr *string
	if transition.ActionKey != nil {
		s := transition.ActionKey.String()
		actionKeyPtr = &s
	}
	// We may or may not have a to state.
	var toStateKeyPtr *string
	if transition.ToStateKey != nil {
		s := transition.ToStateKey.String()
		toStateKeyPtr = &s
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
		classKey.String(),
		transition.Key.String(),
		fromStateKeyPtr,
		transition.EventKey.String(),
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
func RemoveTransition(dbOrTx DbOrTx, modelKey string, classKey identity.Key, transitionKey identity.Key) (err error) {

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
		classKey.String(),
		transitionKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryTransitions loads all transition from the database
func QueryTransitions(dbOrTx DbOrTx, modelKey string) (transitions map[identity.Key][]model_state.Transition, err error) {

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var classKey identity.Key
			var transition model_state.Transition
			if err = scanTransition(scanner, &classKey, &transition); err != nil {
				return errors.WithStack(err)
			}
			if transitions == nil {
				transitions = map[identity.Key][]model_state.Transition{}
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
