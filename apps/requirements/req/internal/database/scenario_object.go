package database

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanObject(scanner Scanner, scenarioKeyPtr *identity.Key, object *model_scenario.Object) (err error) {
	var objectKeyStr string
	var scenarioKeyStr string
	var classKeyStr string

	if err = scanner.Scan(
		&objectKeyStr,
		&scenarioKeyStr,
		&object.ObjectNumber,
		&object.Name,
		&object.NameStyle,
		&classKeyStr,
		&object.Multi,
		&object.UmlComment,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	// Parse the object key string into an identity.Key.
	object.Key, err = identity.ParseKey(objectKeyStr)
	if err != nil {
		return err
	}

	// Parse the scenario key string into an identity.Key.
	*scenarioKeyPtr, err = identity.ParseKey(scenarioKeyStr)
	if err != nil {
		return err
	}

	// Parse the class key string into an identity.Key.
	object.ClassKey, err = identity.ParseKey(classKeyStr)
	if err != nil {
		return err
	}

	return nil
}

// LoadObject loads a scenario object from the database
func LoadObject(dbOrTx DbOrTx, modelKey string, objectKey identity.Key) (scenarioKey identity.Key, object model_scenario.Object, err error) {

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
		objectKey.String())
	if err != nil {
		return identity.Key{}, model_scenario.Object{}, errors.WithStack(err)
	}

	return scenarioKey, object, nil
}

// AddObject adds a scenario object to the database.
func AddObject(dbOrTx DbOrTx, modelKey string, scenarioKey identity.Key, object model_scenario.Object) (err error) {
	return AddObjects(dbOrTx, modelKey, map[identity.Key][]model_scenario.Object{
		scenarioKey: {object},
	})
}

// UpdateObject updates a scenario object in the database.
func UpdateObject(dbOrTx DbOrTx, modelKey string, object model_scenario.Object) (err error) {

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
		object.Key.String(),
		object.ObjectNumber,
		object.Name,
		object.NameStyle,
		object.ClassKey.String(),
		object.Multi,
		object.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveObject deletes a scenario object from the database.
func RemoveObject(dbOrTx DbOrTx, modelKey string, objectKey identity.Key) (err error) {

	// Delete the data.
	_, err = dbExec(dbOrTx, `
			DELETE FROM
				scenario_object
			WHERE
				model_key = $1
			AND
				scenario_object_key = $2`,
		modelKey,
		objectKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryObjects loads all scenario objects from the database grouped by scenario key
func QueryObjects(dbOrTx DbOrTx, modelKey string) (objects map[identity.Key][]model_scenario.Object, err error) {

	objects = make(map[identity.Key][]model_scenario.Object)

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var scenarioKey identity.Key
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

// AddObjects adds multiple scenario objects to the database in a single insert.
func AddObjects(dbOrTx DbOrTx, modelKey string, objects map[identity.Key][]model_scenario.Object) (err error) {
	// Count total objects.
	count := 0
	for _, objs := range objects {
		count += len(objs)
	}
	if count == 0 {
		return nil
	}

	// Build the bulk insert query.
	query := `INSERT INTO scenario_object (model_key, scenario_object_key, scenario_key, object_number, name, name_style, class_key, multi, uml_comment) VALUES `
	args := make([]interface{}, 0, count*9)
	i := 0
	for scenarioKey, objList := range objects {
		for _, obj := range objList {
			if i > 0 {
				query += ", "
			}
			base := i * 9
			query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9)
			args = append(args, modelKey, obj.Key.String(), scenarioKey.String(), obj.ObjectNumber, obj.Name, obj.NameStyle, obj.ClassKey.String(), obj.Multi, obj.UmlComment)
			i++
		}
	}

	_, err = dbExec(dbOrTx, query, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
