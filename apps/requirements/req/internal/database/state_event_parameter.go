package database

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/pkg/errors"
)

// Populate a golang value from a database row.
func scanEventParameter(scanner Scanner, eventKeyPtr *identity.Key, name *string, sortOrder *int) (err error) {
	var eventKeyStr string
	var parameterKeyStr string // Read but not used on return.

	if err = scanner.Scan(
		&eventKeyStr,
		&parameterKeyStr,
		name,
		sortOrder,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	// Parse the event key string into an identity.Key.
	*eventKeyPtr, err = identity.ParseKey(eventKeyStr)
	if err != nil {
		return err
	}

	return nil
}

// LoadEventParameter loads an event parameter name from the database.
func LoadEventParameter(dbOrTx DbOrTx, modelKey string, eventKey identity.Key, parameterKey string) (name string, err error) {
	var loadedEventKey identity.Key
	var sortOrder int

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanEventParameter(scanner, &loadedEventKey, &name, &sortOrder); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			event_key     ,
			parameter_key ,
			name          ,
			sort_order
		FROM
			event_parameter
		WHERE
			model_key     = $1
		AND
			event_key     = $2
		AND
			parameter_key = $3`,
		modelKey,
		eventKey.String(),
		parameterKey)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return name, nil
}

// AddEventParameter adds a single event parameter name to the database.
func AddEventParameter(dbOrTx DbOrTx, modelKey string, eventKey identity.Key, name string) (err error) {
	return AddEventParameters(dbOrTx, modelKey, map[identity.Key][]string{
		eventKey: {name},
	})
}

// UpdateEventParameter updates an event parameter name row in the database.
func UpdateEventParameter(dbOrTx DbOrTx, modelKey string, eventKey identity.Key, parameterKey string, name string, sortOrder int) (err error) {
	// Update the data.
	err = dbExec(dbOrTx, `
		UPDATE
			event_parameter
		SET
			name       = $4 ,
			sort_order = $5
		WHERE
			model_key     = $1
		AND
			event_key     = $2
		AND
			parameter_key = $3`,
		modelKey,
		eventKey.String(),
		parameterKey,
		name,
		sortOrder)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveEventParameter deletes an event parameter from the database.
func RemoveEventParameter(dbOrTx DbOrTx, modelKey string, eventKey identity.Key, parameterKey string) (err error) {
	// Delete the data.
	err = dbExec(dbOrTx, `
		DELETE FROM
			event_parameter
		WHERE
			model_key     = $1
		AND
			event_key     = $2
		AND
			parameter_key = $3`,
		modelKey,
		eventKey.String(),
		parameterKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryEventParameters loads all event parameter names from the database, grouped by event key.
func QueryEventParameters(dbOrTx DbOrTx, modelKey string) (names map[identity.Key][]string, err error) {
	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var eventKey identity.Key
			var name string
			var sortOrder int
			if err = scanEventParameter(scanner, &eventKey, &name, &sortOrder); err != nil {
				return errors.WithStack(err)
			}
			if names == nil {
				names = map[identity.Key][]string{}
			}
			eventNames := names[eventKey]
			eventNames = append(eventNames, name)
			names[eventKey] = eventNames
			return nil
		},
		`SELECT
			event_key     ,
			parameter_key ,
			name          ,
			sort_order
		FROM
			event_parameter
		WHERE
			model_key = $1
		ORDER BY event_key, sort_order`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return names, nil
}

// AddEventParameters adds multiple event parameter names to the database in a single insert.
func AddEventParameters(dbOrTx DbOrTx, modelKey string, names map[identity.Key][]string) (err error) {
	// Count total parameter names.
	count := 0
	for _, ns := range names {
		count += len(ns)
	}
	if count == 0 {
		return nil
	}

	// Build the bulk insert query.
	var qb strings.Builder
	qb.WriteString(`INSERT INTO event_parameter (model_key, event_key, parameter_key, name, sort_order) VALUES `)
	args := make([]any, 0, count*5)
	i := 0
	for eventKey, nameList := range names {
		for paramIdx, name := range nameList {
			if i > 0 {
				qb.WriteString(", ")
			}

			paramKey := identity.NormalizeSubKey(name)

			base := i * 5
			fmt.Fprintf(&qb, "($%d, $%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4, base+5)
			args = append(args, modelKey, eventKey.String(), paramKey, name, paramIdx)
			i++
		}
	}

	err = dbExec(dbOrTx, qb.String(), args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
