package database

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_class"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanClass(scanner Scanner, subdomainKeyPtr *string, class *model_class.Class) (err error) {
	var actorKeyPtr, superclassOfKey, subclassOfKey *string

	if err = scanner.Scan(
		subdomainKeyPtr,
		&class.Key,
		&class.Name,
		&class.Details,
		&actorKeyPtr,
		&superclassOfKey,
		&subclassOfKey,
		&class.UmlComment,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	if actorKeyPtr != nil {
		class.ActorKey = *actorKeyPtr
	}
	if superclassOfKey != nil {
		class.SuperclassOfKey = *superclassOfKey
	}
	if subclassOfKey != nil {
		class.SubclassOfKey = *subclassOfKey
	}

	return nil
}

// LoadClass loads a class from the database
func LoadClass(dbOrTx DbOrTx, modelKey, classKey string) (subdomainKey string, class model_class.Class, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return "", model_class.Class{}, err
	}
	classKey, err = identity.PreenKey(classKey)
	if err != nil {
		return "", model_class.Class{}, err
	}

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanClass(scanner, &subdomainKey, &class); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			subdomain_key     ,
			class_key         ,
			name              ,
			details           ,
			actor_key         ,
			superclass_of_key ,
			subclass_of_key   ,
			uml_comment
		FROM
			class
		WHERE
			class_key = $2
		AND
			model_key = $1`,
		modelKey,
		classKey)
	if err != nil {
		return "", model_class.Class{}, errors.WithStack(err)
	}

	return subdomainKey, class, nil
}

// AddClass adds a class to the database.
func AddClass(dbOrTx DbOrTx, modelKey, subdomainKey string, class model_class.Class) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}
	subdomainKey, err = identity.PreenKey(subdomainKey)
	if err != nil {
		return err
	}
	classKey, err := identity.PreenKey(class.Key)
	if err != nil {
		return err
	}

	// We may or may not be an actor.
	var actorKeyPtr *string
	if class.ActorKey != "" {
		actorKey, err := identity.PreenKey(class.ActorKey)
		if err != nil {
			return err
		}
		actorKeyPtr = &actorKey
	}
	// We may or may not be a superclass.
	var superclassOfKeyPtr *string
	if class.SuperclassOfKey != "" {
		superclassOfKey, err := identity.PreenKey(class.SuperclassOfKey)
		if err != nil {
			return err
		}
		superclassOfKeyPtr = &superclassOfKey
	}
	// We may or may not be a subclass.
	var subclassOfKeyPtr *string
	if class.SubclassOfKey != "" {
		subclassOfKey, err := identity.PreenKey(class.SubclassOfKey)
		if err != nil {
			return err
		}
		subclassOfKeyPtr = &subclassOfKey
	}

	// Add the data.
	_, err = dbExec(dbOrTx, `
		INSERT INTO class
			(
				model_key         ,
				subdomain_key     ,
				class_key         ,
				name              ,
				details           ,
				actor_key         ,
				superclass_of_key ,
				subclass_of_key   ,
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
		subdomainKey,
		classKey,
		class.Name,
		class.Details,
		actorKeyPtr,
		superclassOfKeyPtr,
		subclassOfKeyPtr,
		class.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// UpdateClass updates a class in the database.
func UpdateClass(dbOrTx DbOrTx, modelKey string, class model_class.Class) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}
	classKey, err := identity.PreenKey(class.Key)
	if err != nil {
		return err
	}
	// We may or may not be an actor.
	var actorKeyPtr *string
	if class.ActorKey != "" {
		actorKey, err := identity.PreenKey(class.ActorKey)
		if err != nil {
			return err
		}
		actorKeyPtr = &actorKey
	}
	// We may or may not be a superclass.
	var superclassOfKeyPtr *string
	if class.SuperclassOfKey != "" {
		superclassOfKey, err := identity.PreenKey(class.SuperclassOfKey)
		if err != nil {
			return err
		}
		superclassOfKeyPtr = &superclassOfKey
	}
	// We may or may not be a subclass.
	var subclassOfKeyPtr *string
	if class.SubclassOfKey != "" {
		subclassOfKey, err := identity.PreenKey(class.SubclassOfKey)
		if err != nil {
			return err
		}
		subclassOfKeyPtr = &subclassOfKey
	}
	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			class
		SET
			name              = $3 ,
			details           = $4 ,
			actor_key         = $5 ,
			superclass_of_key = $6 ,
			subclass_of_key   = $7 ,
			uml_comment       = $8
		WHERE
			model_key = $1
		AND
			class_key = $2`,
		modelKey,
		classKey,
		class.Name,
		class.Details,
		actorKeyPtr,
		superclassOfKeyPtr,
		subclassOfKeyPtr,
		class.UmlComment)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveClass deletes a class from the database.
func RemoveClass(dbOrTx DbOrTx, modelKey, classKey string) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}
	classKey, err = identity.PreenKey(classKey)
	if err != nil {
		return err
	}

	// Delete the data.
	_, err = dbExec(dbOrTx, `
		DELETE FROM
			class
		WHERE
			model_key = $1
		AND
			class_key = $2`,
		modelKey,
		classKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryClasses loads all classes from the database
func QueryClasses(dbOrTx DbOrTx, modelKey string) (classes map[string][]model_class.Class, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return nil, err
	}

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var subdomainKey string
			var class model_class.Class
			if err = scanClass(scanner, &subdomainKey, &class); err != nil {
				return errors.WithStack(err)
			}
			if classes == nil {
				classes = map[string][]model_class.Class{}
			}
			subdomainClasses := classes[subdomainKey]
			subdomainClasses = append(subdomainClasses, class)
			classes[subdomainKey] = subdomainClasses

			return nil
		},
		`SELECT
			subdomain_key     ,
			class_key         ,
			name              ,
			details           ,
			actor_key         ,
			superclass_of_key ,
			subclass_of_key   ,
			uml_comment
		FROM
			class
		WHERE
			model_key = $1
		ORDER BY subdomain_key, class_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return classes, nil
}
