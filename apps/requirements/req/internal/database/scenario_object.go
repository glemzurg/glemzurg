package database

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanObject(scanner Scanner, scenarioKeyPtr *string, object *model_scenario.Object) (err error) {
	if err = scanner.Scan(
		&object.Key,
		scenarioKeyPtr,
		&object.ObjectNumber,
		&object.Name,
		&object.NameStyle,
		&object.ClassKey,
		&object.Multi,
		&object.UmlComment,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	return nil
}

// LoadObject loads a scenario object from the database
func LoadObject(dbOrTx DbOrTx, modelKey, objectKey string) (scenarioKey string, object model_scenario.Object, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return "", model_scenario.Object{}, err
	}
	objectKey, err = identity.PreenKey(objectKey)
	if err != nil {
		return "", model_scenario.Object{}, err
	}

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanObject(scanner, &scenarioKey, &object); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			scenario_object_key,
			scenario_key,
			object_number,
			name,
			name_style,
			class_key,
			multi,
			uml_comment
		FROM
			scenario_object
		WHERE
			scenario_object_key = $2
		AND
			model_key = $1`,
		modelKey,
		objectKey)
	if err != nil {
		return "", model_scenario.Object{}, errors.WithStack(err)
	}

	return scenarioKey, object, nil
}

// AddObject adds a scenario object to the database.
func AddObject(dbOrTx DbOrTx, modelKey, scenarioKey string, object model_scenario.Object) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}
	objectKey, err := identity.PreenKey(object.Key)
	if err != nil {
		return err
	}
	scenarioKey, err = identity.PreenKey(scenarioKey)
	if err != nil {
		return err
	}
	classKey, err := identity.PreenKey(object.ClassKey)
	if err != nil {
		return err
	}
	// Add the data.
	_, err = dbExec(dbOrTx, `
			INSERT INTO scenario_object
				(
					model_key,
					scenario_object_key,
					scenario_key,
					object_number,
					name,
					name_style,
					class_key,
					multi,
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
		objectKey,
		scenarioKey,
		object.ObjectNumber,
		object.Name,
		object.NameStyle,
		classKey,
		object.Multi,
		object.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// UpdateObject updates a scenario object in the database.
func UpdateObject(dbOrTx DbOrTx, modelKey string, object model_scenario.Object) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}
	objectKey, err := identity.PreenKey(object.Key)
	if err != nil {
		return err
	}
	classKey, err := identity.PreenKey(object.ClassKey)
	if err != nil {
		return err
	}

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			scenario_object
		SET
			object_number = $3,
			name = $4,
			name_style = $5,
			class_key = $6,
			multi = $7,
			uml_comment = $8
		WHERE
			model_key = $1
		AND
			scenario_object_key = $2`,
		modelKey,
		objectKey,
		object.ObjectNumber,
		object.Name,
		object.NameStyle,
		classKey,
		object.Multi,
		object.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveObject deletes a scenario object from the database.
func RemoveObject(dbOrTx DbOrTx, modelKey, objectKey string) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}
	objectKey, err = identity.PreenKey(objectKey)
	if err != nil {
		return err
	}

	// Delete the data.
	_, err = dbExec(dbOrTx, `
			DELETE FROM
				scenario_object
			WHERE
				model_key = $1
			AND
				scenario_object_key = $2`,
		modelKey,
		objectKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryObjects loads all scenario objects from the database grouped by scenario key
func QueryObjects(dbOrTx DbOrTx, modelKey string) (objects map[string][]model_scenario.Object, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return nil, err
	}

	objects = make(map[string][]model_scenario.Object)

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var scenarioKey string
			var object model_scenario.Object
			if err = scanObject(scanner, &scenarioKey, &object); err != nil {
				return errors.WithStack(err)
			}
			objects[scenarioKey] = append(objects[scenarioKey], object)
			return nil
		},
		`SELECT
			scenario_object_key,
			scenario_key,
			object_number,
			name,
			name_style,
			class_key,
			multi,
			uml_comment
		FROM
			scenario_object
		WHERE
			model_key = $1
		ORDER BY scenario_key, object_number, scenario_object_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return objects, nil
}
