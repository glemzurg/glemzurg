package database

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanGuard(scanner Scanner, classKeyPtr *string, guard *requirements.Guard) (err error) {
	if err = scanner.Scan(
		classKeyPtr,
		&guard.Key,
		&guard.Name,
		&guard.Details,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	return nil
}

// LoadGuard loads a guard from the database
func LoadGuard(dbOrTx DbOrTx, modelKey, guardKey string) (classKey string, guard requirements.Guard, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return "", requirements.Guard{}, err
	}
	guardKey, err = requirements.PreenKey(guardKey)
	if err != nil {
		return "", requirements.Guard{}, err
	}

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
		guardKey)
	if err != nil {
		return "", requirements.Guard{}, errors.WithStack(err)
	}

	return classKey, guard, nil
}

// AddGuard adds a guard to the database.
func AddGuard(dbOrTx DbOrTx, modelKey, classKey string, guard requirements.Guard) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	classKey, err = requirements.PreenKey(classKey)
	if err != nil {
		return err
	}
	guardKey, err := requirements.PreenKey(guard.Key)
	if err != nil {
		return err
	}

	// Add the data.
	_, err = dbExec(dbOrTx, `
			INSERT INTO guard
				(
					model_key   ,
					class_key   ,
					guard_key   ,
					name        ,
					details
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
		classKey,
		guardKey,
		guard.Name,
		guard.Details)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// UpdateGuard updates a guard in the database.
func UpdateGuard(dbOrTx DbOrTx, modelKey, classKey string, guard requirements.Guard) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	classKey, err = requirements.PreenKey(classKey)
	if err != nil {
		return err
	}
	guardKey, err := requirements.PreenKey(guard.Key)
	if err != nil {
		return err
	}

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
		classKey,
		guardKey,
		guard.Name,
		guard.Details)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveGuard deletes a guard from the database.
func RemoveGuard(dbOrTx DbOrTx, modelKey, classKey, guardKey string) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	classKey, err = requirements.PreenKey(classKey)
	if err != nil {
		return err
	}
	guardKey, err = requirements.PreenKey(guardKey)
	if err != nil {
		return err
	}

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
		classKey,
		guardKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryGuards loads all guard from the database
func QueryGuards(dbOrTx DbOrTx, modelKey string) (guards map[string][]requirements.Guard, err error) {

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
			var guard requirements.Guard
			if err = scanGuard(scanner, &classKey, &guard); err != nil {
				return errors.WithStack(err)
			}
			if guards == nil {
				guards = map[string][]requirements.Guard{}
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
