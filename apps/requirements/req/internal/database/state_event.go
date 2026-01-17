package database

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"

	"github.com/lib/pq"
	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanEvent(scanner Scanner, classKeyPtr *identity.Key, event *model_state.Event) (err error) {
	var classKeyStr string
	var eventKeyStr string
	var parametersAsList []string

	if err = scanner.Scan(
		&classKeyStr,
		&eventKeyStr,
		&event.Name,
		&event.Details,
		pq.Array(&parametersAsList),
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

	// Parse the event key string into an identity.Key.
	event.Key, err = identity.ParseKey(eventKeyStr)
	if err != nil {
		return err
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
func LoadEvent(dbOrTx DbOrTx, modelKey string, eventKey identity.Key) (classKey identity.Key, event model_state.Event, err error) {

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
		eventKey.String())
	if err != nil {
		return identity.Key{}, model_state.Event{}, errors.WithStack(err)
	}

	return classKey, event, nil
}

// AddEvent adds a event to the database.
func AddEvent(dbOrTx DbOrTx, modelKey string, classKey identity.Key, event model_state.Event) (err error) {
	return AddEvents(dbOrTx, modelKey, map[identity.Key][]model_state.Event{
		classKey: {event},
	})
}

// UpdateEvent updates a event in the database.
func UpdateEvent(dbOrTx DbOrTx, modelKey string, classKey identity.Key, event model_state.Event) (err error) {

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
		classKey.String(),
		event.Key.String(),
		event.Name,
		event.Details,
		pq.Array(parametersAsList))
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveEvent deletes a event from the database.
func RemoveEvent(dbOrTx DbOrTx, modelKey string, classKey identity.Key, eventKey identity.Key) (err error) {

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
		classKey.String(),
		eventKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryEvents loads all event from the database
func QueryEvents(dbOrTx DbOrTx, modelKey string) (events map[identity.Key][]model_state.Event, err error) {

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var classKey identity.Key
			var event model_state.Event
			if err = scanEvent(scanner, &classKey, &event); err != nil {
				return errors.WithStack(err)
			}
			if events == nil {
				events = map[identity.Key][]model_state.Event{}
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

// AddEvents adds multiple events to the database in a single insert.
func AddEvents(dbOrTx DbOrTx, modelKey string, events map[identity.Key][]model_state.Event) (err error) {
	// Count total events.
	count := 0
	for _, evts := range events {
		count += len(evts)
	}
	if count == 0 {
		return nil
	}

	// Build the bulk insert query.
	query := `INSERT INTO event (model_key, class_key, event_key, name, details, parameters) VALUES `
	args := make([]interface{}, 0, count*6)
	i := 0
	for classKey, eventList := range events {
		for _, event := range eventList {
			if i > 0 {
				query += ", "
			}
			base := i * 6
			query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4, base+5, base+6)

			// Flatten parameters.
			var parametersAsList []string
			for _, param := range event.Parameters {
				parametersAsList = append(parametersAsList, param.Name)
				parametersAsList = append(parametersAsList, param.Source)
			}

			args = append(args, modelKey, classKey.String(), event.Key.String(), event.Name, event.Details, pq.Array(parametersAsList))
			i++
		}
	}

	_, err = dbExec(dbOrTx, query, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
