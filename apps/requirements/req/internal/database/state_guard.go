package database

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanGuard(scanner Scanner, classKeyPtr *identity.Key, guard *model_state.Guard) (err error) {
	var classKeyStr string
	var guardKeyStr string

	if err = scanner.Scan(
		&classKeyStr,
		&guardKeyStr,
		&guard.Name,
		&guard.Logic.Description,
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

	// Parse the guard key string into an identity.Key.
	guard.Key, err = identity.ParseKey(guardKeyStr)
	if err != nil {
		return err
	}

	return nil
}

// LoadGuard loads a guard from the database
func LoadGuard(dbOrTx DbOrTx, modelKey string, guardKey identity.Key) (classKey identity.Key, guard model_state.Guard, err error) {

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanGuard(scanner, &classKey, &guard); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			class_key   ,
			guard_key   ,
			name        ,
			details
		FROM
			guard
		WHERE
			guard_key = $2
		AND
			model_key = $1`,
		modelKey,
		guardKey.String())
	if err != nil {
		return identity.Key{}, model_state.Guard{}, errors.WithStack(err)
	}

	return classKey, guard, nil
}

// AddGuard adds a guard to the database.
func AddGuard(dbOrTx DbOrTx, modelKey string, classKey identity.Key, guard model_state.Guard) (err error) {
	return AddGuards(dbOrTx, modelKey, map[identity.Key][]model_state.Guard{
		classKey: {guard},
	})
}

// UpdateGuard updates a guard in the database.
func UpdateGuard(dbOrTx DbOrTx, modelKey string, classKey identity.Key, guard model_state.Guard) (err error) {

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			guard
		SET
			name                  = $4 ,
			details               = $5
		WHERE
			class_key = $2
		AND
			guard_key = $3
		AND
			model_key = $1`,
		modelKey,
		classKey.String(),
		guard.Key.String(),
		guard.Name,
		guard.Logic.Description)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveGuard deletes a guard from the database.
func RemoveGuard(dbOrTx DbOrTx, modelKey string, classKey identity.Key, guardKey identity.Key) (err error) {

	// Delete the data.
	_, err = dbExec(dbOrTx, `
		DELETE FROM
			guard
		WHERE
			class_key = $2
		AND
			guard_key = $3
		AND
			model_key = $1`,
		modelKey,
		classKey.String(),
		guardKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryGuards loads all guard from the database
func QueryGuards(dbOrTx DbOrTx, modelKey string) (guards map[identity.Key][]model_state.Guard, err error) {

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var classKey identity.Key
			var guard model_state.Guard
			if err = scanGuard(scanner, &classKey, &guard); err != nil {
				return errors.WithStack(err)
			}
			if guards == nil {
				guards = map[identity.Key][]model_state.Guard{}
			}
			classGuards := guards[classKey]
			classGuards = append(classGuards, guard)
			guards[classKey] = classGuards
			return nil
		},
		`SELECT
			class_key   ,
			guard_key   ,
			name        ,
			details
		FROM
			guard
		WHERE
			model_key = $1
		ORDER BY class_key, guard_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return guards, nil
}

// AddGuards adds multiple guards to the database in a single insert.
func AddGuards(dbOrTx DbOrTx, modelKey string, guards map[identity.Key][]model_state.Guard) (err error) {
	// Count total guards.
	count := 0
	for _, gds := range guards {
		count += len(gds)
	}
	if count == 0 {
		return nil
	}

	// Build the bulk insert query.
	query := `INSERT INTO guard (model_key, class_key, guard_key, name, details) VALUES `
	args := make([]interface{}, 0, count*5)
	i := 0
	for classKey, guardList := range guards {
		for _, guard := range guardList {
			if i > 0 {
				query += ", "
			}
			base := i * 5
			query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4, base+5)
			args = append(args, modelKey, classKey.String(), guard.Key.String(), guard.Name, guard.Logic.Description)
			i++
		}
	}

	_, err = dbExec(dbOrTx, query, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
