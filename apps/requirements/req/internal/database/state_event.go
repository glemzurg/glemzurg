package database

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"

	"github.com/lib/pq"
	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanEvent(scanner Scanner, classKeyPtr *string, event *model_state.Event) (err error) {
	var parametersAsList []string

	if err = scanner.Scan(
		classKeyPtr,
		&event.Key,
		&event.Name,
		&event.Details,
		pq.Array(&parametersAsList),
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	// Construct parameters.
	for i := 0; i < len(parametersAsList); i += 2 {
		event.Parameters = append(event.Parameters, model_state.EventParameter{
			Name:   parametersAsList[i],
			Source: parametersAsList[i+1],
		})
	}

	return nil
}

// LoadEvent loads a event from the database
func LoadEvent(dbOrTx DbOrTx, modelKey, eventKey string) (classKey string, event model_state.Event, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return "", model_state.Event{}, err
	}
	eventKey, err = identity.PreenKey(eventKey)
	if err != nil {
		return "", model_state.Event{}, err
	}

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanEvent(scanner, &classKey, &event); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			class_key  ,
			event_key  ,
			name       ,
			details    ,
	        parameters
		FROM
			event
		WHERE
			event_key = $2
		AND
			model_key = $1`,
		modelKey,
		eventKey)
	if err != nil {
		return "", model_state.Event{}, errors.WithStack(err)
	}

	return classKey, event, nil
}

// AddEvent adds a event to the database.
func AddEvent(dbOrTx DbOrTx, modelKey, classKey string, event model_state.Event) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}
	classKey, err = identity.PreenKey(classKey)
	if err != nil {
		return err
	}
	eventKey, err := identity.PreenKey(event.Key)
	if err != nil {
		return err
	}

	// Flatten parameters.
	var parametersAsList []string
	for _, param := range event.Parameters {
		parametersAsList = append(parametersAsList, param.Name)
		parametersAsList = append(parametersAsList, param.Source)
	}

	// Add the data.
	_, err = dbExec(dbOrTx, `
			INSERT INTO event
				(
					model_key  ,
					class_key  ,
					event_key  ,
					name       ,
					details    ,
					parameters
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
		eventKey,
		event.Name,
		event.Details,
		pq.Array(parametersAsList))
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// UpdateEvent updates a event in the database.
func UpdateEvent(dbOrTx DbOrTx, modelKey, classKey string, event model_state.Event) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}
	classKey, err = identity.PreenKey(classKey)
	if err != nil {
		return err
	}
	eventKey, err := identity.PreenKey(event.Key)
	if err != nil {
		return err
	}

	// Flatten parameters.
	var parametersAsList []string
	for _, param := range event.Parameters {
		parametersAsList = append(parametersAsList, param.Name)
		parametersAsList = append(parametersAsList, param.Source)
	}

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			event
		SET
			name       = $4 ,
			details    = $5 ,
			parameters = $6
		WHERE
			class_key = $2
		AND
			event_key = $3
		AND
			model_key = $1`,
		modelKey,
		classKey,
		eventKey,
		event.Name,
		event.Details,
		pq.Array(parametersAsList))
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveEvent deletes a event from the database.
func RemoveEvent(dbOrTx DbOrTx, modelKey, classKey, eventKey string) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}
	classKey, err = identity.PreenKey(classKey)
	if err != nil {
		return err
	}
	eventKey, err = identity.PreenKey(eventKey)
	if err != nil {
		return err
	}

	// Delete the data.
	_, err = dbExec(dbOrTx, `
		DELETE FROM
			event
		WHERE
			class_key = $2
		AND
			event_key = $3
		AND
			model_key = $1`,
		modelKey,
		classKey,
		eventKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryEvents loads all event from the database
func QueryEvents(dbOrTx DbOrTx, modelKey string) (events map[string][]model_state.Event, err error) {

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
			var event model_state.Event
			if err = scanEvent(scanner, &classKey, &event); err != nil {
				return errors.WithStack(err)
			}
			if events == nil {
				events = map[string][]model_state.Event{}
			}
			classEvents := events[classKey]
			classEvents = append(classEvents, event)
			events[classKey] = classEvents
			return nil
		},
		`SELECT
			class_key  ,
			event_key  ,
			name       ,
			details    ,
			parameters
		FROM
			event
		WHERE
			model_key = $1
		ORDER BY class_key, event_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return events, nil
}
