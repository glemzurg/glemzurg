package database

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_scenario"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanScenarioObject(scanner Scanner, scenarioKeyPtr *string, scenarioObject *model_scenario.ScenarioObject) (err error) {
	if err = scanner.Scan(
		&scenarioObject.Key,
		scenarioKeyPtr,
		&scenarioObject.ObjectNumber,
		&scenarioObject.Name,
		&scenarioObject.NameStyle,
		&scenarioObject.ClassKey,
		&scenarioObject.Multi,
		&scenarioObject.UmlComment,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	return nil
}

// LoadScenarioObject loads a scenario object from the database
func LoadScenarioObject(dbOrTx DbOrTx, modelKey, scenarioObjectKey string) (scenarioKey string, scenarioObject model_scenario.ScenarioObject, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return "", model_scenario.ScenarioObject{}, err
	}
	scenarioObjectKey, err = requirements.PreenKey(scenarioObjectKey)
	if err != nil {
		return "", model_scenario.ScenarioObject{}, err
	}

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanScenarioObject(scanner, &scenarioKey, &scenarioObject); err != nil {
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
		scenarioObjectKey)
	if err != nil {
		return "", model_scenario.ScenarioObject{}, errors.WithStack(err)
	}

	return scenarioKey, scenarioObject, nil
}

// AddScenarioObject adds a scenario object to the database.
func AddScenarioObject(dbOrTx DbOrTx, modelKey, scenarioKey string, scenarioObject model_scenario.ScenarioObject) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	scenarioObjectKey, err := requirements.PreenKey(scenarioObject.Key)
	if err != nil {
		return err
	}
	scenarioKey, err = requirements.PreenKey(scenarioKey)
	if err != nil {
		return err
	}
	classKey, err := requirements.PreenKey(scenarioObject.ClassKey)
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
		scenarioObjectKey,
		scenarioKey,
		scenarioObject.ObjectNumber,
		scenarioObject.Name,
		scenarioObject.NameStyle,
		classKey,
		scenarioObject.Multi,
		scenarioObject.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// UpdateScenarioObject updates a scenario object in the database.
func UpdateScenarioObject(dbOrTx DbOrTx, modelKey string, scenarioObject model_scenario.ScenarioObject) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	scenarioObjectKey, err := requirements.PreenKey(scenarioObject.Key)
	if err != nil {
		return err
	}
	classKey, err := requirements.PreenKey(scenarioObject.ClassKey)
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
		scenarioObjectKey,
		scenarioObject.ObjectNumber,
		scenarioObject.Name,
		scenarioObject.NameStyle,
		classKey,
		scenarioObject.Multi,
		scenarioObject.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveScenarioObject deletes a scenario object from the database.
func RemoveScenarioObject(dbOrTx DbOrTx, modelKey, scenarioObjectKey string) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return err
	}
	scenarioObjectKey, err = requirements.PreenKey(scenarioObjectKey)
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
		scenarioObjectKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryScenarioObjects loads all scenario objects from the database grouped by scenario key
func QueryScenarioObjects(dbOrTx DbOrTx, modelKey string) (scenarioObjects map[string][]model_scenario.ScenarioObject, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = requirements.PreenKey(modelKey)
	if err != nil {
		return nil, err
	}

	scenarioObjects = make(map[string][]model_scenario.ScenarioObject)

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var scenarioKey string
			var scenarioObject model_scenario.ScenarioObject
			if err = scanScenarioObject(scanner, &scenarioKey, &scenarioObject); err != nil {
				return errors.WithStack(err)
			}
			scenarioObjects[scenarioKey] = append(scenarioObjects[scenarioKey], scenarioObject)
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

	return scenarioObjects, nil
}
